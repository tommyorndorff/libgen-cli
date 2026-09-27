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

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ciehanski/libgen-cli/libgen"
)

type state int

const (
	stateInput state = iota
	stateSearching
	stateResults
	stateDownloading
	stateDone
	stateError
)

const numResults = 25

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)
	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
)

// model is the root Bubble Tea model for the libgen TUI. It's a simple
// state machine: input -> searching -> results -> downloading -> done/error,
// with "n" from results/done/error looping back to input for a new search.
type model struct {
	state      state
	outputPath string

	query    textinput.Model
	spinner  spinner.Model
	progress progress.Model
	table    table.Model

	books []*libgen.Book
	err   error

	selectedBook *libgen.Book
	progressCh   chan progressMsg

	width  int
	height int
}

func newModel(outputPath string) model {
	ti := textinput.New()
	ti.Placeholder = "search Library Genesis..."
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 60

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	pr := progress.New(progress.WithDefaultGradient())

	tbl := table.New(
		table.WithColumns(tableColumns(40)),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	return model{
		state:      stateInput,
		outputPath: outputPath,
		query:      ti,
		spinner:    sp,
		progress:   pr,
		table:      tbl,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// searchResultMsg carries the outcome of a background libgen.Search call.
type searchResultMsg struct {
	mirror string
	books  []*libgen.Book
	err    error
}

// progressMsg carries download progress from the background download
// goroutine into the Bubble Tea event loop via progressCh.
type progressMsg struct {
	read, total int64
	done        bool
	err         error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.resizeTable()
		if w := m.width - 10; w > 0 {
			m.progress.Width = min(w, 60)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.state == stateInput {
				return m, tea.Quit
			}
		case "esc":
			if m.state != stateInput && m.state != stateDownloading {
				return m.resetToInput(), nil
			}
		}
		return m.handleKey(msg)

	case spinner.TickMsg:
		if m.state == stateSearching {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case searchResultMsg:
		m.state = stateResults
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
			return m, nil
		}
		m.books = msg.books
		if len(m.books) == 0 {
			m.state = stateError
			m.err = fmt.Errorf("no results found from %s", msg.mirror)
			return m, nil
		}
		m.table.SetRows(booksToRows(m.books))
		m.table.SetCursor(0)
		return m, nil

	case progressMsg:
		if msg.err != nil {
			m.state = stateError
			m.err = msg.err
			return m, nil
		}
		if msg.done {
			m.state = stateDone
			return m, nil
		}
		var pct float64
		if msg.total > 0 {
			pct = float64(msg.read) / float64(msg.total)
		}
		cmd := m.progress.SetPercent(pct)
		return m, tea.Batch(cmd, m.listenForProgress())

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		m.progress = newModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateInput:
		if msg.String() == "enter" {
			query := strings.TrimSpace(m.query.Value())
			if query == "" {
				return m, nil
			}
			m.state = stateSearching
			m.err = nil
			return m, tea.Batch(m.spinner.Tick, doSearch(query))
		}
		var cmd tea.Cmd
		m.query, cmd = m.query.Update(msg)
		return m, cmd

	case stateResults:
		switch msg.String() {
		case "enter":
			cursor := m.table.Cursor()
			if cursor < 0 || cursor >= len(m.books) {
				return m, nil
			}
			m.selectedBook = m.books[cursor]
			m.state = stateDownloading
			m.progressCh = make(chan progressMsg, 8)
			_ = m.progress.SetPercent(0)
			return m, tea.Batch(startDownload(m.selectedBook, m.outputPath, m.progressCh), m.listenForProgress())
		case "n":
			return m.resetToInput(), nil
		}
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd

	case stateDone, stateError:
		if msg.String() == "n" {
			return m.resetToInput(), nil
		}
	}

	return m, nil
}

func (m model) resetToInput() model {
	m.state = stateInput
	m.query.SetValue("")
	m.query.Focus()
	m.err = nil
	m.books = nil
	m.selectedBook = nil
	return m
}

// listenForProgress returns a tea.Cmd that blocks on the progress channel
// for a single message. Update re-issues it after every non-terminal
// progressMsg so the model keeps draining the channel.
func (m model) listenForProgress() tea.Cmd {
	ch := m.progressCh
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return progressMsg{done: true}
		}
		return msg
	}
}

