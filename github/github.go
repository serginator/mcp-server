package github

import (
	"context"
	"fmt"
	"mcp-server/tools"

	"github.com/google/go-github/v63/github"
)

// GithubClient is a client for the Github API
// It implements the tools.GithubTool interface
type GithubClient struct {
	client *github.Client
}

// NewGithubClient creates a new GithubClient
// It takes a token as an argument and returns a new GithubClient
// The token is used to authenticate with the Github API
func NewGithubClient(token string) *GithubClient {
	client := github.NewClient(nil).WithAuthToken(token)
	return &GithubClient{client: client}
}

// GetPullRequest gets a pull request from a repository
// It takes the owner, repo, and pull request number as arguments
// It returns a string representation of the pull request and an error if any
func (c *GithubClient) GetPullRequest(owner string, repo string, pullRequestNumber int) (string, error) {
	pr, _, err := c.client.PullRequests.Get(context.Background(), owner, repo, pullRequestNumber)
	if err != nil {
		return "", err
	}
	return pr.String(), nil
}

// GetPullRequestDiff gets the diff of a pull request from a repository
// It takes the owner, repo, and pull request number as arguments
// It returns the diff as a string and an error if any
func (c *GithubClient) GetPullRequestDiff(owner string, repo string, pullRequestNumber int) (string, error) {
	// GitHub API supports getting PR diff in different formats
	// We'll use the unified diff format which is most readable for analysis
	diff, _, err := c.client.PullRequests.GetRaw(context.Background(), owner, repo, pullRequestNumber, github.RawOptions{
		Type: github.Diff,
	})
	if err != nil {
		return "", err
	}
	return string(diff), nil
}

// CreateIssue creates an issue in a repository
func (c *GithubClient) CreateIssue(owner string, repo string, title string, body string) (string, error) {
	issueRequest := &github.IssueRequest{
		Title: &title,
		Body:  &body,
	}
	issue, _, err := c.client.Issues.Create(context.Background(), owner, repo, issueRequest)
	if err != nil {
		return "", err
	}
	return issue.String(), nil
}

// CreatePullRequest creates a pull request in a repository
func (c *GithubClient) CreatePullRequest(owner string, repo string, title string, body string, head string, base string) (string, error) {
	newPR := &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Head:  &head,
		Base:  &base,
	}
	pr, _, err := c.client.PullRequests.Create(context.Background(), owner, repo, newPR)
	if err != nil {
		return "", err
	}
	return pr.String(), nil
}

// GetComments gets the comments from an issue
func (c *GithubClient) GetComments(owner string, repo string, issueNumber int) (string, error) {
	comments, _, err := c.client.Issues.ListComments(context.Background(), owner, repo, issueNumber, nil)
	if err != nil {
		return "", err
	}
	var result string
	for _, comment := range comments {
		result += comment.String() + "\n"
	}
	return result, nil
}

// AddComment adds a comment to an issue
func (c *GithubClient) AddComment(owner string, repo string, issueNumber int, body string) (string, error) {
	comment := &github.IssueComment{
		Body: &body,
	}
	newComment, _, err := c.client.Issues.CreateComment(context.Background(), owner, repo, issueNumber, comment)
	if err != nil {
		return "", err
	}
	return newComment.String(), nil
}

// AssignCopilot assigns copilot to an issue or pull request
func (c *GithubClient) AssignCopilot(owner string, repo string, issueNumber int, assignees []string) (string, error) {
	issue, _, err := c.client.Issues.AddAssignees(context.Background(), owner, repo, issueNumber, assignees)
	if err != nil {
		return "", err
	}
	return issue.String(), nil
}

// CreateBranch creates a branch in a repository
func (c *GithubClient) CreateBranch(owner string, repo string, branchName string, sha string) (string, error) {
	ref := &github.Reference{
		Ref: github.String("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: &sha,
		},
	}
	newRef, _, err := c.client.Git.CreateRef(context.Background(), owner, repo, ref)
	if err != nil {
		return "", err
	}
	return newRef.String(), nil
}

// CreateRepository creates a new repository
func (c *GithubClient) CreateRepository(name string, description string, private bool) (string, error) {
	repo := &github.Repository{
		Name:        &name,
		Description: &description,
		Private:     &private,
	}
	newRepo, _, err := c.client.Repositories.Create(context.Background(), "", repo)
	if err != nil {
		return "", err
	}
	return newRepo.String(), nil
}

// GetCommit gets a commit from a repository
func (c *GithubClient) GetCommit(owner string, repo string, sha string) (string, error) {
	commit, _, err := c.client.Git.GetCommit(context.Background(), owner, repo, sha)
	if err != nil {
		return "", err
	}
	return commit.String(), nil
}

// GetIssue gets an issue from a repository
func (c *GithubClient) GetIssue(owner string, repo string, issueNumber int) (string, error) {
	issue, _, err := c.client.Issues.Get(context.Background(), owner, repo, issueNumber)
	if err != nil {
		return "", err
	}
	return issue.String(), nil
}

// GetReleaseByTag gets a release by tag from a repository
func (c *GithubClient) GetReleaseByTag(owner string, repo string, tagName string) (string, error) {
	release, _, err := c.client.Repositories.GetReleaseByTag(context.Background(), owner, repo, tagName)
	if err != nil {
		return "", err
	}
	return release.String(), nil
}

