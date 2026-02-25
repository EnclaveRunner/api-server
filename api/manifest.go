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

	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultMaxRetries = 3
	defaultRetention  = "24h"
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
	params := make([]*pb.Parameter, len(blueprint.Spec.Params))
	for i, p := range blueprint.Spec.Params {
		params[i] = &pb.Parameter{Value: &pb.Parameter_Dbl{Dbl: p}}
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

	// Register the task in database first
	registeredTask, err := server.db.RegisterTask(
		defaultMaxRetries,
		defaultRetention,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register task in database")

		return PostManifest500Response{}, nil
	}

	task.TaskId = registeredTask.TaskID.String()

	// Enqueue the task for processing
	_, err = server.queueClient.EnqueueTask(registeredTask.TaskID, task)
	if err != nil {
		log.Error().Err(err).Msg("Failed to enqueue task")

		return PostManifest500Response{}, nil
	}

	// Create response with task ID from database
	taskID := registeredTask.TaskID.String()
	responseBody := fmt.Sprintf("taskId: %s\n", taskID)

	return PostManifest201TextyamlResponse{
		Body:          strings.NewReader(responseBody),
		ContentLength: int64(len(responseBody)),
	}, nil
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
