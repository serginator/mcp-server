package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mcp-server/tools"
	"os"
	"strings"
)

// MCPServer implements the Model Context Protocol server
type MCPServer struct {
	Github tools.GithubTool
	Jira   tools.JiraTool
	Notion tools.NotionTool
}

// MCPRequest represents an MCP JSON-RPC request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP JSON-RPC response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents content in a tool result
type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Start starts the MCP server
func (s *MCPServer) Start() {
	log.Println("Starting MCP server...")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var request MCPRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			s.sendError(request.ID, -32700, "Parse error", nil)
			continue
		}

		s.handleRequest(request)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Error reading from stdin: %v", err)
	}
}

// handleRequest processes an MCP request
func (s *MCPServer) handleRequest(request MCPRequest) {
	switch request.Method {
	case "initialize":
		s.handleInitialize(request)
	case "tools/list":
		s.handleToolsList(request)
	case "tools/call":
		s.handleToolCall(request)
	default:
		s.sendError(request.ID, -32601, "Method not found", nil)
	}
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(request MCPRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "mcp-integration-server",
			"version": "1.0.0",
		},
	}
	s.sendResponse(request.ID, result)
}

// handleToolsList handles the tools/list request
func (s *MCPServer) handleToolsList(request MCPRequest) {
	tools := s.getAvailableTools()
	result := map[string]interface{}{
		"tools": tools,
	}
	s.sendResponse(request.ID, result)
}

// handleToolCall handles the tools/call request
func (s *MCPServer) handleToolCall(request MCPRequest) {
	params, ok := request.Params.(map[string]interface{})
	if !ok {
		s.sendError(request.ID, -32602, "Invalid params", nil)
		return
	}

	name, ok := params["name"].(string)
	if !ok {
		s.sendError(request.ID, -32602, "Missing tool name", nil)
		return
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	result, err := s.executeTool(name, arguments)
	if err != nil {
		s.sendResponse(request.ID, ToolResult{
			Content: []ToolContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		})
		return
	}

	s.sendResponse(request.ID, ToolResult{
		Content: []ToolContent{{Type: "text", Text: result}},
		IsError: false,
	})
}

// sendResponse sends a JSON-RPC response
func (s *MCPServer) sendResponse(id interface{}, result interface{}) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.sendJSON(response)
}

// sendError sends a JSON-RPC error response
func (s *MCPServer) sendError(id interface{}, code int, message string, data interface{}) {
	response := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.sendJSON(response)
}

// sendJSON sends a JSON message to stdout
func (s *MCPServer) sendJSON(v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	fmt.Println(string(data))
}

