package services

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/leonunix/onyx_storage/dashboard/backend/internal/domain"
	"github.com/leonunix/onyx_storage/dashboard/backend/internal/system"
)

// ConfigService reads/writes the onyx-storage TOML config file
// and communicates with the running engine via Unix socket IPC.
type ConfigService struct {
	configPath string
	socketPath string
	runner     *system.Runner
}

func NewConfigService(configPath, socketPath string, runner *system.Runner) *ConfigService {
	return &ConfigService{configPath: configPath, socketPath: socketPath, runner: runner}
}

// Read parses the TOML config file into a structured EngineConfig.
func (s *ConfigService) Read() (domain.EngineConfig, error) {
	var cfg domain.EngineConfig

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet — return empty config (bare mode)
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config: %w", err)
	}

	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// Write serialises EngineConfig to TOML and atomically replaces the config file.
func (s *ConfigService) Write(cfg domain.EngineConfig) error {
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(s.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Atomic write: temp file + rename
	tmp := s.configPath + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmp, s.configPath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}

// Reload sends a "reload" command to the running engine via Unix socket IPC.
func (s *ConfigService) Reload() error {
	resp, err := s.sendCommand("reload")
	if err != nil {
		return fmt.Errorf("reload: %w", err)
	}
	_ = resp
	return nil
}

// RestartService restarts the onyx-storage systemd service.
// Required for changes that cannot be hot-reloaded (shard count, devices, etc.).
func (s *ConfigService) RestartService(ctx context.Context) error {
	_, err := s.runner.Run(ctx, "systemctl", "restart", "onyx-storage")
	return err
}

// Mode queries the current engine mode (bare/standby/active) via IPC.
func (s *ConfigService) Mode() (string, error) {
	lines, err := s.sendCommand("mode")
	if err != nil {
		return "unknown", err
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed, nil
		}
	}
	return "unknown", nil
}

func (s *ConfigService) sendCommand(command string) ([]string, error) {
	if _, err := os.Stat(s.socketPath); err != nil {
		return nil, fmt.Errorf("socket not found: %w", err)
	}

	conn, err := net.Dial("unix", s.socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if _, err := fmt.Fprintf(conn, "%s\n", command); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conn)
	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "ok" || strings.HasPrefix(line, "ok ") {
			break
		}
		if strings.HasPrefix(line, "error:") {
			return nil, errors.New(strings.TrimSpace(strings.TrimPrefix(line, "error:")))
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}
