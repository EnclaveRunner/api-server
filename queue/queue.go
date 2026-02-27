package queue

import (
	"api-server/config"
	"api-server/orm"
	pb "api-server/proto_gen"
	"errors"
	"fmt"

	"github.com/hibiken/asynq"
	"google.golang.org/protobuf/proto"
)

const (
	TaskTypeNormal   = "job:normal"
	TaskQueueDefault = "default"
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
	task *pb.Task,
	opts ...asynq.Option,
) (*asynq.TaskInfo, error) {
	payload, err := proto.Marshal(task)
	if err != nil {
		return nil, &GenericError{
			fmt.Errorf("failed to marshal task proto: %w", err),
		}
	}

	queueTask := asynq.NewTask(TaskTypeNormal, payload, opts...)
	taskInfo, err := q.client.Enqueue(queueTask, opts...)
	if err != nil {
		return nil, &GenericError{err}
	}

	return taskInfo, nil
}

func (q *QueueClient) GetTask(id string) (*asynq.TaskInfo, error) {
	taskInfo, err := q.inspector.GetTaskInfo(TaskQueueDefault, id)
	if err != nil {
		if errors.Is(err, asynq.ErrTaskNotFound) {
			return nil, &TaskNotFoundError{
				Id: id,
			}
		}

		return nil, &GenericError{
			err,
		}
	}

	return taskInfo, nil
}

func (q *QueueClient) GetAllTasks() ([]*asynq.TaskInfo, error) {
	allTasks := []*asynq.TaskInfo{}
	//nolint:mnd // For now just retrieve all tasks
	pageSize := asynq.PageSize(999)

	tasks, err := q.inspector.ListActiveTasks(TaskQueueDefault, pageSize)
	if err != nil {
		return nil, &GenericError{err}
	}
	allTasks = append(allTasks, tasks...)

	tasks, err = q.inspector.ListArchivedTasks(TaskQueueDefault, pageSize)
	if err != nil {
		return nil, &GenericError{err}
	}
	allTasks = append(allTasks, tasks...)

	tasks, err = q.inspector.ListCompletedTasks(TaskQueueDefault, pageSize)
	if err != nil {
		return nil, &GenericError{err}
	}
	allTasks = append(allTasks, tasks...)

	tasks, err = q.inspector.ListPendingTasks(TaskQueueDefault, pageSize)
	if err != nil {
		return nil, &GenericError{err}
	}
	allTasks = append(allTasks, tasks...)

	tasks, err = q.inspector.ListRetryTasks(TaskQueueDefault, pageSize)
	if err != nil {
		return nil, &GenericError{err}
	}
	allTasks = append(allTasks, tasks...)

	tasks, err = q.inspector.ListScheduledTasks(TaskQueueDefault, pageSize)
	if err != nil {
		return nil, &GenericError{err}
	}
	allTasks = append(allTasks, tasks...)

	return allTasks, nil
}
