package main

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/dropbox-to-s3-sync/internal/config"
	"github.com/yourusername/dropbox-to-s3-sync/internal/dropbox"
	s3client "github.com/yourusername/dropbox-to-s3-sync/internal/s3"
	"github.com/yourusername/dropbox-to-s3-sync/internal/syncer"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	// Init clients
	dbx := dropbox.NewClient(cfg.DropboxAccessToken, cfg.DropboxSourcePath)
	s3c, err := s3client.NewClient(ctx, cfg.AWSRegion, cfg.S3BucketName, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey)
	if err != nil {
		log.Fatalf("s3 init: %v", err)
	}

	s := syncer.New(dbx, s3c, cfg.S3Prefix)

	interval := time.Duration(cfg.SyncIntervalSec) * time.Second
	if interval <= 0 {
		interval = 0
	}

	log.Printf("Starting Dropbox to S3 sync: source='%s' bucket='%s' prefix='%s' region='%s' interval=%ds", cfg.DropboxSourcePath, cfg.S3BucketName, cfg.S3Prefix, cfg.AWSRegion, cfg.SyncIntervalSec)
	if interval == 0 {
		if err := s.RunOnce(ctx); err != nil {
			log.Fatalf("sync error: %v", err)
		}
		return
	}

	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if err := s.RunOnce(ctx); err != nil {
			log.Printf("sync error: %v", err)
		}
		<-t.C
	}
}