package api

import "time"

// WorkItem represents an Azure DevOps work item.
type WorkItem struct {
	ID     int                    `json:"id"`
	Rev    int                    `json:"rev"`
	Fields map[string]interface{} `json:"fields"`
	URL    string                 `json:"url"`
}

// Title returns the work item title.
func (w *WorkItem) Title() string {
	return stringField(w.Fields, "System.Title")
}

// State returns the work item state.
func (w *WorkItem) State() string {
	return stringField(w.Fields, "System.State")
}

// WorkItemType returns the type (Task, User Story, Bug, etc.).
func (w *WorkItem) WorkItemType() string {
	return stringField(w.Fields, "System.WorkItemType")
}

// AssignedTo returns the display name of the assigned user.
func (w *WorkItem) AssignedTo() string {
	v, ok := w.Fields["System.AssignedTo"]
	if !ok || v == nil {
		return ""
	}
	if m, ok := v.(map[string]interface{}); ok {
		if name, ok := m["displayName"].(string); ok {
			return name
		}
	}
	return ""
}

// Description returns the work item description (may contain HTML).
func (w *WorkItem) Description() string {
	return stringField(w.Fields, "System.Description")
}

// AreaPath returns the area path.
func (w *WorkItem) AreaPath() string {
	return stringField(w.Fields, "System.AreaPath")
}

// IterationPath returns the iteration path.
func (w *WorkItem) IterationPath() string {
	return stringField(w.Fields, "System.IterationPath")
}

// Tags returns the tags string.
func (w *WorkItem) Tags() string {
	return stringField(w.Fields, "System.Tags")
}

func stringField(fields map[string]interface{}, key string) string {
	v, ok := fields[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// PullRequest represents an Azure DevOps pull request.
type PullRequest struct {
	PullRequestID int         `json:"pullRequestId"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	Status        string      `json:"status"`
	CreatedBy     IdentityRef `json:"createdBy"`
	SourceRefName string      `json:"sourceRefName"`
	TargetRefName string      `json:"targetRefName"`
	MergeStatus   string      `json:"mergeStatus"`
	Reviewers     []Reviewer  `json:"reviewers"`
	Repository    RepoRef     `json:"repository"`
	CreationDate  time.Time   `json:"creationDate"`
	IsDraft       bool        `json:"isDraft"`
}

// SourceBranch returns the source branch name without refs/heads/ prefix.
func (pr *PullRequest) SourceBranch() string {
	return trimRefPrefix(pr.SourceRefName)
}

// TargetBranch returns the target branch name without refs/heads/ prefix.
func (pr *PullRequest) TargetBranch() string {
	return trimRefPrefix(pr.TargetRefName)
}

func trimRefPrefix(ref string) string {
	const prefix = "refs/heads/"
	if len(ref) > len(prefix) {
		return ref[len(prefix):]
	}
	return ref
}

// IdentityRef represents a user identity.
type IdentityRef struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	ID          string `json:"id"`
}

// Reviewer represents a PR reviewer with their vote.
type Reviewer struct {
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
	ID          string `json:"id"`
	Vote        int    `json:"vote"` // 10=approved, 5=approved with suggestions, -5=waiting, -10=rejected, 0=no vote
}

// VoteLabel returns a human-readable label for the reviewer's vote.
func (r *Reviewer) VoteLabel() string {
	switch r.Vote {
	case 10:
		return "Approved"
	case 5:
		return "Approved with suggestions"
	case -5:
		return "Waiting for author"
	case -10:
		return "Rejected"
	default:
		return "No vote"
	}
}

// RepoRef is a reference to a repository within a PR.
type RepoRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PipelineRun represents a pipeline execution.
type PipelineRun struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	State        string    `json:"state"`  // completed, inProgress, canceling
	Result       string    `json:"result"` // succeeded, failed, canceled
	Pipeline     Pipeline  `json:"pipeline"`
	CreatedDate  time.Time `json:"createdDate"`
	FinishedDate time.Time `json:"finishedDate"`
}

// Pipeline represents a pipeline definition.
type Pipeline struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Repository represents a git repository.
type Repository struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	DefaultBranch string `json:"defaultBranch"`
	RemoteURL     string `json:"remoteUrl"`
	SSHURL        string `json:"sshUrl"`
	Size          int64  `json:"size"`
	WebURL        string `json:"webUrl"`
}

// WorkItemTypeState represents a valid state for a work item type.
type WorkItemTypeState struct {
	Name     string `json:"name"`
	Category string `json:"category"` // Proposed, InProgress, Completed, Resolved, Removed
	Color    string `json:"color"`
}

// PatchOperation is a JSON Patch operation for work item updates.
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// WIQLResult holds the response from a WIQL query.
type WIQLResult struct {
	WorkItems []WIQLWorkItemRef `json:"workItems"`
}

// WIQLWorkItemRef is a reference to a work item returned by WIQL.
type WIQLWorkItemRef struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// Profile represents the authenticated user's profile.
type Profile struct {
	DisplayName  string `json:"displayName"`
	PublicAlias  string `json:"publicAlias"`
	EmailAddress string `json:"emailAddress"`
	ID           string `json:"id"`
}

// ConnectionData holds response from the _apis/connectionData endpoint.
type ConnectionData struct {
	AuthenticatedUser IdentityRef `json:"authenticatedUser"`
}

// CreatePRRequest is the payload for creating a pull request.
type CreatePRRequest struct {
	SourceRefName string           `json:"sourceRefName"`
	TargetRefName string           `json:"targetRefName"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	WorkItemRefs  []WorkItemRefReq `json:"workItemRefs,omitempty"`
	IsDraft       bool             `json:"isDraft,omitempty"`
}

// WorkItemRefReq references a work item when creating a PR.
type WorkItemRefReq struct {
	ID  string `json:"id"`
	URL string `json:"url,omitempty"`
}
