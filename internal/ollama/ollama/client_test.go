package ollama

import (
	"context"
	"ollama-desktop/internal/ollama"
	"testing"
)

func TestClient_SearchPreview(t *testing.T) {
	q := "falcon"
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	items, err := client.SearchPreview(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		t.Log(item)
	}
}

func TestClient_Search(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Search(context.Background(), ollama.SearchRequest{
		Q: "falcon",
		P: 1,
		C: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Items {
		t.Log(item)
	}
}

func TestClient_Library(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	list, err := client.Library(context.Background(), ollama.LibraryRequest{
		Q:    "falcon",
		Sort: "featured",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		t.Log(item)
	}
}

func TestClient_ModelTags(t *testing.T) {
	model := "falcon"
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.ModelTags(context.Background(), model)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response.Model)
	for _, item := range response.Tags {
		t.Log(item)
	}
}

func TestClient_ModelInfo(t *testing.T) {
	model := "falcon"
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.ModelInfo(context.Background(), model, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response.Model)
	for _, item := range response.Tags {
		t.Log(item)
	}
	t.Log(response.Readme)
}
