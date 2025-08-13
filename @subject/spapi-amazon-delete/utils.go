package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// AddTimestampToFilename takes a filename and returns a new filename
// with a timestamp in the format YYYYMMDD-HHMMSS inserted before the extension.
func AddTimestampToFilename(filename string) string {
	// Get the current time
	now := time.Now()
	// Format the timestamp
	timestamp := now.Format("20060102-150405")

	// Get the file extension
	ext := filepath.Ext(filename)
	// Get the filename without the extension
	baseName := strings.TrimSuffix(filename, ext)

	// Create the new filename
	return fmt.Sprintf("%s-%s%s", baseName, timestamp, ext)
}
