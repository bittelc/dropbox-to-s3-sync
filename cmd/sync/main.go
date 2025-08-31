package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"dropbox-to-s3-sync/internal/config"
	"dropbox-to-s3-sync/internal/dropbox"
	s3client "dropbox-to-s3-sync/internal/s3"
	"dropbox-to-s3-sync/internal/syncer"
)

func main() {
	// Change working directory to project root for .env file access
	if err := setWorkingDir(); err != nil {
		log.Printf("Warning: Could not set working directory: %v", err)
	}

	log.Println("Loading configuration from .env file and environment variables...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	ctx := context.Background()
	// Init clients
	dbx := dropbox.NewClient(cfg.DropboxRefreshToken, cfg.DropboxRootNs, cfg.DropboxMemberId, cfg.DropboxAppKey, cfg.DropboxAppSecret)
	s3c, err := s3client.NewClient(ctx, cfg.AWSRegion, cfg.S3BucketName, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey)
	if err != nil {
		log.Fatalf("s3 init: %v", err)
	}
	s := syncer.New(dbx, s3c, cfg.S3Prefix)
	if err := s.RunOnce(ctx, cfg.DropboxSourcePath); err != nil {
		log.Fatalf("sync error: %v", err)
	}
	log.Println("Sync completed successfully")
}

// setWorkingDir attempts to set the working directory to the project root
// to ensure .env file can be found regardless of where the binary is run from.
func setWorkingDir() error {
	// Get the executable directory
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(execPath)

	// Check if we're in the project root (containing .env)
	if _, err := os.Stat(filepath.Join(execDir, ".env")); err == nil {
		return os.Chdir(execDir)
	}

	// Check if we're in /cmd/sync and need to go up two levels
	if filepath.Base(execDir) == "sync" && filepath.Base(filepath.Dir(execDir)) == "cmd" {
		projectRoot := filepath.Dir(filepath.Dir(execDir))
		return os.Chdir(projectRoot)
	}

	// Already in the right place or can't determine project root
	return nil
}
