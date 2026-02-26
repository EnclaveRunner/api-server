package api

import (
	"api-server/orm"
	"context"
	"errors"
	"time"

	"github.com/EnclaveRunner/shareddeps/utils"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

var (
	// ErrRequestBodyRequired is returned when request body is missing
	ErrRequestBodyRequired = errors.New("request body is required")
	// ErrInvalidTaskIDFormat is returned when task ID format is invalid
	ErrInvalidTaskIDFormat = errors.New("invalid task ID format")
)

// GetTasksList implements the GET /tasks/list endpoint
func (s *Server) GetTasksList(
	ctx context.Context,
	request GetTasksListRequestObject,
) (GetTasksListResponseObject, error) {
	tasks, err := s.queueClient.GetAllTasks()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list tasks")

		return GetTasksList500Response{}, nil
	}

	taskStates := make([]TaskState, len(tasks))
	for _, task := range tasks {
		taskStates = append(taskStates, taskToTaskState(*task))
	}

	return GetTasksList200JSONResponse(taskStates), nil
}

// GetTasksTask implements the GET /tasks/task endpoint
func (s *Server) GetTasksTask(
	ctx context.Context,
	request GetTasksTaskRequestObject,
) (GetTasksTaskResponseObject, error) {
	if request.Params.Id == "" {
		return GetTasksTask400JSONResponse{}, ErrRequestBodyRequired
	}

	task, err := s.queueClient.GetTask(request.Params.Id)
	if err != nil {
		if !errors.Is(err, asynq.ErrTaskNotFound) {
			log.Error().
				Err(err).
				Str("id", request.Params.Id).
				Msg("Failed to retrieve task")

			return GetTasksTask500Response{}, nil
		}

		return GetTasksTask404JSONResponse{}, nil
	}

	state := taskToTaskState(*task)

	logs, err := s.db.GetLogsOfTask(ctx, request.Params.Id)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve logs of task")
	}

	state.Logs = utils.Ptr(dbLogsToJsonLogs(logs))

	return GetTasksTask200JSONResponse(state), nil
}

func taskToTaskState(task asynq.TaskInfo) TaskState {
	state := TaskState{
		Id:         task.ID,
		Retries:    task.Retried,
		MaxRetries: task.MaxRetry,
		State:      task.State.String(),
		Retention:  task.Retention.String(),
	}

	if task.Result != nil {
		state.ResultPayload = utils.Ptr(string(task.Result))
	}

	if task.LastErr != "" {
		state.LastError = &task.LastErr
	}

	if !task.LastFailedAt.IsZero() {
		state.LastFailedAt = utils.Ptr(task.LastFailedAt.Format(time.RFC3339))
	}

	if !task.NextProcessAt.IsZero() {
		state.NextProcessAt = utils.Ptr(task.NextProcessAt.Format(time.RFC3339))
	}

	if !task.CompletedAt.IsZero() {
		state.CompletedAt = utils.Ptr(task.CompletedAt.Format(time.RFC3339))
	}

	return state
}

func dbLogsToJsonLogs(dbLogs []orm.TaskLog) []TaskLog {
	jsonLogs := make([]TaskLog, len(dbLogs))

	for _, log := range dbLogs {
		jsonLogs = append(jsonLogs, TaskLog{
			Issuer:    log.Issuer,
			Level:     log.Level,
			Message:   log.Message,
			Timestamp: log.Timestamp.Format(time.RFC3339),
		})
	}

	return jsonLogs
}
