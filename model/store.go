package model

import (
	"database/sql"
)

// Making the Tools connecting to connect to the data base in the functions
type Tools struct {
	DB *sql.DB
}

// The store struct to put all the func - tools - in one place
type Store struct {
	// Stored interface to make the SQL DB search tools
	Stored interface {
		//AutherDBSearch: make a search about an article using the author name
		AutherDBSearch(AuthorName string) ([]string, error)

		//ArticleDBSearch: make a search about an article using the article Title
		ArticleDBSearch(ArticleTitle string) ([]string, error)
	}
}

// Make a connecting to the store
func NewStore(db *sql.DB) Store {
	if db == nil {
		panic("nil pointer passed to NewStore")
	}
	res := &Tools{DB: db}

	return Store{
		Stored: res,
	}
}
