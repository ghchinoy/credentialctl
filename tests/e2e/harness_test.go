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

package e2e_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"debug/elf"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// CmdResult captures the output and exit state of an executed command.
type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// RepoRoot returns the absolute path to the repository root directory.
func RepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate repository root containing go.mod starting from %s", dir)
		}
		dir = parent
	}
}

// BinaryPath returns the path to the compiled credentialctl binary.
func BinaryPath(t *testing.T) string {
	t.Helper()
	root := RepoRoot(t)
	return filepath.Join(root, "bin", "credentialctl")
}

// EnsureBuiltBinary ensures the credentialctl binary is compiled before tests execute.
func EnsureBuiltBinary(t *testing.T) string {
	t.Helper()
	root := RepoRoot(t)
	bin := BinaryPath(t)

	// Check if binary already exists and is executable
	if fi, err := os.Stat(bin); err == nil && (fi.Mode()&0111 != 0) {
		return bin
	}

	// Fetch native lib if missing
	libDarwin := filepath.Join(root, "third_party", "credentio", "lib", "libcredentio_c.dylib")
	libLinux := filepath.Join(root, "third_party", "credentio", "lib", "libcredentio_c.so")
	if _, errD := os.Stat(libDarwin); os.IsNotExist(errD) {
		if _, errL := os.Stat(libLinux); os.IsNotExist(errL) {
			fetchCmd := exec.Command("./scripts/fetch-credentio-lib.sh", "0.1.5")
			fetchCmd.Dir = root
			if out, err := fetchCmd.CombinedOutput(); err != nil {
				t.Fatalf("failed to fetch credentio native library: %v, output: %s", err, string(out))
			}
		}
	}

	// Build binary using make build
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = root
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build credentialctl binary via make build: %v, output: %s", err, string(out))
	}

	return bin
}

// RunCmd executes a command and returns the captured result.
func RunCmd(t *testing.T, dir string, name string, args ...string) CmdResult {
	t.Helper()
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	return CmdResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
		Err:      err,
	}
}

// RunCredentialctl executes the compiled credentialctl binary with arguments.
func RunCredentialctl(t *testing.T, workingDir string, args ...string) CmdResult {
	t.Helper()
	bin := EnsureBuiltBinary(t)
	return RunCmd(t, workingDir, bin, args...)
}

// GetBinaryRPATHs extracts embedded LC_RPATH load commands from Mach-O or ELF binaries.
func GetBinaryRPATHs(t *testing.T, binPath string) []string {
	t.Helper()
	var rpaths []string

	if runtime.GOOS == "darwin" {
		res := RunCmd(t, "", "otool", "-l", binPath)
		if res.ExitCode == 0 {
			lines := strings.Split(res.Stdout, "\n")
			re := regexp.MustCompile(`path\s+([^\s]+)`)
			for i, line := range lines {
				if strings.Contains(line, "cmd LC_RPATH") && i+2 < len(lines) {
					pathLine := lines[i+2]
					match := re.FindStringSubmatch(pathLine)
					if len(match) > 1 {
						rpaths = append(rpaths, match[1])
					}
				}
			}
		}
		return rpaths
	} else if runtime.GOOS == "linux" {
		elfFile, err := elf.Open(binPath)
		if err == nil {
			defer elfFile.Close()
			dynStrings, err := elfFile.DynString(elf.DT_RUNPATH)
			if err == nil {
				rpaths = append(rpaths, dynStrings...)
			}
			rpathStrings, err := elfFile.DynString(elf.DT_RPATH)
			if err == nil {
				rpaths = append(rpaths, rpathStrings...)
			}
			return rpaths
		}
	}

	return rpaths
}

// UnpackTarGz extracts a .tar.gz archive into the target directory.
func UnpackTarGz(t *testing.T, srcFile string, destDir string) error {
	t.Helper()
	f, err := os.Open(srcFile)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file %s: %w", srcFile, err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar header: %w", err)
		}

		target := filepath.Join(destDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// EmDashAuditResult represents the count and locations of em dashes in markdown prose.
type EmDashAuditResult struct {
	TotalEmDashes int
	Occurrences   []string
}

// AuditEmDashes scans a Markdown file for em dashes (— / \u2014 / &mdash;), excluding code blocks.
func AuditEmDashes(t *testing.T, filePath string) EmDashAuditResult {
	t.Helper()
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file for em dash audit %s: %v", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	inCodeBlock := false
	var occurrences []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		if strings.Contains(line, "—") || strings.Contains(line, "\u2014") || strings.Contains(line, "&mdash;") {
			occurrences = append(occurrences, fmt.Sprintf("%s:%d: %s", filepath.Base(filePath), i+1, strings.TrimSpace(line)))
		}
	}

	return EmDashAuditResult{
		TotalEmDashes: len(occurrences),
		Occurrences:   occurrences,
	}
}

