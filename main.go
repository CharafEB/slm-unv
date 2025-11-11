package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/CharafEB/slm-unv/middlewares"
	"github.com/CharafEB/slm-unv/model"
	"github.com/CharafEB/slm-unv/tools"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to PostgreSQL Database
	conninfo := os.Getenv("DB")
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	log.Println("Connected to PostgreSQL database")

	// Create store
	store := model.NewStore(db)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)

	// Create Application with all dependencies
	app := &middlewares.Application{
		Database: store,
		Srv:      server,
	}
	// Create Tools
	toolsApp := &tools.Tools{
		Application: app,
	}
	
	//initialize tools
	toolsApp.AddTools()
	log.Println("Tools initialized")

	// Run the server over stdin/stdout, until the client disconnects.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}

	//To make Developer fill good about him self so that he will work better (very important)
	log.Println("Application shut down Goodbye lovely Developer!")

}
