package state

import "github.com/michael-rose/workman/internal/config"

type Worktree struct {
	Name   string
	Branch string
	Path   string
}

type AppState struct {
	Config            *config.Config
	SelectedRepoIndex int
	SelectedWTIndex   int
	ActivePane        Pane // "repos" or "worktrees"
	Worktrees         []Worktree
}

type Pane string

const (
	ReposPane     Pane = "repos"
	WorktreesPane Pane = "worktrees"
)

func New(cfg *config.Config) *AppState {
	return &AppState{
		Config:            cfg,
		SelectedRepoIndex: 0,
		SelectedWTIndex:   0,
		ActivePane:        ReposPane,
		Worktrees:         []Worktree{},
	}
}

func (s *AppState) GetSelectedRepo() *config.Repository {
	if len(s.Config.Repositories) == 0 {
		return nil
	}
	if s.SelectedRepoIndex >= len(s.Config.Repositories) {
		s.SelectedRepoIndex = len(s.Config.Repositories) - 1
	}
	return &s.Config.Repositories[s.SelectedRepoIndex]
}

func (s *AppState) NextRepo() {
	if len(s.Config.Repositories) > 0 {
		s.SelectedRepoIndex = (s.SelectedRepoIndex + 1) % len(s.Config.Repositories)
		s.SelectedWTIndex = 0
	}
}

func (s *AppState) PrevRepo() {
	if len(s.Config.Repositories) > 0 {
		s.SelectedRepoIndex--
		if s.SelectedRepoIndex < 0 {
			s.SelectedRepoIndex = len(s.Config.Repositories) - 1
		}
		s.SelectedWTIndex = 0
	}
}

func (s *AppState) NextWorktree() {
	if len(s.Worktrees) > 0 {
		s.SelectedWTIndex = (s.SelectedWTIndex + 1) % len(s.Worktrees)
	}
}

func (s *AppState) PrevWorktree() {
	if len(s.Worktrees) > 0 {
		s.SelectedWTIndex--
		if s.SelectedWTIndex < 0 {
			s.SelectedWTIndex = len(s.Worktrees) - 1
		}
	}
}

func (s *AppState) TogglePane() {
	if s.ActivePane == ReposPane {
		s.ActivePane = WorktreesPane
	} else {
		s.ActivePane = ReposPane
	}
}
