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

package inspectview

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
)

type TabIndex int

const (
	TabOverview TabIndex = iota
	TabAssertions
	TabSignature
	TabValidation
	TabRawJSON
)

type BackToFolderMsg struct{}

// Model represents the file inspector state.
type Model struct {
	Item       engine.FileItem
	ActiveTab  TabIndex
	Viewport   viewport.Model
	Width      int
	Height     int
	Ready      bool
	TabContent string
}

// NewModel creates a new file inspector model.
func NewModel(item engine.FileItem, width, height int) Model {
	vpHeight := max(10, height-9)
	vpWidth := max(40, width-4)
	vp := viewport.New(vpWidth, vpHeight)
	vp.SetContent("")

	m := Model{
		Item:      item,
		ActiveTab: TabOverview,
		Viewport:  vp,
		Width:     width,
		Height:    height,
		Ready:     true,
	}
	m.updateTabContent()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles tab switching and viewport scrolling.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Viewport.Width = max(40, msg.Width-4)
		m.Viewport.Height = max(10, msg.Height-9)
		m.updateTabContent()

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "backspace", "h":
			return m, func() tea.Msg {
				return BackToFolderMsg{}
			}
		case "tab", "right", "l":
			m.ActiveTab = (m.ActiveTab + 1) % 5
			m.updateTabContent()
		case "shift+tab", "left":
			m.ActiveTab = (m.ActiveTab + 4) % 5
			m.updateTabContent()
		case "1":
			m.ActiveTab = TabOverview
			m.updateTabContent()
		case "2":
			m.ActiveTab = TabAssertions
			m.updateTabContent()
		case "3":
			m.ActiveTab = TabSignature
			m.updateTabContent()
		case "4":
			m.ActiveTab = TabValidation
			m.updateTabContent()
		case "5":
			m.ActiveTab = TabRawJSON
			m.updateTabContent()
		default:
			var cmd tea.Cmd
			m.Viewport, cmd = m.Viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateTabContent() {
	var content string
	rep := m.Item.Report

	switch m.ActiveTab {
	case TabOverview:
		content = m.renderOverview(rep)
	case TabAssertions:
		content = m.renderAssertions(rep)
	case TabSignature:
		content = m.renderSignature(rep)
	case TabValidation:
		content = m.renderValidation(rep)
	case TabRawJSON:
		content = m.renderRawJSON(rep)
	}

	m.TabContent = content
	m.Viewport.SetContent(content)
	m.Viewport.GotoTop()
}

func (m Model) renderOverview(rep *credentio.ProvenanceReport) string {
	var sb strings.Builder
	sb.WriteString(theme.StyleSubHeader.Render("ASSET OVERVIEW") + "\n\n")

	renderRow := func(label, value string) {
		lbl := lipgloss.NewStyle().Width(18).Foreground(theme.ColorMuted).Bold(true).Render(label + ":")
		sb.WriteString(fmt.Sprintf("  %s %s\n", lbl, value))
	}

	renderRow("Filename", m.Item.Filename)
	renderRow("Full Path", m.Item.Path)
	renderRow("File Size", m.Item.HumanSize())
	renderRow("Media Type", m.Item.MediaType)
	renderRow("Status", theme.RenderBadge(m.Item.Badge()))

	if rep != nil {
		if rep.SpecVersion != "" {
			renderRow("Spec Version", fmt.Sprintf("C2PA %s", rep.SpecVersion))
		}
		if rep.CoreSeconds > 0 {
			renderRow("Core Engine", fmt.Sprintf("%.2f ms", rep.CoreSeconds*1000))
		}
		if rep.ElapsedSeconds > 0 {
			renderRow("Wall Time", fmt.Sprintf("%.2f ms", rep.ElapsedSeconds*1000))
		}

		if rep.ActiveManifest != nil {
			mft := rep.ActiveManifest
			sb.WriteString("\n" + theme.StyleSubHeader.Render("ACTIVE MANIFEST") + "\n\n")
			renderRow("Manifest Label", theme.StyleID.Render(mft.Label))
			if mft.Title != "" {
				renderRow("Title", mft.Title)
			}
			if mft.Format != "" {
				renderRow("Format", mft.Format)
			}
			if mft.ClaimGenerator != "" {
				renderRow("Claim Generator", theme.StyleAccent.Render(mft.ClaimGenerator))
			}
			renderRow("Is Update", fmt.Sprintf("%t", mft.IsUpdateManifest))
			renderRow("Assertions", fmt.Sprintf("%d attached", len(mft.Assertions)))
			renderRow("Validation Items", fmt.Sprintf("%d recorded", len(mft.ValidationStatuses)))
		} else if !rep.HasCredentials {
			sb.WriteString("\n" + theme.StyleWarn.Render("  [No C2PA Manifest Store found in this asset]") + "\n")
		}
	} else if m.Item.Err != "" {
		sb.WriteString("\n" + theme.StyleFail.Render("  Validation Error: "+m.Item.Err) + "\n")
	}

	return sb.String()
}

func (m Model) renderAssertions(rep *credentio.ProvenanceReport) string {
	if rep == nil || rep.ActiveManifest == nil || len(rep.ActiveManifest.Assertions) == 0 {
		return "\n  " + theme.StyleMuted.Render("No assertions found in active manifest.")
	}

	var sb strings.Builder
	sb.WriteString(theme.StyleSubHeader.Render(fmt.Sprintf("ASSERTIONS (%d)", len(rep.ActiveManifest.Assertions))) + "\n\n")

	for i, a := range rep.ActiveManifest.Assertions {
		kindBadge := lipgloss.NewStyle().
			Background(lipgloss.Color("#2E2E3E")).
			Foreground(theme.ColorID).
			Padding(0, 1).
			Render(string(a.Kind))

		labelStr := theme.StyleAccent.Bold(true).Render(a.Label)
		sb.WriteString(fmt.Sprintf("  %d. %s  %s\n", i+1, labelStr, kindBadge))
		if a.Summary != "" {
			sb.WriteString(fmt.Sprintf("     %s %s\n", theme.StyleMuted.Render("Summary:"), a.Summary))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderSignature(rep *credentio.ProvenanceReport) string {
	if rep == nil || rep.ActiveManifest == nil || rep.ActiveManifest.Signature == nil {
		return "\n  " + theme.StyleMuted.Render("No signature information present.")
	}

	sig := rep.ActiveManifest.Signature
	var sb strings.Builder
	sb.WriteString(theme.StyleSubHeader.Render("CRYPTOGRAPHIC SIGNATURE") + "\n\n")

	renderRow := func(label, value string) {
		lbl := lipgloss.NewStyle().Width(18).Foreground(theme.ColorMuted).Bold(true).Render(label + ":")
		sb.WriteString(fmt.Sprintf("  %s %s\n", lbl, value))
	}

	issuer := sig.Issuer
	if issuer == "" {
		issuer = "—"
	}
	renderRow("Signer / Issuer", theme.StylePass.Render(issuer))

	alg := sig.Algorithm
	if alg == "" {
		alg = "—"
	}
	renderRow("Algorithm", theme.StyleID.Render(alg))

	timeStr := "—"
	if sig.Time != nil {
		timeStr = sig.Time.Format(time.RFC1123)
	}
	renderRow("Signing Time", timeStr)

	if sig.CertChainSummary != "" {
		renderRow("Certificate SN", theme.StyleMuted.Render(sig.CertChainSummary))
	}

	return sb.String()
}

func (m Model) renderValidation(rep *credentio.ProvenanceReport) string {
	if rep == nil || rep.ActiveManifest == nil || len(rep.ActiveManifest.ValidationStatuses) == 0 {
		return "\n  " + theme.StyleMuted.Render("No validation status entries recorded.")
	}

	var sb strings.Builder
	sb.WriteString(theme.StyleSubHeader.Render(fmt.Sprintf("VALIDATION STATUSES (%d)", len(rep.ActiveManifest.ValidationStatuses))) + "\n\n")

	for i, st := range rep.ActiveManifest.ValidationStatuses {
		sevBadge := theme.RenderSeverity(st.Severity)
		codeStr := theme.StyleBold.Render(st.Code)
		sb.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, sevBadge, codeStr))
		if st.Explanation != "" {
			sb.WriteString(fmt.Sprintf("     %s\n", st.Explanation))
		}
		if st.URL != "" {
			sb.WriteString(fmt.Sprintf("     %s %s\n", theme.StyleMuted.Render("Spec:"), theme.StyleID.Render(st.URL)))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderRawJSON(rep *credentio.ProvenanceReport) string {
	if rep == nil || rep.RawJSON == "" {
		return "\n  " + theme.StyleMuted.Render("No raw JSON payload available.")
	}

	var unmarshaled interface{}
	if err := json.Unmarshal([]byte(rep.RawJSON), &unmarshaled); err == nil {
		indented, err := json.MarshalIndent(unmarshaled, "", "  ")
		if err == nil {
			return string(indented)
		}
	}
	return rep.RawJSON
}

// View renders the file inspector screen.
func (m Model) View() string {
	var sb strings.Builder

	// Header banner
	headerText := fmt.Sprintf(" CREDENTIALCTL  •  INSPECTING: %s ", m.Item.Filename)
	sb.WriteString(theme.StyleHeader.Render(headerText) + "\n")

	// Tabs Bar
	tabs := []string{"1. Overview", "2. Assertions", "3. Signature", "4. Validation", "5. Raw JSON"}
	var renderedTabs []string
	for i, t := range tabs {
		if TabIndex(i) == m.ActiveTab {
			renderedTabs = append(renderedTabs, theme.StyleActiveTab.Render(t))
		} else {
			renderedTabs = append(renderedTabs, theme.StyleInactiveTab.Render(t))
		}
	}
	sb.WriteString(" " + lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...) + "\n\n")

	// Viewport Content
	sb.WriteString(m.Viewport.View() + "\n\n")

	// Footer / Keymap
	scrollPercent := fmt.Sprintf("%3.f%%", m.Viewport.ScrollPercent()*100)
	helpBar := fmt.Sprintf(" [1-5/Tab] Tabs  •  [↑/↓/PgUp/PgDn] Scroll (%s)  •  [Esc/Backspace] Back to Folder  •  [q] Quit ", scrollPercent)
	sb.WriteString(theme.StyleMuted.Render(helpBar) + "\n")

	return sb.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
