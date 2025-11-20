package api

import (
	pb "api-server/proto_gen"
	"api-server/queue"
	"api-server/schema"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// all valid manifest are required to be able to marshal to this struct
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
var ErrInvalidIdentifier = fmt.Errorf("invalid identifier format")

// PostManifest implements StrictServerInterface.
func (s *Server) PostManifest(
	ctx context.Context,
	mreq PostManifestRequestObject,
) (PostManifestResponseObject, error) {
	baseManifest, err := unmarshalManifest(mreq.Body)
	if err != nil {
		return PostManifest400JSONResponse{GenericBadRequestJSONResponse{}}, nil
	}

	switch baseManifest.Kind {
	case "Blueprint":
		return s.processBlueprint(mreq.Body)
	default:
		return PostManifest400JSONResponse{GenericBadRequestJSONResponse{}}, nil
	}
}

func (s *Server) processBlueprint(
	data io.Reader,
) (PostManifestResponseObject, error) {
	var blueprint schema.Blueprint
	var fullIdentifier Identifier

	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&blueprint); err != nil {
		return PostManifest400JSONResponse{GenericBadRequestJSONResponse{}}, nil
	}

	fullIdentifier, err := parseSource(blueprint.Spec.Artifact.Source)
	if err != nil {
		return PostManifest400JSONResponse{GenericBadRequestJSONResponse{}}, nil
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
		return PostManifest400JSONResponse{GenericBadRequestJSONResponse{}}, nil
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
	// try marshalling as BaseManifest
	var manifest BaseManifest
	decoder := json.NewDecoder(data)
	if err := decoder.Decode(&manifest); err != nil {
		return BaseManifest{}, err
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
	//nolint:mnd // ignore magic numbers for split counts
	if len(data) < 2 {
		return id, ErrInvalidIdentifier
	}

	if len(data) == 3 && data[1] == "hash" { //nolint:gomagicnumber
		id.Name = data[0]
		id.Hash = data[2]
	} else {
		id.Name = data[0]
		id.Tag = data[1]
	}

	return id, nil
}