// getAvailableTools returns the list of available tools
func (s *MCPServer) getAvailableTools() []Tool {
	return []Tool{
		// GitHub tools
		{
			Name:        "github_get_pull_request",
			Description: "Get details of a specific pull request",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":  map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":   map[string]interface{}{"type": "string", "description": "Repository name"},
					"number": map[string]interface{}{"type": "integer", "description": "Pull request number"},
				},
				"required": []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_get_pull_request_diff",
			Description: "Get the diff of a specific pull request for analysis",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":  map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":   map[string]interface{}{"type": "string", "description": "Repository name"},
					"number": map[string]interface{}{"type": "integer", "description": "Pull request number"},
				},
				"required": []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_create_issue",
			Description: "Create a new issue in a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":  map[string]interface{}{"type": "string", "description": "Repository name"},
					"title": map[string]interface{}{"type": "string", "description": "Issue title"},
					"body":  map[string]interface{}{"type": "string", "description": "Issue body"},
				},
				"required": []string{"owner", "repo", "title"},
			},
		},
		{
			Name:        "github_create_pull_request",
			Description: "Create a new pull request",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":  map[string]interface{}{"type": "string", "description": "Repository name"},
					"title": map[string]interface{}{"type": "string", "description": "Pull request title"},
					"body":  map[string]interface{}{"type": "string", "description": "Pull request body"},
					"head":  map[string]interface{}{"type": "string", "description": "Source branch"},
					"base":  map[string]interface{}{"type": "string", "description": "Target branch"},
				},
				"required": []string{"owner", "repo", "title", "head", "base"},
			},
		},
		{
			Name:        "github_get_issue",
			Description: "Get details of a specific issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":  map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":   map[string]interface{}{"type": "string", "description": "Repository name"},
					"number": map[string]interface{}{"type": "integer", "description": "Issue number"},
				},
				"required": []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_list_branches",
			Description: "List all branches in a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":  map[string]interface{}{"type": "string", "description": "Repository name"},
				},
				"required": []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_list_commits",
			Description: "List commits in a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":  map[string]interface{}{"type": "string", "description": "Repository name"},
				},
				"required": []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_search_repositories",
			Description: "Search for repositories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search query"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "github_search_issues",
			Description: "Search for issues across repositories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search query"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "github_get_workflows",
			Description: "Get workflows for a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":  map[string]interface{}{"type": "string", "description": "Repository name"},
				},
				"required": []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_run_workflow",
			Description: "Trigger a workflow run",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":      map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":       map[string]interface{}{"type": "string", "description": "Repository name"},
					"workflowID": map[string]interface{}{"type": "string", "description": "Workflow ID"},
					"ref":        map[string]interface{}{"type": "string", "description": "Git reference"},
				},
				"required": []string{"owner", "repo", "workflowID", "ref"},
			},
		},
		{
			Name:        "github_add_comment",
			Description: "Add a comment to an issue or pull request",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":  map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":   map[string]interface{}{"type": "string", "description": "Repository name"},
					"number": map[string]interface{}{"type": "integer", "description": "Issue or pull request number"},
					"body":   map[string]interface{}{"type": "string", "description": "Comment body"},
				},
				"required": []string{"owner", "repo", "number", "body"},
			},
		},
		{
			Name:        "github_get_comments",
			Description: "Get comments from an issue or pull request",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":  map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":   map[string]interface{}{"type": "string", "description": "Repository name"},
					"number": map[string]interface{}{"type": "integer", "description": "Issue or pull request number"},
				},
				"required": []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_assign_copilot",
			Description: "Assign users to an issue or pull request",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":     map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":      map[string]interface{}{"type": "string", "description": "Repository name"},
					"number":    map[string]interface{}{"type": "integer", "description": "Issue or pull request number"},
					"assignees": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Array of usernames to assign"},
				},
				"required": []string{"owner", "repo", "number", "assignees"},
			},
		},
		{
			Name:        "github_create_branch",
			Description: "Create a new branch in a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":      map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":       map[string]interface{}{"type": "string", "description": "Repository name"},
					"branchName": map[string]interface{}{"type": "string", "description": "Name for the new branch"},
					"sha":        map[string]interface{}{"type": "string", "description": "SHA of the commit to branch from"},
				},
				"required": []string{"owner", "repo", "branchName", "sha"},
			},
		},
		{
			Name:        "github_create_repository",
			Description: "Create a new repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":        map[string]interface{}{"type": "string", "description": "Repository name"},
					"description": map[string]interface{}{"type": "string", "description": "Repository description"},
					"private":     map[string]interface{}{"type": "boolean", "description": "Whether the repository should be private"},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "github_get_commit",
			Description: "Get details of a specific commit",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":  map[string]interface{}{"type": "string", "description": "Repository name"},
					"sha":   map[string]interface{}{"type": "string", "description": "Commit SHA"},
				},
				"required": []string{"owner", "repo", "sha"},
			},
		},
		{
			Name:        "github_get_release_by_tag",
			Description: "Get release information by tag",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":   map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":    map[string]interface{}{"type": "string", "description": "Repository name"},
					"tagName": map[string]interface{}{"type": "string", "description": "Tag name"},
				},
				"required": []string{"owner", "repo", "tagName"},
			},
		},
		{
			Name:        "github_get_tag",
			Description: "Get tag information",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner":   map[string]interface{}{"type": "string", "description": "Repository owner"},
					"repo":    map[string]interface{}{"type": "string", "description": "Repository name"},
					"tagName": map[string]interface{}{"type": "string", "description": "Tag name"},
				},
				"required": []string{"owner", "repo", "tagName"},
			},
		},
		{
			Name:        "github_search_code",
			Description: "Search for code in repositories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search query"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "github_search_pull_requests",
			Description: "Search for pull requests",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{"type": "string", "description": "Search query"},
				},
				"required": []string{"query"},
			},
		},

		// Jira tools
		{
			Name:        "jira_get_ticket",
			Description: "Get details of a Jira ticket",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ticketID": map[string]interface{}{"type": "string", "description": "Jira ticket ID"},
				},
				"required": []string{"ticketID"},
			},
		},
		{
			Name:        "jira_search_tickets",
			Description: "Search for Jira tickets using JQL",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"jql": map[string]interface{}{"type": "string", "description": "JQL query string"},
				},
				"required": []string{"jql"},
			},
		},
		{
			Name:        "jira_create_ticket",
			Description: "Create a new Jira ticket",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"projectKey":  map[string]interface{}{"type": "string", "description": "Project key"},
					"summary":     map[string]interface{}{"type": "string", "description": "Ticket summary"},
					"description": map[string]interface{}{"type": "string", "description": "Ticket description"},
				},
				"required": []string{"projectKey", "summary"},
			},
		},

		// Notion tools
		{
			Name:        "notion_search_pages",
			Description: "Search for Notion pages by title",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{"type": "string", "description": "Page title to search for"},
				},
				"required": []string{"title"},
			},
		},
		{
			Name:        "notion_get_page",
			Description: "Get a Notion page by URL",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{"type": "string", "description": "Page URL"},
				},
				"required": []string{"url"},
			},
		},
		{
			Name:        "notion_get_database",
			Description: "Get a Notion database by ID",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"databaseID": map[string]interface{}{"type": "string", "description": "Database ID"},
				},
				"required": []string{"databaseID"},
			},
		},
		{
			Name:        "notion_create_page",
			Description: "Create a new Notion page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parentID": map[string]interface{}{"type": "string", "description": "Parent page/database ID"},
					"title":    map[string]interface{}{"type": "string", "description": "Page title"},
					"content":  map[string]interface{}{"type": "string", "description": "Page content"},
				},
				"required": []string{"parentID", "title"},
			},
		},
		{
			Name:        "notion_create_database",
			Description: "Create a new Notion database",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"parentPageID": map[string]interface{}{"type": "string", "description": "Parent page ID"},
					"title":        map[string]interface{}{"type": "string", "description": "Database title"},
				},
				"required": []string{"parentPageID", "title"},
			},
		},
		{
			Name:        "notion_update_page",
			Description: "Update an existing Notion page",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pageID":  map[string]interface{}{"type": "string", "description": "Page ID to update"},
					"title":   map[string]interface{}{"type": "string", "description": "New page title"},
					"content": map[string]interface{}{"type": "string", "description": "New page content"},
				},
				"required": []string{"pageID"},
			},
		},
		{
			Name:        "notion_update_database",
			Description: "Update an existing Notion database",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"databaseID": map[string]interface{}{"type": "string", "description": "Database ID to update"},
					"title":      map[string]interface{}{"type": "string", "description": "New database title"},
				},
				"required": []string{"databaseID", "title"},
			},
		},
	}
}

