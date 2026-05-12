package main

import (
	"context"
	"log"

	"terraform-provider-dev-docs/internal/docs"
	"terraform-provider-dev-docs/internal/server"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	repoURL := "https://github.com/hashicorp/web-unified-docs.git"
	cacheDir := ".docs_cache"

	// Initialize Docs Manager (versions will be resolved per-call)
	dm := docs.NewManager(repoURL, cacheDir, nil)

	handler := server.NewHandler(dm)

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "terraform-provider-docs",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_docs",
		Description: "List documentation files for a provider",
	}, handler.HandleListDocs)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "read_doc",
		Description: "Read the content of a specific documentation file",
	}, handler.HandleReadDoc)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_docs",
		Description: "Search for keywords in provider documentation",
	}, handler.HandleSearchDocs)

	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
