package api

import (
	pb "api-server/proto_gen"
	"bytes"
	"cmp"
	"context"
	"errors"
	"io"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	VersionHashPrefix              = "hash:"
	MetadataFieldMaxSize           = 1024 * 10       // 10KiB
	ArtifactRegistryMaxMessageSize = 1024 * 1024 * 3 // 3MiB: Make sure to keep below gRPC max message size of 4MiB
)

// GetV1Artifact implements StrictServerInterface.
func (server *Server) GetV1Artifact(ctx context.Context, request GetV1ArtifactRequestObject) (GetV1ArtifactResponseObject, error) {
	artifactQueryResponse, err := server.registryClient.QueryArtifacts(ctx, &pb.ArtifactQuery{})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to query artifacts")
	}

	artifactsPaginated := paginate(artifactQueryResponse.Artifacts, *request.Params.Limit, *request.Params.Offset, cmpArtifacts)

	responseArtifacts := make([]Artifact, len(artifactsPaginated))
	for i, artifact := range artifactsPaginated {
		responseArtifacts[i] = artifactToArtifact(artifact)
	}

	return GetV1Artifact200JSONResponse(responseArtifacts), nil
}

// GetV1ArtifactNamespace implements StrictServerInterface.
func (server *Server) GetV1ArtifactNamespace(ctx context.Context, request GetV1ArtifactNamespaceRequestObject) (GetV1ArtifactNamespaceResponseObject, error) {
	artifactQueryResponse, err := server.registryClient.QueryArtifacts(ctx, &pb.ArtifactQuery{
		Namespace: &request.Namespace,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to query artifacts")
	}

	artifactsPaginated := paginate(artifactQueryResponse.Artifacts, *request.Params.Limit, *request.Params.Offset, cmpArtifacts)

	responseArtifacts := make([]Artifact, len(artifactsPaginated))
	for i, artifact := range artifactsPaginated {
		responseArtifacts[i] = artifactToArtifact(artifact)
	}

	return GetV1ArtifactNamespace200JSONResponse(responseArtifacts), nil
}

// GetV1ArtifactNamespaceName implements StrictServerInterface.
func (server *Server) GetV1ArtifactNamespaceName(ctx context.Context, request GetV1ArtifactNamespaceNameRequestObject) (GetV1ArtifactNamespaceNameResponseObject, error) {
	artifactQueryResponse, err := server.registryClient.QueryArtifacts(ctx, &pb.ArtifactQuery{
		Namespace: &request.Namespace,
		Name:      &request.Name,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to query artifacts")
	}

	artifactsPaginated := paginate(artifactQueryResponse.Artifacts, *request.Params.Limit, *request.Params.Offset, cmpArtifacts)

	responseArtifacts := make([]Artifact, len(artifactsPaginated))
	for i, artifact := range artifactsPaginated {
		responseArtifacts[i] = artifactToArtifact(artifact)
	}

	return GetV1ArtifactNamespaceName200JSONResponse(responseArtifacts), nil
}

// GetV1ArtifactNamespaceNameHashHash implements [StrictServerInterface].
func (server *Server) GetV1ArtifactNamespaceNameHashHash(
	ctx context.Context,
	request GetV1ArtifactNamespaceNameHashHashRequestObject,
) (GetV1ArtifactNamespaceNameHashHashResponseObject, error) {
	artifactResponse, err := server.registryClient.GetArtifact(ctx, &pb.ArtifactIdentifier{
		Package: &pb.PackageName{
			Namespace: request.Namespace,
			Name:      request.Name,
		},
		Identifier: &pb.ArtifactIdentifier_VersionHash{
			VersionHash: request.Hash,
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return GetV1ArtifactNamespaceNameHashHash404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get artifact")

		return &GetV1ArtifactNamespaceNameHashHash500Response{}, nil
	}

	return GetV1ArtifactNamespaceNameHashHash200JSONResponse(artifactToArtifact(artifactResponse)), nil
}

// GetV1ArtifactNamespaceNameTagTag implements [StrictServerInterface].
func (server *Server) GetV1ArtifactNamespaceNameTagTag(
	ctx context.Context,
	request GetV1ArtifactNamespaceNameTagTagRequestObject,
) (GetV1ArtifactNamespaceNameTagTagResponseObject, error) {
	artifactResponse, err := server.registryClient.GetArtifact(ctx, &pb.ArtifactIdentifier{
		Package: &pb.PackageName{
			Namespace: request.Namespace,
			Name:      request.Name,
		},
		Identifier: &pb.ArtifactIdentifier_Tag{
			Tag: request.Tag,
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return GetV1ArtifactNamespaceNameTagTag404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to get artifact")

		return &GetV1ArtifactNamespaceNameTagTag500Response{}, nil
	}

	return GetV1ArtifactNamespaceNameTagTag200JSONResponse(artifactToArtifact(artifactResponse)), nil
}

// GetV1ArtifactRawNamespaceNameHashHash implements [StrictServerInterface].
func (server *Server) GetV1ArtifactRawNamespaceNameHashHash(
	ctx context.Context,
	request GetV1ArtifactRawNamespaceNameHashHashRequestObject,
) (GetV1ArtifactRawNamespaceNameHashHashResponseObject, error) {
	pullRequest := &pb.ArtifactIdentifier{
		Package: &pb.PackageName{
			Namespace: request.Namespace,
			Name:      request.Name,
		},
		Identifier: &pb.ArtifactIdentifier_VersionHash{
			VersionHash: request.Hash,
		},
	}

	stream, err := server.registryClient.PullArtifact(ctx, pullRequest)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return GetV1ArtifactRawNamespaceNameHashHash404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to pull artifact")

		return &GenericInternalServerErrorResponse{}, nil
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

			return &GenericInternalServerErrorResponse{}, nil
		}

		_, err = buffer.Write(chunk.Data)
		if err != nil {
			log.Error().Err(err).Msg("Failed to write artifact chunk")

			return &GenericInternalServerErrorResponse{}, nil
		}
	}

	return GetV1ArtifactRawNamespaceNameHashHash200ApplicationoctetStreamResponse{
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentLength: int64(buffer.Len()),
	}, nil
}

// GetV1ArtifactRawNamespaceNameTagTag implements [StrictServerInterface].
func (server *Server) GetV1ArtifactRawNamespaceNameTagTag(
	ctx context.Context,
	request GetV1ArtifactRawNamespaceNameTagTagRequestObject,
) (GetV1ArtifactRawNamespaceNameTagTagResponseObject, error) {
	pullRequest := &pb.ArtifactIdentifier{
		Package: &pb.PackageName{
			Namespace: request.Namespace,
			Name:      request.Name,
		},
		Identifier: &pb.ArtifactIdentifier_Tag{
			Tag: request.Tag,
		},
	}

	stream, err := server.registryClient.PullArtifact(ctx, pullRequest)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return GetV1ArtifactRawNamespaceNameTagTag404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to pull artifact")

		return &GenericInternalServerErrorResponse{}, nil
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

			return &GenericInternalServerErrorResponse{}, nil
		}

		_, err = buffer.Write(chunk.Data)
		if err != nil {
			log.Error().Err(err).Msg("Failed to write artifact chunk")

			return &GenericInternalServerErrorResponse{}, nil
		}
	}

	return GetV1ArtifactRawNamespaceNameTagTag200ApplicationoctetStreamResponse{
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentLength: int64(buffer.Len()),
	}, nil
}

// PostV1ArtifactRawNamespaceName implements [StrictServerInterface].
func (server *Server) PostV1ArtifactRawNamespaceName(
	ctx context.Context,
	request PostV1ArtifactRawNamespaceNameRequestObject,
) (PostV1ArtifactRawNamespaceNameResponseObject, error) {
	stream, err := server.registryClient.UploadArtifact(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create upload artifact stream")

		return &GenericInternalServerErrorResponse{}, nil
	}

	//nolint:errcheck // CloseSend never returns an error
	defer stream.CloseSend()

	metadata := &pb.UploadMetadata{
		Fqn: &pb.PackageName{
			Namespace: request.Namespace,
			Name:      request.Name,
		},
	}

	err = stream.Send(&pb.UploadArtifactRequest{
		Request: &pb.UploadArtifactRequest_Metadata{Metadata: metadata},
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send upload artifact metadata")

		return &GenericInternalServerErrorResponse{}, nil
	}

	for {
		buff := make([]byte, ArtifactRegistryMaxMessageSize)
		readBytes, err := io.ReadFull(request.Body, buff)

		isEOF := errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, io.EOF)
		if err != nil && !isEOF {
			log.Error().Err(err).Msg("Failed to read artifact file")

			return &GenericInternalServerErrorResponse{}, nil
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

			return &GenericInternalServerErrorResponse{}, nil
		}

		if lastChunk {
			break
		}
	}

	artifact, err := stream.CloseAndRecv()
	if err != nil {
		// Check if it's a gRPC AlreadyExists error
		if status.Code(err) == codes.AlreadyExists {
			log.Warn().Err(err).Msg("Artifact already exists")

			return PostV1ArtifactRawNamespaceName409JSONResponse{
				Error: "An artifact already exists with this fully qualified name and version hash",
			}, nil
		}

		log.Error().Err(err).Msg("Failed to finalize artifact upload")

		return &GenericInternalServerErrorResponse{}, nil
	}

	return PostV1ArtifactRawNamespaceName201JSONResponse{
		VersionHash: artifact.VersionHash,
	}, nil
}

// PatchV1ArtifactNamespaceNameHashHash implements [StrictServerInterface].
func (server *Server) PatchV1ArtifactNamespaceNameHashHash(
	ctx context.Context,
	request PatchV1ArtifactNamespaceNameHashHashRequestObject,
) (PatchV1ArtifactNamespaceNameHashHashResponseObject, error) {
	requestParams := pb.SetTagsRequest{
		Artifact: &pb.ArtifactIdentifier{
			Package: &pb.PackageName{
				Namespace: request.Namespace,
				Name:      request.Name,
			},
			Identifier: &pb.ArtifactIdentifier_VersionHash{
				VersionHash: request.Hash,
			},
		},
	}

	if request.Body.Tags == nil {
		requestParams.Tags = []string{}
	} else {
		requestParams.Tags = *request.Body.Tags
	}

	artifactResponse, err := server.registryClient.SetTags(ctx, &requestParams)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return PatchV1ArtifactNamespaceNameHashHash404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		case codes.InvalidArgument:
			return PatchV1ArtifactNamespaceNameHashHash400JSONResponse{
				GenericBadRequestJSONResponse{
					Error: "Invalid tags provided: " + err.Error(),
				},
			}, nil
		default:
			log.Error().Err(err).Msg("Failed to set artifact tags")
			return &GenericInternalServerErrorResponse{}, nil
		}
	}

	return PatchV1ArtifactNamespaceNameHashHash200JSONResponse(artifactToArtifact(artifactResponse)), nil
}