// executeTool executes the specified tool with given arguments
func (s *MCPServer) executeTool(name string, args map[string]interface{}) (string, error) {
	switch name {
	// GitHub tools
	case "github_get_pull_request":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		number, _ := args["number"].(float64)
		return s.Github.GetPullRequest(owner, repo, int(number))

	case "github_get_pull_request_diff":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		number, _ := args["number"].(float64)
		return s.Github.GetPullRequestDiff(owner, repo, int(number))

	case "github_create_issue":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		title, _ := args["title"].(string)
		body, _ := args["body"].(string)
		return s.Github.CreateIssue(owner, repo, title, body)

	case "github_create_pull_request":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		title, _ := args["title"].(string)
		body, _ := args["body"].(string)
		head, _ := args["head"].(string)
		base, _ := args["base"].(string)
		return s.Github.CreatePullRequest(owner, repo, title, body, head, base)

	case "github_get_issue":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		number, _ := args["number"].(float64)
		return s.Github.GetIssue(owner, repo, int(number))

	case "github_list_branches":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		return s.Github.ListBranches(owner, repo)

	case "github_list_commits":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		return s.Github.ListCommits(owner, repo)

	case "github_search_repositories":
		query, _ := args["query"].(string)
		return s.Github.SearchRepositories(query)

	case "github_search_issues":
		query, _ := args["query"].(string)
		return s.Github.SearchIssues(query)

	case "github_get_workflows":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		return s.Github.GetWorkflows(owner, repo)

	case "github_run_workflow":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		workflowID, _ := args["workflowID"].(string)
		ref, _ := args["ref"].(string)
		return s.Github.RunWorkflow(owner, repo, workflowID, ref)

	case "github_add_comment":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		number, _ := args["number"].(float64)
		body, _ := args["body"].(string)
		return s.Github.AddComment(owner, repo, int(number), body)

	case "github_get_comments":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		number, _ := args["number"].(float64)
		return s.Github.GetComments(owner, repo, int(number))

	case "github_assign_copilot":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		number, _ := args["number"].(float64)
		assignees, _ := args["assignees"].([]interface{})
		assigneeStrs := make([]string, len(assignees))
		for i, assignee := range assignees {
			assigneeStrs[i], _ = assignee.(string)
		}
		return s.Github.AssignCopilot(owner, repo, int(number), assigneeStrs)

	case "github_create_branch":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		branchName, _ := args["branchName"].(string)
		sha, _ := args["sha"].(string)
		return s.Github.CreateBranch(owner, repo, branchName, sha)

	case "github_create_repository":
		name, _ := args["name"].(string)
		description, _ := args["description"].(string)
		private, _ := args["private"].(bool)
		return s.Github.CreateRepository(name, description, private)

	case "github_get_commit":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		sha, _ := args["sha"].(string)
		return s.Github.GetCommit(owner, repo, sha)

	case "github_get_release_by_tag":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		tagName, _ := args["tagName"].(string)
		return s.Github.GetReleaseByTag(owner, repo, tagName)

	case "github_get_tag":
		owner, _ := args["owner"].(string)
		repo, _ := args["repo"].(string)
		tagName, _ := args["tagName"].(string)
		return s.Github.GetTag(owner, repo, tagName)

	case "github_search_code":
		query, _ := args["query"].(string)
		return s.Github.SearchCode(query)

	case "github_search_pull_requests":
		query, _ := args["query"].(string)
		return s.Github.SearchPullRequests(query)

	// Jira tools
	case "jira_get_ticket":
		ticketID, _ := args["ticketID"].(string)
		return s.Jira.GetTicketByID(ticketID)

	case "jira_search_tickets":
		jql, _ := args["jql"].(string)
		return s.Jira.SearchTickets(jql)

	case "jira_create_ticket":
		projectKey, _ := args["projectKey"].(string)
		summary, _ := args["summary"].(string)
		description, _ := args["description"].(string)
		return s.Jira.CreateTicket(projectKey, summary, description)

	// Notion tools
	case "notion_search_pages":
		title, _ := args["title"].(string)
		return s.Notion.SearchPagesByTitle(title)

	case "notion_get_page":
		url, _ := args["url"].(string)
		return s.Notion.GetPageByURL(url)

	case "notion_get_database":
		databaseID, _ := args["databaseID"].(string)
		return s.Notion.GetDatabase(databaseID)

	case "notion_create_page":
		parentID, _ := args["parentID"].(string)
		title, _ := args["title"].(string)
		content, _ := args["content"].(string)
		return s.Notion.CreatePage(parentID, title, content)

	case "notion_create_database":
		parentPageID, _ := args["parentPageID"].(string)
		title, _ := args["title"].(string)
		return s.Notion.CreateDatabase(parentPageID, title)

	case "notion_update_page":
		pageID, _ := args["pageID"].(string)
		title, _ := args["title"].(string)
		content, _ := args["content"].(string)
		return s.Notion.UpdatePage(pageID, title, content)

	case "notion_update_database":
		databaseID, _ := args["databaseID"].(string)
		title, _ := args["title"].(string)
		return s.Notion.UpdateDatabase(databaseID, title)

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
