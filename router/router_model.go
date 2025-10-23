package router

import (
	"context"
	"encoding/json"
	"log"

	"github.com/CharafEB/slm-unv/types"
	"github.com/mark3labs/mcphost/sdk"
	"github.com/redis/go-redis/v9"
)

type Mcphost struct {
	Host    *sdk.MCPHost
	Redis   *redis.Client
	Channel string
}

func (h *Mcphost) Model(ctx context.Context, prompt string) (string, error) {

	response, err := h.Host.Prompt(ctx, "who are you?")
	if err != nil {
		log.Fatalf("Prompt error: %v", err)
	}

	return response, nil

}

// ProcessMessage processes a message based on its label
func (r *Mcphost) ProcessMessage(ctx context.Context, label, input string) (string, error) {
	const ModelUnv = "qwen2.5:v1"
	const ModelGeneral = "qwen2.5:1.5b-instruct"

	switch label {
	case "__label__general":
		response, err := r.Model(ctx, input)
		if err != nil {
			return "Internal error, please try later", err
		}
		return response, nil

	case "__label__author_article":
		return r.Model(ctx, input)

	case "__label__article_search":
		return r.Model(ctx, input)

	case "__label__university":
		response, err := r.Model(ctx, input)
		if err != nil {
			return "Internal error, please try later", err
		}
		return response, nil

	default:
		return "Unknown label", nil
	}
}

// Start begins listening to Redis channel and processing input
func (r *Mcphost) Start(ctx context.Context) error {
	// Subscribe to Redis channel
	pubsub := r.Redis.Subscribe(ctx, r.Channel)
	defer pubsub.Close()

	log.Printf(" Subscribed to Redis channel '%s'", r.Channel)
	log.Println(" Listening for input...")

	for {
		select {
		case <-ctx.Done():
			log.Println(" Stopping message listener...")
			return nil

		default:
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				log.Printf(" Error receiving message: %v", err)
				continue
			}

			// Parse message
			var input types.Input
			if err := json.Unmarshal([]byte(msg.Payload), &input); err != nil {
				log.Printf("Error decoding message: %v", err)
				continue
			}

			log.Printf("Received: Label=%s, Input=%s", input.Label, input.Input)

			// Process with model
			response, err := r.ProcessMessage(ctx, input.Label, input.Input)
			if err != nil {
				log.Printf("Model error: %v", err)
				continue
			}

			log.Printf(" Response: %s", response)
		}
	}
}
