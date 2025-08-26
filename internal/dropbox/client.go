package dropbox

import (
	"context"
	"fmt"
	"io"
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
	files files.Client
	root  string
}

// NewClient creates a new Dropbox client with access token and source root path.
func NewClient(accessToken string, sourceRoot string) *Client {
	cfg := dropbox.Config{Token: accessToken}
	fc := files.New(cfg)
	return &Client{files: fc, root: sourceRoot}
}

// ListFiles recursively lists files under the configured root path and returns relative paths.
func (c *Client) ListFiles(ctx context.Context) ([]File, error) {
	arg := &files.ListFolderArg{
		Path:      c.root,
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
			rel := trimPrefixPath(fullPath, c.root)
			accum = append(accum, File{
				FullPath:       fullPath,
				RelativePath:   rel,
				ServerModified: fm.ServerModified,
				Size:           int64(fm.Size),
			})
		}
	}
	appendEntries(res.Entries)
	for res.HasMore {
		res, err = c.files.ListFolderContinue(&files.ListFolderContinueArg{Cursor: res.Cursor})
		if err != nil {
			return nil, fmt.Errorf("dropbox list_folder/continue: %w", err)
		}
		appendEntries(res.Entries)
	}
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