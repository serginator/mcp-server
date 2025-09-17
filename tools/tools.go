package tools

// NotionTool is the interface for the Notion tools
// It defines the methods that can be used to interact with the Notion API.
type NotionTool interface {
	SearchPagesByTitle(title string) (string, error)
	GetPageByURL(url string) (string, error)
	GetDatabase(databaseID string) (string, error)
	CreatePage(parentID string, title string, content string) (string, error)
	CreateDatabase(parentPageID string, title string) (string, error)
	UpdatePage(pageID string, title string, content string) (string, error)
	UpdateDatabase(databaseID string, title string) (string, error)
}

// JiraTool is the interface for the Jira tools
// It defines the methods that can be used to interact with the Jira API.
type JiraTool interface {
	SearchTickets(query string) (string, error)
	GetTicketByID(ticketID string) (string, error)
	CreateTicket(projectKey string, summary string, description string) (string, error)
}

// GithubTool is the interface for the Github tools
// It defines the methods that can be used to interact with the Github API.
type GithubTool interface {
	GetPullRequest(owner string, repo string, pullRequestNumber int) (string, error)
	GetPullRequestDiff(owner string, repo string, pullRequestNumber int) (string, error)
	CreateIssue(owner string, repo string, title string, body string) (string, error)
	CreatePullRequest(owner string, repo string, title string, body string, head string, base string) (string, error)
	GetComments(owner string, repo string, issueNumber int) (string, error)
	AddComment(owner string, repo string, issueNumber int, body string) (string, error)
	AssignCopilot(owner string, repo string, issueNumber int, assignees []string) (string, error)
	CreateBranch(owner string, repo string, branchName string, sha string) (string, error)
	CreateRepository(name string, description string, private bool) (string, error)
	GetCommit(owner string, repo string, sha string) (string, error)
	GetIssue(owner string, repo string, issueNumber int) (string, error)
	GetReleaseByTag(owner string, repo string, tagName string) (string, error)
	GetTag(owner string, repo string, tagName string) (string, error)
	ListBranches(owner string, repo string) (string, error)
	ListCommits(owner string, repo string) (string, error)
	GetWorkflows(owner string, repo string) (string, error)
	RunWorkflow(owner string, repo string, workflowID string, ref string) (string, error)
	RunFailedJobs(owner string, repo string, runID int64) (string, error)
	CreateCommit(owner string, repo string, message string, tree string, parents []string) (string, error)
	Push(owner string, repo string, ref string, sha string) (string, error)
	SearchCode(query string) (string, error)
	SearchIssues(query string) (string, error)
	SearchPullRequests(query string) (string, error)
	SearchRepositories(query string) (string, error)
}
