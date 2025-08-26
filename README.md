# Dropbox to S3 Sync

A Golang application that synchronizes files from a Dropbox directory to an Amazon S3 bucket.

## Overview

This application provides a seamless way to sync files between Dropbox and AWS S3. It monitors a specified Dropbox directory and ensures that any changes (additions, modifications, or deletions) are reflected in the target S3 bucket.

## Features

- One-way synchronization from Dropbox to S3
- Support for file additions, modifications, and deletions
- Configurable sync intervals
- Detailed logging for sync operations
- Optional file filtering based on patterns

## Requirements

- Go 1.16+
- Dropbox API credentials
- AWS credentials with S3 access
- An S3 bucket for destination

## Installation

```bash
go get github.com/yourusername/dropbox-to-s3-sync
```

## Configuration

The application requires the following environment variables to be set:

- `DROPBOX_ACCESS_TOKEN`: Your Dropbox access token
- `AWS_ACCESS_KEY_ID`: Your AWS access key
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret key
- `AWS_REGION`: AWS region of your S3 bucket
- `S3_BUCKET_NAME`: Name of the destination S3 bucket
- `DROPBOX_SOURCE_PATH`: Path to the Dropbox directory to sync
- `SYNC_INTERVAL`: Time interval between syncs (in seconds, default: 300)

## Usage

```bash
dropbox-to-s3-sync
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.