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
	"os"
	"path/filepath"
	"testing"
)

func TestIsSupportedMedia(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"image.jpg", true},
		{"photo.JPEG", true},
		{"banner.PNG", true},
		{"clip.mp4", true},
		{"audio.mp3", true},
		{"document.pdf", false},
		{"data.json", false},
		{"script.go", false},
	}

	for _, tt := range tests {
		if got := IsSupportedMedia(tt.path); got != tt.expected {
			t.Errorf("IsSupportedMedia(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestScanDirectory(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tempDir, "test1.jpg"), []byte("jpg data"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "test2.png"), []byte("png data"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "ignore.txt"), []byte("text data"), 0644)

	subDir := filepath.Join(tempDir, "sub")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "nested.webp"), []byte("webp data"), 0644)

	// Non-recursive scan
	items, err := ScanDirectory(tempDir, false)
	if err != nil {
		t.Fatalf("ScanDirectory non-recursive failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 files non-recursive, got %d", len(items))
	}

	// Recursive scan
	recItems, err := ScanDirectory(tempDir, true)
	if err != nil {
		t.Fatalf("ScanDirectory recursive failed: %v", err)
	}
	if len(recItems) != 3 {
		t.Errorf("expected 3 files recursive, got %d", len(recItems))
	}
}
