// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

// DownloadResult contains the result of a download operation.
type DownloadResult struct {
	Path   string // Path to the downloaded (and decompressed) file
	SHA256 string // SHA256 hash of the final decompressed file
}

// DownloadOptions configures the download behavior.
type DownloadOptions struct {
	ExpectedSHA256 string // Optional: expected SHA256 for verification
	DestDir        string // Destination directory (default: temp dir)
}

// DownloadImage downloads an image from a URL, automatically decompressing if needed.
// Supports .xz, .gz, and .zip compression.
func DownloadImage(ctx context.Context, url string, opts *DownloadOptions) (*DownloadResult, error) {
	if opts == nil {
		opts = &DownloadOptions{}
	}

	// Create destination directory
	destDir := opts.DestDir
	if destDir == "" {
		var err error
		destDir, err = os.MkdirTemp("", "turingpi-download-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	// Download file
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Determine filename and compression type
	filename := filepath.Base(url)
	compression := detectCompression(url, resp.Header.Get("Content-Type"))

	// Save the downloaded file
	downloadPath := filepath.Join(destDir, filename)
	downloadFile, err := os.Create(downloadPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create download file: %w", err)
	}

	_, err = io.Copy(downloadFile, resp.Body)
	downloadFile.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to save download: %w", err)
	}

	// Decompress if needed
	finalPath := downloadPath
	if compression != "" {
		finalPath, err = decompress(downloadPath, compression)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress: %w", err)
		}
		// Remove the compressed file
		os.Remove(downloadPath)
	}

	// Calculate SHA256
	sha256Hash, err := calculateSHA256(finalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	// Verify SHA256 if expected
	if opts.ExpectedSHA256 != "" && sha256Hash != opts.ExpectedSHA256 {
		os.Remove(finalPath)
		return nil, fmt.Errorf("SHA256 mismatch: expected %s, got %s", opts.ExpectedSHA256, sha256Hash)
	}

	return &DownloadResult{
		Path:   finalPath,
		SHA256: sha256Hash,
	}, nil
}

// detectCompression determines the compression type from URL or content type.
func detectCompression(url, contentType string) string {
	url = strings.ToLower(url)

	if strings.HasSuffix(url, ".xz") {
		return "xz"
	}
	if strings.HasSuffix(url, ".gz") || strings.HasSuffix(url, ".gzip") {
		return "gz"
	}
	if strings.HasSuffix(url, ".zip") {
		return "zip"
	}

	// Check content type
	contentType = strings.ToLower(contentType)
	if strings.Contains(contentType, "xz") || strings.Contains(contentType, "x-xz") {
		return "xz"
	}
	if strings.Contains(contentType, "gzip") {
		return "gz"
	}
	if strings.Contains(contentType, "zip") {
		return "zip"
	}

	return ""
}

// decompress decompresses a file based on the compression type.
func decompress(path, compression string) (string, error) {
	outputPath := strings.TrimSuffix(path, "."+compression)
	if outputPath == path {
		outputPath = path + ".decompressed"
	}

	switch compression {
	case "xz":
		return decompressXZ(path, outputPath)
	case "gz":
		return decompressGzip(path, outputPath)
	case "zip":
		return decompressZip(path, outputPath)
	default:
		return "", fmt.Errorf("unsupported compression: %s", compression)
	}
}

// decompressXZ decompresses an XZ file.
func decompressXZ(src, dst string) (string, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	reader, err := xz.NewReader(srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to create xz reader: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, reader)
	if err != nil {
		return "", fmt.Errorf("failed to decompress xz: %w", err)
	}

	return dst, nil
}

// decompressGzip decompresses a gzip file.
func decompressGzip(src, dst string) (string, error) {
	srcFile, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	reader, err := gzip.NewReader(srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, reader)
	if err != nil {
		return "", fmt.Errorf("failed to decompress gzip: %w", err)
	}

	return dst, nil
}

// decompressZip extracts the first file from a zip archive.
func decompressZip(src, dst string) (string, error) {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return "", fmt.Errorf("failed to open zip: %w", err)
	}
	defer reader.Close()

	if len(reader.File) == 0 {
		return "", fmt.Errorf("zip archive is empty")
	}

	// Extract the first file (assuming single-file archive for images)
	file := reader.File[0]
	srcFile, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return "", fmt.Errorf("failed to extract from zip: %w", err)
	}

	return dst, nil
}

// calculateSHA256 calculates the SHA256 hash of a file.
func calculateSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CalculateFileSHA256 is an exported version for external use.
func CalculateFileSHA256(path string) (string, error) {
	return calculateSHA256(path)
}
