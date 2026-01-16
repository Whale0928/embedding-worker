package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB GORM DB 인스턴스 생성
func NewDB(cfg *DBConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{
		// FK 제약 생성 안 함
		DisableForeignKeyConstraintWhenMigrating: true,
		// 로그 레벨 (개발 중엔 Info, 운영에선 Warn)
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("DB 연결 실패: %w", err)
	}

	return db, nil
}
