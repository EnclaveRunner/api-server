package api

import (
	pb "api-server/proto_gen"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	VersionHashPrefix              = "hash:"
	MetadataFieldMaxSize           = 1024 * 10       // 10KiB
	ArtifactRegistryMaxMessageSize = 1024 * 1024 * 3 // 3MiB: Make sure to keep below gRPC max message size of 4MiB
)

func artifactToArtifact(artifact *pb.Artifact) Artifact {
	return Artifact{
		Fqn: FQN{
			Source: artifact.Fqn.Source,
			Author: artifact.Fqn.Author,
			Name:   artifact.Fqn.Name,
		},
		VersionHash: artifact.VersionHash,
		Tags:        artifact.Tags,
		Pulls:       int(artifact.Metadata.Pulls),
		CreatedAt:   artifact.Metadata.Created.AsTime(),
	}
}

// DeleteArtifact implements StrictServerInterface.
func (server *Server) DeleteArtifact(
	ctx context.Context,
	request DeleteArtifactRequestObject,
) (DeleteArtifactResponseObject, error) {
	artifactIdentifier := &pb.ArtifactIdentifier{Fqn: &pb.FullyQualifiedName{
		Source: request.Body.Fqn.Source,
		Author: request.Body.Fqn.Author,
		Name:   request.Body.Fqn.Name,
	}}

	if after, ok := strings.CutPrefix(request.Body.Identifier, VersionHashPrefix); ok {
		artifactIdentifier.Identifier = &pb.ArtifactIdentifier_VersionHash{
			VersionHash: after,
		}
	} else {
		artifactIdentifier.Identifier = &pb.ArtifactIdentifier_Tag{
			Tag: request.Body.Identifier,
		}
	}

	artifact, err := server.registryClient.DeleteArtifact(ctx, artifactIdentifier)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return DeleteArtifact404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete artifact")

		return &DeleteArtifact500Response{}, nil
	}

	return DeleteArtifact200JSONResponse(artifactToArtifact(artifact)), nil
}

// DeleteArtifactTag implements StrictServerInterface.
func (server *Server) DeleteArtifactTag(
	ctx context.Context,
	request DeleteArtifactTagRequestObject,
) (DeleteArtifactTagResponseObject, error) {
	removeTagRequest := &pb.AddRemoveTagRequest{
		Fqn: &pb.FullyQualifiedName{
			Source: request.Body.Fqn.Source,
			Author: request.Body.Fqn.Author,
			Name:   request.Body.Fqn.Name,
		},
		VersionHash: request.Body.VersionHash,
		Tag:         request.Body.Tag,
	}

	_, err := server.registryClient.RemoveTag(ctx, removeTagRequest)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return DeleteArtifactTag404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to remove tag from artifact")

		return &DeleteArtifactTag500Response{}, nil
	}

	// DeleteArtifactTag returns 200 with no body on success
	return DeleteArtifactTag200Response{}, nil
}

// GetArtifact implements StrictServerInterface.
func (server *Server) GetArtifact(
	ctx context.Context,
	request GetArtifactRequestObject,
) (GetArtifactResponseObject, error) {
	artifactIdentifier := &pb.ArtifactIdentifier{Fqn: &pb.FullyQualifiedName{
		Source: request.Params.Source,
		Author: request.Params.Author,
		Name:   request.Params.Name,
	}}

	if after, ok := strings.CutPrefix(request.Params.Identifier, VersionHashPrefix); ok {
		artifactIdentifier.Identifier = &pb.ArtifactIdentifier_VersionHash{
			VersionHash: after,
		}
	} else {
		artifactIdentifier.Identifier = &pb.ArtifactIdentifier_Tag{
			Tag: request.Params.Identifier,
		}
	}

	artifact, err := server.registryClient.GetArtifact(ctx, artifactIdentifier)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return GetArtifact404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get artifact")

		return &GetArtifact500Response{}, nil
	}

	return GetArtifact200JSONResponse(artifactToArtifact(artifact)), nil
}

// GetArtifactList implements StrictServerInterface.
func (server *Server) GetArtifactList(
	ctx context.Context,
	request GetArtifactListRequestObject,
) (GetArtifactListResponseObject, error) {
	query := &pb.ArtifactQuery{}

	if request.Params.Source != nil {
		query.Source = request.Params.Source
	}

	if request.Params.Author != nil {
		query.Author = request.Params.Author
	}

	if request.Params.Name != nil {
		query.Name = request.Params.Name
	}

	response, err := server.registryClient.QueryArtifacts(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query artifacts")

		return &GetArtifactList500Response{}, nil
	}

	artifacts := make([]Artifact, len(response.Artifacts))
	for i, artifact := range response.Artifacts {
		artifacts[i] = artifactToArtifact(artifact)
	}

	return GetArtifactList200JSONResponse(artifacts), nil
}

