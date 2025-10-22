package tools

import (
	"context"
	"fmt"

	"github.com/CharafEB/slm-unv/middlewares"
	"github.com/CharafEB/slm-unv/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tools struct {
	*middlewares.Application
}

type OllamaTool struct {
	Type     string             `json:"type"`
	Function OllamaToolFunction `json:"function"`
}

type OllamaToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GetAvailableTools 
func (app *Tools) GetAvailableTools() []OllamaTool {
	return []OllamaTool{
		{
			Type: "function",
			Function: OllamaToolFunction{
				Name:        "author_article",
				Description: "Search for articles by author name. Use this for: books, articles, exams, professors, departments, classrooms",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "The name of the author, professor, or department to search for",
						},
					},
					"required": []string{"name"},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaToolFunction{
				Name:        "article_search",
				Description: "Search for articles by title or keyword",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"description": "The title or keyword to search for in articles",
						},
					},
					"required": []string{"title"},
				},
			},
		},
	}
}

// ExecuteTool 
func (app *Tools) ExecuteTool(ctx context.Context, toolName string, arguments map[string]interface{}) (string, error) {
	switch toolName {
	case "author_article":
		name, ok := arguments["name"].(string)
		if !ok || name == "" {
			return "", fmt.Errorf("missing or invalid 'name' parameter")
		}

		results, err := app.Database.Stored.AutherDBSearch(name)
		if err != nil {
			return fmt.Sprintf("Error searching for author '%s': %v", name, err), err
		}

		if len(results) == 0 {
			return fmt.Sprintf("No articles found for author: %s", name), nil
		}

		output := fmt.Sprintf("Found %d articles by '%s':\n", len(results), name)
		for i, result := range results {
			output += fmt.Sprintf("%d. %s\n", i+1, result)
		}
		return output, nil

	case "article_search":
		title, ok := arguments["title"].(string)
		if !ok || title == "" {
			return "", fmt.Errorf("missing or invalid 'title' parameter")
		}

		results, err := app.Database.Stored.ArticleDBSearch(title)
		if err != nil {
			return fmt.Sprintf("Error searching for article '%s': %v", title, err), err
		}

		if len(results) == 0 {
			return fmt.Sprintf("No articles found with title: %s", title), nil
		}

		output := fmt.Sprintf("Found %d articles matching '%s':\n", len(results), title)
		for i, result := range results {
			output += fmt.Sprintf("%d. %s\n", i+1, result)
		}
		return output, nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

// The Author tool
func (app *Tools) AuthorTool(ctx context.Context, req *mcp.CallToolRequest, input types.AuthorInput) (
	*mcp.CallToolResult,
	types.AuthorOutput,
	error,
) {
	var result []string
	//Check if the author name != null
	if input.Name == "" {
		return nil, types.AuthorOutput{Found: false, Results: []string{"the input name is nill"}}, nil
	}
	result, err := app.Database.Stored.AutherDBSearch(input.Name)

	//Check that we don't have any errors or nil result
	if err != nil {
		return nil, types.AuthorOutput{Found: false, Results: []string{"there is an err"}}, err
	} else if result == nil {
		return nil, types.AuthorOutput{Found: false, Results: []string{"there is no result"}}, nil
	}

	return nil, types.AuthorOutput{Found: true, Results: result}, nil
}

// The Article tool
func (app *Tools) ArticleTool(ctx context.Context, req *mcp.CallToolRequest, input types.ArticleInput) (
	*mcp.CallToolResult,
	types.ArticleOutput,
	error,
) {
	var result []string
	//Check if the Title name != null
	if input.Title == "" {
		return nil, types.ArticleOutput{Found: false, Results: []string{"sorry there is no article whit this title  "}}, nil
	}
	result, err := app.Database.Stored.ArticleDBSearch(input.Title)

	//Check that we don't have any errors or nil result
	if err != nil || result == nil {
		return nil, types.ArticleOutput{Found: false, Results: []string{"sorry there is no article whit this title"}}, nil
	}

	return nil, types.ArticleOutput{Found: true, Results: result}, nil
}

// Creating all tools in one place to make avery thing clean and clear
func (app *Tools) AddTools() {
	mcp.AddTool(app.Application.Srv, &mcp.Tool{Name: "author_article", Description: ""}, app.AuthorTool)

	mcp.AddTool(app.Application.Srv, &mcp.Tool{Name: "article_search", Description: ""}, app.ArticleTool)
}