// PatchV1ArtifactNamespaceNameTagTag implements [StrictServerInterface].
func (server *Server) PatchV1ArtifactNamespaceNameTagTag(
	ctx context.Context,
	request PatchV1ArtifactNamespaceNameTagTagRequestObject,
) (PatchV1ArtifactNamespaceNameTagTagResponseObject, error) {
	requestParams := pb.SetTagsRequest{
		Artifact: &pb.ArtifactIdentifier{
			Package: &pb.PackageName{
				Namespace: request.Namespace,
				Name:      request.Name,
			},
			Identifier: &pb.ArtifactIdentifier_Tag{
				Tag: request.Tag,
			},
		},
	}

	if request.Body.Tags == nil {
		requestParams.Tags = []string{}
	} else {
		requestParams.Tags = *request.Body.Tags
	}

	artifactResponse, err := server.registryClient.SetTags(ctx, &requestParams)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			return PatchV1ArtifactNamespaceNameTagTag404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		case codes.InvalidArgument:
			return PatchV1ArtifactNamespaceNameTagTag400JSONResponse{
				GenericBadRequestJSONResponse{
					Error: "Invalid tags provided: " + err.Error(),
				},
			}, nil
		default:
			log.Error().Err(err).Msg("Failed to set artifact tags")
			return &GenericInternalServerErrorResponse{}, nil
		}
	}

	return PatchV1ArtifactNamespaceNameTagTag200JSONResponse(artifactToArtifact(artifactResponse)), nil
}

