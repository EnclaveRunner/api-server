package queue

import (
	"api-server/config"
	pb "api-server/proto_gen"
	"fmt"

	"github.com/hibiken/asynq"
	"google.golang.org/protobuf/proto"
)

var Q Queue

const (
	TASK_TYPE_NORMAL = "job:normal"
	LOGGING_INTERVAL = 100 // seconds
)

type Queue struct {
	Client    *asynq.Client
	Inspector *asynq.Inspector
}

func Init() {
	redisOpt := asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%d", config.Cfg.Redis.Host, config.Cfg.Redis.Port),
		DB:   config.Cfg.Redis.DB,
	}

	Q.Client = asynq.NewClient(redisOpt)
	Q.Inspector = asynq.NewInspector(redisOpt)
}

func (q *Queue) EnqueueTask(
	task *pb.Task,
	opts ...asynq.Option,
) (*asynq.Task, error) {
	payload, err := proto.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task to protobuf: %w", err)
	}

	queueTask := asynq.NewTask(TASK_TYPE_NORMAL, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue task: %w", err)
	}

	return queueTask, nil
}
