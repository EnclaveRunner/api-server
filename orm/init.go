package orm

import (
	"api-server/config"
	"context"
	"fmt"
	"strings"

	"github.com/EnclaveRunner/shareddeps/auth"
	"github.com/casbin/casbin/v2/persist"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"

	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() persist.Adapter {
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

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	adapter, err := gormadapter.NewAdapterByDBUseTableName(DB, "casbin", "rules")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create casbin adapter")
	}

	log.Debug().Msg("Successfully connected to the database")

	// Run database migrations
	err = DB.AutoMigrate(&User{}, &Auth_Basic{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	// hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(config.Cfg.Admin.Password), HashCost)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash admin password")
	}

	// generate / update default user
	DB.Save(&User{Username: config.Cfg.Admin.Username})
	adminUser, _ := gorm.G[User](
		DB,
	).Where(&User{Username: config.Cfg.Admin.Username}).
		First(context.Background())
	DB.Save(&Auth_Basic{UserID: adminUser.ID, Password: hash})

	auth.AddUserToGroup(adminUser.ID.String(), "enclaveAdmin")

	return adapter
}
