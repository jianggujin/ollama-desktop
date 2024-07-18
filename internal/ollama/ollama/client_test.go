package ollama

import (
	"context"
	"testing"
)

func TestClient_SearchPreview(t *testing.T) {
	q := "qwen"
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
	q := "qwen"
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Search(context.Background(), q, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range result.Items {
		t.Log(item)
	}
}

func TestClient_Library(t *testing.T) {
	//<option disabled="" value="">Sort By</option>
	//
	//<option value="featured" selected="">Featured</option>
	//<option value="popular">Most popular</option>
	//
	//<option value="newest">Newest</option>
	q := "qwen"
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	list, err := client.Library(context.Background(), q, "featured")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range list {
		t.Log(item)
	}
}

func TestClient_ModelTags(t *testing.T) {
	model := "gemma2"
	client, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.ModelTags(context.Background(), model)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(response.Model)
}
