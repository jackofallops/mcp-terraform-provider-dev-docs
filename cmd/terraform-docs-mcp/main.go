package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"terraform-provider-dev-docs/internal/docs"
	"terraform-provider-dev-docs/internal/server"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func getWriteableCacheDir(requested string) string {
	// Attempt to create and write to the requested directory
	if err := os.MkdirAll(requested, 0755); err == nil {
		testFile := filepath.Join(requested, ".write_test")
		if wErr := os.WriteFile(testFile, []byte("test"), 0644); wErr == nil {
			_ = os.Remove(testFile)
			return requested
		}
	}

	// Fallback to system temp directory
	tempFallback := filepath.Join(os.TempDir(), "terraform-provider-docs-cache")
	log.Printf("warning: requested cache directory %q is not writeable. Falling back to system temporary directory: %s", requested, tempFallback)
	_ = os.MkdirAll(tempFallback, 0755)
	return tempFallback
}

func main() {
	repoURL := "https://github.com/hashicorp/web-unified-docs.git"

	cacheDir := os.Getenv("TERRAFORM_DOCS_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = ".docs_cache"
	}

	// Resolve writeable cache directory (handles read-only filesystems)
	cacheDir = getWriteableCacheDir(cacheDir)

	// Initialize Docs Manager (versions will be resolved per-call)
	dm := docs.NewManager(repoURL, cacheDir, nil)

	// Perform startup sync to clone/update the repository
	log.Println("Initializing documentation cache on startup...")
	if err := dm.Sync(context.Background()); err != nil {
		log.Fatalf("Failed to initialize documentation cache: %v", err)
	}

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

	log.Println("Starting MCP server...")
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
