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
		return nil, fmt.Errorf("فشل في إنشاء stdin: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("فشل في إنشاء stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("فشل في إنشاء stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("فشل في تشغيل mcphost: %w", err)
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

	// إرسال الطلب
	_, err := fmt.Fprintf(c.stdin, "%s\n", prompt)
	if err != nil {
		return "", fmt.Errorf("فشل في إرسال الطلب: %w", err)
	}

	// قراءة الاستجابة
	scanner := bufio.NewScanner(c.stdout)
	var response strings.Builder
	var lineCount int

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		// تخطي الأسطر الفارغة في البداية
		if lineCount == 1 && strings.TrimSpace(line) == "" {
			continue
		}

		// توقف عند سطر فارغ بعد الاستجابة
		if strings.TrimSpace(line) == "" && lineCount > 1 {
			break
		}

		// توقف عند prompt جديد (عادة يبدأ بـ >)
		if strings.HasPrefix(line, ">") {
			break
		}

		response.WriteString(line)
		response.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("خطأ في القراءة: %w", err)
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
		return "", fmt.Errorf("انتهت مهلة الانتظار بعد %v", timeout)
	case result := <-resultChan:
		return result.response, result.err
	}
}

// close the session
func (c *MCPHostClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// محاولة الإغلاق بشكل نظيف
	fmt.Fprintln(c.stdin, "exit")

	// انتظار لمدة ثانية
	done := make(chan error, 1)
	go func() {
		done <- c.cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		// إذا لم يتوقف، أوقفه بالقوة
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
		fmt.Printf("❌ خطأ: %v\n", err)
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
