// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var supportedExtensions = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".avif": "image/avif",
	".heic": "image/heic",
	".heif": "image/heif",
	".tif":  "image/tiff",
	".tiff": "image/tiff",
	".mp4":  "video/mp4",
	".mov":  "video/quicktime",
	".m4a":  "audio/mp4",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".c2pa": "application/x-c2pa-manifest-store",
}

// IsSupportedMedia returns true if the file extension is a supported C2PA format.
func IsSupportedMedia(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := supportedExtensions[ext]
	return ok
}

// DetectMediaType returns the IANA MIME type for a given media file path.
func DetectMediaType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if mime, ok := supportedExtensions[ext]; ok {
		return mime
	}
	return ""
}

// SupportedExtensionsList returns a human-readable list of supported file extensions.
func SupportedExtensionsList() string {
	exts := make([]string, 0, len(supportedExtensions))
	for ext := range supportedExtensions {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return strings.Join(exts, ", ")
}

// ScanDirectory discovers all supported media files within a directory.
func ScanDirectory(dir string, recursive bool) ([]FileItem, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(absDir)
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		// Single file target
		if IsSupportedMedia(absDir) {
			return []FileItem{
				{
					Path:      absDir,
					Filename:  filepath.Base(absDir),
					SizeBytes: fi.Size(),
					MediaType: DetectMediaType(absDir),
				},
			}, nil
		}
		return nil, nil
	}

	var items []FileItem

	if !recursive {
		entries, err := os.ReadDir(absDir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			fullPath := filepath.Join(absDir, entry.Name())
			if IsSupportedMedia(fullPath) {
				info, err := entry.Info()
				size := int64(0)
				if err == nil {
					size = info.Size()
				}
				items = append(items, FileItem{
					Path:      fullPath,
					Filename:  entry.Name(),
					SizeBytes: size,
					MediaType: DetectMediaType(fullPath),
				})
			}
		}
	} else {
		err := filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip unreadable paths
			}
			if d.IsDir() {
				if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasPrefix(d.Name(), ".") {
				return nil
			}
			if IsSupportedMedia(path) {
				info, err := d.Info()
				size := int64(0)
				if err == nil {
					size = info.Size()
				}
				relPath, relErr := filepath.Rel(absDir, path)
				displayName := path
				if relErr == nil {
					displayName = relPath
				}
				items = append(items, FileItem{
					Path:      path,
					Filename:  displayName,
					SizeBytes: size,
					MediaType: DetectMediaType(path),
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return strings.ToLower(items[i].Filename) < strings.ToLower(items[j].Filename)
	})

	return items, nil
}
