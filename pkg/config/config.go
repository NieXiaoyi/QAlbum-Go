package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Storage  StorageConfig  `yaml:"storage"`
	JWT      JWTConfig      `yaml:"jwt"`
	WeChat   WeChatConfig   `yaml:"wechat"`
	Task     TaskConfig     `yaml:"task"`
	Backup   BackupConfig   `yaml:"backup"`
	Cleanup  CleanupConfig  `yaml:"cleanup"`
	Audit    AuditConfig    `yaml:"audit"`
}

type ServerConfig struct {
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	Env          string `yaml:"env"`
}

type StorageConfig struct {
	BasePath string `yaml:"base_path"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireHour int    `yaml:"expire_hour"`
}

type WeChatConfig struct {
	AppID  string `yaml:"app_id"`
	Secret string `yaml:"secret"`
}

type TaskConfig struct {
	WorkerCount    int `yaml:"worker_count"`
	QueueSize      int `yaml:"queue_size"`
	CoverSemaphore int `yaml:"cover_semaphore"`
	CoverTimeout   int `yaml:"cover_timeout"`
}

type BackupConfig struct {
	Enabled          bool `yaml:"enabled"`
	UploadGraceHours int  `yaml:"upload_grace_hours"`
}

type CleanupConfig struct {
	Enabled bool `yaml:"enabled"`
	Hour    int  `yaml:"hour"`
}

type AuditConfig struct {
	Enabled       bool `yaml:"enabled"`
	IntervalHours int  `yaml:"interval_hours"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}
