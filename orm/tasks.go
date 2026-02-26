package orm

import (
	"context"

	"gorm.io/gorm"
)

func (db *DB) GetLogsOfTask(ctx context.Context, id string) ([]TaskLog, error) {
	logs, err := gorm.G[TaskLog](db.dbGorm).Where(&TaskLog{
		TaskID: id,
	}).Find(ctx)
	if err != nil {
		return nil, &GenericError{Inner: err}
	}

	return logs, nil
}
