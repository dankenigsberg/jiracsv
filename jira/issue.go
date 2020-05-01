package jira

import (
	"errors"
	"net/http"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

// IssueApprovals represents a Jira Issue Approvals
type IssueApprovals struct {
	Development   bool
	Product       bool
	Quality       bool
	Experience    bool
	Documentation bool
	Support       bool
}

// Issue represents a Jira Issue
type Issue struct {
	jira.Issue
	Link         string
	ParentLink   string
	LinkedIssues IssueCollection
	StoryPoints  int
	Approvals    IssueApprovals
	QAContact    string
	Acceptance   string
	Owner        string
	Impediment   bool
	Comments     []*Comment
}

// Comment represents Jira Issue Comment
type Comment struct {
	*jira.Comment
	Created time.Time
	Updated time.Time
}

// IssueCollection is a collection of Jira Issues
type IssueCollection []*Issue

const (
	// DeliveryOwnerRegExp is the Regular Expression used to collect the Epic Delivery Owner
	DeliveryOwnerRegExp = `\W*(Delivery Owner|DELIVERY OWNER)\W*:\W*\[~([a-zA-Z0-9]*)\]`
)

var (
	// ErrorAuthentication is returned when the authentication failed
	ErrorAuthentication = errors.New("Access Unauthorized: check basic authentication")
)

// IssueType represent an Issue Type
type IssueType string

const (
	// IssueTypeInitiative represents the Issue Type Initiative
	IssueTypeInitiative IssueType = "Initiative"

	// IssueTypeEpic represents the Issue Type Epic
	IssueTypeEpic IssueType = "Epic"

	// IssueTypeStory represents the Issue Type Story
	IssueTypeStory IssueType = "Story"

	// IssueTypeTask represents the Issue Type Task
	IssueTypeTask IssueType = "Task"
)

// IssueStatus represent an Issue Status
type IssueStatus string

const (
	// IssueStatusDone represents the Issue Status Done
	IssueStatusDone IssueStatus = "Done"

	// IssueStatusObsolete represents the Issue Status Done
	IssueStatusObsolete IssueStatus = "Obsolete"

	// IssueStatusInProgress represents the Issue Status Done
	IssueStatusInProgress IssueStatus = "In Progress"

	// IssueStatusFeatureComplete represents the Issue Status Done
	IssueStatusFeatureComplete IssueStatus = "Feature Complete"

	// IssueStatusCodeReview represents the Issue Status Done
	IssueStatusCodeReview IssueStatus = "Code Review"

	// IssueStatusQEReview represents the Issue Status Done
	IssueStatusQEReview IssueStatus = "QE Review"
)

// IssueResolution represent an Issue Resolution
type IssueResolution string

const (
	// IssueResolutionDone represents the Issue Resolution Done
	IssueResolutionDone IssueResolution = "Done"
)

// IssuePriority represent an Issue Resolution
type IssuePriority string

const (
	// IssuePriorityUnprioritized represents the Issue Priority Unprioritized
	IssuePriorityUnprioritized IssuePriority = "Unprioritized"
)

func jiraReturnError(ret *jira.Response, err error) error {
	if err == nil {
		return nil
	}

	if ret.Response.StatusCode == http.StatusForbidden || ret.Response.StatusCode == http.StatusUnauthorized {
		return ErrorAuthentication
	}

	return err
}

// NewIssueCollection creates and returns a new Jira Issue Collection
func NewIssueCollection(size int) IssueCollection {
	return make([]*Issue, size)
}

// Approved returns true if all approvals are true
func (a *IssueApprovals) Approved() bool {
	return a.Development == true && a.Product == true && a.Quality == true && a.Experience == true && a.Documentation == true
}

// IsActive returns true if the issue is currently worked on
func (i *Issue) IsActive() bool {
	switch IssueStatus(i.Fields.Status.Name) {
	case IssueStatusInProgress:
		return true
	case IssueStatusFeatureComplete:
		return true
	case IssueStatusCodeReview:
		return true
	case IssueStatusQEReview:
		return true
	}

	return false
}

// IsType returns true if the issue is of the relevant type
func (i *Issue) IsType(tp IssueType) bool {
	if IssueType(i.Issue.Fields.Type.Name) == tp {
		return true
	}

	return false
}

// InStatus returns true if the issue is in the relevant status
func (i *Issue) InStatus(status IssueStatus) bool {
	return i.Fields.Status != nil && IssueStatus(i.Fields.Status.Name) == status
}

// IsResolved returns true if the issue Resolution is Done
func (i *Issue) IsResolved() bool {
	if i.Fields.Resolution != nil && IssueResolution(i.Fields.Resolution.Name) == IssueResolutionDone {
		return true
	}

	return false
}

// IsPrioritized returns true if the issue Priority has been set
func (i *Issue) IsPrioritized() bool {
	if i.Fields.Priority != nil {
		switch IssuePriority(i.Fields.Priority.Name) {
		case IssuePriorityUnprioritized:
			return false
		case "":
			return false
		}
	}

	return true
}

// HasStoryPoints returns true if the issue has Story Points defined
func (i *Issue) HasStoryPoints() bool {
	if i.StoryPoints > NoStoryPoints {
		return true
	}

	return false
}

// HasComponent returns true if the issue has the relevant component
func (i *Issue) HasComponent(component string) bool {
	if i.Fields.Components == nil {
		return false
	}

	for _, c := range i.Fields.Components {
		if c.Name == component {
			return true
		}
	}

	return false
}

// FilterByFunction returns jira issues from collection that satisfy the provided function
func (c IssueCollection) FilterByFunction(fn func(*Issue) bool) IssueCollection {
	r := NewIssueCollection(0)

	for _, i := range c {
		if fn(i) {
			r = append(r, i)
		}
	}

	return r
}

// Len returns the number of issues in the collection
func (c IssueCollection) Len() int {
	return len(c)
}

// StoryPoints returns the total number of story points for the issues in the collection
func (c IssueCollection) StoryPoints() int {
	points := 0

	for _, i := range c {
		if i.HasStoryPoints() {
			points += i.StoryPoints
		}
	}

	return points
}
