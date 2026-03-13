package database

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"wellness_tracker/internal/models"
)

// New opens a new Gorm database connection using the provided Postgres DSN.
func New(dsn string) (*gorm.DB, error) {
	gormLogger := logger.New(
		log.New(os.Stdout, "gorm: ", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Warn,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Reasonable defaults; tune as needed.
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return db, nil
}

// AutoMigrate runs database migrations for all models.
func AutoMigrate(db *gorm.DB) error {
	// db.Exec("TRUNCATE TABLE checkins RESTART IDENTITY CASCADE;")
	// return nil
	return db.AutoMigrate(&models.Checkin{}, &models.User{})
}
