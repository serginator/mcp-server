package main

import (
	"log"
	"mcp-server/config"
	"mcp-server/github"
	"mcp-server/jira"
	"mcp-server/notion"
	"mcp-server/server"
)

func main() {
	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	githubClient := github.NewGithubClient(cfg.GithubToken)
	jiraClient, err := jira.NewJiraClient(cfg.JiraURL, cfg.JiraUsername, cfg.JiraToken)
	if err != nil {
		log.Fatalf("Error creating Jira client: %v", err)
	}
	notionClient := notion.NewNotionClient(cfg.NotionToken)
	log.Println("Starting MCP server...")
	srv := &server.MCPServer{
		Github: githubClient,
		Jira:   jiraClient,
		Notion: notionClient,
	}
	srv.Start()
}
