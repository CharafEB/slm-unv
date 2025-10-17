package model

import "github.com/lib/pq"

func (db *Tools) AutherDBSearch(AuthorName string) ([]string, error) {
	var response []string

	query := "SELECT ARRAY(SELECT title FROM articles WHERE similarity(author, $1) > 0.3)"

	err := db.DB.QueryRow(query, AuthorName).Scan(pq.Array(&response))
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (db *Tools) ArticleDBSearch(ArticleTitle string) ([]string, error) {

	var response []string
	// we had to  use similarity() in case where the user made an mistake sapling the  name to anabel this in you data base you have to => ' CREATE EXTENSION IF NOT EXISTS pg_trgm' to work

	query := "SELECT array_to_json(ARRAY(SELECT title FROM articles WHERE similarity(author, $1) > 0.3))"
	// Execute the main query
	err := db.DB.QueryRow(query, ArticleTitle).Scan(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
