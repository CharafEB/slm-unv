package main

import (
	"context"
	"log"
	"os"

	"database/sql"

	"github.com/CharafEB/slm-unv/middlewares"
	"github.com/CharafEB/slm-unv/model"
	"github.com/CharafEB/slm-unv/tools"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
 
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	conninfo := os.Getenv("DB")
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("There is an err in Ping: ", err)
	}
	store := model.NewStore(db)

	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)

	app := middlewares.Application{
		Srv:      server,
		Database: store,
	}

	appTools := tools.Tools{
		Application: &app,
	}

	appTools.AddTools()

	if err := app.Srv.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

