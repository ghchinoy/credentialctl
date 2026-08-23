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

package folderview

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
)

type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterSigned
	FilterUnsigned
	FilterInvalid
)

func (f FilterMode) String() string {
	switch f {
	case FilterSigned:
		return "Signed Only"
	case FilterUnsigned:
		return "Unsigned Only"
	case FilterInvalid:
		return "Invalid Only"
	default:
		return "All Files"
	}
}

// Msg definitions
type ScanCompletedMsg struct {
	Summary *engine.FolderScanSummary
}

type ScanProgressMsg struct {
	Item      engine.FileItem
	Completed int
	Total     int
}

type InspectFileMsg struct {
	Item engine.FileItem
}

// Model represents the folder review state.
type Model struct {
	Directory    string
	Recursive    bool
	Validator    *engine.ValidatorService
	Items        []engine.FileItem
	Filtered     []engine.FileItem
	Cursor       int
	Filter       FilterMode
	Scanning     bool
	Progress     int
	Total        int
	Spinner      spinner.Model
	Width        int
	Height       int
	ShowHelp     bool
	SummaryStats engine.FolderScanSummary
}

// NewModel creates an initial folderview model.
func NewModel(dir string, recursive bool, validator *engine.ValidatorService) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.ColorAccent)

	absDir, _ := filepath.Abs(dir)
	return Model{
		Directory: absDir,
		Recursive: recursive,
		Validator: validator,
		Spinner:   s,
		Scanning:  true,
		Width:     80,
		Height:    24,
	}
}

// Init starts the initial directory scan.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		m.startScan(),
	)
}

func (m Model) startScan() tea.Cmd {
	return func() tea.Msg {
		summary, err := m.Validator.ScanFolder(m.Directory, m.Recursive, nil)
		if err != nil {
			return ScanCompletedMsg{Summary: &engine.FolderScanSummary{Directory: m.Directory}}
		}
		return ScanCompletedMsg{Summary: summary}
	}
}

// Update handles UI events and messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	case spinner.TickMsg:
		if m.Scanning {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ScanCompletedMsg:
		m.Scanning = false
		if msg.Summary != nil {
			m.Items = msg.Summary.Files
			m.SummaryStats = *msg.Summary
		}
		m.applyFilter()
		if m.Cursor >= len(m.Filtered) {
			m.Cursor = max(0, len(m.Filtered)-1)
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Filtered)-1 {
				m.Cursor++
			}
		case "home", "g":
			m.Cursor = 0
		case "end", "G":
			if len(m.Filtered) > 0 {
				m.Cursor = len(m.Filtered) - 1
			}
		case "tab", "f":
			m.Filter = (m.Filter + 1) % 4
			m.applyFilter()
			m.Cursor = 0
		case "r":
			m.Scanning = true
			m.Items = nil
			m.Filtered = nil
			m.Cursor = 0
			cmds = append(cmds, m.Spinner.Tick, m.startScan())
		case "enter":
			if len(m.Filtered) > 0 && m.Cursor < len(m.Filtered) {
				selected := m.Filtered[m.Cursor]
				return m, func() tea.Msg {
					return InspectFileMsg{Item: selected}
				}
			}
		case "?":
			m.ShowHelp = !m.ShowHelp
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) applyFilter() {
	if m.Filter == FilterAll {
		m.Filtered = m.Items
		return
	}
	m.Filtered = nil
	for _, item := range m.Items {
		switch m.Filter {
		case FilterSigned:
			if item.Badge() == credentio.BadgeSigned {
				m.Filtered = append(m.Filtered, item)
			}
		case FilterUnsigned:
			if item.Badge() == credentio.BadgeUnsigned {
				m.Filtered = append(m.Filtered, item)
			}
		case FilterInvalid:
			if item.Badge() == credentio.BadgeInvalid {
				m.Filtered = append(m.Filtered, item)
			}
		}
	}
}

