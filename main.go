package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CharafEB/slm-unv/middlewares"
	"github.com/CharafEB/slm-unv/model"
	"github.com/CharafEB/slm-unv/router"
	"github.com/CharafEB/slm-unv/tools"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mark3labs/mcphost/sdk"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// ⚠️ تحديد عنوان Ollama بشكل صريح
	if os.Getenv("OLLAMA_HOST") == "" {
		os.Setenv("OLLAMA_HOST", "http://localhost:11434")
	}

	// Setup context with cancellation on interrupt signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDISADD_GO"),
		Username: os.Getenv("REDISUSERNAME"),
		Password: os.Getenv("REDISPASSWORD"),
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Error connecting to Redis:", err)
	}
	//Mcphost server
	host, err := sdk.New(ctx, &sdk.Options{
		Model:      "ollama:qwen2.5:v1",
		ConfigFile: "",
		MaxSteps:   5,
		Streaming:  false,
		Quiet:      true,
	})
	if err != nil {
		log.Fatalf("MCPHost error: %v", err)
	}
	defer host.Close()

	redisapp := router.Mcphost{
		Host:    host,
		Redis:   rdb,
		Channel: "news_channel",
	}

	redisapp.Start(ctx)
	log.Println("Connected to Redis")

	// Create Application with all dependencies
	app := &middlewares.Application{
		Database: store,
	}

	// Create Tools
	toolsApp := &tools.Tools{
		Application: app,
	}
	toolsApp.AddTools()
	log.Println("Tools initialized")

	log.Println("Application shut down Goodbye lovely user!")
}
