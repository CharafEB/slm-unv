package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)


type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
}


type ChatMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}


type ToolCall struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Function Function `json:"function"`
}


type Function struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}


type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}


type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type OllamaChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Tools    []Tool        `json:"tools,omitempty"`
	Stream   bool          `json:"stream"`
}

type OllamaChatResponse struct {
	Model         string      `json:"model"`
	Message       ChatMessage `json:"message"`
	Done          bool        `json:"done"`
	TotalDuration int64       `json:"total_duration,omitempty"`
}


type OllamaModel struct {
	Name       string    `json:"name"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
}

type OllamaListResponse struct {
	Models []OllamaModel `json:"models"`
}

func NewOllamaClient(baseURL string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *OllamaClient) ChatWithTools(ctx context.Context, model string, messages []ChatMessage, tools []Tool) (*OllamaChatResponse, error) {
	reqBody := OllamaChatRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
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

// ListModels الحصول على قائمة النماذج المتاحة
func (c *OllamaClient) ListModels(ctx context.Context) ([]OllamaModel, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error: status %d: %s", resp.StatusCode, string(body))
	}

	var result OllamaListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Models, nil
}

// تعريف الأدوات المتاحة
func getAvailableTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_current_time",
				Description: "Get the current time in a specific timezone",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"timezone": map[string]interface{}{
							"type":        "string",
							"description": "Timezone name (e.g., 'Africa/Algiers', 'UTC')",
						},
					},
					"required": []string{"timezone"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "search_university_info",
				Description: "Search for information about university departments, professors, courses",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query (e.g., 'computer science department', 'Dr. Ahmed')",
						},
						"category": map[string]interface{}{
							"type":        "string",
							"description": "Category: department, professor, course, classroom",
							"enum":        []string{"department", "professor", "course", "classroom"},
						},
					},
					"required": []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "calculate",
				Description: "Perform mathematical calculations",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"expression": map[string]interface{}{
							"type":        "string",
							"description": "Mathematical expression (e.g., '2 + 2', 'sqrt(16)')",
						},
					},
					"required": []string{"expression"},
				},
			},
		},
	}
}

// تنفيذ الأدوات
func executeTool(toolName string, arguments string) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	switch toolName {
	case "get_current_time":
		timezone, _ := args["timezone"].(string)
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			loc = time.UTC
		}
		return fmt.Sprintf("Current time in %s: %s", timezone, time.Now().In(loc).Format("2006-01-02 15:04:05")), nil

	case "search_university_info":
		query, _ := args["query"].(string)
		category, _ := args["category"].(string)
		// محاكاة بحث في قاعدة البيانات
		return fmt.Sprintf("Search results for '%s' in category '%s':\n- Department of Computer Science\n- Building A, Floor 3", query, category), nil

	case "calculate":
		expression, _ := args["expression"].(string)
		// محاكاة حساب بسيط (في الواقع استخدم مكتبة آمنة)
		return fmt.Sprintf("Result of '%s': 42 (mock result)", expression), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}
}

func main() {
	client := NewOllamaClient("http://localhost:11434")
	ctx := context.Background()

	log.Println("🔍 Checking Ollama connection...")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatalf("❌ Failed to connect to Ollama: %v", err)
	}

	log.Println("✅ Connected to Ollama")
	log.Printf("📋 Available models: %d\n", len(models))

	modelName := "qwen2.5:v1"
	tools := getAvailableTools()

	log.Printf("\n💬 Starting chat with %s", modelName)
	log.Println("🛠️  Available tools:")
	for _, tool := range tools {
		log.Printf("  - %s: %s", tool.Function.Name, tool.Function.Description)
	}
	log.Println(strings.Repeat("=", 60))

	reader := bufio.NewReader(os.Stdin)
	messages := []ChatMessage{}

	for {
		fmt.Print("\n👤 You: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}
		if strings.ToLower(input) == "exit" {
			log.Println("👋 Goodbye!")
			return
		}

		// إضافة رسالة المستخدم
		messages = append(messages, ChatMessage{
			Role:    "user",
			Content: input,
		})

		// إرسال للنموذج مع الأدوات
		fmt.Print("🤖 Assistant: ")
		response, err := client.ChatWithTools(ctx, modelName, messages, tools)
		if err != nil {
			log.Printf("\n❌ Error: %v\n", err)
			continue
		}

		// التحقق من استدعاء أدوات
		if len(response.Message.ToolCalls) > 0 {
			log.Println("\n🔧 Tool calls detected:")
			for _, toolCall := range response.Message.ToolCalls {
				log.Printf("  Calling: %s(%s)", toolCall.Function.Name, toolCall.Function.Arguments)
				result, err := executeTool(toolCall.Function.Name, toolCall.Function.Arguments)
				if err != nil {
					log.Printf("  ❌ Error: %v", err)
					continue
				}
				log.Printf("  ✅ Result: %s", result)

				// إضافة نتيجة الأداة للمحادثة
				messages = append(messages, ChatMessage{
					Role:    "tool",
					Content: result,
				})
			}

			// إعادة الإرسال مع نتائج الأدوات
			response, err = client.ChatWithTools(ctx, modelName, messages, tools)
			if err != nil {
				log.Printf("\n❌ Error: %v\n", err)
				continue
			}
		}

		fmt.Println(response.Message.Content)
		messages = append(messages, response.Message)
	}
}
