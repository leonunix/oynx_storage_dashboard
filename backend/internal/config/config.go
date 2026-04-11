package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server     ServerConfig
	Auth       AuthConfig
	Database   DatabaseConfig
	Onyx       OnyxConfig
	Command    CommandConfig
	Operations OperationsConfig
}

type ServerConfig struct {
	Address        string
	AllowedOrigins []string
}

type AuthConfig struct {
	JWTSecret         string
	TokenTTL          time.Duration
	BootstrapUsername string
	BootstrapPassword string
	BootstrapRole     string
}

type DatabaseConfig struct {
	Path string
}

type OnyxConfig struct {
	ConfigPath string
	BinaryPath string
	SocketPath string
}

type CommandConfig struct {
	ExecTimeout      time.Duration
	StorageOpTimeout time.Duration
}

type OperationsConfig struct {
	AllowDestructiveDM bool
}

func Load() Config {
	onyxConfigPath := getenv("ONYX_STORAGE_CONFIG", "config/default.toml")
	return Config{
		Server: ServerConfig{
			Address:        getenv("ONYX_DASHBOARD_ADDR", ":8080"),
			AllowedOrigins: csvOrDefault("ONYX_DASHBOARD_ALLOWED_ORIGINS", "http://localhost:5173"),
		},
		Auth: AuthConfig{
			JWTSecret:         getenv("ONYX_DASHBOARD_JWT_SECRET", "change-me-in-production"),
			TokenTTL:          time.Duration(getenvInt("ONYX_DASHBOARD_TOKEN_TTL_HOURS", 12)) * time.Hour,
			BootstrapUsername: getenv("ONYX_DASHBOARD_ADMIN_USER", "admin"),
			BootstrapPassword: getenv("ONYX_DASHBOARD_ADMIN_PASSWORD", "onyx-admin"),
			BootstrapRole:     getenv("ONYX_DASHBOARD_ADMIN_ROLE", "admin"),
		},
		Database: DatabaseConfig{
			Path: getenv("ONYX_DASHBOARD_DB_PATH", "var/dashboard.db"),
		},
		Onyx: OnyxConfig{
			ConfigPath: onyxConfigPath,
			BinaryPath: getenv("ONYX_STORAGE_BIN", "onyx-storage"),
			SocketPath: getenv("ONYX_STORAGE_SOCKET", detectOnyxSocketPath(onyxConfigPath)),
		},
		Command: CommandConfig{
			ExecTimeout:      time.Duration(getenvInt("ONYX_DASHBOARD_EXEC_TIMEOUT_SECONDS", 10)) * time.Second,
			StorageOpTimeout: time.Duration(getenvInt("ONYX_DASHBOARD_STORAGE_OP_TIMEOUT_SECONDS", 120)) * time.Second,
		},
		Operations: OperationsConfig{
			AllowDestructiveDM: getenvBool("ONYX_DASHBOARD_ALLOW_DM_MUTATIONS", false),
		},
	}
}

func detectOnyxSocketPath(configPath string) string {
	const fallback = "/var/run/onyx-storage.sock"

	type serviceSection struct {
		SocketPath string `toml:"socket_path"`
	}
	type onyxToml struct {
		Service serviceSection `toml:"service"`
	}

	if strings.TrimSpace(configPath) == "" {
		return fallback
	}

	var cfg onyxToml
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return fallback
	}

	if path := strings.TrimSpace(cfg.Service.SocketPath); path != "" {
		return path
	}
	return fallback
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getenvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func csvOrDefault(key, fallback string) []string {
	raw := getenv(key, fallback)
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
