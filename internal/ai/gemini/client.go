package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"spender-backend/v2/internal/domain"

	"google.golang.org/genai"
)

type Client struct {
	genaiClient *genai.Client
	modelName   string
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	// The new SDK can take nil if the key is in the environment,
	// or you can pass it explicitly via Config
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		genaiClient: client,
		modelName:   "gemini-2.5-flash", // Use the latest stable model
	}, nil
}

func (c *Client) ProcessText(ctx context.Context, rawText string) (*domain.Transaction, error) {
	// Setup the System Instruction and JSON requirement in the Config
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: `You are a financial data parser. Your goal is to first identify that message is transaction message or not if yes convert unstructured bank notifications or SMS into a valid JSON object.
				
				Rules:
				1. 'type': Must be "credited" (money received) or "debited" (money spent).
				2. 'amount': Must be a float.
				3. 'isTransaction': boolean value represents is this a transaction message or not. should be zero or one.
				
				JSON Schema Example:
				{
					"type": "credited",
					"amount": 45.50,
					"isTransaction": 1,
				}`},
			},
		},
		ResponseMIMEType: "application/json",
	}

	result, err := c.genaiClient.Models.GenerateContent(
		ctx,
		c.modelName,
		genai.Text(rawText),
		config,
	)
	if err != nil {
		return nil, fmt.Errorf("gemini execution error: %w", err)
	}

	// The new SDK provides a helper result.Text()
	jsonStr := result.Text()

	var tx domain.Transaction
	if err := json.Unmarshal([]byte(jsonStr), &tx); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON: %w. Raw: %s", err, jsonStr)
	}

	// This is not a valid transaction.
	if tx.IsTransaction == 0 {
		return nil, fmt.Errorf("Not a transaction message Dropping it")
	}

	return &tx, nil
}
