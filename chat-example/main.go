package main

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
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
	fmt.Printf("Creating Doubao model with config:\n")
	fmt.Printf("  BaseURL: %s\n", modelConfig.BaseURL)
	fmt.Printf("  APIKey: %s\n", modelConfig.APIKey)
	fmt.Printf("  Model: %s\n", modelConfig.Model)

	model, err := ark.NewChatModel(ctx, modelConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Doubao model %s: %w", modelName, err)
	}
	return model, nil
}

func main() {
	ctx := context.Background()
	baseUrl := "https://ark.cn-beijing.aliyuncs.com"
	apiKey := "1234567890"
	modelName := "ark-1.5"
	model, err := NewChatModel(ctx, baseUrl, apiKey, modelName)
	if err != nil {
		fmt.Printf("failed to create model: %v\n", err)
		return
	}
	resp, err := model.Generate(ctx, []*schema.Message{
		{
			Role:    schema.User,
			Content: "你好",
		},
	})
	if err != nil {
		fmt.Printf("failed to chat: %v\n", err)
		return
	}
	fmt.Printf("chat resp: %v\n", resp)
}
