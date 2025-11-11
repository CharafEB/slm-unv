package tools

import (
	"context"

	"github.com/CharafEB/slm-unv/middlewares"
	"github.com/CharafEB/slm-unv/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tools struct {
	*middlewares.Application
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