// DeleteV1ArtifactNamespaceNameHashHash implements [StrictServerInterface].
func (server *Server) DeleteV1ArtifactNamespaceNameHashHash(
	ctx context.Context,
	request DeleteV1ArtifactNamespaceNameHashHashRequestObject,
) (DeleteV1ArtifactNamespaceNameHashHashResponseObject, error) {
	artifactResponse, err := server.registryClient.DeleteArtifact(ctx, &pb.ArtifactIdentifier{
		Package: &pb.PackageName{
			Namespace: request.Namespace,
			Name:      request.Name,
		},
		Identifier: &pb.ArtifactIdentifier_VersionHash{
			VersionHash: request.Hash,
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return DeleteV1ArtifactNamespaceNameHashHash404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete artifact")

		return &DeleteV1ArtifactNamespaceNameHashHash500Response{}, nil
	}

	return DeleteV1ArtifactNamespaceNameHashHash200JSONResponse(artifactToArtifact(artifactResponse)), nil
}

// DeleteV1ArtifactNamespaceNameTagTag implements [StrictServerInterface].
func (server *Server) DeleteV1ArtifactNamespaceNameTagTag(
	ctx context.Context,
	request DeleteV1ArtifactNamespaceNameTagTagRequestObject,
) (DeleteV1ArtifactNamespaceNameTagTagResponseObject, error) {
	artifactResponse, err := server.registryClient.DeleteArtifact(ctx, &pb.ArtifactIdentifier{
		Package: &pb.PackageName{
			Namespace: request.Namespace,
			Name:      request.Name,
		},
		Identifier: &pb.ArtifactIdentifier_Tag{
			Tag: request.Tag,
		},
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return DeleteV1ArtifactNamespaceNameTagTag404JSONResponse{
				GenericNotFoundJSONResponse{
					Error: "Artifact not found",
				},
			}, nil
		}

		log.Error().Err(err).Msg("Failed to delete artifact")

		return &DeleteV1ArtifactNamespaceNameTagTag500Response{}, nil
	}

	return DeleteV1ArtifactNamespaceNameTagTag200JSONResponse(artifactToArtifact(artifactResponse)), nil
}

func cmpArtifacts(a, b *pb.Artifact) int {
	if a.Package.Namespace != b.Package.Namespace {
		return cmp.Compare(a.Package.Namespace, b.Package.Namespace)
	}

	if a.Package.Name != b.Package.Name {
		return cmp.Compare(a.Package.Name, b.Package.Name)
	}

	return cmp.Compare(a.VersionHash, b.VersionHash)
}

func artifactToArtifact(artifact *pb.Artifact) Artifact {
	return Artifact{
		Package: PackageName{
			Namespace: artifact.Package.Namespace,
			Name:      artifact.Package.Name,
		},
		VersionHash: artifact.VersionHash,
		Tags:        artifact.Tags,
		Pulls:       int(artifact.Metadata.Pulls),
		CreatedAt:   artifact.Metadata.Created.AsTime(),
	}
}
