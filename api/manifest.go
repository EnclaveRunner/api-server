package api

import (
	pb "api-server/proto_gen"
	"api-server/schema"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultMaxRetries = 3
	defaultRetention  = 24 * time.Hour
)

// all valid manifests are required to be able to unmarshal to this struct
type BaseManifest struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind"       yaml:"kind"`
	Metadata   map[string]any `json:"metadata"   yaml:"metadata"`
	Spec       map[string]any `json:"spec"       yaml:"spec"`
}

// ErrInvalidIdentifier is returned when an identifier cannot be parsed
var ErrInvalidIdentifier = errors.New("invalid identifier format")

// PostManifest implements StrictServerInterface.
func (server *Server) PostManifest(
	ctx context.Context,
	mreq PostManifestRequestObject,
) (PostManifestResponseObject, error) {
	body, err := io.ReadAll(mreq.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")

		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Failed to read body",
			},
		}, nil
	}
	baseManifest, err := unmarshalManifest(body)
	if err != nil {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid YAML format: " + err.Error(),
			},
		}, nil
	}

	switch baseManifest.Kind {
	case "Blueprint":
		return server.processBlueprint(ctx, body)
	default:
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Unsupported manifest kind: " + baseManifest.Kind,
			},
		}, nil
	}
}

func (server *Server) processBlueprint(
	ctx context.Context,
	data []byte,
) (PostManifestResponseObject, error) {
	var blueprint schema.Blueprint

	// Copy data
	decoder := yaml.NewDecoder(bytes.NewBuffer(data))
	if err := decoder.Decode(&blueprint); err != nil {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid Blueprint YAML format: " + err.Error(),
			},
		}, nil
	}

	fullIdentifier, err := parseSource(blueprint.Spec.Source)
	if err != nil {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid artifact source format: " + err.Error(),
			},
		}, nil
	}

	// Convert params to proto Parameters
	params := make([]*pb.Val, len(blueprint.Spec.Params))
	for i, p := range blueprint.Spec.Params {
		params[i] = anyToProtoVal(p)
	}

	// Convert env to proto EnvironmentVariables
	envVars := make([]*pb.EnvironmentVariable, len(blueprint.Spec.Env))
	for i, e := range blueprint.Spec.Env {
		key, value := "", ""
		if e.Key != nil {
			key = *e.Key
		}
		if e.Value != nil {
			value = *e.Value
		}
		envVars[i] = &pb.EnvironmentVariable{Key: key, Value: value}
	}

	task := &pb.Task{
		Function:             fullIdentifier,
		Parameters:           params,
		Arguments:            blueprint.Spec.Args,
		EnvironmentVariables: envVars,
	}

	// Check that artifact exists
	_, err = server.registryClient.GetArtifact(ctx, fullIdentifier.Artifact)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return PostManifest400JSONResponse{
				GenericBadRequestJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get artifact")

		return &PostManifest500Response{}, nil
	}

	taskOptions := []asynq.Option{}
	if blueprint.Spec.Retention != nil {
		retention, err := time.ParseDuration(*blueprint.Spec.Retention)
		if err != nil {
			return &PostManifest400JSONResponse{
				GenericBadRequestJSONResponse{
					Error: "Retention string invalid: " + err.Error(),
				},
			}, nil
		}

		taskOptions = append(taskOptions, asynq.Retention(retention))
	} else {
		taskOptions = append(taskOptions, asynq.Retention(defaultRetention))
	}

	if blueprint.Spec.Retries != nil {
		taskOptions = append(taskOptions, asynq.MaxRetry(*blueprint.Spec.Retries))
	} else {
		taskOptions = append(taskOptions, asynq.MaxRetry(defaultMaxRetries))
	}

	// Enqueue the task for processing
	taskInfo, err := server.queueClient.EnqueueTask(
		task,
		taskOptions...,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to enqueue task")

		return PostManifest500Response{}, nil
	}

	return PostManifest201TextResponse(taskInfo.ID), nil
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
			fields = append(fields, &pb.RecordField{Name: k, Value: anyToProtoVal(fv)})
		}

		return &pb.Val{Value: &pb.Val_RecordVal{RecordVal: &pb.RecordVal{Fields: fields}}}
	default:
		// nil or any unrecognised type → option<T> none
		return &pb.Val{Value: &pb.Val_OptionVal{OptionVal: &pb.OptionVal{}}}
	}
}

func unmarshalManifest(data []byte) (BaseManifest, error) {
	// try unmarshalling as BaseManifest
	var manifest BaseManifest

	decoder := yaml.NewDecoder(bytes.NewBuffer(data))
	if err := decoder.Decode(&manifest); err != nil {
		return BaseManifest{}, fmt.Errorf("failed to decode manifest YAML: %w", err)
	}

	return manifest, nil
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
