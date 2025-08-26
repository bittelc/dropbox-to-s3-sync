package syncer

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/yourusername/dropbox-to-s3-sync/internal/dropbox"
	s3client "github.com/yourusername/dropbox-to-s3-sync/internal/s3"
)

// Syncer orchestrates syncing files from Dropbox to S3.
type Syncer struct {
	Dropbox  *dropbox.Client
	S3       *s3client.Client
	S3Prefix string
}

// New creates a new Syncer.
func New(dbx *dropbox.Client, s3c *s3client.Client, s3Prefix string) *Syncer {
	return &Syncer{Dropbox: dbx, S3: s3c, S3Prefix: s3Prefix}
}

// RunOnce performs a single sync pass: upload new/changed files and delete removed files.
func (s *Syncer) RunOnce(ctx context.Context) error {
	log.Printf("Sync pass started")
	files, err := s.Dropbox.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("list dropbox files: %w", err)
	}
	// Build set of desired keys
	desiredKeys := make(map[string]dropbox.File, len(files))
	for _, f := range files {
		key := s.keyFor(f.RelativePath)
		desiredKeys[key] = f
	}
	// Upload or update files
	for key, f := range desiredKeys {
		head, _ := s.S3.Head(ctx, key)
		var needUpload bool = false
		if head == nil {
			needUpload = true
		} else {
			// Compare last modified and size
			s3lm := time.Time{}
			if head.LastModified != nil {
				s3lm = *head.LastModified
			}
			var s3size int64 = 0
			if head.ContentLength != nil {
				s3size = *head.ContentLength
			}
			if f.ServerModified.After(s3lm) || f.Size != s3size {
				needUpload = true
			}
		}
		if needUpload {
			if err := s.uploadFile(ctx, key, f); err != nil {
				return err
			}
		}
	}
	// Delete extraneous keys under the prefix only
	existing, err := s.S3.ListKeys(ctx, s.S3Prefix)
	if err != nil {
		return fmt.Errorf("list s3 keys: %w", err)
	}
	for _, key := range existing {
		if _, ok := desiredKeys[key]; !ok {
			if err := s.S3.Delete(ctx, key); err != nil {
				return fmt.Errorf("delete s3 %s: %w", key, err)
			}
		}
	}
	log.Printf("Sync pass completed: %d files considered", len(files))
	return nil
}

func (s *Syncer) keyFor(relative string) string {
	k := toS3Key(relative)
	return s.S3Prefix + k
}

func (s *Syncer) uploadFile(ctx context.Context, key string, f dropbox.File) error {
	log.Printf("Uploading %s -> %s", f.FullPath, key)
	rc, _, err := s.Dropbox.Download(ctx, f.FullPath)
	if err != nil {
		return fmt.Errorf("download dropbox %s: %w", f.FullPath, err)
	}
	defer rc.Close()
	if err := s.S3.Put(ctx, key, rc, -1, f.ServerModified); err != nil {
		return fmt.Errorf("put s3 %s: %w", key, err)
	}
	return nil
}

func toS3Key(relative string) string {
	// Normalize Windows-style backslashes just in case
	k := strings.ReplaceAll(relative, "\\", "/")
	return k
}