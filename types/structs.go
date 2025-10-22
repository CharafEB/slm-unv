package types

import (
	"io"
	"os/exec"
	"sync"
)

// Author struct
type AuthorInput struct {
	Name string `json:"name" jsonschema_description:"the auther name to search on.`
}

type AuthorOutput struct {
	Found   bool     `json:"found"`
	Results []string `json:"results,omitempty"`
}

// Article struct
type ArticleInput struct {
	Title string `json:"name" jsonschema_description:"the auther name to search on.`
}

type ArticleOutput struct {
	Found   bool     `json:"found"`
	Results []string `json:"results,omitempty"`
}

type Input struct {
	Label string `json:"label"`
	Input string `json:"input"`
}

type Output struct {
	Response string `json:"response"`
}

type MCPHostClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
}