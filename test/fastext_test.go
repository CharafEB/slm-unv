package test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func runModel(modelName, prompt string) (string, error) {
	body := fmt.Sprintf(`{"model": "%s", "prompt": "%s"}`, modelName, prompt)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	return string(data), nil
}

func TestModel(t *testing.T) {
	response, err := runModel("qwen-sql:latest", "hi")
	if err != nil {
		t.Log(err)
	}
	t.Logf("Response from model: %s", response)
}
