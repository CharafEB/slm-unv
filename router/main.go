package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CharafEB/slm-unv/middlewares"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/mark3labs/mcphost/sdk"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Explicitly set the Ollama host address
	if os.Getenv("OLLAMA_HOST") == "" {
		os.Setenv("OLLAMA_HOST", "http://localhost:11434")
	}

	// Get the MCPHOST_MODEL from the environment variables, with a default value
	mcphostModel := os.Getenv("MCPHOST_MODEL")
	if mcphostModel == "" {
		mcphostModel = os.Getenv("MCPHOST_MODEL")
	}

	// Setup context with cancellation on interrupt signal
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	//Mcphost server : you had to make more then one connection to the host if you
	//like to chang the model for better output's for that add new host in the Mcphost struct then
	//creat a new host "sdk.New()" for an other model then pass it to use
	host, err := sdk.New(ctx, &sdk.Options{
		Model:      os.Getenv("MCPHOST_MODEL"),
		ConfigFile: "",
		MaxSteps:   5,
		Streaming:  false,
		Quiet:      true,
	})
	if err != nil {
		log.Fatalf("MCPHost error: %v", err)
	}
	defer host.Close()
	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDISADD_GO"),
		Username: os.Getenv("REDISUSERNAME"),
		Password: os.Getenv("REDISPASSWORD"),
	})
	defer rdb.Close()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Println("Error connecting to Redis:", err)
		return
	}

	app := middlewares.Mcphost{
		Host:    host,
		Redis:   rdb,
		Channel: os.Getenv("REDIS_CHANNEL"),
	}

	redisapp := Mcphost{
		Mcphost: &app,
	}
	log.Println("Connected to Redis, starting redis app")
	if err := redisapp.Start(ctx); err != nil {
		log.Println("Redis app error:", err)
	}

	//To make Developer fill good about him self so that he will work better (very important)
	log.Println("Application shut down Goodbye lovely Developer!")

}
