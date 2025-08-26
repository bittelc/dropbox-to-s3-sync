package s3client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithy "github.com/aws/smithy-go"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Client wraps AWS S3 operations used by the syncer.
type Client struct {
	bucket string
	cli    *s3.Client
}

// NewClient creates a new S3 client for the given region and bucket.
func NewClient(ctx context.Context, region string, bucket string, accessKey string, secretKey string) (*Client, error) {
	var cfg aws.Config
	var err error
	if accessKey != "" && secretKey != "" {
		creds := aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: accessKey, SecretAccessKey: secretKey}, nil
		}))
		cfg, err = config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(creds), config.WithRegion(region))
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(region))
	}
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}
	cli := s3.NewFromConfig(cfg)
	return &Client{bucket: bucket, cli: cli}, nil
}

// Head returns object metadata or nil if not found.
func (c *Client) Head(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	out, err := c.cli.HeadObject(ctx, &s3.HeadObjectInput{Bucket: &c.bucket, Key: &key})
	if err != nil {
		var re *smithyhttp.ResponseError
		if errors.As(err, &re) {
			if re.Response.StatusCode == 404 {
				return nil, nil
			}
		}
		var ae smithy.APIError
		if errors.As(err, &ae) {
			code := ae.ErrorCode()
			if code == "NoSuchKey" || code == "NotFound" {
				return nil, nil
			}
		}
		return nil, err
	}
	return out, nil
}

// Put uploads data to the given key, using the uploader for multipart if needed.
func (c *Client) Put(ctx context.Context, key string, body io.Reader, contentLength int64, lastModified time.Time) error {
	contentType := mime.TypeByExtension(filepath.Ext(key))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	uploader := manager.NewUploader(c.cli)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      &c.bucket,
		Key:         &key,
		Body:        body,
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"source":          "dropbox",
			"source-modified": lastModified.UTC().Format(time.RFC3339),
		},
	})
	return err
}

// ListKeys lists all keys under the given prefix.
func (c *Client) ListKeys(ctx context.Context, prefix string) ([]string, error) {
	keys := make([]string, 0, 128)
	var token *string
	for {
		out, err := c.cli.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            &c.bucket,
			Prefix:            aws.String(prefix),
			ContinuationToken: token,
		})
		if err != nil {
			return nil, err
		}
		for _, o := range out.Contents {
			if o.Key != nil {
				keys = append(keys, *o.Key)
			}
		}
		if out.IsTruncated != nil && *out.IsTruncated && out.NextContinuationToken != nil {
			token = out.NextContinuationToken
			continue
		}
		break
	}
	return keys, nil
}

// Delete removes the object at key.
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.cli.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &c.bucket, Key: &key})
	return err
}