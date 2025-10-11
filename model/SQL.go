package model

import (
	"context"
	"fmt"
)

func (db *Tools) PerformDatabaseSearch(ctx context.Context, query string, args ...interface{}) ([]string, error) {
	var response []string

	// Execute the main query
	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	//Getting the result for the tables
	for rows.Next() {
		var title, author, content string
		if err := rows.Scan(&title, &author, &content); err != nil {
			return nil, err
		}
		response = append(response, fmt.Sprintf("Title: %s, Author: %s, Content: %s", title, author, content))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return response, nil
}