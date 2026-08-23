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
	"fmt"
	"github.com/ghchinoy/credentio-contributions/go"
)

// FileItem represents a single media asset file and its C2PA validation state.
type FileItem struct {
	Path      string                      `json:"path"`
	Filename  string                      `json:"filename"`
	SizeBytes int64                       `json:"size_bytes"`
	MediaType string                      `json:"media_type"`
	Validated bool                        `json:"validated"`
	Report    *credentio.ProvenanceReport `json:"report,omitempty"`
	Err       string                      `json:"error,omitempty"`
}

// HumanSize returns a formatted string of the file size (e.g. 1.25 MB).
func (f FileItem) HumanSize() string {
	const unit = 1024.0
	if f.SizeBytes < unit {
		return fmt.Sprintf("%d B", f.SizeBytes)
	}
	div, exp := int64(unit), 0
	for n := f.SizeBytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(f.SizeBytes)/float64(div), "KMGTPE"[exp])
}

// Badge returns the BadgeState for the file.
func (f FileItem) Badge() credentio.BadgeState {
	if f.Report == nil {
		return credentio.BadgeUnsigned
	}
	return f.Report.Badge()
}

// Signer returns the signature issuer or a fallback dash.
func (f FileItem) Signer() string {
	if f.Report != nil && f.Report.ActiveManifest != nil && f.Report.ActiveManifest.Signature != nil {
		if f.Report.ActiveManifest.Signature.Issuer != "" {
			return f.Report.ActiveManifest.Signature.Issuer
		}
	}
	return "—"
}

// Generator returns the claim generator or a fallback dash.
func (f FileItem) Generator() string {
	if f.Report != nil && f.Report.ActiveManifest != nil {
		if f.Report.ActiveManifest.ClaimGenerator != "" {
			return f.Report.ActiveManifest.ClaimGenerator
		}
	}
	return "—"
}

// FolderScanSummary aggregates results of validating multiple files in a folder.
type FolderScanSummary struct {
	Directory     string     `json:"directory"`
	TotalFiles    int        `json:"total_files"`
	SignedCount   int        `json:"signed_count"`
	UnsignedCount int        `json:"unsigned_count"`
	InvalidCount  int        `json:"invalid_count"`
	ErrorCount    int        `json:"error_count"`
	DurationSec   float64    `json:"duration_seconds"`
	Files         []FileItem `json:"files"`
}

// RecalculateCounts tallies the status counters across all files.
func (s *FolderScanSummary) RecalculateCounts() {
	s.TotalFiles = len(s.Files)
	s.SignedCount = 0
	s.UnsignedCount = 0
	s.InvalidCount = 0
	s.ErrorCount = 0
	for _, f := range s.Files {
		if f.Err != "" {
			s.ErrorCount++
		}
		if !f.Validated {
			continue
		}
		switch f.Badge() {
		case credentio.BadgeSigned:
			s.SignedCount++
		case credentio.BadgeUnsigned:
			s.UnsignedCount++
		case credentio.BadgeInvalid:
			s.InvalidCount++
		}
	}
}