// GetTag gets a tag from a repository
func (c *GithubClient) GetTag(owner string, repo string, tagName string) (string, error) {
	// There is no direct way to get a tag by name.
	// We need to list all tags and find the one with the matching name.
	tags, _, err := c.client.Repositories.ListTags(context.Background(), owner, repo, nil)
	if err != nil {
		return "", err
	}
	for _, tag := range tags {
		if *tag.Name == tagName {
			return fmt.Sprintf("Tag: %s\nCommit: %s", *tag.Name, *tag.Commit.SHA), nil
		}
	}
	return "Tag not found", nil
}

// ListBranches lists the branches of a repository
func (c *GithubClient) ListBranches(owner string, repo string) (string, error) {
	branches, _, err := c.client.Repositories.ListBranches(context.Background(), owner, repo, nil)
	if err != nil {
		return "", err
	}
	var result string
	for _, branch := range branches {
		result += fmt.Sprintf("Branch: %s\nSHA: %s\n\n", *branch.Name, *branch.Commit.SHA)
	}
	return result, nil
}

// ListCommits lists the commits of a repository
func (c *GithubClient) ListCommits(owner string, repo string) (string, error) {
	commits, _, err := c.client.Repositories.ListCommits(context.Background(), owner, repo, nil)
	if err != nil {
		return "", err
	}
	var result string
	for _, commit := range commits {
		result += commit.String() + "\n"
	}
	return result, nil
}

// GetWorkflows gets the workflows of a repository
func (c *GithubClient) GetWorkflows(owner string, repo string) (string, error) {
	workflows, _, err := c.client.Actions.ListWorkflows(context.Background(), owner, repo, nil)
	if err != nil {
		return "", err
	}
	var result string
	for _, workflow := range workflows.Workflows {
		result += fmt.Sprintf("Workflow: %s\nID: %d\nState: %s\n\n",
			*workflow.Name, *workflow.ID, *workflow.State)
	}
	return result, nil
}

// RunWorkflow runs a workflow in a repository
func (c *GithubClient) RunWorkflow(owner string, repo string, workflowID string, ref string) (string, error) {
	opts := github.CreateWorkflowDispatchEventRequest{
		Ref: ref,
	}
	_, err := c.client.Actions.CreateWorkflowDispatchEventByFileName(context.Background(), owner, repo, workflowID, opts)
	if err != nil {
		return "", err
	}
	return "Workflow run successfully", nil
}

// RunFailedJobs runs the failed jobs of a workflow
func (c *GithubClient) RunFailedJobs(owner string, repo string, runID int64) (string, error) {
	// TODO: Implement this method
	return "", nil
}

// CreateCommit creates a commit in a repository
func (c *GithubClient) CreateCommit(owner string, repo string, message string, tree string, parents []string) (string, error) {
	// TODO: Implement this method
	return "", nil
}

// Push pushes to a repository
func (c *GithubClient) Push(owner string, repo string, ref string, sha string) (string, error) {
	// TODO: Implement this method
	return "", nil
}

// SearchCode searches for code in a repository
func (c *GithubClient) SearchCode(query string) (string, error) {
	opts := &github.SearchOptions{
		Sort:  "indexed",
		Order: "desc",
	}
	result, _, err := c.client.Search.Code(context.Background(), query, opts)
	if err != nil {
		return "", err
	}

	var output string
	for _, codeResult := range result.CodeResults {
		output += fmt.Sprintf("File: %s\nRepo: %s\nURL: %s\n\n",
			*codeResult.Name, *codeResult.Repository.FullName, *codeResult.HTMLURL)
	}
	return output, nil
}

// SearchIssues searches for issues in a repository
func (c *GithubClient) SearchIssues(query string) (string, error) {
	opts := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
	}
	result, _, err := c.client.Search.Issues(context.Background(), query, opts)
	if err != nil {
		return "", err
	}

	var output string
	for _, issue := range result.Issues {
		output += fmt.Sprintf("Title: %s\nNumber: %d\nState: %s\nURL: %s\n\n",
			*issue.Title, *issue.Number, *issue.State, *issue.HTMLURL)
	}
	return output, nil
}

// SearchPullRequests searches for pull requests in a repository
func (c *GithubClient) SearchPullRequests(query string) (string, error) {
	// GitHub API treats pull requests as issues, so we'll search for issues with is:pr
	fullQuery := query + " is:pr"
	opts := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
	}
	result, _, err := c.client.Search.Issues(context.Background(), fullQuery, opts)
	if err != nil {
		return "", err
	}

	var output string
	for _, pr := range result.Issues {
		output += fmt.Sprintf("Title: %s\nNumber: %d\nState: %s\nURL: %s\n\n",
			*pr.Title, *pr.Number, *pr.State, *pr.HTMLURL)
	}
	return output, nil
}

// SearchRepositories searches for repositories
func (c *GithubClient) SearchRepositories(query string) (string, error) {
	opts := &github.SearchOptions{
		Sort:  "stars",
		Order: "desc",
	}
	result, _, err := c.client.Search.Repositories(context.Background(), query, opts)
	if err != nil {
		return "", err
	}

	var output string
	for _, repo := range result.Repositories {
		output += fmt.Sprintf("Name: %s\nDescription: %s\nStars: %d\nURL: %s\n\n",
			*repo.FullName, repo.GetDescription(), *repo.StargazersCount, *repo.HTMLURL)
	}
	return output, nil
}

var _ tools.GithubTool = &GithubClient{}
