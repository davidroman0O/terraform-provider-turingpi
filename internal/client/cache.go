// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	// CacheLocationLocal stores images in the local filesystem.
	CacheLocationLocal = "local"
	// CacheLocationBMC stores images on the BMC via SFTP.
	CacheLocationBMC = "bmc"
	// CacheLocationNone disables caching.
	CacheLocationNone = "none"

	// BMC cache directory
	bmcCacheDir = "/tmp/tpi-cache"
)

// ImageCache manages cached images for the Turing Pi provider.
type ImageCache struct {
	client   *Client
	localDir string
}

// NewImageCache creates a new image cache manager.
func NewImageCache(client *Client) (*ImageCache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	localDir := filepath.Join(homeDir, ".cache", "terraform-provider-turingpi")
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &ImageCache{
		client:   client,
		localDir: localDir,
	}, nil
}

// GetCachedImagePath returns the path to a cached image, or empty string if not cached.
func (c *ImageCache) GetCachedImagePath(sha256 string, location string) (string, error) {
	switch location {
	case CacheLocationLocal:
		return c.getLocalCachePath(sha256)
	case CacheLocationBMC:
		return c.getBMCCachePath(sha256)
	case CacheLocationNone:
		return "", nil
	default:
		return "", fmt.Errorf("unknown cache location: %s", location)
	}
}

// CacheImage stores an image in the specified cache location.
// Returns the path where the image was cached.
func (c *ImageCache) CacheImage(localPath, sha256, location string) (string, error) {
	switch location {
	case CacheLocationLocal:
		return c.cacheLocally(localPath, sha256)
	case CacheLocationBMC:
		return c.cacheToBMC(localPath, sha256)
	case CacheLocationNone:
		return localPath, nil
	default:
		return "", fmt.Errorf("unknown cache location: %s", location)
	}
}

// getLocalCachePath checks if an image exists in the local cache.
func (c *ImageCache) getLocalCachePath(sha256 string) (string, error) {
	path := filepath.Join(c.localDir, sha256+".img")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if os.IsNotExist(err) {
		return "", nil
	} else {
		return "", err
	}
}

// getBMCCachePath checks if an image exists in the BMC cache.
func (c *ImageCache) getBMCCachePath(sha256 string) (string, error) {
	remotePath := fmt.Sprintf("%s/%s.img", bmcCacheDir, sha256)

	files, err := c.client.ListDirectory(bmcCacheDir)
	if err != nil {
		// Directory might not exist yet
		return "", nil
	}

	expectedName := sha256 + ".img"
	for _, f := range files {
		if f.Name == expectedName {
			return remotePath, nil
		}
	}

	return "", nil
}

// cacheLocally copies an image to the local cache.
func (c *ImageCache) cacheLocally(srcPath, sha256 string) (string, error) {
	destPath := filepath.Join(c.localDir, sha256+".img")

	// Check if already cached
	if _, err := os.Stat(destPath); err == nil {
		return destPath, nil
	}

	// Copy file
	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(destPath)
		return "", fmt.Errorf("failed to copy to cache: %w", err)
	}

	return destPath, nil
}

// cacheToBMC uploads an image to the BMC cache.
func (c *ImageCache) cacheToBMC(localPath, sha256 string) (string, error) {
	remotePath := fmt.Sprintf("%s/%s.img", bmcCacheDir, sha256)

	// Ensure cache directory exists on BMC
	_, err := c.client.ExecuteCommand(fmt.Sprintf("mkdir -p %s", bmcCacheDir))
	if err != nil {
		return "", fmt.Errorf("failed to create BMC cache directory: %w", err)
	}

	// Check if already cached
	existingPath, err := c.getBMCCachePath(sha256)
	if err != nil {
		return "", err
	}
	if existingPath != "" {
		return existingPath, nil
	}

	// Upload to BMC
	if err := c.client.UploadFile(localPath, remotePath); err != nil {
		return "", fmt.Errorf("failed to upload to BMC: %w", err)
	}

	return remotePath, nil
}

// CleanLocalCache removes all cached images from the local cache.
func (c *ImageCache) CleanLocalCache() error {
	entries, err := os.ReadDir(c.localDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".img" {
			path := filepath.Join(c.localDir, entry.Name())
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove cached file: %w", err)
			}
		}
	}

	return nil
}

// CleanBMCCache removes all cached images from the BMC cache.
func (c *ImageCache) CleanBMCCache() error {
	_, err := c.client.ExecuteCommand(fmt.Sprintf("rm -rf %s", bmcCacheDir))
	if err != nil {
		return fmt.Errorf("failed to clean BMC cache: %w", err)
	}
	return nil
}
