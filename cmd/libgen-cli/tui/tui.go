// Copyright © 2026 Ryan Ciehanski <ryan@ciehanski.com>
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

// Package tui implements a full-screen, interactive terminal UI for
// libgen-cli built on Bubble Tea. It lets a user search, browse, and
// download Library Genesis resources without leaving a single persistent
// screen, as an alternative to the one-shot Cobra subcommands.
package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the Bubble Tea program and blocks until the user quits.
func Run(outputPath string) error {
	p := tea.NewProgram(newModel(outputPath), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
