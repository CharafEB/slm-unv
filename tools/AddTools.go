package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/CharafEB/slm-unv/middlewares"
	"github.com/CharafEB/slm-unv/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tools struct {
	*middlewares.Application
}

// Using the function to make an search in the Database for data
func (app *Tools) SearchDatabase(ctx context.Context, req *mcp.CallToolRequest, input types.DBSearchInput) (
	*mcp.CallToolResult, types.DBSearchOutput, error,
) {

	// Basic validation
	if input.Term == "" {
		return nil, types.DBSearchOutput{Found: false, Source: "database"}, fmt.Errorf("search term cannot be empty")
	}

	// Define searchable columns
	allowedColumns := []string{"title", "author", "category", "content"}
	columnsToSearch := input.Columns
	if len(columnsToSearch) == 0 {
		columnsToSearch = allowedColumns
	}

	// Construct the query safely
	var whereClauses []string
	var args []interface{}
	for _, col := range columnsToSearch {
		// Ensure the column is allowed to be searched
		isAllowed := false
		for _, allowed := range allowedColumns {
			if col == allowed {
				isAllowed = true
				break
			}
		}
		if !isAllowed {
			return nil, types.DBSearchOutput{Found: false, Source: "database"}, fmt.Errorf("invalid column to search: %s", col)
		}
		whereClauses = append(whereClauses, fmt.Sprintf("%s ILIKE '%%' || $%d || '%%'", col, len(args)+1))
		args = append(args, input.Term)
	}

	query := fmt.Sprintf("SELECT title, author, content FROM articles WHERE %s", strings.Join(whereClauses, " OR "))

	results, err := app.Database.Stored.PerformDatabaseSearch(context.Background(), query, args)
	if err != nil {
		return nil, types.DBSearchOutput{
			Found:  false,
			Source: "database",
		}, err
	}

	if len(results) > 0 {
		return nil, types.DBSearchOutput{
			Found:   true,
			Results: results,
			Source:  "database",
		}, nil
	}

	return nil, types.DBSearchOutput{
		Found:  false,
		Source: "database",
	}, nil
}

func (app *Tools) AddTools() {
	// Creating all tools in one place to make avery thing clean and clear

	mcp.AddTool(app.Application.Srv, &mcp.Tool{Name: "search_database", Description: "Search the articles table in the database. The table has the following columns: id, title, author, category, published_date, content, views."}, app.SearchDatabase)
}