// GetArtifactUpload implements StrictServerInterface.
func (server *Server) GetArtifactUpload(
	ctx context.Context,
	request GetArtifactUploadRequestObject,
) (GetArtifactUploadResponseObject, error) {
	pullRequest := &pb.ArtifactIdentifier{
		Fqn: &pb.FullyQualifiedName{
			Source: request.Params.Source,
			Author: request.Params.Author,
			Name:   request.Params.Name,
		},
	}

	if after, ok := strings.CutPrefix(request.Params.Identifier, VersionHashPrefix); ok {
		pullRequest.Identifier = &pb.ArtifactIdentifier_VersionHash{
			VersionHash: after,
		}
	} else {
		pullRequest.Identifier = &pb.ArtifactIdentifier_Tag{
			Tag: request.Params.Identifier,
		}
	}

	stream, err := server.registryClient.PullArtifact(ctx, pullRequest)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return GetArtifactUpload404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to pull artifact")

		return &GetArtifactUpload500Response{}, nil
	}

	// Read all chunks from the stream
	var buffer bytes.Buffer
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("Failed to receive artifact chunk")

			return &GetArtifactUpload500Response{}, nil
		}

		_, err = buffer.Write(chunk.Data)
		if err != nil {
			log.Error().Err(err).Msg("Failed to write artifact chunk")

			return &GetArtifactUpload500Response{}, nil
		}
	}

	return GetArtifactUpload200ApplicationoctetStreamResponse{
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentLength: int64(buffer.Len()),
	}, nil
}

// HeadArtifact implements StrictServerInterface.
func (server *Server) HeadArtifact(
	ctx context.Context,
	request HeadArtifactRequestObject,
) (HeadArtifactResponseObject, error) {
	artifactIdentifier := &pb.ArtifactIdentifier{Fqn: &pb.FullyQualifiedName{
		Source: request.Params.Source,
		Author: request.Params.Author,
		Name:   request.Params.Name,
	}}

	if after, ok := strings.CutPrefix(request.Params.Identifier, VersionHashPrefix); ok {
		artifactIdentifier.Identifier = &pb.ArtifactIdentifier_VersionHash{
			VersionHash: after,
		}
	} else {
		artifactIdentifier.Identifier = &pb.ArtifactIdentifier_Tag{
			Tag: request.Params.Identifier,
		}
	}

	_, err := server.registryClient.GetArtifact(ctx, artifactIdentifier)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return HeadArtifact404Response{}, nil
		}

		log.Error().Err(err).Msg("Failed to check artifact existence")

		return HeadArtifact500Response{}, nil
	}

	return HeadArtifact200Response{}, nil
}

// PostArtifactTag implements StrictServerInterface.
func (server *Server) PostArtifactTag(
	ctx context.Context,
	request PostArtifactTagRequestObject,
) (PostArtifactTagResponseObject, error) {
	addTagRequest := &pb.AddRemoveTagRequest{
		Fqn: &pb.FullyQualifiedName{
			Source: request.Body.Fqn.Source,
			Author: request.Body.Fqn.Author,
			Name:   request.Body.Fqn.Name,
		},
		VersionHash: request.Body.VersionHash,
		Tag:         request.Body.NewTag,
	}

	_, err := server.registryClient.AddTag(ctx, addTagRequest)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return PostArtifactTag404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to add tag to artifact")

		return &PostArtifactTag500Response{}, nil
	}

	// PostArtifactTag returns 201 with no body on success
	return PostArtifactTag201Response{}, nil
}

func readWithMaxLength(
	reader io.Reader,
	maxLength int,
) (value string, tooLong bool, err error) {
	buff := make([]byte, maxLength)
	readBytes, err := io.ReadFull(reader, buff)
	if err == nil {
		// No io.EOF, means there is more that maxLength bytes
		return "", true, nil
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		log.Debug().Err(err).Msg("Failed to read field")
		//nolint:wrapcheck // Propagate read errors to be handled by caller
		return "", false, err
	}

	return string(buff[:readBytes]), false, nil
}

