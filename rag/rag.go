package rag

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/philippgille/chromem-go"
	"github.com/sashabaranov/go-openai"
)

type OpenAIConfig struct {
	EmbedModel string
	Endpoint   string
	LLMModel   string
	Token      string
}

type RAG struct {
	db *chromem.DB

	config        OpenAIConfig
	embeddingFunc chromem.EmbeddingFunc
}

func New(filename string, config *OpenAIConfig) (*RAG, error) {
	if config == nil {
		config = &OpenAIConfig{
			// https://platform.openai.com/docs/guides/embeddings#embedding-models
			EmbedModel: "text-embedding-3-small",
			Endpoint:   "https://api.openai.com/v1",
			// https://platform.openai.com/docs/model
			LLMModel: "gpt-4o-mini",
			Token:    os.Getenv("OPENAI_API_KEY"),
		}
	}

	db := chromem.NewDB()

	if filename != ":memory:" {
		var err error

		db, err = chromem.NewPersistentDB(filename, true)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
	}

	return &RAG{
		db: db,

		config:        *config,
		embeddingFunc: chromem.NewEmbeddingFuncOpenAICompat(config.Endpoint, config.Token, config.EmbedModel, nil),
	}, nil
}

func (r *RAG) AddDocument(id string, document string) error {
	collection, err := r.db.GetOrCreateCollection("documents", nil, r.embeddingFunc)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = collection.AddDocument(ctx, chromem.Document{
		ID:      id,
		Content: document,
	})
	if err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

func (r *RAG) Search(query string) ([]chromem.Result, error) {
	collection, err := r.db.GetOrCreateCollection("documents", nil, r.embeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	results, err := collection.Query(ctx, query, 1, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	return results, nil
}

//go:embed system.md
var systemPrompt string

func (r *RAG) Ask(query string) (string, error) {
	results, err := r.Search(query)
	if err != nil {
		return "", fmt.Errorf("failed to search: %w", err)
	}

	config := openai.DefaultConfig(r.config.Token)
	config.BaseURL = r.config.Endpoint
	client := openai.NewClientWithConfig(config)

	userPrompt := "Query: " + query + "\n\nDocuments:\n\n"
	for _, result := range results {
		userPrompt += fmt.Sprintf("- ID: %s\n```markdown\n%s\n```\n", result.ID, result.Content)
	}

	fmt.Println(userPrompt)

	response, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: r.config.LLMModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: systemPrompt,
				},
				{
					Role:    "user",
					Content: userPrompt,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	return response.Choices[0].Message.Content, nil
}
