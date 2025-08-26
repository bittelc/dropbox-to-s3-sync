package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// AppConfig holds application configuration loaded from environment variables.
type AppConfig struct {
	DropboxAccessToken string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	S3BucketName       string
	S3Prefix           string
	DropboxSourcePath  string
	SyncIntervalSec    int
}

// Load reads configuration from environment variables and validates required fields.
func Load() (AppConfig, error) {
	cfg := AppConfig{
		DropboxAccessToken: strings.TrimSpace(os.Getenv("DROPBOX_ACCESS_TOKEN")),
		AWSAccessKeyID:     strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID")),
		AWSSecretAccessKey: strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY")),
		AWSRegion:          strings.TrimSpace(os.Getenv("AWS_REGION")),
		S3BucketName:       strings.TrimSpace(os.Getenv("S3_BUCKET_NAME")),
		S3Prefix:           normalizePrefix(strings.TrimSpace(os.Getenv("S3_PREFIX"))),
		DropboxSourcePath:  strings.TrimSpace(os.Getenv("DROPBOX_SOURCE_PATH")),
		SyncIntervalSec:    300,
	}

	if v := strings.TrimSpace(os.Getenv("SYNC_INTERVAL")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.SyncIntervalSec = n
		}
	}

	missing := make([]string, 0)
	if cfg.DropboxAccessToken == "" {
		missing = append(missing, "DROPBOX_ACCESS_TOKEN")
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