package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 存储所有应用程序的配置
type Config struct {
	User      UserConfig      `mapstructure:"user"`
	Location  LocationConfig  `mapstructure:"location"`
	LLM       LLMConfig       `mapstructure:"llm"`
	SignIn    SignInConfig    `mapstructure:"signin"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

// UserConfig 存储用户凭据
type UserConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// LocationConfig 存储地理位置信息
type LocationConfig struct {
	Longitude float64 `mapstructure:"longitude"`
	Latitude  float64 `mapstructure:"latitude"`
}

// LLMConfig 存储 LLM API 的配置
type LLMConfig struct {
	APIKey   string `mapstructure:"api_key"`
	Endpoint string `mapstructure:"endpoint"`
	Model    string `mapstructure:"model"`
}

// SignInConfig 存储签到相关的配置
type SignInConfig struct {
	BaseURL       string        `mapstructure:"base_url"`
	RetryTimes    int           `mapstructure:"retry_times"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`
}

// SchedulerConfig 存储定时任务的配置
type SchedulerConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Cron     string `mapstructure:"cron"`
	Timezone string `mapstructure:"timezone"`
}

// LoggingConfig 存储日志相关的配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Debug      bool   `mapstructure:"debug"` // <--- 添加 Debug 字段
}

// LoadConfig 从文件和环境变量中读取配置
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}