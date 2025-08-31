package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AppConfig holds application configuration loaded from environment variables.
type AppConfig struct {
	DropboxRefreshToken string
	DropboxAppKey       string
	DropboxAppSecret    string
	DropboxSourcePath   string
	DropboxRootNs       string
	DropboxMemberId     string
	AWSAccessKeyID      string
	AWSSecretAccessKey  string
	AWSRegion           string
	S3BucketName        string
	S3Prefix            string
}

// Load reads configuration from .env file and environment variables, and validates required fields.
func Load() (AppConfig, error) {
	// First try to load from .env file
	envVars := make(map[string]string)
	loadEnvFile(envVars)

	// Create config with fallback order: env vars, then .env file
	cfg := AppConfig{
		DropboxRefreshToken: getConfigValue(envVars, "DROPBOX_REFRESH_TOKEN"),
		DropboxAppKey:       getConfigValue(envVars, "DROPBOX_APP_KEY"),
		DropboxAppSecret:    getConfigValue(envVars, "DROPBOX_APP_SECRET"),
		DropboxSourcePath:   getConfigValue(envVars, "DROPBOX_SOURCE_PATH"),
		DropboxRootNs:       getConfigValue(envVars, "DROPBOX_SOURCE_ROOT_NS"),
		// DropboxMemberId:     getConfigValue(envVars, "DROPBOX_MEMBER_ID"),
		AWSAccessKeyID:     getConfigValue(envVars, "AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: getConfigValue(envVars, "AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          getConfigValue(envVars, "AWS_REGION"),
		S3BucketName:       getConfigValue(envVars, "S3_BUCKET_NAME"),
		S3Prefix:           normalizePrefix(getConfigValue(envVars, "S3_PREFIX")),
	}

	missing := make([]string, 0)
	if cfg.DropboxRefreshToken == "" {
		missing = append(missing, "DROPBOX_REFRESH_TOKEN")
	}
	if cfg.DropboxAppKey == "" {
		missing = append(missing, "DROPBOX_APP_KEY")
	}
	if cfg.DropboxAppSecret == "" {
		missing = append(missing, "DROPBOX_APP_SECRET")
	}
	if cfg.AWSAccessKeyID == "" {
		missing = append(missing, "AWS_ACCESS_KEY_ID")
	}
	if cfg.AWSSecretAccessKey == "" {
		missing = append(missing, "AWS_SECRET_ACCESS_KEY")
	}
	if cfg.AWSRegion == "" {
		missing = append(missing, "AWS_REGION")
	}
	if cfg.S3BucketName == "" {
		missing = append(missing, "S3_BUCKET_NAME")
	}
	if cfg.DropboxSourcePath == "" {
		missing = append(missing, "DROPBOX_SOURCE_PATH")
	}
	if cfg.DropboxRootNs == "" {
		missing = append(missing, "DROPBOX_SOURCE_ROOT_NS")
	}
	// if cfg.DropboxMemberId == "" {
	// 	missing = append(missing, "DROPBOX_MEMBER_ID")
	// }
	if len(missing) > 0 {
		return AppConfig{}, fmt.Errorf("missing required environment variables: %v", missing)
	}

	return cfg, nil
}

func normalizePrefix(p string) string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return ""
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

// loadEnvFile reads the .env file and populates the provided map with key-value pairs.
func loadEnvFile(envVars map[string]string) {
	file, err := os.Open(".env")
	if err != nil {
		// File doesn't exist or can't be read, silently continue
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		envVars[key] = value
	}
}

// getConfigValue retrieves a configuration value, checking environment variables first,
// then falling back to the provided map of values loaded from .env.
func getConfigValue(envVars map[string]string, key string) string {
	// First check environment variable
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	// Fall back to .env file value
	return strings.TrimSpace(envVars[key])
}
