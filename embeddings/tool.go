package main

import (
    "context"
    "fmt"
    "github.com/mark3labs/mcphost/sdk"
    "log"
)

func main() {
    ctx := context.Background()

    // Create MCPHost with custom options
    host, err := sdk.New(ctx, &sdk.Options{
        Model:      "ollama:qwen2.5:1.5b-instruct ", // ⚠️ Note the addition of ollama: at the beginning
        ConfigFile: "",                              // Use default path ~/.mcphost.yml
        MaxSteps:   15,                              // Maximum thinking steps
        Streaming:  false,                           // تعطيل البث المباشر
        Quiet:      true,                            // إخفاء رسائل debug
    })
    if err != nil {
        log.Fatalf("فشل في إنشاء MCPHost: %v", err)
    }
    defer host.Close()

    // إرسال prompt
    response, err := host.Prompt(ctx, "who are you?")
    if err != nil {
        log.Fatalf("فشل في إرسال Prompt: %v", err)
    }

    fmt.Println("الإجابة:", response)
}
