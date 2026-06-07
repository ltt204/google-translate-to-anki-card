package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/genai"
)

func main() {
	godotenv.Load()
	os.MkdirAll("audio", 0755)
	ctx := context.Background()
	client := &http.Client{}
	genaiClient, err := genai.NewClient(ctx, &genai.ClientConfig{})
	if err != nil {
		log.Fatal("Failed to create genai client: ", err)
	}

	config := &genai.GenerateContentConfig{Temperature: genai.Ptr[float32](0.5)}
	chat, err := genaiClient.Chats.Create(ctx, os.Getenv("LLM_MODEL"), config, nil)
	if err != nil {
		log.Fatal("Failed to create chat: ", err)
	}

	records := TransformToAnkiFormat(ctx, client, chat, "saved_translations.csv")
	PushToAnki(client, records)
}
