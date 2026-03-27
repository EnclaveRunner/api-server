package api

import (
	"api-server/orm"
	"api-server/proto_gen"
	"api-server/queue"
	"time"

	"github.com/EnclaveRunner/shareddeps/auth"
)

// ensure that we've conformed to the `ServerInterface` with
// a compile-time check
var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	authModule     auth.AuthModule
	db             orm.DB
	maxRetries     int
	retention      time.Duration
	queueClient    queue.QueueClient
	registryClient proto_gen.RegistryServiceClient
}

func NewServer(
	authModule auth.AuthModule,
	db orm.DB,
	maxRetries int,
	retention time.Duration,
	queueClient queue.QueueClient,
	registryClient proto_gen.RegistryServiceClient,
) *Server {
	return &Server{
		db:             db,
		authModule:     authModule,
		registryClient: registryClient,
		queueClient:    queueClient,
		maxRetries:     maxRetries,
		retention:      retention,
	}
}
