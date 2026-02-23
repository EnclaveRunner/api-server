package orm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	// Task statuses
	TaskStatusStarted   = "STARTED"
	TaskActionSubmitted = "SUBMITTED"
	TaskActionEnqueued  = "ENQUEUED"

	// Log statuses and issuers
	LogStatusInfo   = "INFO"
	LogIssuerSystem = "SYSTEM"

	// Default values
	DefaultRetention = "24h"
)

// RegisterTask creates a new virtual task with the specified parameters.
func (db *DB) RegisterTask(
	maxRetries int,
	retention string,
) (*VirtualTask, error) {
	// Set defaults if not provided
	if maxRetries < 0 {
		log.Warn().
			Int("provided_max_retries", maxRetries).
			Msg("Invalid max_retries value, defaulting to 0")
		maxRetries = 0
	}

	_, err := time.ParseDuration(retention)
	if retention == "" || err != nil {
		log.Warn().
			Str("provided_retention", retention).
			Msg("Invalid or missing retention value, defaulting to 24h")
		retention = DefaultRetention
	}

	task := &VirtualTask{
		RunnerHost:    "",
		Retries:       0,
		MaxRetries:    maxRetries,
		Retention:     retention,
		Status:        TaskStatusStarted,
		LastAction:    TaskActionSubmitted,
		ResultPayload: nil,
	}

	// Execute task registration and logging in a transaction
	err = db.dbGorm.Transaction(func(tx *gorm.DB) error {
		// Save the task to database
		if err := tx.Create(task).Error; err != nil {
			log.Error().Err(err).Msg("Failed to register task")

			return err
		}

		// Add log entry for task submission
		taskLog := &TaskLog{
			TaskID: task.TaskID,
			Status: LogStatusInfo,
			Issuer: LogIssuerSystem,
			Payload: []byte(
				"Task registered with max_retries=" + strconv.Itoa(
					maxRetries,
				) + " and retention=" + retention,
			),
		}

		if err := tx.Create(taskLog).Error; err != nil {
			log.Error().
				Err(err).
				Str("task_id", task.TaskID.String()).
				Msg("Failed to create task log entry")

			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register task in transaction: %w", err)
	}

	log.Info().
		Str("task_id", task.TaskID.String()).
		Msg("Task registered successfully")

	return task, nil
}

// EnqueueTask updates the last_action field of a task and adds a log entry.
func (db *DB) EnqueueTask(taskID uuid.UUID) error {
	task, err := gorm.G[VirtualTask](db.dbGorm).
		Where(&VirtualTask{TaskID: taskID}).
		First(context.Background())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn().
				Str("task_id", taskID.String()).
				Msg("Task not found for enqueue operation")

			return gorm.ErrRecordNotFound
		}

		log.Error().
			Err(err).
			Str("task_id", taskID.String()).
			Msg("Failed to find task for enqueue operation")

		return fmt.Errorf("failed to find task for enqueue operation: %w", err)
	}

	// Execute task update and logging in a transaction
	err = db.dbGorm.Transaction(func(tx *gorm.DB) error {
		// Update the task's last_action field
		task.LastAction = TaskActionEnqueued
		if err := tx.Save(&task).Error; err != nil {
			log.Error().
				Err(err).
				Str("task_id", taskID.String()).
				Msg("Failed to update task last_action")

			return err
		}

		// Add log entry for task enqueue
		taskLog := &TaskLog{
			TaskID:  taskID,
			Status:  LogStatusInfo,
			Issuer:  LogIssuerSystem,
			Payload: nil,
		}

		if err := tx.Create(taskLog).Error; err != nil {
			log.Error().
				Err(err).
				Str("task_id", taskID.String()).
				Msg("Failed to create enqueue log entry")

			return err
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to enqueue task in transaction: %w", err)
	}

	log.Info().Str("task_id", taskID.String()).Msg("Task enqueued successfully")

	return nil
}

// GetAllTasks retrieves all virtual tasks from the database.
func (db *DB) GetAllTasks(ctx context.Context) ([]VirtualTask, error) {
	tasks, err := gorm.G[VirtualTask](db.dbGorm).
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}

	return tasks, nil
}

// GetTaskByID retrieves a single virtual task by its ID.
func (db *DB) GetTaskByID(
	ctx context.Context,
	taskID uuid.UUID,
) (*VirtualTask, error) {
	task, err := gorm.G[VirtualTask](db.dbGorm).
		Where(&VirtualTask{TaskID: taskID}).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}

		return nil, fmt.Errorf("failed to fetch task: %w", err)
	}

	return &task, nil
}
