package test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

type MCPHostClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
}

// Create a new mcp host server to post to 	
func NewMCPHostClient(model string) (*MCPHostClient, error) {
	cmd := exec.Command("mcphost", "-m", model)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start mcphost: %w", err)
	}

	client := &MCPHostClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}

	// انتظار حتى يكون جاهزاً (يمكن قراءة أول سطرين من المخرجات)
	time.Sleep(500 * time.Millisecond)

	return client, nil
}

//Send prompt from user to slm 
func (c *MCPHostClient) SendPrompt(prompt string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()


	_, err := fmt.Fprintf(c.stdin, "%s\n", prompt)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}


	scanner := bufio.NewScanner(c.stdout)
	var response strings.Builder
	var lineCount int

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++


		if lineCount == 1 && strings.TrimSpace(line) == "" {
			continue
		}


		if strings.TrimSpace(line) == "" && lineCount > 1 {
			break
		}


		if strings.HasPrefix(line, ">") {
			break
		}

		response.WriteString(line)
		response.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading output: %w", err)
	}

	return strings.TrimSpace(response.String()), nil
}

// send prompt with time out  
func (c *MCPHostClient) SendPromptWithTimeout(prompt string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resultChan := make(chan struct {
		response string
		err      error
	}, 1)

	go func() {
		response, err := c.SendPrompt(prompt)
		resultChan <- struct {
			response string
			err      error
		}{response, err}
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("timeout occurred after %v", timeout)
	case result := <-resultChan:
		return result.response, result.err
	}
}

// close the session
func (c *MCPHostClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()


	fmt.Fprintln(c.stdin, "exit")


	done := make(chan error, 1)
	go func() {
		done <- c.cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):

		return c.cmd.Process.Kill()
	}
}

// Handel Errors
func (c *MCPHostClient) ReadErrors() (string, error) {
	data, err := io.ReadAll(c.stderr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func TestModel(t *testing.T) {

	client, err := NewMCPHostClient("ollama:qwen2.5:v1")
	if err != nil {
		t.Logf("Error: %v\n", err)
		return
	}
	defer client.Close()

	//Example 01:
	response, err := client.SendPromptWithTimeout(
		"ما هي عاصمة الجزائر؟",
		30*time.Second,
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	t.Logf("Responce: %s\n\n", response)

	//Example 02:
	response, err = client.SendPromptWithTimeout(
		"ما هو الطقس اليوم في سطيف؟",
		30*time.Second,
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	t.Logf("Response: %s\n\n", response)

	response, err = client.SendPromptWithTimeout(
		"احسب 156 * 789",
		30*time.Second,
	)
	if err != nil {
		t.Logf("Error: %v\n", err)
		return
	}
	t.Logf("response: %s\n\n", response)

}
