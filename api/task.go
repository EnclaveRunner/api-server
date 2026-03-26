package api

import (
	"api-server/orm"
	pb "api-server/proto_gen"
	"api-server/queue"
	"cmp"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/EnclaveRunner/shareddeps/utils"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var (
	// ErrRequestBodyRequired is returned when request body is missing
	ErrRequestBodyRequired = errors.New("request body is required")
	// ErrInvalidTaskIDFormat is returned when task ID format is invalid
	ErrInvalidTaskIDFormat = errors.New("invalid task ID format")
	// ErrInvalidIdentifier is returned when artifact source identifier format is
	// invalid
	ErrInvalidIdentifier = errors.New("invalid identifier format")
)

// GetV1Task implements [StrictServerInterface].
func (server *Server) GetV1Task(
	ctx context.Context,
	request GetV1TaskRequestObject,
) (GetV1TaskResponseObject, error) {
	tasks, err := server.queueClient.GetAllTasks()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list tasks")

		return GetV1Task500Response{}, nil
	}

	if request.Params.State != nil {
		// Filter tasks matching state parameter
		tasks = slices.Collect(func(yield func(*asynq.TaskInfo) bool) {
			for _, task := range tasks {
				if task.State.String() == *request.Params.State {
					if !yield(task) {
						return
					}
				}
			}
		})
	}

	taskPage := paginate(
		tasks,
		*request.Params.Limit,
		*request.Params.Offset,
		func(a, b *asynq.TaskInfo) int {
			return cmp.Compare(a.ID, b.ID)
		},
	)

	taskPageTransformed := make([]Task, len(taskPage))
	for i, task := range taskPage {
		state, err := taskToTaskResponse(task)
		if err != nil {
			log.Error().Err(err).Str("id", task.ID).Msg("Failed to transform task")

			return GetV1Task500Response{}, nil
		}

		taskPageTransformed[i] = state
	}

	return GetV1Task200JSONResponse(taskPageTransformed), nil
}

// PostV1Task implements [StrictServerInterface].
func (server *Server) PostV1Task(
	ctx context.Context,
	request PostV1TaskRequestObject,
) (PostV1TaskResponseObject, error) {
	fullIdentifier, err := parseSource(request.Body.Source)
	if err != nil {
		return PostV1Task400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid artifact source format: " + err.Error(),
			},
		}, nil
	}

	task := &pb.Task{
		Function: fullIdentifier,
	}

	// Convert params to proto Parameters
	if request.Body.Params != nil {
		params := make([]*pb.Val, len(*request.Body.Params))
		for i, p := range *request.Body.Params {
			params[i] = anyToProtoVal(p)
		}

		task.Parameters = params
	}

	// Convert env to proto EnvironmentVariables
	if request.Body.Env != nil {
		envVars := make([]*pb.EnvironmentVariable, len(*request.Body.Env))
		for i, e := range *request.Body.Env {
			envVars[i] = &pb.EnvironmentVariable{Key: e.Key, Value: e.Value}
		}

		task.EnvironmentVariables = envVars
	}

	if request.Body.Args != nil {
		task.Arguments = *request.Body.Args
	}

	// Check that artifact exists
	_, err = server.registryClient.GetArtifact(ctx, fullIdentifier.Artifact)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return PostV1Task400JSONResponse{
				GenericBadRequestJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get artifact")

		return &PostV1Task500Response{}, nil
	}

	taskOptions := []asynq.Option{}
	if request.Body.Retention != nil && *request.Body.Retention != "" {
		retention, err := time.ParseDuration(*request.Body.Retention)
		if err != nil {
			return &PostV1Task400JSONResponse{
				GenericBadRequestJSONResponse{
					Error: "Retention string invalid: " + err.Error(),
				},
			}, nil
		}

		taskOptions = append(taskOptions, asynq.Retention(retention))
	} else {
		taskOptions = append(taskOptions, asynq.Retention(server.retention))
	}

	if request.Body.Retries != nil {
		taskOptions = append(taskOptions, asynq.MaxRetry(*request.Body.Retries))
	} else {
		taskOptions = append(taskOptions, asynq.MaxRetry(server.maxRetries))
	}

	// Enqueue the task for processing
	taskInfo, err := server.queueClient.EnqueueTask(
		task,
		taskOptions...,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to enqueue task")

		return PostV1Task500Response{}, nil
	}

	return PostV1Task201JSONResponse{
		Id:        taskInfo.ID,
		Source:    request.Body.Source,
		Params:    request.Body.Params,
		Args:      request.Body.Args,
		Env:       request.Body.Env,
		Callback:  request.Body.Callback,
		Retention: utils.Ptr(taskInfo.Retention.String()),
		Retries:   &taskInfo.MaxRetry,
		Status:    TaskStatus{},
	}, nil
}