// View renders the folder review screen.
func (m Model) View() string {
	var sb strings.Builder

	// Header banner
	headerText := fmt.Sprintf(" CREDENTIALCTL  •  C2PA FOLDER REVIEW: %s ", m.Directory)
	sb.WriteString(theme.StyleHeader.Render(headerText) + "\n")

	// Stats and filter bar
	stats := fmt.Sprintf(" Total: %d  |  Signed: %d  |  Unsigned: %d  |  Invalid: %d ",
		m.SummaryStats.TotalFiles, m.SummaryStats.SignedCount, m.SummaryStats.UnsignedCount, m.SummaryStats.InvalidCount)
	
	filterBadge := lipgloss.NewStyle().
		Background(lipgloss.Color("#2E2E3E")).
		Foreground(theme.ColorID).
		Padding(0, 1).
		Render(fmt.Sprintf("Filter: %s (f/tab)", m.Filter.String()))

	timing := ""
	if m.SummaryStats.DurationSec > 0 {
		timing = theme.StyleMuted.Render(fmt.Sprintf(" (%.2fs)", m.SummaryStats.DurationSec))
	}

	if m.Scanning {
		sb.WriteString(fmt.Sprintf("\n %s Scanning media assets in directory...%s\n\n", m.Spinner.View(), timing))
	} else {
		sb.WriteString(fmt.Sprintf("\n %s  %s%s\n\n", theme.StyleCard.Padding(0, 1).Render(stats), filterBadge, timing))
	}

	// Table Header
	colStatus := lipgloss.NewStyle().Width(14).Bold(true).Render("STATUS")
	colName := lipgloss.NewStyle().Width(28).Bold(true).Render("FILENAME")
	colSize := lipgloss.NewStyle().Width(10).Bold(true).Render("SIZE")
	colMime := lipgloss.NewStyle().Width(14).Bold(true).Render("TYPE")
	colSigner := lipgloss.NewStyle().Width(24).Bold(true).Render("SIGNER / GENERATOR")
	colTime := lipgloss.NewStyle().Width(10).Bold(true).Render("WALL TIME")

	tableHeader := lipgloss.JoinHorizontal(lipgloss.Left, colStatus, colName, colSize, colMime, colSigner, colTime)
	sb.WriteString(" " + theme.StyleMuted.Render(tableHeader) + "\n")
	sb.WriteString(" " + theme.StyleMuted.Render(strings.Repeat("─", max(80, m.Width-4))) + "\n")

	// Table Rows
	if len(m.Filtered) == 0 {
		if m.Scanning {
			sb.WriteString(theme.StyleMuted.Render("  Discovering assets...") + "\n")
		} else {
			sb.WriteString(theme.StyleMuted.Render("  No media files match current filter.") + "\n")
		}
	} else {
		// Calculate viewport window
		maxRows := max(5, m.Height-10)
		start := 0
		if m.Cursor >= maxRows {
			start = m.Cursor - maxRows + 1
		}
		end := min(len(m.Filtered), start+maxRows)

		for i := start; i < end; i++ {
			item := m.Filtered[i]
			statusCell := theme.RenderBadge(item.Badge())
			
			nameStr := item.Filename
			if len(nameStr) > 26 {
				nameStr = nameStr[:23] + "..."
			}
			nameCell := lipgloss.NewStyle().Width(28).Render(nameStr)
			sizeCell := lipgloss.NewStyle().Width(10).Render(item.HumanSize())
			
			mimeStr := item.MediaType
			if len(mimeStr) > 12 {
				mimeStr = mimeStr[:11] + "…"
			}
			mimeCell := theme.StyleMuted.Width(14).Render(mimeStr)

			signerGen := item.Signer()
			if signerGen == "—" {
				signerGen = item.Generator()
			}
			if len(signerGen) > 22 {
				signerGen = signerGen[:20] + "..."
			}
			signerCell := theme.StyleID.Width(24).Render(signerGen)

			wallTimeStr := "—"
			if item.Report != nil && item.Report.ElapsedSeconds > 0 {
				wallTimeStr = fmt.Sprintf("%.1fms", item.Report.ElapsedSeconds*1000)
			}
			timeCell := theme.StyleMuted.Width(10).Render(wallTimeStr)

			rowStr := lipgloss.JoinHorizontal(lipgloss.Left,
				lipgloss.NewStyle().Width(14).Render(statusCell),
				nameCell,
				sizeCell,
				mimeCell,
				signerCell,
				timeCell,
			)

			if i == m.Cursor {
				rowStr = lipgloss.NewStyle().
					Background(lipgloss.Color("#2E2E3E")).
					Bold(true).
					Render(" " + rowStr + " ")
			} else {
				rowStr = "  " + rowStr
			}

			sb.WriteString(rowStr + "\n")
		}
	}

	// Footer / Keymap
	sb.WriteString("\n")
	helpBar := " [↑/↓/j/k] Navigate  •  [Enter] Inspect File  •  [f/Tab] Filter  •  [r] Rescan  •  [q] Quit "
	sb.WriteString(theme.StyleMuted.Render(helpBar) + "\n")

	return sb.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
