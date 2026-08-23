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

package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui/folderview"
	"github.com/ghchinoy/credentialctl/internal/ui/inspectview"
)

type ViewState int

const (
	StateFolder ViewState = iota
	StateInspect
)

// AppModel manages view switching between folder review and file inspection.
type AppModel struct {
	State        ViewState
	FolderModel  folderview.Model
	InspectModel inspectview.Model
	Validator    *engine.ValidatorService
	Width        int
	Height       int
}

// NewAppModel creates the root application model.
func NewAppModel(targetPath string, recursive bool, validator *engine.ValidatorService) AppModel {
	fm := folderview.NewModel(targetPath, recursive, validator)
	return AppModel{
		State:       StateFolder,
		FolderModel: fm,
		Validator:   validator,
		Width:       80,
		Height:      24,
	}
}

// NewInspectAppModel starts directly in file inspector mode for a single file.
func NewInspectAppModel(item engine.FileItem, validator *engine.ValidatorService) AppModel {
	fm := folderview.NewModel(".", false, validator)
	im := inspectview.NewModel(item, 80, 24)
	return AppModel{
		State:        StateInspect,
		FolderModel:  fm,
		InspectModel: im,
		Validator:    validator,
		Width:        80,
		Height:       24,
	}
}

func (a AppModel) Init() tea.Cmd {
	if a.State == StateInspect {
		return a.InspectModel.Init()
	}
	return a.FolderModel.Init()
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width = msg.Width
		a.Height = msg.Height
		var cmd tea.Cmd
		a.FolderModel, cmd = a.FolderModel.Update(msg)
		cmds = append(cmds, cmd)
		if a.State == StateInspect {
			a.InspectModel, cmd = a.InspectModel.Update(msg)
			cmds = append(cmds, cmd)
		}
		return a, tea.Batch(cmds...)

	case folderview.InspectFileMsg:
		a.State = StateInspect
		a.InspectModel = inspectview.NewModel(msg.Item, a.Width, a.Height)
		return a, a.InspectModel.Init()

	case inspectview.BackToFolderMsg:
		a.State = StateFolder
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return a, tea.Quit
		}
	}

	if a.State == StateFolder {
		var cmd tea.Cmd
		a.FolderModel, cmd = a.FolderModel.Update(msg)
		cmds = append(cmds, cmd)
	} else if a.State == StateInspect {
		var cmd tea.Cmd
		a.InspectModel, cmd = a.InspectModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return a, tea.Batch(cmds...)
}

func (a AppModel) View() string {
	if a.State == StateInspect {
		return a.InspectModel.View()
	}
	return a.FolderModel.View()
}
