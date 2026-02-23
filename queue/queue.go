package queue

import (
	"api-server/config"
	"api-server/orm"
	pb "api-server/proto_gen"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"google.golang.org/protobuf/proto"
)

const (
	TaskTypeNormal = "job:normal"
)

type QueueClient struct {
	client    *asynq.Client
	inspector *asynq.Inspector
	db        *orm.DB
}

func NewQueueClient(cfg *config.AppConfig, db *orm.DB) QueueClient {
	redisOpt := asynq.RedisClientOpt{
		Addr: fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		DB:   cfg.Redis.DB,
	}

	return QueueClient{
		client:    asynq.NewClient(redisOpt),
		inspector: asynq.NewInspector(redisOpt),
		db:        db,
	}
}

func (q *QueueClient) EnqueueTask(
	taskID uuid.UUID,
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

	if err := q.db.EnqueueTask(taskID); err != nil {
		return nil, fmt.Errorf("failed to update task in database: %w", err)
	}

	return info, nil
}
