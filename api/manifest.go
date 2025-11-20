package api

import (
	pb "api-server/proto_gen"
	"api-server/queue"
	"api-server/schema"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// all valid manifests are required to be able to unmarshal to this struct
type BaseManifest struct {
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                 `json:"kind"       yaml:"kind"`
	Metadata   map[string]interface{} `json:"metadata"   yaml:"metadata"`
	Spec       map[string]interface{} `json:"spec"       yaml:"spec"`
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

// MaxInputSize defines the maximum allowed size for base64-encoded input
// (500MB)
const MaxInputSize = 500 * 1024 * 1024

// PostManifest implements StrictServerInterface.
func (s *Server) PostManifest(
	ctx context.Context,
	mreq PostManifestRequestObject,
) (PostManifestResponseObject, error) {
	baseManifest, err := unmarshalManifest(mreq.Body)
	if err != nil {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid JSON format: " + err.Error(),
			},
		}, nil
	}

	switch baseManifest.Kind {
	case "Blueprint":
		return s.processBlueprint(mreq.Body)
	default:
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Unsupported manifest kind: " + baseManifest.Kind,
			},
		}, nil
	}
}

func (s *Server) processBlueprint(
	data io.Reader,
) (PostManifestResponseObject, error) {
	var blueprint schema.Blueprint

	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&blueprint); err != nil {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: "Invalid Blueprint JSON format: " + err.Error(),
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

	// Validate input size before decoding to prevent memory exhaustion attacks
	if len(blueprint.Spec.Artifact.Input) > MaxInputSize {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: fmt.Sprintf(
					"Artifact input size exceeds maximum allowed size of %d bytes",
					MaxInputSize,
				),
			},
		}, nil
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

	// Validate decoded size as well
	if len(inputBytes) > MaxInputSize {
		return PostManifest400JSONResponse{
			GenericBadRequestJSONResponse{
				Error: fmt.Sprintf(
					"Decoded artifact input size exceeds maximum allowed size of %d bytes",
					MaxInputSize,
				),
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

	// Enqueue the task for processing
	enqueuedTask, err := queue.Q.EnqueueTask(task)
	if err != nil {
		return PostManifest500Response{}, nil
	}

	// Create response with task ID
	taskID := enqueuedTask.ResultWriter().TaskID()
	responseBody := fmt.Sprintf("taskId: %s\n", taskID)

	return PostManifest201TextyamlResponse{
		Body:          strings.NewReader(responseBody),
		ContentLength: int64(len(responseBody)),
	}, nil
}

func unmarshalManifest(data io.Reader) (BaseManifest, error) {
	// try unmarshalling as BaseManifest
	var manifest BaseManifest
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&manifest); err != nil {
		return BaseManifest{}, fmt.Errorf("failed to decode manifest JSON: %w", err)
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