// PostArtifactUpload implements StrictServerInterface.
func (server *Server) PostArtifactUpload(
	ctx context.Context,
	request PostArtifactUploadRequestObject,
) (PostArtifactUploadResponseObject, error) {
	// First pass: collect all metadata fields
	var source, author, name string
	var tags []string

	for {
		part, err := request.Body.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("Failed to read multipart form")

			return &PostArtifactUpload500Response{}, nil
		}

		switch part.FormName() {
		case "source":
			if source != "" {
				log.Error().Msg("Multiple source fields provided")

				return PostArtifactUpload400JSONResponse{
					GenericBadRequestJSONResponse{
						Error: "Multiple source fields provided",
					},
				}, nil
			}

			data, tooLong, err := readWithMaxLength(part, MetadataFieldMaxSize)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read source field")

				return &PostArtifactUpload500Response{}, nil
			}

			if tooLong {
				log.Error().Msg("Source field too long")

				return PostArtifactUpload413JSONResponse{
					GenericTooLargeJSONResponse{
						Error: "Source field is too long",
					},
				}, nil
			}

			source = data
		case "author":
			if author != "" {
				log.Error().Msg("Multiple author fields provided")

				return PostArtifactUpload400JSONResponse{
					GenericBadRequestJSONResponse{
						Error: "Multiple author fields provided",
					},
				}, nil
			}

			data, tooLong, err := readWithMaxLength(part, MetadataFieldMaxSize)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read author field")

				return &PostArtifactUpload500Response{}, nil
			}

			if tooLong {
				log.Error().Msg("Author field too long")

				return PostArtifactUpload413JSONResponse{
					GenericTooLargeJSONResponse{
						Error: "Author field is too long",
					},
				}, nil
			}

			author = data
		case "name":
			if name != "" {
				log.Error().Msg("Multiple name fields provided")

				return PostArtifactUpload400JSONResponse{
					GenericBadRequestJSONResponse{
						Error: "Multiple name fields provided",
					},
				}, nil
			}

			data, tooLong, err := readWithMaxLength(part, MetadataFieldMaxSize)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read name field")

				return &PostArtifactUpload500Response{}, nil
			}

			if tooLong {
				log.Error().Msg("Name field too long")

				return PostArtifactUpload413JSONResponse{
					GenericTooLargeJSONResponse{
						Error: "Name field is too long",
					},
				}, nil
			}

			name = data
		case "tag":
			data, tooLong, err := readWithMaxLength(
				part,
				MetadataFieldMaxSize-len(strings.Join(tags, "")),
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read tags field")

				return &PostArtifactUpload500Response{}, nil
			}

			if tooLong {
				log.Error().Msg("Tags field too long")

				return PostArtifactUpload413JSONResponse{
					GenericTooLargeJSONResponse{
						Error: "Tags field is too long",
					},
				}, nil
			}

			tags = append(tags, data)
		case "file":
			// Process file immediately - don't store the part reader
			return server.uploadArtifactWithFile(
				ctx,
				source,
				author,
				name,
				tags,
				part,
			)
		default:
			log.Error().
				Str("field_name", part.FormName()).
				Msg("Unexpected form field encountered")

			return &PostArtifactUpload400JSONResponse{
				GenericBadRequestJSONResponse{
					Error: "Encountered an unexpected form field: " + part.FormName(),
				},
			}, nil
		}
	}

	// File part not provided
	log.Error().Msg("File part not provided")

	return PostArtifactUpload400JSONResponse{
		GenericBadRequestJSONResponse{
			Error: "File part is required",
		},
	}, nil
}

func (server *Server) uploadArtifactWithFile(
	ctx context.Context,
	source, author, name string,
	tags []string,
	fileReader io.Reader,
) (PostArtifactUploadResponseObject, error) {
	stream, err := server.registryClient.UploadArtifact(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create upload artifact stream")

		return &PostArtifactUpload500Response{}, nil
	}

	//nolint:errcheck // CloseSend never returns an error
	defer stream.CloseSend()

	metadata := &pb.UploadMetadata{
		Fqn: &pb.FullyQualifiedName{
			Source: source,
			Author: author,
			Name:   name,
		},
		Tags: tags,
	}

	err = stream.Send(&pb.UploadArtifactRequest{
		Request: &pb.UploadArtifactRequest_Metadata{Metadata: metadata},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send upload artifact metadata")

		return &PostArtifactUpload500Response{}, nil
	}

	for {
		buff := make([]byte, ArtifactRegistryMaxMessageSize)
		readBytes, err := io.ReadFull(fileReader, buff)

		isEOF := errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)
		if err != nil && !isEOF {
			log.Error().Err(err).Msg("Failed to read artifact file")

			return &PostArtifactUpload500Response{}, nil
		}

		lastChunk := errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)

		if readBytes == 0 {
			break
		}

		err = stream.Send(&pb.UploadArtifactRequest{
			Request: &pb.UploadArtifactRequest_Content{
				Content: &pb.ArtifactContent{
					Data: buff[:readBytes],
				},
			},
		})

		log.Debug().Int("sent_bytes", readBytes).Msg("Sent artifact chunk")
		if err != nil {
			log.Error().Err(err).Msg("Failed to send artifact chunk")

			return &PostArtifactUpload500Response{}, nil
		}

		if lastChunk {
			break
		}
	}

	artifact, err := stream.CloseAndRecv()
	if err != nil {
		log.Error().Err(err).Msg("Failed to finalize artifact upload")

		return &PostArtifactUpload500Response{}, nil
	}

	return PostArtifactUpload201JSONResponse(artifactToArtifact(artifact)), nil
}