// GetV1TaskId implements [StrictServerInterface].
func (server *Server) GetV1TaskId(
	ctx context.Context,
	request GetV1TaskIdRequestObject,
) (GetV1TaskIdResponseObject, error) {
	task, err := server.queueClient.GetTask(request.Id)
	if err != nil {
		if errors.Is(err, &queue.TaskNotFoundError{}) {
			return GetV1TaskId404JSONResponse{GenericNotFoundJSONResponse{
				Error: err.Error(),
			}}, nil
		}

		log.Error().
			Err(err).
			Str("id", request.Id).
			Msg("Failed to retrieve task")

		return GetV1TaskId500Response{}, nil
	}

	var taskPayload pb.Task
	if err := proto.Unmarshal(task.Payload, &taskPayload); err != nil {
		log.Error().
			Err(err).
			Str("id", request.Id).
			Msg("Failed to unmarshall task payload")
	}

	state, err := taskToTaskResponse(task)
	if err != nil {
		log.Error().Err(err).Str("id", request.Id).Msg("Failed to transform task")

		return GetV1TaskId500Response{}, nil
	}

	return GetV1TaskId200JSONResponse(state), nil
}

// GetV1TaskIdLogs implements [StrictServerInterface].
func (server *Server) GetV1TaskIdLogs(
	ctx context.Context,
	request GetV1TaskIdLogsRequestObject,
) (GetV1TaskIdLogsResponseObject, error) {
	if _, err := server.queueClient.GetTask(request.Id); err != nil {
		if !errors.Is(err, &queue.TaskNotFoundError{}) {
			log.Error().
				Err(err).
				Str("id", request.Id).
				Msg("Failed to retrieve task")

			return GetV1TaskIdLogs500Response{}, nil
		}

		return GetV1TaskIdLogs404JSONResponse{}, nil
	}

	logs, err := server.db.GetLogsOfTask(ctx, request.Id)
	if err != nil {
		log.Error().
			Err(err).
			Str("id", request.Id).
			Msg("Failed to retrieve logs of task")

		return GetV1TaskIdLogs500Response{}, nil
	}

	return GetV1TaskIdLogs200JSONResponse(dbLogsToJsonLogs(logs)), nil
}

func taskToTaskResponse(task *asynq.TaskInfo) (Task, error) {
	var taskPayload pb.Task
	if err := proto.Unmarshal(task.Payload, &taskPayload); err != nil {
		//nolint:wrapcheck // Error is not used but only logged later on, no error wrap needed
		return Task{}, err
	}

	state := Task{
		Id:        task.ID,
		Source:    serializeSource(taskPayload.Function),
		Retries:   &task.MaxRetry,
		Retention: utils.Ptr(task.Retention.String()),
		Args:      &taskPayload.Arguments,
		Callback:  utils.Ptr(""), // Currently not in task proto
		Status: TaskStatus{
			Retries:       task.Retried,
			State:         task.State.String(),
			LastError:     &task.LastErr,
			LastFailedAt:  &task.LastFailedAt,
			NextProcessAt: &task.NextProcessAt,
			CompletedAt:   &task.CompletedAt,
		},
	}

	if taskPayload.Parameters != nil {
		params := make([]any, len(taskPayload.Parameters))

		for i, param := range taskPayload.Parameters {
			params[i] = protoValToAny(param)
		}

		state.Params = &params
	}

	if taskPayload.EnvironmentVariables != nil {
		envVars := make(
			[]EnvironmentVariable,
			len(taskPayload.EnvironmentVariables),
		)

		for i, envVar := range taskPayload.EnvironmentVariables {
			envVars[i] = EnvironmentVariable{
				Key:   envVar.Key,
				Value: envVar.Value,
			}
		}

		state.Env = &envVars
	}

	if task.Result != nil {
		state.Status.ResultPayload = utils.Ptr(
			base64.StdEncoding.EncodeToString(task.Result),
		)
	}

	return state, nil
}

func serializeSource(source *pb.FunctionIdentifier) string {
	serialized := ""

	serialized += source.Artifact.Package.Namespace
	serialized += ":"
	serialized += source.Artifact.Package.Name
	serialized += "/"
	serialized += source.Interface
	serialized += "/"
	serialized += source.Name
	serialized += "@"
	if tag, ok := source.Artifact.Identifier.(*pb.ArtifactIdentifier_Tag); ok {
		serialized += tag.Tag
	} else {
		serialized += "hash:"
		//nolint:forcetypeassert // The only other option for the oneof is VersionHash, so this type assertion is safe
		serialized += source.Artifact.Identifier.(*pb.ArtifactIdentifier_VersionHash).VersionHash
	}

	return serialized
}

func dbLogsToJsonLogs(dbLogs []orm.TaskLog) []TaskLog {
	jsonLogs := make([]TaskLog, len(dbLogs))

	for i, log := range dbLogs {
		jsonLogs[i] = TaskLog{
			Issuer:    log.Issuer,
			Level:     log.Level,
			Message:   log.Message,
			Timestamp: log.Timestamp,
		}
	}

	return jsonLogs
}

