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
	"testing"

	"github.com/ghchinoy/credentio-contributions/go"
)

func TestFileItem_HumanSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{5242880, "5.0 MB"},
	}

	for _, tt := range tests {
		item := FileItem{SizeBytes: tt.bytes}
		if got := item.HumanSize(); got != tt.expected {
			t.Errorf("HumanSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestFolderScanSummary_RecalculateCounts(t *testing.T) {
	summary := FolderScanSummary{
		Files: []FileItem{
			{
				Validated: true,
				Report: &credentio.ProvenanceReport{
					HasCredentials: true,
					ActiveManifest: &credentio.Manifest{
						ValidationStatuses: []credentio.ValidationStatus{},
					},
				},
			},
			{
				Validated: true,
				Report: &credentio.ProvenanceReport{
					HasCredentials: false,
				},
			},
			{
				Validated: true,
				Report: &credentio.ProvenanceReport{
					HasCredentials: true,
					ActiveManifest: &credentio.Manifest{
						ValidationStatuses: []credentio.ValidationStatus{
							{Code: "claimSignature.invalid", Severity: credentio.SeverityError},
						},
					},
				},
			},
		},
	}

	summary.RecalculateCounts()

	if summary.TotalFiles != 3 {
		t.Errorf("expected 3 total files, got %d", summary.TotalFiles)
	}
	if summary.SignedCount != 1 {
		t.Errorf("expected 1 signed file, got %d", summary.SignedCount)
	}
	if summary.UnsignedCount != 1 {
		t.Errorf("expected 1 unsigned file, got %d", summary.UnsignedCount)
	}
	if summary.InvalidCount != 1 {
		t.Errorf("expected 1 invalid file, got %d", summary.InvalidCount)
	}
}
