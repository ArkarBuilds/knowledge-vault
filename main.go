package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
	openai "github.com/sashabaranov/go-openai"
)

func insertItem(db *sql.DB, title string, content string, embedding []float32) (int, error) {
	query := `
	        INSERT INTO knowledge_items (title, content, embedding)
			VALUES ($1, $2, $3)
			RETURNING id;
			`
	var id int

	err := db.QueryRow(query, title, content, pgvector.NewVector(embedding)).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func generateEmbedding(client *openai.Client, text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	}

	resp, err := client.CreateEmbeddings(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return resp.Data[0].Embedding, nil
}

func main() {
	connStr := "postgres://postgres:secretpassword@localhost:5432/knowledge_vault?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Missing OpenAI API key")
	}
	openAIClient := openai.NewClient(apiKey)
	fmt.Println("Successfully connected to the database!")

	noteTitle := "The Psychology of Money - Note 1"
	noteContent := "Wealth is what you don't see. It is the cars not purchased, the diamonds not bought, the watches not worn."

	fmt.Println("Generating embedding from OpenAI...")

	// 4. Chain the embedding generation to the database insertion
	vectorArray, err := generateEmbedding(openAIClient, noteContent)
	if err != nil {
		log.Fatalf("Failed to generate embedding: %v", err)
	}

	newID, err := insertItem(db, noteTitle, noteContent, vectorArray)
	if err != nil {
		log.Fatalf("Failed to insert knowledge item: %v", err)
	}

	fmt.Printf("Successfully inserted note '%s' with ID: %d and 1536-dimensional vector!\n", noteTitle, newID)
}
