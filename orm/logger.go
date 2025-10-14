// Modified from https://github.com/go-gorm/gorm/blob/master/logger/logger.go
package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type zerologConnector struct{}

// LogMode log mode
func (l *zerologConnector) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info print info
func (l *zerologConnector) Info(ctx context.Context, msg string, data ...interface{}) {
	log.Info().Ctx(ctx).Str("component", "gorm").Msgf(msg, data...)
}

// Warn print warn messages
func (l *zerologConnector) Warn(ctx context.Context, msg string, data ...interface{}) {
	log.Warn().Ctx(ctx).Str("component", "gorm").Msgf(msg, data...)
}

// Error print error messages
func (l *zerologConnector) Error(ctx context.Context, msg string, data ...interface{}) {
	log.Err(errors.New(fmt.Sprintf(msg, data...))).Ctx(ctx).Str("component", "gorm").Msg("Database error")
}

// Trace print sql message
func (l *zerologConnector) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && (!errors.Is(err, logger.ErrRecordNotFound)):
		sql, rows := fc()
		if rows == -1 {
			log.Trace().Str("component", "gorm").Err(err).Msg(sql)
		} else {
			log.Trace().Str("component", "gorm").Int64("rows", rows).Err(err).Msg(sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// ParamsFilter filter params
func (l *zerologConnector) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}

type traceRecorder struct {
	Interface
	BeginAt      time.Time
	SQL          string
	RowsAffected int64
	Err          error
}

// New trace recorder
func (l *traceRecorder) New() *traceRecorder {
	return &traceRecorder{Interface: l.Interface, BeginAt: time.Now()}
}

// Trace implement logger interface
func (l *traceRecorder) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.BeginAt = begin
	l.SQL, l.RowsAffected = fc()
	l.Err = err
}

func (l *traceRecorder) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if RecorderParamsFilter == nil {
		return sql, params
	}
	return RecorderParamsFilter(ctx, sql, params...)
}
