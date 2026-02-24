package api

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	tasks, err := s.db.GetAllTasks(ctx)
	if err != nil {
		return GetTasksList500Response{}, fmt.Errorf(
			"failed to get all tasks: %w",
			err,
		)
	}

	taskStates := make([]TaskState, len(tasks))
	for i := range tasks {
		taskStates[i] = TaskState{
			Id:            tasks[i].TaskID.String(),
			CreatedOn:     tasks[i].CreatedOn.Format("2006-01-02-15:04"),
			LastAction:    tasks[i].LastAction,
			RunnerHost:    tasks[i].RunnerHost,
			Retries:       tasks[i].Retries,
			MaxRetries:    tasks[i].MaxRetries,
			Retention:     tasks[i].Retention,
			Status:        tasks[i].Status,
			ResultPayload: string(tasks[i].ResultPayload),
		}
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

	taskID, err := uuid.Parse(request.Params.Id)
	if err != nil {
		return GetTasksTask400JSONResponse{}, ErrInvalidTaskIDFormat
	}

	task, err := s.db.GetTaskByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return GetTasksTask404JSONResponse{}, nil
		}

		return GetTasksTask500Response{}, fmt.Errorf(
			"failed to get task by ID: %w",
			err,
		)
	}

	taskState := TaskState{
		Id:            task.TaskID.String(),
		CreatedOn:     task.CreatedOn.Format("2006-01-02-15:04"),
		LastAction:    task.LastAction,
		RunnerHost:    task.RunnerHost,
		Retries:       task.Retries,
		MaxRetries:    task.MaxRetries,
		Retention:     task.Retention,
		Status:        task.Status,
		ResultPayload: string(task.ResultPayload),
	}

	return GetTasksTask200JSONResponse(taskState), nil
}
