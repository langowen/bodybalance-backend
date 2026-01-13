package config

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
)

// Config содержит всю конфигурацию приложения.
type Config struct {
	Database    DatabaseConfig `yaml:"database"` // Конфигурация базы данных.
	HTTPServer  HTTPServer     `yaml:"http_server"`
	Media       Media          `yaml:"media"`
	Docs        Docs           `yaml:"swagger"`
	Redis       Redis          `yaml:"redis"`
	LogLevel    string         `yaml:"log_level" env:"LOG_LEVEL" env-default:"Info"`   // Режим логирования debug, info, warn, error
	PatchLog    string         `yaml:"patch_log" env:"PATCH_LOG" env-default:""`       // Путь к папке для логов, если не указано, то логи будут в stdout
	PatchConfig string         `env:"PATCH_CONFIG" env-default:"./config/config.yaml"` // Путь к конфигурационному файлу.
	Env         string         `env:"ENV" env-default:"dev"`                           //dev, prod, local
	Debug       bool           `env:"DEBUG" env-default:"false"`                       // Режим отладки pprof
}

// DatabaseConfig содержит конфигурацию для работы с базой данных.
type DatabaseConfig struct {
	Timeout  time.Duration `yaml:"timeout" env:"BD_TIMEOUT" env-default:"20s"` // Тайм-аут для операций с базой данных.
	Host     string        `yaml:"host" env:"BD_HOST" env-required:"true"`
	Port     int           `yaml:"port" env:"BD_PORT" env-required:"true"`
	User     string        `yaml:"user" env:"BD_USER" env-required:"true"`
	Password string        `yaml:"password" env:"BD_PASSWORD" env-required:"true"`
	DBName   string        `yaml:"dbname" env:"BD_DBNAME" env-required:"true"`
	SSLMode  string        `yaml:"sslmode" env:"BD_SSL_MODE" env-default:"disable"`
	Schema   string        `yaml:"schema" env:"BD_SCHEMA" env-default:"dev"` //dev, public
}

type HTTPServer struct {
	Port             string        `yaml:"port" env:"HTTP_PORT" env-default:"8083"`
	Timeout          time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"10m"`
	IdleTimeout      time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	SigningKey       string        `yaml:"signing_key" env:"HTTP_SIGNING_KEY" env-default:"MY_SIGNING_KEY"`
	TokenTTL         time.Duration `yaml:"token_ttl" env:"HTTP_TOKEN_TTL" env-default:"1440h"`
	TimeLimitError   time.Duration `yaml:"time_limit" env:"HTTP_TIME_LIMIT" env-default:"10m"`
	MaxErrorCount    int           `yaml:"max_error_count" env:"HTTP_MAX_ERROR_COUNT" env-default:"5"`
	MaxErrorDuration time.Duration `yaml:"max_error_duration" env:"HTTP_MAX_ERROR_DURATION" env-default:"1m"`
	BanDuration      time.Duration `yaml:"ban_duration" env:"HTTP_BAN_DURATION" env-default:"5m"`
}

type Media struct {
	BaseURL     string `yaml:"base_url" env:"BASE_URL" env-default:"http://localhost:8083"`
	VideoPatch  string `yaml:"video_patch" env:"VIDEO_PATCH" env-default:"data/video"`
	ImagesPatch string `yaml:"images_patch" env:"IMAGES_PATCH" env-default:"data/img"`
}

type Docs struct {
	User     string `yaml:"user" env:"DOCS_USER" env-required:"true"`
	Password string `yaml:"password" env:"DOCS_PASSWORD" env-required:"true"`
}

type Redis struct {
	Host     string        `yaml:"host" env:"REDIS_HOST" env-default:"redis:6379"`
	Password string        `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	DB       int           `yaml:"db" env:"REDIS_DB" env-default:"1"`
	CacheTTL time.Duration `yaml:"ttl" env:"REDIS_TTL" env-default:"24h"`
	Enable   bool          `yaml:"enabled" env:"REDIS_ENABLED" env-default:"true"`
}

var (
	instance *Config
	once     sync.Once
)

// MustGetConfig возвращает экземпляр конфигурации.
// Конфигурация загружается только один раз при первом вызове.
func MustGetConfig() *Config {
	once.Do(func() {
		instance = &Config{} // Инициализация instance

		// Загружаем переменные окружения из .env файла
		_ = godotenv.Load(".env")

		//сначала загружаем переменные окружения
		err := cleanenv.ReadEnv(instance)
		if err != nil {
			log.Fatal("Error reading env", sl.Err(err))
		}

		// Затем загружаем переменные окружения из YAML файла
		_ = cleanenv.ReadConfig(instance.PatchConfig, instance)

	})
	return instance
}

// LogValue определяет форматирование при выводе в лог
func (c *Config) LogValue() logging.Value {
	return logging.GroupValue(
		// Database
		logging.StringAttr("db_host", c.Database.Host),
		logging.IntAttr("db_port", c.Database.Port),
		logging.StringAttr("db_user", c.Database.User),
		logging.StringAttr("db_password", "REDACTED"),
		logging.StringAttr("db_name", c.Database.DBName),
		logging.StringAttr("db_ssl_mode", c.Database.SSLMode),
		logging.StringAttr("db_timeout", formatDuration(c.Database.Timeout)),
		logging.StringAttr("db_schema", c.Database.Schema),

		//HTTPServer
		logging.StringAttr("http_host", c.HTTPServer.Port),
		logging.StringAttr("http_timeout", formatDuration(c.HTTPServer.Timeout)),
		logging.StringAttr("http_idle_timeout", formatDuration(c.HTTPServer.IdleTimeout)),
		logging.StringAttr("http_signing_key", "REDACTED"),
		logging.StringAttr("token_ttl", formatDuration(c.HTTPServer.TokenTTL)),

		//Media
		logging.StringAttr("base_url", c.Media.BaseURL),
		logging.StringAttr("video_patch", c.Media.VideoPatch),
		logging.StringAttr("images_patch", c.Media.ImagesPatch),

		//Docs
		logging.StringAttr("docs_user", c.Docs.User),
		logging.StringAttr("docs_password", "REDACTED"),

		//redis
		logging.StringAttr("redis_host", c.Redis.Host),
		logging.StringAttr("redis_password", "REDACTED"),
		logging.IntAttr("redis_db", c.Redis.DB),
		logging.StringAttr("redis_ttl", formatDuration(c.Redis.CacheTTL)),
		logging.BoolAttr("redis_enabled", c.Redis.Enable),

		// General
		logging.StringAttr("log_level", c.LogLevel),
		logging.StringAttr("patch_log", c.PatchLog),
		logging.StringAttr("config_path", c.PatchConfig),
		logging.StringAttr("env", c.Env),
	)
}

// Форматирует Duration в читаемый вид (например, "20s", "1h30m")
func formatDuration(d time.Duration) string {
	return d.String() // Используем встроенный метод String()
}