// anyToProtoVal converts a value produced by YAML/JSON unmarshalling into a
// proto Val. Only the types that the standard library decoders can produce are
// handled:
//
//	bool            → BoolVal
//	int             → S64Val
//	int64           → S64Val
//	float64         → F64Val
//	string          → StringVal
//	[]interface{}   → ListVal  (recursive)
//	map[string]any  → RecordVal (recursive)
//	nil             → OptionVal{Value: nil}  (none)
func anyToProtoVal(v any) *pb.Val {
	switch val := v.(type) {
	case bool:
		return &pb.Val{Value: &pb.Val_BoolVal{BoolVal: val}}
	case int:
		return &pb.Val{Value: &pb.Val_S64Val{S64Val: int64(val)}}
	case int64:
		return &pb.Val{Value: &pb.Val_S64Val{S64Val: val}}
	case float64:
		return &pb.Val{Value: &pb.Val_F64Val{F64Val: val}}
	case string:
		return &pb.Val{Value: &pb.Val_StringVal{StringVal: val}}
	case []interface{}:
		elems := make([]*pb.Val, len(val))
		for i, elem := range val {
			elems[i] = anyToProtoVal(elem)
		}

		return &pb.Val{Value: &pb.Val_ListVal{ListVal: &pb.ListVal{Values: elems}}}
	case map[string]interface{}:
		fields := make([]*pb.RecordField, 0, len(val))
		for k, fv := range val {
			fields = append(
				fields,
				&pb.RecordField{Name: k, Value: anyToProtoVal(fv)},
			)
		}

		return &pb.Val{
			Value: &pb.Val_RecordVal{RecordVal: &pb.RecordVal{Fields: fields}},
		}
	default:
		// nil or any unrecognised type → option<T> none
		return &pb.Val{Value: &pb.Val_OptionVal{OptionVal: &pb.OptionVal{}}}
	}
}

// protoValToAny converts a proto Val back into JSON-serializable Go values.
// It handles only the subset of variants produced by anyToProtoVal.
func protoValToAny(v *pb.Val) any {
	if v == nil {
		return nil
	}

	switch val := v.Value.(type) {
	case *pb.Val_BoolVal:
		return val.BoolVal
	case *pb.Val_S64Val:
		return val.S64Val
	case *pb.Val_F64Val:
		return val.F64Val
	case *pb.Val_StringVal:
		return val.StringVal
	case *pb.Val_ListVal:
		if val.ListVal == nil {
			return []interface{}{}
		}

		elems := make([]interface{}, len(val.ListVal.Values))
		for i, elem := range val.ListVal.Values {
			elems[i] = protoValToAny(elem)
		}

		return elems
	case *pb.Val_RecordVal:
		obj := make(map[string]interface{}, 0)
		if val.RecordVal == nil {
			return obj
		}

		obj = make(map[string]interface{}, len(val.RecordVal.Fields))
		for _, field := range val.RecordVal.Fields {
			if field == nil {
				continue
			}

			obj[field.Name] = protoValToAny(field.Value)
		}

		return obj
	case *pb.Val_OptionVal:
		if val.OptionVal == nil || val.OptionVal.Value == nil {
			return nil
		}

		return protoValToAny(val.OptionVal.Value)
	default:
		return nil
	}
}

// parses namespace:package/interface/function@hash<hash>|version into
// Identifier struct
func parseSource(identifier string) (*pb.FunctionIdentifier, error) {
	functionIdentifier := &pb.FunctionIdentifier{
		Artifact: &pb.ArtifactIdentifier{
			Package: &pb.PackageName{},
		},
	}

	//nolint:mnd // Two parts: <namespace>:<rest>
	split := strings.SplitN(identifier, ":", 2)
	//nolint:mnd // Two parts: <namespace>:<rest>
	if len(split) != 2 {
		return functionIdentifier, fmt.Errorf(
			"%w: Missing namespace",
			ErrInvalidIdentifier,
		)
	}

	functionIdentifier.Artifact.Package.Namespace = split[0]
	identifier = split[1]

	//nolint:mnd // 3 parts: <name>/<interface>/<rest>
	split = strings.SplitN(identifier, "/", 3)
	//nolint:mnd // 3 parts: <name>/<interface>/<rest>
	if len(split) != 3 {
		return functionIdentifier, fmt.Errorf(
			"%w: Identifier incomplete. name/interface/function needed",
			ErrInvalidIdentifier,
		)
	}

	functionIdentifier.Artifact.Package.Name = split[0]
	functionIdentifier.Interface = split[1]
	identifier = split[2]

	//nolint:mnd // 2 parts: <function>@<rest>
	split = strings.SplitN(identifier, "@", 2)
	//nolint:mnd // 2 parts: <function>@<rest>
	if len(split) != 2 {
		return functionIdentifier, fmt.Errorf(
			"%w: Missing version",
			ErrInvalidIdentifier,
		)
	}

	functionIdentifier.Name = split[0]
	identifier = split[1]

	if strings.HasPrefix(identifier, "hash:") {
		// Hash version
		functionIdentifier.Artifact.Identifier = &pb.ArtifactIdentifier_VersionHash{
			VersionHash: strings.TrimPrefix(identifier, "hash:"),
		}
	} else {
		// tag version
		functionIdentifier.Artifact.Identifier = &pb.ArtifactIdentifier_Tag{
			Tag: identifier,
		}
	}

	return functionIdentifier, nil
}
