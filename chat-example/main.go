package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/sherlockhua/koala/logs"
)

// NewChatModel creates a new chat model with Doubao provider
func NewChatModel(ctx context.Context, baseUrl, apiKey, modelName string) (model.ToolCallingChatModel, error) {
	// Create model configuration
	modelConfig := &ark.ChatModelConfig{
		BaseURL: baseUrl,
		APIKey:  apiKey,
		Model:   modelName,
	}

	// Print detailed configuration information for debugging
	logs.Infof(ctx, "Creating Doubao model with config:")
	logs.Infof(ctx, "BaseURL: %s\n", modelConfig.BaseURL)
	logs.Infof(ctx, "APIKey: %s\n", modelConfig.APIKey)
	logs.Infof(ctx, "Model: %s\n", modelConfig.Model)

	model, err := ark.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Doubao model %s: %w", modelName, err)
	}
	return model, nil
}

func main() {
	ctx := context.Background()
	baseUrl := "https://api.deepseek.com"
	apiKey := "sk-898cf368e3bf42008bc9ad8caafb719f"
	modelName := "deepseek-chat"
	model, err := NewChatModel(ctx, baseUrl, apiKey, modelName)
	if err != nil {
		fmt.Printf("failed to create model: %v\n", err)
		return
	}
	resp, err := model.Generate(ctx, []*schema.Message{
		{
			Role:    schema.Assistant,
			Content: fmt.Sprint(ai_summary_prompt, meetContent),
		},
	})
	if err != nil {
		fmt.Printf("failed to chat: %v\n", err)
		return
	}
	fmt.Printf("%v\n", resp.Content)

	var jsonOutput = extractJSON(resp.Content)
	fmt.Printf("JSON Output: %s\n", jsonOutput)
}

func extractJSON(input string) string {
	// 去除```json和```标记
	start := strings.Index(input, "```json")
	if start == -1 {
		start = strings.Index(input, "{")
		if start == -1 {
			return input
		}
	} else {
		start += len("```json")
	}

	end := strings.LastIndex(input, "```")
	if end == -1 {
		end = len(input)
	} else {
		end = strings.LastIndex(input, "```")
	}

	return strings.TrimSpace(input[start:end])
}
