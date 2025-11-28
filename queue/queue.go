package queue

import (
	"api-server/config"
	pb "api-server/proto_gen"
	"fmt"

	"github.com/hibiken/asynq"
	"google.golang.org/protobuf/proto"
)

const (
	TaskTypeNormal = "job:normal"
)

type QueueClient struct {
	client    *asynq.Client
	inspector *asynq.Inspector
}

func NewQueueClient(cfg *config.AppConfig) QueueClient {
	redisOpt := asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	}

	return QueueClient{
		client:    asynq.NewClient(redisOpt),
		inspector: asynq.NewInspector(redisOpt),
	}
}

func (q *QueueClient) EnqueueTask(
	task *pb.Task,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	payload, err := proto.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task to protobuf: %w", err)
	}

	queueTask := asynq.NewTask(TaskTypeNormal, payload)

	info, err := q.client.Enqueue(queueTask, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return info, nil
}
