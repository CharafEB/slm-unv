package test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/CharafEB/slm-unv/model"
	"github.com/CharafEB/slm-unv/types"
	_ "github.com/lib/pq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Tool struct {
	store model.Store
}

func (app *Tool) AuthorTool(ctx context.Context, req *mcp.CallToolRequest, input types.AuthorInput) (
	*mcp.CallToolResult,
	types.AuthorOutput,
	error,
) {
	var result []string
	//Check if the author name != null
	if input.Name == "" {
		return nil, types.AuthorOutput{Found: false, Results: []string{"the input name is nill"}}, nil
	}
	result, err := app.store.Stored.AutherDBSearch(input.Name)

	//Check that we don't have any errors or nil result
	if err != nil {
		return nil, types.AuthorOutput{Found: false, Results: []string{"there is an err"}}, err
	} else if result == nil {
		return nil, types.AuthorOutput{Found: false, Results: []string{"there is no result"}}, nil
	}

	return nil, types.AuthorOutput{Found: true, Results: result}, nil
}

func TestAutherDBSearch(t *testing.T) {
	db, err := sql.Open("postgres", "postgresql://work_on_user:bqXCYvGTJBx9AbLRoML6i0Kg9TFuoUe2@dpg-d3jo7bl6ubrc73d0brtg-a.frankfurt-postgres.render.com/work_on")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := model.NewStore(db)

	app := Tool{
		store: store,
	}

	app.AuthorTool(context.Background(), &mcp.CallToolRequest{}, types.AuthorInput{Name: "Alice Johnson"})

}
