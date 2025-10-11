package model

import (
	"context"
	"database/sql"
)

//Making the Tools connecting to connect to the data base in the functions
type Tools struct {
	DB *sql.DB
}

//The store struct to put all the func - tools - in one place  
type Store struct {
	// Stored interface to make the SQL DB search tools 
	Stored interface {
		PerformDatabaseSearch(ctx context.Context, query string, args ...interface{}) ([]string, error)
	}

}

//Make a connecting to the store
func NewStore(db *sql.DB) Store {
	if db == nil {
		panic("nil pointer passed to NewStore")
	}
	res := &Tools{DB: db}

	return Store{
		Stored: res,
	}
}