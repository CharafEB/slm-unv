package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/CharafEB/slm-unv/types"
)
//Publisher: do publish the slm response to the user channel
func (r *Mcphost) Publisher(ctx context.Context, output ,userchannel string) error {
	if err := r.Redis.Publish(ctx, userchannel, output).Err(); err != nil {
		return err
	}
	return nil
}
func (h *Mcphost) Model(ctx context.Context, prompt string) (string, error) {

	//you had to add more model for more labels so that you can make a good resolute
	response, err := h.Host.Prompt(ctx, "answer fast no thinking "+prompt)
	if err != nil {
		log.Fatalf("Prompt error: %v", err)
	}

	return response, nil

}

// ProcessMessage processes a message based on its label
func (r *Mcphost) ProcessMessage(ctx context.Context, label, input , userchannel string) error {

	switch label {
	case "__label__general":
		response, err := r.Model(ctx, input)
		if err != nil {
			return fmt.Errorf("there is an internal err error: %v", err)
		}
		if err := r.Publisher(ctx, response , userchannel); err != nil {
			return err
		}
		return nil

	case "__label__author_article":
		response, err := r.Model(ctx, input)
		if err != nil {
			return fmt.Errorf("there is an internal err error: %v", err)
		}
		if err := r.Publisher(ctx, response , userchannel); err != nil {
			return err
		}
		return nil

	case "__label__article_search":
		response, err := r.Model(ctx, input)
		if err != nil {
			return fmt.Errorf("there is an internal err error: %v", err)
		}
		if err := r.Publisher(ctx, response , userchannel); err != nil {
			return err
		}
		return nil

	case "__label__university":
		response, err := r.Model(ctx, input)
		if err != nil {
			return fmt.Errorf("there is an internal err error: %v", err)
		}
		if err := r.Publisher(ctx, response , userchannel); err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("unknown label")
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
			err := r.Redis.Publish(ctx, r.Channel, "ping").Err()
			if err != nil {
				log.Printf("Error publishing message: %v", err)
				continue
			}

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

			// Process with model
			if err := r.ProcessMessage(ctx, input.Label, input.Input , input.UserChannel); err != nil {

				log.Printf("Model error: %v", err)
				continue
			}

		}
	}
}
