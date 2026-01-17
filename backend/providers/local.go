package providers

import (
	"context"
	"fmt"

	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/logging"

	"github.com/tmc/langchaingo/llms/openai"
)

// OpenAIClientConfig holds configuration for creating an OpenAI-compatible client
type OpenAIClientConfig struct {
	Token    string
	Model    string
	BaseURL  string
	Provider ProviderType
}

// createOpenAICompatibleClient creates an OpenAI client with the given configuration
// This centralizes the client creation logic used by all local providers
func createOpenAICompatibleClient(cfg OpenAIClientConfig) (*openai.LLM, error) {
	client, err := openai.New(
		openai.WithToken(cfg.Token),
		openai.WithModel(cfg.Model),
		openai.WithBaseURL(cfg.BaseURL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", cfg.Provider, err)
	}
	return client, nil
}

// mustCreateOpenAIClient creates an OpenAI client or logs error and returns nil
// Used during provider initialization where failure is logged but doesn't crash
func mustCreateOpenAIClient(cfg OpenAIClientConfig) *openai.LLM {
	client, err := createOpenAICompatibleClient(cfg)
	if err != nil {
		logging.Error("Failed to create provider client",
			"provider", cfg.Provider,
			"error", err.Error(),
		)
		return nil
	}
	return client
}

// LMStudioProvider implements the Provider interface for LM Studio
// LM Studio provides an OpenAI-compatible API on localhost:1234
type LMStudioProvider struct {
	client  *openai.LLM
	model   string
	baseURL string
	name    ProviderType
}

func (p LMStudioProvider) New() Provider {
	model := config.Config.LMStudioModel
	baseURL := config.Config.LMStudioServerURL

	client := mustCreateOpenAIClient(OpenAIClientConfig{
		Token:    "lm-studio", // LM Studio doesn't require an API key
		Model:    model,
		BaseURL:  baseURL,
		Provider: ProviderLMStudio,
	})

	return LMStudioProvider{
		client:  client,
		model:   model,
		baseURL: baseURL,
		name:    ProviderLMStudio,
	}
}

func (p LMStudioProvider) Name() ProviderType {
	return p.name
}

func (p LMStudioProvider) Summary(query string, n int) (string, error) {
	return Summary(p.client, p.model, query, n)
}

func (p LMStudioProvider) DockerImageName(task string) (string, error) {
	return DockerImageName(p.client, p.model, task)
}

func (p LMStudioProvider) NextTask(args NextTaskOptions) *database.Task {
	return localModelNextTask(p.client, p.model, args, true)
}

// LocalAIProvider implements the Provider interface for LocalAI
// LocalAI provides an OpenAI-compatible API
type LocalAIProvider struct {
	client  *openai.LLM
	model   string
	baseURL string
	name    ProviderType
}

func (p LocalAIProvider) New() Provider {
	model := config.Config.LocalAIModel
	baseURL := config.Config.LocalAIServerURL

	client := mustCreateOpenAIClient(OpenAIClientConfig{
		Token:    "local-ai", // LocalAI doesn't require an API key
		Model:    model,
		BaseURL:  baseURL,
		Provider: ProviderLocalAI,
	})

	return LocalAIProvider{
		client:  client,
		model:   model,
		baseURL: baseURL,
		name:    ProviderLocalAI,
	}
}

func (p LocalAIProvider) Name() ProviderType {
	return p.name
}

func (p LocalAIProvider) Summary(query string, n int) (string, error) {
	return Summary(p.client, p.model, query, n)
}

func (p LocalAIProvider) DockerImageName(task string) (string, error) {
	return DockerImageName(p.client, p.model, task)
}

func (p LocalAIProvider) NextTask(args NextTaskOptions) *database.Task {
	return localModelNextTask(p.client, p.model, args, true)
}

// OpenAICompatibleProvider is a generic provider for any OpenAI-compatible API
// Works with: vLLM, text-generation-webui, llama.cpp server, etc.
type OpenAICompatibleProvider struct {
	client  *openai.LLM
	model   string
	baseURL string
	name    ProviderType
}

func (p OpenAICompatibleProvider) New() Provider {
	model := config.Config.OpenAICompatibleModel
	baseURL := config.Config.OpenAICompatibleServerURL
	apiKey := config.Config.OpenAICompatibleAPIKey

	client := mustCreateOpenAIClient(OpenAIClientConfig{
		Token:    apiKey,
		Model:    model,
		BaseURL:  baseURL,
		Provider: ProviderOpenAICompatible,
	})

	return OpenAICompatibleProvider{
		client:  client,
		model:   model,
		baseURL: baseURL,
		name:    ProviderOpenAICompatible,
	}
}

func (p OpenAICompatibleProvider) Name() ProviderType {
	return p.name
}

func (p OpenAICompatibleProvider) Summary(query string, n int) (string, error) {
	return Summary(p.client, p.model, query, n)
}

func (p OpenAICompatibleProvider) DockerImageName(task string) (string, error) {
	return DockerImageName(p.client, p.model, task)
}

func (p OpenAICompatibleProvider) NextTask(args NextTaskOptions) *database.Task {
	// Use JSON mode (no tool calls) for maximum compatibility
	return localModelNextTask(p.client, p.model, args, false)
}

// localModelNextTask is a shared implementation for local model providers
// It handles both tool-calling models and JSON-response models
func localModelNextTask(client *openai.LLM, model string, args NextTaskOptions, useToolCalls bool) *database.Task {
	logging.Debug("Getting next task from local model", "model", model, "use_tool_calls", useToolCalls)

	prepared, err := PreparePrompt(PromptConfig{
		DockerImage:  args.DockerImage,
		Tasks:        args.Tasks,
		UseToolCalls: useToolCalls,
	})

	if err != nil {
		logging.Error("Failed to prepare prompt", "error", err.Error())
		return defaultAskTask("The conversation history is too long. Please start a new task.")
	}

	task, err := GenerateNextTask(context.Background(), GenerateTaskConfig{
		Client:       client,
		Model:        model,
		Messages:     prepared.Messages,
		UseToolCalls: useToolCalls,
		Temperature:  0.1, // Slightly higher for local models
		TopP:         0.9,
	})

	if err != nil {
		logging.Error("Failed to generate task", "provider", "local", "model", model, "error", err.Error())
		return defaultAskTask("There was an error connecting to the local model. Is it running?")
	}

	return task
}
