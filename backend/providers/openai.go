package providers

import (
	"context"
	"os"

	"github.com/arandu-ai/arandu/config"
	"github.com/arandu-ai/arandu/database"
	"github.com/arandu-ai/arandu/logging"

	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAIProvider struct {
	client  *openai.LLM
	model   string
	baseURL string
	name    ProviderType
}

func (p OpenAIProvider) New() Provider {
	model := config.Config.OpenAIModel
	baseURL := config.Config.OpenAIServerURL

	client, err := openai.New(
		openai.WithToken(config.Config.OpenAIKey),
		openai.WithModel(model),
		openai.WithBaseURL(baseURL),
	)

	if err != nil {
		logging.Error("Failed to create OpenAI client", "error", err.Error())
		os.Exit(1)
	}

	return OpenAIProvider{
		client:  client,
		model:   model,
		baseURL: baseURL,
		name:    ProviderOpenAI,
	}
}

func (p OpenAIProvider) Name() ProviderType {
	return p.name
}

func (p OpenAIProvider) Summary(query string, n int) (string, error) {
	return Summary(p.client, p.model, query, n)
}

func (p OpenAIProvider) DockerImageName(task string) (string, error) {
	return DockerImageName(p.client, p.model, task)
}

func (p OpenAIProvider) NextTask(args NextTaskOptions) *database.Task {
	logging.Debug("Getting next task from OpenAI", "model", p.model)

	prepared, err := PreparePrompt(PromptConfig{
		DockerImage:  args.DockerImage,
		Tasks:        args.Tasks,
		UseToolCalls: true,
	})

	if err != nil {
		logging.Error("Failed to prepare prompt", "error", err.Error())
		return defaultAskTask("The conversation history is too long. Please start a new task.")
	}

	task, err := GenerateNextTask(context.Background(), GenerateTaskConfig{
		Client:       p.client,
		Model:        p.model,
		Messages:     prepared.Messages,
		UseToolCalls: true,
		Temperature:  0.0,
		TopP:         0.2,
	})

	if err != nil {
		logging.Error("Failed to generate task", "provider", "openai", "error", err.Error())
		return defaultAskTask("There was an error getting the next task")
	}

	return task
}
