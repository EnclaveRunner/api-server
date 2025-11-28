package api

import (
	pb "api-server/proto_gen"
	"api-server/schema"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// all valid manifests are required to be able to unmarshal to this struct
type BaseManifest struct {
	APIVersion string         `json:"apiVersion" yaml:"apiVersion"`
	Kind       string         `json:"kind"       yaml:"kind"`
	Metadata   map[string]any `json:"metadata"   yaml:"metadata"`
	Spec       map[string]any `json:"spec"       yaml:"spec"`
}

type Identifier struct {
	Source string `json:"source" yaml:"source"`
	Author string `json:"author" yaml:"author"`
	Name   string `json:"name"   yaml:"name"`
	Hash   string `json:"hash"   yaml:"hash"`
	Tag    string `json:"tag"    yaml:"tag"`
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

	fullIdentifier, err := parseSource(blueprint.Spec.Artifact.Source)
	if err != nil {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid artifact source format: " + err.Error(),
			},
		}, nil
	}

	var tagIdentifier *pb.ArtifactIdentifier_Tag
	var versionIdentifier *pb.ArtifactIdentifier_VersionHash

	if fullIdentifier.Hash == "" {
		tagIdentifier = &pb.ArtifactIdentifier_Tag{
			Tag: fullIdentifier.Tag,
		}
	} else {
		versionIdentifier = &pb.ArtifactIdentifier_VersionHash{
			VersionHash: fullIdentifier.Hash,
		}
	}

	// Decode base64 input to bytes
	inputBytes, err := base64.StdEncoding.DecodeString(
		blueprint.Spec.Artifact.Input,
	)
	if err != nil {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid base64 encoding in artifact input: " + err.Error(),
			},
		}, nil
	}

	task := &pb.Task{
		Artifact: &pb.ArtifactIdentifier{
			Fqn: &pb.FullyQualifiedName{
				Source: fullIdentifier.Source,
				Author: fullIdentifier.Author,
				Name:   fullIdentifier.Name,
			},
		},
		Function: blueprint.Spec.Artifact.Function,
		Input:    inputBytes,
	}

	// Set the correct identifier type
	if tagIdentifier != nil {
		task.Artifact.Identifier = tagIdentifier
	} else {
		task.Artifact.Identifier = versionIdentifier
	}

	// Check that artifact exists
	_, err = server.registryClient.GetArtifact(ctx, task.Artifact)
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

	// Enqueue the task for processing
	taskInfo, err := server.queueClient.EnqueueTask(task)
	if err != nil {
		log.Error().Err(err).Msg("Failed to enqueue task")

		return PostManifest500Response{}, nil
	}

	// Create response with task ID
	taskID := taskInfo.ID
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

// parses <source>/<author>/<name>:<hash:versionhash|tag> into Identifier struct
func parseSource(identifier string) (Identifier, error) {
	var id Identifier
	var data []string

	data = strings.Split(identifier, "/")
	//nolint:mnd // ignore magic numbers for split counts
	if len(data) != 3 {
		return id, ErrInvalidIdentifier
	}

	id.Source = data[0]
	id.Author = data[1]

	data = strings.Split(data[2], ":")
	if len(data) < 2 || len(data) > 3 {
		return id, ErrInvalidIdentifier
	}

	id.Name = data[0]

	// Handle different identifier formats
	switch len(data) {
	case 2: //nolint:mnd // 2 parts means name:tag format
		// Format: name:tag
		id.Tag = data[1]
	case 3: //nolint:mnd // 3 parts means name:hash:versionhash format
		// Format: name:hash:versionhash
		if data[1] != "hash" {
			return id, fmt.Errorf(
				"%w: expected 'hash' but got '%s'",
				ErrInvalidIdentifier,
				data[1],
			)
		}
		id.Hash = data[2]
	default:
		// More than 3 parts is invalid
		return id, ErrInvalidIdentifier
	}

	// Validate required fields and ensure either Tag or Hash is set (but not
	// both)
	if id.Source == "" || id.Author == "" || id.Name == "" {
		return Identifier{}, ErrInvalidIdentifier
	}

	// Must have either Tag or Hash, but not both or neither
	if (id.Tag == "" && id.Hash == "") || (id.Tag != "" && id.Hash != "") {
		return Identifier{}, ErrInvalidIdentifier
	}

	return id, nil
}
