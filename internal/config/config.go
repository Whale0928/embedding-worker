package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 애플리케이션 전체 설정
type Config struct {
	HuggingFace HuggingFaceConfig
	DB          DBConfig
	Vector      VectorConfig
	HttpConfig  EchoHttpConfig
}

// HuggingFaceConfig HuggingFace 관련 설정
type HuggingFaceConfig struct {
	Token     string `mapstructure:"HUGGING_FACE_TOKEN"`
	ModelRepo string `mapstructure:"HUGGING_FACE_MODEL_REPO"`
	CacheDir  string // 환경변수 아님, 코드에서 설정
}

// DBConfig 데이터베이스 연결 설정
type DBConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	Name     string `mapstructure:"DB_NAME"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
}

// VectorConfig 벡터 DB 설정
type VectorConfig struct {
	Host string `mapstructure:"VECTOR_HOST"`
	Port string `mapstructure:"VECTOR_PORT"`
}

type EchoHttpConfig struct {
	Host string `mapstructure:"HOST"`
	Port string `mapstructure:"PORT"`
}

// Load 환경변수에서 전체 설정 로드
func Load() (*Config, error) {
	// .env 파일 읽기 (없어도 OK)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	// 환경변수 자동 바인딩
	viper.AutomaticEnv()

	// 기본값 설정
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("PORT", "8000")
	viper.SetDefault("DB_PORT", "3306")
	viper.SetDefault("VECTOR_HOST", "localhost")
	viper.SetDefault("VECTOR_PORT", "8080")

	cfg := &Config{}

	// Http 설정
	if err := viper.Unmarshal(&cfg.HttpConfig); err != nil {
		return nil, fmt.Errorf("http 설정 로드 실패: %w", err)
	}

	// HuggingFace 설정
	if err := viper.Unmarshal(&cfg.HuggingFace); err != nil {
		return nil, fmt.Errorf("HuggingFace 설정 로드 실패: %w", err)
	}

	// DB 설정
	if err := viper.Unmarshal(&cfg.DB); err != nil {
		return nil, fmt.Errorf("DB 설정 로드 실패: %w", err)
	}

	// Vector 설정
	if err := viper.Unmarshal(&cfg.Vector); err != nil {
		return nil, fmt.Errorf("vector 설정 로드 실패: %w", err)
	}

	// CacheDir 설정 (환경변수 아님)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("홈 디렉토리 조회 실패: %w", err)
	}
	cfg.HuggingFace.CacheDir = filepath.Join(homeDir, ".cache", "embedding-worker")

	// 필수값 검증
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate 필수 설정값 검증
func (c *Config) validate() error {
	if c.HuggingFace.Token == "" {
		return fmt.Errorf("HUGGING_FACE_TOKEN is required")
	}
	if c.HuggingFace.ModelRepo == "" {
		return fmt.Errorf("HUGGING_FACE_MODEL_REPO is required")
	}
	if c.DB.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.DB.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	if c.DB.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	return nil
}

// DSN MySQL 연결 문자열 생성
func (c *DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Name)
}
