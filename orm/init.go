package orm

import (
	"api-server/config"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

type GormLogger interface {
	LogMode(level gorm.Logger)
}

func InitDB() {
	dsn := fmt.Sprintf("host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode='%s'",
		config.Cfg.Database.Host,
		config.Cfg.Database.Port,
		config.Cfg.Database.Username,
		config.Cfg.Database.Password,
		config.Cfg.Database.Database,
		config.Cfg.Database.SSLMode,
	)

	dsn_redacted := strings.ReplaceAll(dsn, config.Cfg.Database.Password, "*****")
	log.Debug().Msgf("Connecting to postgres using the following information: %s", dsn_redacted)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: log.Logger,
	})

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	DB = db
}
