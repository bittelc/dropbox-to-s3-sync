package dropbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

// File represents a file entry in Dropbox under the source path.
type File struct {
	// FullPath is the absolute path in Dropbox (e.g., "/Apps/app/data/file.txt")
	FullPath       string
	RelativePath   string
	ServerModified time.Time
	Size           int64
}

// Client wraps Dropbox files API operations used by the syncer.
type Client struct {
	files        files.Client
	memberId     string
	namespaceId  string
	appKey       string
	appSecret    string
	refreshToken string
}

// NewClient creates a new Dropbox client with app credentials and refresh token.
// It generates a fresh access token on startup.
func NewClient(refreshToken, namespaceId, memberId, appKey, appSecret string) *Client {
	log.Println("Initializing Dropbox client and generating access token...")

	// Create client instance first
	client := &Client{
		memberId:     memberId,
		namespaceId:  namespaceId,
		appKey:       appKey,
		appSecret:    appSecret,
		refreshToken: refreshToken,
	}

	// Generate access token
	accessToken, err := client.generateAccessToken()
	if err != nil {
		log.Fatalf("Failed to generate Dropbox access token: %v", err)
	}
	// Configure Dropbox SDK client with the fresh token
	cfg := dropbox.Config{
		Token: accessToken,
		// AsMemberID: memberId,
	}
	cfg = cfg.WithNamespaceID(namespaceId)
	client.files = files.New(cfg)
	return client
}

// generateAccessToken creates a new access token using the refresh token
func (c *Client) generateAccessToken() (string, error) {
	// Prepare the token request
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.refreshToken)
	data.Set("client_id", c.appKey)
	data.Set("client_secret", c.appSecret)

	// Send the request to Dropbox OAuth API
	resp, err := http.Post(
		"https://api.dropbox.com/oauth2/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("received empty access token")
	}
	return result.AccessToken, nil
}

// ListFiles recursively lists files under the configured root path and returns relative paths.
func (c *Client) ListFiles(ctx context.Context, path string) ([]File, error) {
	arg := &files.ListFolderArg{
		Path:      path,
		Recursive: true,
	}

	accum := make([]File, 0, 128)
	res, err := c.files.ListFolder(arg)
	if err != nil {
		return nil, fmt.Errorf("dropbox list_folder: %w", err)
	}
	appendEntries := func(entries []files.IsMetadata) {
		for _, e := range entries {
			fm, ok := e.(*files.FileMetadata)
			if !ok {
				continue
			}
			fullPath := fm.PathDisplay
			rel := trimPrefixPath(fullPath, path)
			accum = append(accum, File{
				FullPath:       fullPath,
				RelativePath:   rel,
				ServerModified: fm.ServerModified,
				Size:           int64(fm.Size),
			})
		}
	}
	appendEntries(res.Entries)

	// Handle pagination using the HasMore field and list_folder/continue endpoint
	for res.HasMore {
		continueArg := &files.ListFolderContinueArg{
			Cursor: res.Cursor,
		}
		res, err = c.files.ListFolderContinue(continueArg)
		if err != nil {
			return nil, fmt.Errorf("dropbox list_folder/continue: %w", err)
		}
		appendEntries(res.Entries)
	}

	log.Printf("Total files found: %d", len(accum))
	return accum, nil
}

// Download returns a ReadCloser and size for the given full path.
func (c *Client) Download(ctx context.Context, fullPath string) (io.ReadCloser, int64, error) {
	arg := &files.DownloadArg{Path: fullPath}
	_, content, err := c.files.Download(arg)
	if err != nil {
		return nil, 0, fmt.Errorf("dropbox download %s: %w", fullPath, err)
	}
	return content, -1, nil
}

func trimPrefixPath(path string, prefix string) string {
	if prefix == "" || prefix == "/" {
		if len(path) > 0 && path[0] == '/' {
			return path[1:]
		}
		return path
	}
	// normalize trailing slash for prefix matching
	pp := prefix
	if len(pp) > 1 && pp[len(pp)-1] == '/' {
		pp = pp[:len(pp)-1]
	}
	if path == pp {
		return ""
	}
	if len(path) > len(pp) && path[:len(pp)] == pp {
		rel := path[len(pp):]
		if len(rel) > 0 && rel[0] == '/' {
			rel = rel[1:]
		}
		return rel
	}
	return path
}
