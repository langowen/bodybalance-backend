package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"log/slog"
	"sync"
	"time"
)

// Config содержит всю конфигурацию приложения.
type Config struct {
	Telegram    TelegramConfig `yaml:"telegram"` // Конфигурация Telegram.
	Database    DatabaseConfig `yaml:"database"` // Конфигурация базы данных.
	HTTPServer  HTTPServer     `yaml:"http_server"`
	Media       Media          `yaml:"media"`
	LogLevel    string         `yaml:"log_level" env:"LOG_LEVEL" env-default:"Info"`   // Режим логирования debug, info, warn, error
	PatchConfig string         `env:"PATCH_CONFIG" env-default:"./config/config.yaml"` // Путь к конфигурационному файлу.
}

// TelegramConfig содержит конфигурацию для работы с Telegram API.
type TelegramConfig struct {
	Token     string `yaml:"token" env:"TG_TOKEN" env-required:"true"`          // Токен для доступа к API Telegram.
	Host      string `yaml:"host" env:"TG_HOST" env-default:"api.telegram.org"` // Хост API Telegram (по умолчанию "api.telegram.org").
	BatchSize int    `yaml:"batch_size" env:"TG_BATCH_SIZE" env-default:"100"`  // Размер пакета для обработки данных.
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
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_ADDRESS" env-default:"localhost:8083"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"20s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type Media struct {
	BaseURL   string `yaml:"base_url" env:"BASE_URL" env-default:"http://localhost:8083/"` // https://api.7375.org
	VideoPath string `yaml:"video_path" env:"VIDEO_PATH" env-default:"date/video"`
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
		err := godotenv.Load(".env")
		if err != nil {
			slog.Error("Error loading .env file", sl.Err(err))
		}

		//сначала загружаем переменные окружения
		err = cleanenv.ReadEnv(instance)
		if err != nil {
			slog.Error("Error reading env", sl.Err(err))
		}

		// Затем загружаем переменные окружения из YAML файла
		err = cleanenv.ReadConfig(instance.PatchConfig, instance)
		if err != nil {
			slog.Error("Error loading .yaml file", sl.Err(err))
		}

	})
	return instance
}

// LogValue определяет форматирование при выводе в лог
func (c *Config) LogValue() logging.Value {
	return logging.GroupValue(
		// Telegram
		logging.StringAttr("telegram_host", c.Telegram.Host),
		logging.StringAttr("telegram_token", maskToken(c.Telegram.Token)),
		logging.IntAttr("batch_size", c.Telegram.BatchSize),

		// Database
		logging.StringAttr("db_host", c.Database.Host),
		logging.IntAttr("db_port", c.Database.Port),
		logging.StringAttr("db_user", c.Database.User),
		logging.StringAttr("db_name", c.Database.DBName),
		logging.StringAttr("db_ssl_mode", c.Database.SSLMode),
		logging.StringAttr("db_timeout", formatDuration(c.Database.Timeout)), // Форматируем

		//HTTPServer
		logging.StringAttr("http_host", c.HTTPServer.Address),
		logging.StringAttr("http_timeout", formatDuration(c.HTTPServer.Timeout)),
		logging.StringAttr("http_idle_timeout", formatDuration(c.HTTPServer.IdleTimeout)),

		//Media
		logging.StringAttr("base_url", c.Media.BaseURL),

		// General
		logging.StringAttr("log_level", c.LogLevel),
		logging.StringAttr("config_path", c.PatchConfig),
	)
}

// Форматирует Duration в читаемый вид (например, "20s", "1h30m")
func formatDuration(d time.Duration) string {
	return d.String() // Используем встроенный метод String()
}

// Вспомогательная функция для маскировки токенов
func maskToken(token string) string {
	if len(token) <= 5 {
		return "*****"
	}
	return token[:5] + "*****"
}
