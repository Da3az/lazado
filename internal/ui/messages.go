package ui

import (
	"github.com/da3az/lazado/internal/api"
)

// Messages are the tea.Msg types used to communicate between
// async operations and the UI.

// PanelID identifies which panel is active.
type PanelID int

const (
	PanelWorkItems PanelID = iota
	PanelPullRequests
	PanelPipelines
	PanelRepos
)

// PanelNames maps panel IDs to display names.
var PanelNames = map[PanelID]string{
	PanelWorkItems:    "Work Items",
	PanelPullRequests: "Pull Requests",
	PanelPipelines:    "Pipelines",
	PanelRepos:        "Repos",
}

// Data loaded messages

type WorkItemsLoadedMsg struct {
	Items []api.WorkItem
	Err   error
}

type PullRequestsLoadedMsg struct {
	PRs []api.PullRequest
	Err error
}

type PipelinesLoadedMsg struct {
	Runs []api.PipelineRun
	Err  error
}

type ReposLoadedMsg struct {
	Repos []api.Repository
	Err   error
}

type StatesLoadedMsg struct {
	Type   string
	States []api.WorkItemTypeState
	Err    error
}

// Action result messages

type WorkItemCreatedMsg struct {
	Item *api.WorkItem
	Err  error
}

type WorkItemUpdatedMsg struct {
	Item *api.WorkItem
	Err  error
}

type PRCreatedMsg struct {
	PR  *api.PullRequest
	Err error
}

type PRApprovedMsg struct {
	Err error
}

type PRCompletedMsg struct {
	Err error
}

type PipelineTriggeredMsg struct {
	Run *api.PipelineRun
	Err error
}

// Metadata loaded messages

type TeamMembersLoadedMsg struct {
	Members []string
	Err     error
}

type AreaPathsLoadedMsg struct {
	Paths []string
	Err   error
}

type IterationsLoadedMsg struct {
	Paths []string
	Err   error
}

// Comments messages

type CommentsLoadedMsg struct {
	WIID     int
	Comments []api.Comment
	Err      error
}

type CommentAddedMsg struct {
	WIID    int
	Comment *api.Comment
	Err     error
}

// StatusMsg displays a temporary message in the status bar.
type StatusMsg struct {
	Text    string
	IsError bool
}

// RefreshMsg triggers a refresh of the specified panel (or all if -1).
type RefreshMsg struct {
	Panel PanelID
}

// ErrorMsg represents an error to display.
type ErrorMsg struct {
	Err error
}
