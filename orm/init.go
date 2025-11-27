package orm

import (
	"api-server/config"
	"context"
	"fmt"
	"strings"

	"github.com/EnclaveRunner/shareddeps/auth"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"

	"gorm.io/gorm/logger"

	"gorm.io/gorm"
)

type DB struct {
	dbGorm     *gorm.DB
	authModule auth.AuthModule
}

// Initializes a new database connection and runs migrations
func InitDB(cfg *config.AppConfig) *gorm.DB {
	dsn := fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode='%s'",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	dsn_redacted := strings.ReplaceAll(dsn, cfg.Database.Password, "*****")
	log.Debug().
		Msgf("Connecting to postgres using the following information: %s", dsn_redacted)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	log.Debug().Msg("Successfully connected to the database")

	// Run database migrations
	err = db.AutoMigrate(&User{}, &Auth_Basic{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to migrate database")
	}

	return db
}

func NewDB(authModule auth.AuthModule, db *gorm.DB) DB {
	return DB{dbGorm: db, authModule: authModule}
}

func NewCasbinAdapter(db *gorm.DB) *gormadapter.Adapter {
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "casbin", "rules")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create casbin adapter")
	}

	return adapter
}

// InitAdminUser creates the default admin user after auth system is initialized
func (db *DB) InitAdminUser(cfg *config.AppConfig, authModule auth.AuthModule) {
	// hash password
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(cfg.Admin.Password),
		HashCost,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash admin password")
	}

	// generate / update default user
	db.dbGorm.Save(
		&User{
			Username:    cfg.Admin.Username,
			DisplayName: cfg.Admin.DisplayName,
		},
	)
	adminUser, _ := gorm.G[User](
		db.dbGorm,
	).Where(&User{Username: cfg.Admin.Username}).
		First(context.Background())

	err = db.dbGorm.Save(&Auth_Basic{UserID: adminUser.ID, Password: hash}).Error
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create admin auth record")
	}

	err = authModule.AddUserToGroup(adminUser.ID.String(), "enclave_admin")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to add admin user to enclave_admin group")
	}
}
