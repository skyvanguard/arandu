package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/logging"

	"github.com/tmc/langchaingo/llms/ollama"
)

type OllamaProvider struct {
	client  *ollama.LLM
	model   string
	baseURL string
	name    ProviderType
}

func (p OllamaProvider) New() Provider {
	model := config.Config.OllamaModel
	baseURL := config.Config.OllamaServerURL

	client, err := ollama.New(
		ollama.WithModel(model),
		ollama.WithServerURL(baseURL),
		ollama.WithFormat("json"),
	)

	if err != nil {
		logging.Error("Failed to create Ollama client", "error", err.Error())
		os.Exit(1)
	}

	return OllamaProvider{
		client:  client,
		model:   model,
		baseURL: baseURL,
		name:    ProviderOllama,
	}
}

func (p OllamaProvider) Name() ProviderType {
	return p.name
}

func (p OllamaProvider) Summary(query string, n int) (string, error) {
	// Create a client without JSON format for summary
	client, err := ollama.New(
		ollama.WithModel(p.model),
		ollama.WithServerURL(p.baseURL),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create Ollama client: %v", err)
	}
	return Summary(client, p.model, query, n)
}

func (p OllamaProvider) DockerImageName(task string) (string, error) {
	// Create a client without JSON format for Docker image name
	client, err := ollama.New(
		ollama.WithModel(p.model),
		ollama.WithServerURL(p.baseURL),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create Ollama client: %v", err)
	}
	return DockerImageName(client, p.model, task)
}

// Call represents a tool call from a JSON-responding model
type Call struct {
	Tool    string            `json:"tool"`
	Input   map[string]string `json:"tool_input"`
	Message string            `json:"message"`
}

func (p OllamaProvider) NextTask(args NextTaskOptions) *database.Task {
	logging.Debug("Getting next task from Ollama", "model", p.model)

	prepared, err := PreparePrompt(PromptConfig{
		DockerImage:  args.DockerImage,
		Tasks:        args.Tasks,
		UseToolCalls: false, // Ollama uses JSON format
	})

	if err != nil {
		logging.Error("Failed to prepare prompt", "error", err.Error())
		return defaultAskTask("The conversation history is too long. Please start a new task.")
	}

	task, err := GenerateNextTask(context.Background(), GenerateTaskConfig{
		Client:       p.client,
		Model:        p.model,
		Messages:     prepared.Messages,
		UseToolCalls: false,
		Temperature:  0.0,
		TopP:         0.2,
	})

	if err != nil {
		logging.Error("Failed to generate task", "provider", "ollama", "error", err.Error())
		return defaultAskTask("There was an error getting the next task")
	}

	return task
}

// getToolPlaceholder generates the tool description for JSON-based models
func getToolPlaceholder() string {
	bs, err := json.Marshal(Tools)
	if err != nil {
		logging.Error("Failed to marshal tools for placeholder", "error", err.Error())
		os.Exit(1)
	}

	return fmt.Sprintf(`You have access to the following tools:

%s

To use a tool, respond with a JSON object with the following structure:
{
  "tool": <name of the called tool>,
  "tool_input": <parameters for the tool matching the above JSON schema>,
  "message": <a message that will be displayed to the user>
}

Always use a tool. Always reply with valid JSON. Always include a message.
`, string(bs))
}