// MarkAllenRubricScore represents the breakdown of the Mark Allen 8-part README rubric.
type MarkAllenRubricScore struct {
	ClearDescription  int // 1-5
	QuickInstallation int // 1-5
	ImmediateUsage    int // 1-5
	LocalDevSetup     int // 1-5
	PublishDeploy     int // 1-5
	EncourageContrib  int // 1-5
	UseMarkdownWell   int // 1-5
	OptionalExtras    int // 1-5
	TotalScore        int // out of 40
	Feedback          []string
}

// AuditMarkAllenRubric scores a README.md file against Mark Allen's 8-part quality rubric.
func AuditMarkAllenRubric(t *testing.T, readmePath string) MarkAllenRubricScore {
	t.Helper()
	raw, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", readmePath, err)
	}
	content := string(raw)

	score := MarkAllenRubricScore{}

	// 1. Clear Project Description (first 5 lines contains description)
	firstFewLines := strings.Split(content, "\n")
	hasDesc := false
	for i := 0; i < len(firstFewLines) && i < 10; i++ {
		line := strings.TrimSpace(firstFewLines[i])
		if line != "" && !strings.HasPrefix(line, "#") && len(line) > 20 {
			hasDesc = true
			break
		}
	}
	if hasDesc {
		score.ClearDescription = 5
	} else {
		score.ClearDescription = 2
		score.Feedback = append(score.Feedback, "Missing clear one-sentence project description at top")
	}

	// 2. Quick Installation Instructions (brew or download code block)
	if strings.Contains(content, "brew install") || strings.Contains(content, "brew tap") {
		score.QuickInstallation = 5
	} else if strings.Contains(content, "go install") || strings.Contains(content, "make build") {
		score.QuickInstallation = 4
	} else {
		score.QuickInstallation = 2
		score.Feedback = append(score.Feedback, "Missing quick copy-pasteable installation instructions")
	}

	// 3. Immediate Usage Example (working example command)
	if strings.Contains(content, "credentialctl validate") || strings.Contains(content, "credentialctl folder") || strings.Contains(content, "credentialctl version") {
		score.ImmediateUsage = 5
	} else {
		score.ImmediateUsage = 2
		score.Feedback = append(score.Feedback, "Missing immediate usage examples")
	}

	// 4. Local Development Setup (prerequisites, make build, make test)
	if strings.Contains(content, "make build") && strings.Contains(content, "make test") {
		score.LocalDevSetup = 5
	} else if strings.Contains(content, "go build") || strings.Contains(content, "go test") {
		score.LocalDevSetup = 4
	} else {
		score.LocalDevSetup = 2
		score.Feedback = append(score.Feedback, "Missing local development setup and test instructions")
	}

	// 5. Publish / Deploy Process (release / snapshot instructions)
	if strings.Contains(content, "release") || strings.Contains(content, "RELEASING.md") || strings.Contains(content, "goreleaser") {
		score.PublishDeploy = 5
	} else {
		score.PublishDeploy = 3
		score.Feedback = append(score.Feedback, "Missing release/publishing reference")
	}

	// 6. Encourage Contributions
	if strings.Contains(strings.ToLower(content), "contribut") || strings.Contains(content, "LICENSE") {
		score.EncourageContrib = 5
	} else {
		score.EncourageContrib = 3
		score.Feedback = append(score.Feedback, "Missing contributing or license section")
	}

	// 7. Use Markdown Well (proper headers, code fences with language tags)
	codeFenceWithLang := regexp.MustCompile("```(bash|sh|go|yaml|json)")
	if codeFenceWithLang.MatchString(content) && strings.Count(content, "#") >= 5 {
		score.UseMarkdownWell = 5
	} else {
		score.UseMarkdownWell = 3
		score.Feedback = append(score.Feedback, "Markdown hierarchy or language-hinted code fences could be improved")
	}

	// 8. Optional Extras (TUI keybindings table, badges, TOC)
	if strings.Contains(content, "|") && (strings.Contains(content, "Key") || strings.Contains(content, "Shortcut") || strings.Contains(content, "TUI")) {
		score.OptionalExtras = 5
	} else {
		score.OptionalExtras = 3
	}

	score.TotalScore = score.ClearDescription + score.QuickInstallation + score.ImmediateUsage +
		score.LocalDevSetup + score.PublishDeploy + score.EncourageContrib + score.UseMarkdownWell + score.OptionalExtras

	return score
}
