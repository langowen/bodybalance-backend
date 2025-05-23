package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"log"
	"log/slog"
	"sync"
	"time"
)

// Config содержит всю конфигурацию приложения.
type Config struct {
	Database    DatabaseConfig `yaml:"database"` // Конфигурация базы данных.
	HTTPServer  HTTPServer     `yaml:"http_server"`
	Media       Media          `yaml:"media"`
	Docs        Docs           `yaml:"docs"`
	LogLevel    string         `yaml:"log_level" env:"LOG_LEVEL" env-default:"Info"`   // Режим логирования debug, info, warn, error
	PatchConfig string         `env:"PATCH_CONFIG" env-default:"./config/config.yaml"` // Путь к конфигурационному файлу.
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
	Port        string        `yaml:"port" env:"HTTP_PORT" env-default:"8083"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"20s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
}

type Media struct {
	BaseURL   string `yaml:"base_url" env:"BASE_URL" env-default:"http://localhost:8083"` //адрес сервера https://api.7375.org
	VideoPath string `yaml:"video_path" env:"VIDEO_PATH" env-default:"data/video"`
}

type Docs struct {
	User     string `yaml:"user" env:"DOCS_USER" env-required:"true"`
	Password string `yaml:"password" env:"DOCS_PASSWORD" env-required:"true"`
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
			slog.Info("Cant loading .env config file", sl.Err(err))
		}

		//сначала загружаем переменные окружения
		err = cleanenv.ReadEnv(instance)
		if err != nil {
			log.Fatal("Error reading env", sl.Err(err))
		}

		// Затем загружаем переменные окружения из YAML файла
		err = cleanenv.ReadConfig(instance.PatchConfig, instance)
		if err != nil {
			slog.Info("Cant loading .yaml config file", sl.Err(err))
		}

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
		logging.StringAttr("db_name", c.Database.DBName),
		logging.StringAttr("db_ssl_mode", c.Database.SSLMode),
		logging.StringAttr("db_timeout", formatDuration(c.Database.Timeout)), // Форматируем

		//HTTPServer
		logging.StringAttr("http_host", c.HTTPServer.Port),
		logging.StringAttr("http_timeout", formatDuration(c.HTTPServer.Timeout)),
		logging.StringAttr("http_idle_timeout", formatDuration(c.HTTPServer.IdleTimeout)),

		//Media
		logging.StringAttr("base_url", c.Media.BaseURL),

		//Docs
		logging.StringAttr("docs_user", c.Docs.User),

		// General
		logging.StringAttr("log_level", c.LogLevel),
		logging.StringAttr("config_path", c.PatchConfig),
	)
}

// Форматирует Duration в читаемый вид (например, "20s", "1h30m")
func formatDuration(d time.Duration) string {
	return d.String() // Используем встроенный метод String()
}
