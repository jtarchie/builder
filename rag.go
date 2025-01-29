package builder

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/philippgille/chromem-go"
	"github.com/sashabaranov/go-openai"
)

type RAG struct {
	db *chromem.DB

	embedModel    string
	endpoint      string
	llmModel      string
	token         string
	embeddingFunc chromem.EmbeddingFunc
}

func NewRAG(filename string, embedModel string, llmModel string, endpoint string, token string) (*RAG, error) {
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

		embedModel:    embedModel,
		endpoint:      endpoint,
		llmModel:      llmModel,
		token:         token,
		embeddingFunc: chromem.NewEmbeddingFuncOpenAICompat(endpoint, token, embedModel, nil),
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

//go:embed rag/system.md
var systemPrompt string

func (r *RAG) Ask(query string) (string, error) {
	results, err := r.Search(query)
	if err != nil {
		return "", fmt.Errorf("failed to search: %w", err)
	}

	config := openai.DefaultConfig(r.token)
	config.BaseURL = r.endpoint
	client := openai.NewClientWithConfig(config)

	userPrompt := "Query: " + query + "\n\nDocuments:\n\n"
	for _, result := range results {
		userPrompt += fmt.Sprintf("- ID: %s\n```markdown\n%s\n```\n", result.ID, result.Content)
	}

	fmt.Println(userPrompt)

	response, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: r.llmModel,
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
