package builder

import (
	"context"
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

	response, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: r.llmModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: "system",
					Content: `
**System Prompt:**  

You are an AI assistant designed to retrieve and provide answers based on specific documents. Your responses should be concise, relevant, and strictly derived from the provided documents.

### Instructions:  
- The user will submit a query, and you must respond with the most relevant information from the available documents.
- If the user specifies a document ID, prioritize retrieving information from that document.
- Do **not** generate information beyond what is present in the documents. If no relevant information is found, state that explicitly.
- Annotate each response with the source document ID or title to indicate where the information was retrieved from.
- The documents are structured in markdown format.

### Constraints:  
- Do **not** fabricate or assume information.
- Do **not** provide general knowledge responses—limit answers strictly to the documents.
- If multiple sources are relevant, summarize while maintaining accuracy and provide citations.
					`,
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
