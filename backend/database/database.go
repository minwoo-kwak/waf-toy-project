package database

import (
	"path/filepath"
	"waf-backend/models"
	"waf-backend/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(log *logrus.Logger) error {
	// 데이터베이스 파일 경로 설정
	dbPath := utils.GetEnv("DB_PATH", "/data/waf.db")
	
	// 절대 경로로 변환
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		log.WithError(err).Error("Failed to get absolute database path")
		return err
	}
	
	log.WithField("db_path", absPath).Info("Initializing SQLite database")
	
	// GORM logger 설정
	var gormLogger logger.Interface
	if utils.GetEnv("LOG_LEVEL", "info") == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}
	
	// SQLite 데이터베이스 연결
	db, err := gorm.Open(sqlite.Open(absPath), &gorm.Config{
		Logger: gormLogger,
	})
	
	if err != nil {
		log.WithError(err).Error("Failed to connect to SQLite database")
		return err
	}
	
	DB = db
	
	// Auto Migration 실행
	log.Info("Running database migrations")
	if err := db.AutoMigrate(&models.User{}, &models.CustomRule{}); err != nil {
		log.WithError(err).Error("Failed to run database migrations")
		return err
	}
	
	log.Info("Database initialized successfully")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}