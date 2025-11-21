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
	TaskTypeNormal = "job:normal"
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
) (*asynq.TaskInfo, error) {
	payload, err := proto.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task to protobuf: %w", err)
	}

	queueTask := asynq.NewTask(TaskTypeNormal, payload)

	info, err := q.Client.Enqueue(queueTask, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}
