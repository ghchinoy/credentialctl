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
	"sync"
	"time"

	"github.com/ghchinoy/credentio-contributions/go"
)

// ValidatorService wraps the native Credentio validator.
type ValidatorService struct {
	validator *credentio.Validator
	mu        sync.Mutex
}

// NewValidatorService initializes a native Credentio validator.
func NewValidatorService(opts ...credentio.Option) (*ValidatorService, error) {
	v, err := credentio.NewValidator(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize credentio validator: %w", err)
	}
	return &ValidatorService{validator: v}, nil
}

// Close cleans up native validator resources.
func (vs *ValidatorService) Close() error {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	if vs.validator != nil {
		return vs.validator.Close()
	}
	return nil
}

// ValidateFile runs C2PA validation on a single asset file.
func (vs *ValidatorService) ValidateFile(path string, mediaType string) (*credentio.ProvenanceReport, error) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	if vs.validator == nil {
		return nil, fmt.Errorf("validator service is closed")
	}
	return vs.validator.ValidateFile(path, mediaType)
}

// ValidateItem validates a FileItem and updates its Report or Err in-place.
func (vs *ValidatorService) ValidateItem(item *FileItem) {
	report, err := vs.ValidateFile(item.Path, item.MediaType)
	if err != nil {
		item.Err = err.Error()
		item.Validated = true
		return
	}
	item.Report = report
	item.Validated = true
}

// ScanFolder discovers and validates all media files in a directory.
func (vs *ValidatorService) ScanFolder(dir string, recursive bool, onProgress func(item FileItem, completed int, total int)) (*FolderScanSummary, error) {
	items, err := ScanDirectory(dir, recursive)
	if err != nil {
		return nil, err
	}

	summary := &FolderScanSummary{
		Directory:  dir,
		TotalFiles: len(items),
		Files:      make([]FileItem, len(items)),
	}

	startTime := time.Now()
	for i := range items {
		item := items[i]
		vs.ValidateItem(&item)
		summary.Files[i] = item
		if onProgress != nil {
			onProgress(item, i+1, len(items))
		}
	}
	summary.DurationSec = time.Since(startTime).Seconds()
	summary.RecalculateCounts()

	return summary, nil
}
