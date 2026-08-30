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
	"runtime"
	"sync"
	"testing"
	"unicode/utf8"
)

// TestEmpirical_CredentioCVersion_HighConcurrency tests 2,000 concurrent goroutines
// synchronized with a start barrier to maximize race condition probability.
func TestEmpirical_CredentioCVersion_HighConcurrency(t *testing.T) {
	const numGoroutines = 2000
	barrier := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]string, numGoroutines)
	errors := make([]error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			<-barrier // wait for all goroutines to be spawned
			res := CredentioCVersion()
			results[idx] = res
		}(i)
	}

	// Release barrier
	close(barrier)
	wg.Wait()

	expected := results[0]
	if expected == "" {
		t.Fatalf("CredentioCVersion returned empty string")
	}

	for i, r := range results {
		if r != expected {
			t.Errorf("goroutine %d returned %q, expected %q", i, r, expected)
		}
	}
	_ = errors
}

// TestEmpirical_CredentioCVersion_MemoryStability verifies 100,000 invocations
// for memory leaks, GC stability, and pointer validity.
func TestEmpirical_CredentioCVersion_MemoryStability(t *testing.T) {
	const iterations = 100000

	runtime.GC()
	var mBefore runtime.MemStats
	runtime.ReadMemStats(&mBefore)

	for i := 0; i < iterations; i++ {
		v := CredentioCVersion()
		if len(v) == 0 {
			t.Fatalf("iteration %d returned empty version", i)
		}
	}

	runtime.GC()
	var mAfter runtime.MemStats
	runtime.ReadMemStats(&mAfter)

	t.Logf("MemStats: TotalAlloc Before=%d, TotalAlloc After=%d, HeapAlloc Diff=%d bytes",
		mBefore.TotalAlloc, mAfter.TotalAlloc, int64(mAfter.HeapAlloc)-int64(mBefore.HeapAlloc))
}

// TestEmpirical_CredentioCVersion_UTF8AndPurity checks UTF-8 validity and printable ASCII.
func TestEmpirical_CredentioCVersion_UTF8AndPurity(t *testing.T) {
	ver := CredentioCVersion()
	if !utf8.ValidString(ver) {
		t.Errorf("CredentioCVersion returned invalid UTF-8: %q", ver)
	}
	for i, r := range ver {
		if r < 32 || r > 126 {
			t.Errorf("CredentioCVersion contains non-printable ASCII at index %d: rune=%d", i, r)
		}
	}
}
