package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
)

// ReflectionAgent demonstrates a query -> generator -> reflector -> output pattern.
func reflect() {
	client := anthropic.NewClient()
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Enter your query:")
	query, _ := reader.ReadString('\n')
	query = query[:len(query)-1]

	// Step 1: Generator LLM call
	generatorPrompt := fmt.Sprintf("You are a helpful code assistant. Given the following query, generate a solution or plan.\nQuery: %s", query)
	genMsg := anthropic.NewUserMessage(anthropic.NewTextBlock(generatorPrompt))
	genResp, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		Model: anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{genMsg},
	})
	if err != nil {
		fmt.Printf("Generator error: %s\n", err.Error())
		return
	}
	genOutput := genResp.Content[0].Text
	fmt.Println("\n--- Generator Output ---\n", genOutput)

	// Step 2: Reflector LLM call
	reflectorPrompt := fmt.Sprintf("Reflect on the following solution or plan. Suggest improvements, point out flaws, and provide a refined version if possible.\nOriginal Output: %s", genOutput)
	reflMsg := anthropic.NewUserMessage(anthropic.NewTextBlock(reflectorPrompt))
	reflResp, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
		Model: anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{reflMsg},
	})
	if err != nil {
		fmt.Printf("Reflector error: %s\n", err.Error())
		return
	}
	finalOutput := reflResp.Content[0].Text
	fmt.Println("\n--- Final Output (Refined) ---\n", finalOutput)
}
