package jira

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mcp-server/tools"
	"net/http"
	"time"
)

// JiraClient is a client for the Jira API
// It implements the tools.JiraTool interface
type JiraClient struct {
	baseURL    string
	username   string
	token      string
	httpClient *http.Client
}

// JiraIssue represents a Jira issue response
type JiraIssue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description struct {
			Type    string `json:"type"`
			Version int    `json:"version"`
			Content []struct {
				Type    string `json:"type"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content,omitempty"`
			} `json:"content"`
		} `json:"description"`
		Status struct {
			Name string `json:"name"`
		} `json:"status"`
		Assignee *JiraUser `json:"assignee"`
	} `json:"fields"`
}

// JiraUser represents a Jira user
type JiraUser struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// JiraSearchResponse represents the response from Jira search API
type JiraSearchResponse struct {
	Issues []JiraIssue `json:"issues"`
	Total  int         `json:"total"`
}

// JiraCreateIssueRequest represents a request to create a Jira issue
type JiraCreateIssueRequest struct {
	Fields struct {
		Project struct {
			Key string `json:"key"`
		} `json:"project"`
		Summary     string `json:"summary"`
		Description struct {
			Type    string `json:"type"`
			Version int    `json:"version"`
			Content []struct {
				Type    string `json:"type"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"content"`
		} `json:"description"`
		IssueType struct {
			Name string `json:"name"`
		} `json:"issuetype"`
	} `json:"fields"`
}

// NewJiraClient creates a new JiraClient
// It takes a jira url, username and token as arguments and returns a new JiraClient
// The token is used to authenticate with the Jira API
func NewJiraClient(jiraURL, username, token string) (*JiraClient, error) {
	if username == "" {
		return nil, fmt.Errorf("username/email is required for Jira authentication")
	}
	if token == "" {
		return nil, fmt.Errorf("API token is required for Jira authentication")
	}

	// Ensure the URL ends with a slash for proper API endpoint construction
	if jiraURL[len(jiraURL)-1] != '/' {
		jiraURL += "/"
	}

	return &JiraClient{
		baseURL:  jiraURL,
		username: username,
		token:    token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// makeRequest makes an authenticated HTTP request to the Jira API
func (c *JiraClient) makeRequest(method, endpoint string, body []byte) (*http.Response, error) {
	url := c.baseURL + "rest/api/3/" + endpoint

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set up Basic Authentication
	auth := base64.StdEncoding.EncodeToString([]byte(c.username + ":" + c.token))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Accept", "application/json")
	if method == "POST" || method == "PUT" {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// GetTicketByID gets a ticket by its ID
// It takes a ticketID as an argument
// It returns a string representation of the ticket and an error if any
func (c *JiraClient) GetTicketByID(ticketID string) (string, error) {
	if ticketID == "" {
		return "", fmt.Errorf("ticket ID cannot be empty")
	}

	response, err := c.makeRequest("GET", "issue/"+ticketID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to make request for ticket %s: %w", ticketID, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("failed to get ticket %s (HTTP %d): %s", ticketID, response.StatusCode, string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var issue JiraIssue
	if err := json.Unmarshal(body, &issue); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract description text from the content structure
	description := extractDescriptionText(issue.Fields.Description)

	return fmt.Sprintf("ID: %s\nSummary: %s\nStatus: %s\nAssignee: %s\nDescription: %s\n",
		issue.Key,
		issue.Fields.Summary,
		issue.Fields.Status.Name,
		getAssigneeName(issue.Fields.Assignee),
		description), nil
}

// SearchTickets searches for tickets using JQL
func (c *JiraClient) SearchTickets(jql string) (string, error) {
	if jql == "" {
		return "", fmt.Errorf("JQL query cannot be empty")
	}

	searchRequest := map[string]interface{}{
		"jql":        jql,
		"maxResults": 50,
		"fields":     []string{"summary", "status", "assignee"},
	}

	requestBody, err := json.Marshal(searchRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal search request: %w", err)
	}

	response, err := c.makeRequest("POST", "search", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to make search request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("failed to search tickets with JQL '%s' (HTTP %d): %s", jql, response.StatusCode, string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var searchResponse JiraSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return "", fmt.Errorf("failed to parse search response: %w", err)
	}

	if len(searchResponse.Issues) == 0 {
		return "No tickets found matching the query.", nil
	}

	var result string
	for _, issue := range searchResponse.Issues {
		result += fmt.Sprintf("Key: %s\nSummary: %s\nStatus: %s\nAssignee: %s\n\n",
			issue.Key,
			issue.Fields.Summary,
			issue.Fields.Status.Name,
			getAssigneeName(issue.Fields.Assignee))
	}
	return result, nil
}

// CreateTicket creates a new ticket
func (c *JiraClient) CreateTicket(projectKey string, summary string, description string) (string, error) {
	if projectKey == "" {
		return "", fmt.Errorf("project key cannot be empty")
	}
	if summary == "" {
		return "", fmt.Errorf("summary cannot be empty")
	}

	createRequest := JiraCreateIssueRequest{}
	createRequest.Fields.Project.Key = projectKey
	createRequest.Fields.Summary = summary
	createRequest.Fields.IssueType.Name = "Task"

	// Set up description in the Atlassian Document Format
	createRequest.Fields.Description.Type = "doc"
	createRequest.Fields.Description.Version = 1
	if description != "" {
		createRequest.Fields.Description.Content = []struct {
			Type    string `json:"type"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		}{
			{
				Type: "paragraph",
				Content: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{
						Type: "text",
						Text: description,
					},
				},
			},
		}
	}

	requestBody, err := json.Marshal(createRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal create request: %w", err)
	}

	response, err := c.makeRequest("POST", "issue", requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to make create request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		return "", fmt.Errorf("failed to create ticket (HTTP %d): %s", response.StatusCode, string(body))
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var createdIssue struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(body, &createdIssue); err != nil {
		return "", fmt.Errorf("failed to parse create response: %w", err)
	}

	return fmt.Sprintf("Created ticket: %s - %s", createdIssue.Key, summary), nil
}

// Helper function to safely get assignee name
func getAssigneeName(assignee *JiraUser) string {
	if assignee == nil {
		return "Unassigned"
	}
	if assignee.DisplayName != "" {
		return assignee.DisplayName
	}
	return assignee.EmailAddress
}

// extractDescriptionText extracts plain text from Jira's Atlassian Document Format
func extractDescriptionText(description struct {
	Type    string `json:"type"`
	Version int    `json:"version"`
	Content []struct {
		Type    string `json:"type"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content,omitempty"`
	} `json:"content"`
}) string {
	var text string
	for _, content := range description.Content {
		if content.Type == "paragraph" {
			for _, item := range content.Content {
				if item.Type == "text" {
					text += item.Text + " "
				}
			}
			text += "\n"
		}
	}
	return text
}

var _ tools.JiraTool = &JiraClient{}
