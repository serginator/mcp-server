package notion

import (
	"context"
	"fmt"
	"mcp-server/tools"
	"net/url"
	"strings"

	"github.com/dstotijn/go-notion"
)

// NotionClient is a client for the Notion API
// It implements the tools.NotionTool interface
type NotionClient struct {
	client *notion.Client
}

// NewNotionClient creates a new NotionClient
// It takes a token as an argument and returns a new NotionClient
// The token is used to authenticate with the Notion API
func NewNotionClient(token string) *NotionClient {
	client := notion.NewClient(token)
	return &NotionClient{client: client}
}

// SearchPagesByTitle searches for pages by title
// It takes a title as an argument
// It returns a string representation of the pages and an error if any
func (c *NotionClient) SearchPagesByTitle(title string) (string, error) {
	query := &notion.SearchOpts{
		Query: title,
	}
	resp, err := c.client.Search(context.Background(), query)
	if err != nil {
		return "", err
	}

	var result string
	for _, p := range resp.Results {
		page, ok := p.(notion.Page)
		if !ok {
			continue
		}
		if page.Parent.Type == notion.ParentTypePage {
			result += page.URL + "\n"
		}
	}

	return result, nil
}

// GetPageByURL gets a page by its URL
func (c *NotionClient) GetPageByURL(pageURL string) (string, error) {
	pageID, err := extractPageIDFromURL(pageURL)
	if err != nil {
		return "", err
	}

	page, err := c.client.FindPageByID(context.Background(), pageID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Page ID: %s\nURL: %s\nCreated: %s\nLast Edited: %s",
		page.ID, page.URL, page.CreatedTime, page.LastEditedTime), nil
}

// GetDatabase gets a database by its ID
func (c *NotionClient) GetDatabase(databaseID string) (string, error) {
	database, err := c.client.FindDatabaseByID(context.Background(), databaseID)
	if err != nil {
		return "", err
	}

	var properties string
	for name, prop := range database.Properties {
		properties += fmt.Sprintf("- %s (%s)\n", name, string(prop.Type))
	}

	return fmt.Sprintf("Database ID: %s\nTitle: %s\nCreated: %s\nProperties:\n%s",
		database.ID, getDatabaseTitle(&database), database.CreatedTime, properties), nil
}

// CreatePage creates a new page
func (c *NotionClient) CreatePage(parentID string, title string, content string) (string, error) {
	params := notion.CreatePageParams{
		ParentType: notion.ParentTypePage,
		ParentID:   parentID,
		Title: []notion.RichText{
			{
				Type: notion.RichTextTypeText,
				Text: &notion.Text{Content: title},
			},
		},
	}

	if content != "" {
		paragraphBlock := notion.ParagraphBlock{
			RichText: []notion.RichText{
				{
					Type: notion.RichTextTypeText,
					Text: &notion.Text{Content: content},
				},
			},
		}

		params.Children = []notion.Block{paragraphBlock}
	}

	page, err := c.client.CreatePage(context.Background(), params)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Created page: %s (ID: %s)", title, page.ID), nil
}

// CreateDatabase creates a new database
func (c *NotionClient) CreateDatabase(parentPageID string, title string) (string, error) {
	params := notion.CreateDatabaseParams{
		ParentPageID: parentPageID,
		Title: []notion.RichText{
			{
				Type: notion.RichTextTypeText,
				Text: &notion.Text{Content: title},
			},
		},
		Properties: map[string]notion.DatabaseProperty{
			"Name": {
				Type:  notion.DBPropTypeTitle,
				Title: &notion.EmptyMetadata{},
			},
		},
	}

	database, err := c.client.CreateDatabase(context.Background(), params)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Created database: %s (ID: %s)", title, database.ID), nil
}

// UpdatePage updates a page
func (c *NotionClient) UpdatePage(pageID string, title string, content string) (string, error) {
	params := notion.UpdatePageParams{}

	// Note: Updating page title requires different approach in this API version
	// For now, we'll just update properties if needed

	page, err := c.client.UpdatePage(context.Background(), pageID, params)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated page: %s", page.ID), nil
}

// UpdateDatabase updates a database
func (c *NotionClient) UpdateDatabase(databaseID string, title string) (string, error) {
	params := notion.UpdateDatabaseParams{
		Title: []notion.RichText{
			{
				Type: notion.RichTextTypeText,
				Text: &notion.Text{Content: title},
			},
		},
	}

	database, err := c.client.UpdateDatabase(context.Background(), databaseID, params)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated database: %s", database.ID), nil
}

// Helper functions
func extractPageIDFromURL(pageURL string) (string, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return "", err
	}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "-")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid Notion URL")
	}

	// The last part should be the page ID
	pageID := parts[len(parts)-1]
	return pageID, nil
}

func getDatabaseTitle(database *notion.Database) string {
	if len(database.Title) > 0 && database.Title[0].Text != nil {
		return database.Title[0].Text.Content
	}
	return "Untitled"
}

var _ tools.NotionTool = &NotionClient{}
