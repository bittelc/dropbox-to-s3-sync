# Dropbox to S3 Sync

A Golang application that performs a one-time synchronization of files from a Dropbox directory to an Amazon S3 bucket.

## Overview

This application provides a simple and efficient way to copy files from Dropbox to AWS S3. It scans a specified Dropbox directory and transfers all files to the target S3 bucket, preserving the directory structure.

## Features

- One-time, one-way synchronization from Dropbox to S3
- Support for transferring file hierarchy and preserving structure
- Detailed logging for sync operations
- Efficient handling of large files and directories
- Support for Dropbox Business/Team accounts

## Requirements

- Go 1.18+
- Dropbox API credentials
- AWS credentials with S3 access
- An S3 bucket for destination

## Installation

Clone the repository:

```bash
git clone https://github.com/yourusername/dropbox-to-s3-sync.git
cd dropbox-to-s3-sync
```

Build the application:

```bash
go build -o dropbox-s3-sync ./cmd/sync
```

## Configuration

The application requires the following environment variables to be set:

### Required Environment Variables

| Variable | Description |
|----------|-------------|
| `DROPBOX_ACCESS_TOKEN` | Your Dropbox API access token |
| `DROPBOX_SOURCE_PATH` | Path to the Dropbox directory to sync (e.g., `/Photos`) |
| `DROPBOX_ROOT_NS` | Dropbox namespace ID for team accounts (see authentication section) |
| `AWS_ACCESS_KEY_ID` | Your AWS access key |
| `AWS_SECRET_ACCESS_KEY` | Your AWS secret key |
| `AWS_REGION` | AWS region of your S3 bucket (e.g., `us-west-2`) |
| `S3_BUCKET_NAME` | Name of the destination S3 bucket |

### Optional Environment Variables

| Variable | Description |
|----------|-------------|
| `S3_PREFIX` | (Optional) Prefix to prepend to objects in S3 (e.g., `backup/dropbox/`) |

You can set these in your environment or create a `.env` file at the root of the project.

## Dropbox Authentication Setup

### For Individual Accounts

1. Create a Dropbox app in the [Dropbox Developer Console](https://www.dropbox.com/developers/apps)
2. Generate an access token from the app settings page
3. Set the `DROPBOX_ACCESS_TOKEN` environment variable with this token
4. Leave `DROPBOX_SOURCE_ROOT_NS` blank

### For Dropbox Business/Team Accounts

Dropbox for Business requires a horribly complex authentication process:

1. **Create a Dropbox App**
   - Create a Dropbox "App" in the Developer Console
   - Grant all necessary permissions
      - files.metadata.read
      - files.content.read
      - team_member.read
      - TODO, more
   - Note the app key and app secret

2. **Get a One-Time Authorization Code**
   - Paste the following URL into a browser (replace values in < >):
     ```
     https://www.dropbox.com/oauth2/authorize?client_id=<App key from Dropbox App Console>&response_type=code&token_access_type=offline&scope=files.metadata.read files.content.write&state=<random_csrf_string>
     ```
   - You'll receive an authorization code (valid for only a few seconds)

3. **Exchange the Auth Code for a Refresh Token and Access Token**
   - Run this immediately after getting the auth code:
     ```bash
     curl https://api.dropbox.com/oauth2/token \
       -d code=<AUTH_CODE_FROM_PREVIOUS_STEP> \
       -d grant_type=authorization_code \
       -d client_id=$DROPBOX_APP_KEY \
       -d client_secret=$DROPBOX_APP_SECRET
     ```
   - This returns a refresh token and access token

4. **Generate New Access Tokens as Needed**
   - Using your refresh token:
     ```bash
      curl https://api.dropbox.com/oauth2/token \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d grant_type=refresh_token \
        -d refresh_token=$DROPBOX_REFRESH_TOKEN \
        -d client_id=$DROPBOX_APP_KEY \
        -d client_secret=$DROPBOX_APP_SECRET
     ```

5. **Get Dropbox Member ID for User to Impersonate**
   - Find the user ID in your Dropbox admin console
   - This ID looks like `dbmid:...`

6. **Get Team Space Root ID**
   - Run:
     ```bash
     curl -X POST https://api.dropboxapi.com/2/users/get_current_account \
       -H "Authorization: Bearer $DROPBOX_ACCESS_TOKEN" \
       -H "Dropbox-API-Select-User: $DROPBOX_MEMBER_ID" | jq
     ```
   - Find the namespace ID in the response
   - Set this value as `DROPBOX_ROOT_NS`

7. **Validate Access**
   - Test your setup with:
     ```bash
      curl -X POST https://api.dropboxapi.com/2/files/list_folder \
        -H "Authorization: Bearer $DROPBOX_ACCESS_TOKEN" \
        -H "Dropbox-API-Path-Root: {\".tag\": \"root\", \"root\": \"$DROPBOX_ROOT_NS\"}" \
        -H "Dropbox-API-Select-User: $DROPBOX_MEMBER_ID" \
        -H "Content-Type: application/json" \
        -d '{"path": ""}'
     ```

## Usage

Run the application to perform a one-time sync:

```bash
./dropbox-s3-sync
```

Example output:

```
2023-07-15T10:30:45 Loading configuration from .env file and environment variables...
2023-07-15T10:30:45 Configuration loaded successfully.
2023-07-15T10:30:45 Starting one-time Dropbox to S3 sync: source='/Photos' bucket='my-backup-bucket' prefix='dropbox-backup/' region='us-west-2'
2023-07-15T10:32:15 Sync completed successfully
```

## Docker Usage

You can also run the application in Docker:

```bash
docker build -t dropbox-s3-sync .
docker run --env-file .env dropbox-s3-sync
```

## Common Use Cases

- Create a backup of Dropbox content to S3
- Migrate files from Dropbox to S3
- Archive Dropbox content in S3 for long-term storage
- Create a one-time export of Dropbox files to S3 for data processing

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify your Dropbox access token is valid and has not expired
   - For team accounts, ensure you have the correct namespace ID

2. **Permission Errors**
   - Ensure your AWS credentials have permission to write to the target S3 bucket
   - Check that your Dropbox app has the necessary scopes enabled

3. **Rate Limiting**
   - For large transfers, you might hit Dropbox API rate limits
   - The app will automatically retry with backoff, but very large syncs may fail

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