func doSearch(query string) tea.Cmd {
	return func() tea.Msg {
		mirror, err := libgen.GetWorkingSearchMirror()
		if err != nil {
			return searchResultMsg{err: fmt.Errorf("error finding a working mirror: %w", err)}
		}
		books, err := libgen.Search(&libgen.SearchOptions{
			Query:        query,
			SearchMirror: mirror,
			Results:      numResults,
			Print:        false,
		})
		if err != nil {
			return searchResultMsg{err: fmt.Errorf("error completing search query: %w", err)}
		}
		return searchResultMsg{mirror: mirror.String(), books: books}
	}
}

func startDownload(book *libgen.Book, outputPath string, ch chan progressMsg) tea.Cmd {
	return func() tea.Msg {
		go func() {
			defer close(ch)
			if err := libgen.GetDownloadURL(book, false); err != nil {
				ch <- progressMsg{err: err}
				return
			}
			err := libgen.DownloadBookProgress(book, outputPath, func(read, total int64) {
				ch <- progressMsg{read: read, total: total}
			})
			if err != nil {
				ch <- progressMsg{err: err}
				return
			}
			ch <- progressMsg{done: true}
		}()
		return nil
	}
}

// tableColumns builds the results table's column set, giving the Title
// column all the width not claimed by the other fixed-width columns.
func tableColumns(titleWidth int) []table.Column {
	if titleWidth < 20 {
		titleWidth = 20
	}
	return []table.Column{
		{Title: "ID", Width: 8},
		{Title: "Title", Width: titleWidth},
		{Title: "Author", Width: 20},
		{Title: "Ext", Width: 5},
		{Title: "Size", Width: 10},
	}
}

// resizeTable recomputes the results table's column widths and height to
// fit the current terminal size, called on every tea.WindowSizeMsg.
func (m *model) resizeTable() {
	const fixedColsWidth = 8 + 20 + 5 + 10 // ID + Author + Ext + Size
	const tablePadding = 8                 // inter-column spacing/margins

	tableWidth := m.width - 4
	titleWidth := tableWidth - fixedColsWidth - tablePadding
	m.table.SetColumns(tableColumns(titleWidth))
	m.table.SetWidth(tableWidth)

	const chromeHeight = 8 // title + blank lines + help line + margins
	tableHeight := m.height - chromeHeight
	if tableHeight < 3 {
		tableHeight = 3
	}
	m.table.SetHeight(tableHeight)
}

func booksToRows(books []*libgen.Book) []table.Row {
	rows := make([]table.Row, 0, len(books))
	for _, b := range books {
		author := b.Author
		if author == "" {
			author = "N/A"
		}
		rows = append(rows, table.Row{
			b.ID,
			b.Title,
			author,
			b.Extension,
			libgen.FormatFilesize(b.Filesize),
		})
	}
	return rows
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("libgen-cli"))
	b.WriteString("\n\n")

	switch m.state {
	case stateInput:
		b.WriteString("Search Library Genesis:\n\n")
		b.WriteString(m.query.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter search · q quit"))

	case stateSearching:
		fmt.Fprintf(&b, "%s Searching for %q...\n", m.spinner.View(), m.query.Value())

	case stateResults:
		b.WriteString(m.table.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("↑/↓ navigate · enter download · n new search · esc back · q quit"))

	case stateDownloading:
		title := ""
		if m.selectedBook != nil {
			title = m.selectedBook.Title
		}
		fmt.Fprintf(&b, "Downloading: %s\n\n", title)
		b.WriteString(m.progress.View())

	case stateDone:
		title, author := "", ""
		if m.selectedBook != nil {
			title, author = m.selectedBook.Title, m.selectedBook.Author
		}
		b.WriteString(successStyle.Render(fmt.Sprintf("[OK] %s by %s", title, author)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("n new search · q quit"))

	case stateError:
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("n new search · q quit"))
	}

	b.WriteString("\n")
	return b.String()
}
