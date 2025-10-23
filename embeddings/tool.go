package main

import (
    "context"
    "fmt"
    "github.com/mark3labs/mcphost/sdk"
    "log"
)

func main() {
    ctx := context.Background()

    // إنشاء MCPHost مع خيارات مخصصة
    host, err := sdk.New(ctx, &sdk.Options{
        Model:      "ollama:qwen2.5:1.5b-instruct ", // ⚠️ لاحظ إضافة ollama: في البداية
        ConfigFile: "",                              // استخدام المسار الافتراضي ~/.mcphost.yml
        MaxSteps:   15,                              // عدد خطوات التفكير القصوى
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
