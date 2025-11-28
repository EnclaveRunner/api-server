package api

import (
	"api-server/orm"
	"api-server/proto_gen"
	"api-server/queue"

	"github.com/EnclaveRunner/shareddeps/auth"
)

// ensure that we've conformed to the `ServerInterface` with
// a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	db             orm.DB
	authModule     auth.AuthModule
	registryClient proto_gen.RegistryServiceClient
	queueClient    queue.QueueClient
}

func NewServer(
	db orm.DB,
	authModule auth.AuthModule,
	registryClient proto_gen.RegistryServiceClient,
	queueClient queue.QueueClient,
) *Server {
	return &Server{
		db:             db,
		authModule:     authModule,
		registryClient: registryClient,
		queueClient:    queueClient,
	}
}

type EmptyInternalServerError struct{}

func (e *EmptyInternalServerError) Error() string {
	return "internal server error"
}
