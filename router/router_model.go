package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/CharafEB/slm-unv/tools"
	"github.com/CharafEB/slm-unv/types"
	"github.com/redis/go-redis/v9"
)

type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}

// ChatMessage
type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall
type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function
type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// OllamaChatRequest
type OllamaChatRequest struct {
	Model    string             `json:"model"`
	Messages []ChatMessage      `json:"messages"`
	Tools    []tools.OllamaTool `json:"tools,omitempty"`
	Stream   bool               `json:"stream"`
}

// OllamaChatResponse responces from Ollama API
type OllamaChatResponse struct {
	Model         string      `json:"model"`
	Message       ChatMessage `json:"message"`
	Done          bool        `json:"done"`
	TotalDuration int64       `json:"total_duration,omitempty"`
}

// OllamaRouter manages communication with Ollama API and Redis
type OllamaRouter struct {
	client  *OllamaClient
	redis   *redis.Client
	tools   *tools.Tools
	channel string
}

// NewOllamaClient build a new OllamaClient
func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// NewOllamaRouter build a new OllamaRouter
func NewOllamaRouter(ollamaURL string, redisClient *redis.Client, toolsApp *tools.Tools) *OllamaRouter {
	return &OllamaRouter{
		client:  NewOllamaClient(ollamaURL),
		redis:   redisClient,
		tools:   toolsApp,
		channel: "news_channel",
	}
}

// ChatWithTools send chat request with tools to Ollama API
func (c *OllamaClient) ChatWithTools(ctx context.Context, model string, messages []ChatMessage, toolsList []tools.OllamaTool) (*OllamaChatResponse, error) {
	reqBody := OllamaChatRequest{
		Model:    model,
		Messages: messages,
		Tools:    toolsList,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error: status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result OllamaChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// ProcessMessage processes a message based on its label
func (r *OllamaRouter) ProcessMessage(ctx context.Context, label, input string) (string, error) {
	const ModelUnv = "qwen2.5:v1"
	const ModelGeneral = "qwen2.5:1.5b-instruct"
	messages := []ChatMessage{
		{
			Role:    "user",
			Content: input,
		},
	}

	switch label {
	case "__label__general":
		response, err := r.client.ChatWithTools(ctx, ModelGeneral, messages, nil)
		if err != nil {
			return "Internal error, please try later", err
		}
		return response.Message.Content, nil

	case "__label__author_article":
		return r.processWithTools(ctx, ModelUnv, messages)

	case "__label__article_search":
		return r.processWithTools(ctx, ModelUnv, messages)

	case "__label__university":
		response, err := r.client.ChatWithTools(ctx, ModelGeneral, messages, nil)
		if err != nil {
			return "Internal error, please try later", err
		}
		return response.Message.Content, nil

	default:
		return "Unknown label", nil
	}
}

// processWithTools process message using tools
func (r *OllamaRouter) processWithTools(ctx context.Context, model string, messages []ChatMessage) (string, error) {
	availableTools := r.tools.GetAvailableTools()

	response, err := r.client.ChatWithTools(ctx, model, messages, availableTools)
	if err != nil {
		return "Internal error, please try later", err
	}

	// check for tool calls
	if len(response.Message.ToolCalls) > 0 {
		var results strings.Builder

		for _, toolCall := range response.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				log.Printf("Error parsing tool arguments: %v", err)
				continue
			}

			result, err := r.tools.ExecuteTool(ctx, toolCall.Function.Name, args)
			if err != nil {
				log.Printf("Error executing tool %s: %v", toolCall.Function.Name, err)
				results.WriteString(fmt.Sprintf("Error: %v\n", err))
				continue
			}

			results.WriteString(result)
			results.WriteString("\n")
		}

		return results.String(), nil
	}

	return response.Message.Content, nil
}

// Start begins listening to Redis channel and processing messages
func (r *OllamaRouter) Start(ctx context.Context) error {
	// Subscribe to Redis channel
	pubsub := r.redis.Subscribe(ctx, r.channel)
	defer pubsub.Close()

	log.Printf(" Subscribed to Redis channel '%s'", r.channel)
	log.Println(" Listening for messages...")

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
