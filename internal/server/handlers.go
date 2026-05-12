package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"terraform-provider-dev-docs/internal/docs"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Handler struct {
	Docs *docs.Manager
}

func NewHandler(dm *docs.Manager) *Handler {
	return &Handler{Docs: dm}
}

type ListDocsArgs struct {
	Provider string `json:"provider" jsonschema:"The provider (framework, sdk, core)"`
	Path     string `json:"path"     jsonschema:"Subpath to list"`
	Version  string `json:"version,omitempty" jsonschema:"Git tag or branch for this provider"`
}

type ReadDocArgs struct {
	Provider string `json:"provider" jsonschema:"The provider (framework, sdk, core)"`
	Path     string `json:"path"     jsonschema:"Path to the .md file"`
	Version  string `json:"version,omitempty" jsonschema:"Git tag or branch for this provider"`
}

type SearchDocsArgs struct {
	Provider string `json:"provider" jsonschema:"The provider (framework, sdk, core)"`
	Query    string `json:"query"              jsonschema:"Search keyword"`
	Version  string `json:"version,omitempty" jsonschema:"Git tag or branch for this provider"`
}

func (h *Handler) HandleListDocs(ctx context.Context, req *mcp.CallToolRequest, args ListDocsArgs) (*mcp.CallToolResult, any, error) {
	provider := args.Provider
	if provider == "" {
		provider = "framework"
	}

	files, err := h.Docs.List(ctx, provider, args.Path, args.Version)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("error listing docs: %v", err)}}, IsError: true}, nil, nil
	}

	text := fmt.Sprintf("Files in %s/%s:\n%s", provider, args.Path, strings.Join(files, "\n"))
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

func (h *Handler) HandleReadDoc(ctx context.Context, req *mcp.CallToolRequest, args ReadDocArgs) (*mcp.CallToolResult, any, error) {
	allowed := map[string]bool{"framework": true, "sdk": true, "core": true}
	if !allowed[args.Provider] {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("unsupported provider: %s", args.Provider)}}, IsError: true}, nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if args.Path == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "path is required"}}, IsError: true}, nil, nil
	}

	content, err := h.Docs.Read(ctx, args.Provider, args.Path, args.Version)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("error reading doc: %v", err)}}, IsError: true}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}, nil, nil
}

func (h *Handler) HandleSearchDocs(ctx context.Context, req *mcp.CallToolRequest, args SearchDocsArgs) (*mcp.CallToolResult, any, error) {
	allowed := map[string]bool{"framework": true, "sdk": true, "core": true}
	if !allowed[args.Provider] {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("unsupported provider: %s", args.Provider)}}, IsError: true}, nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if args.Query == "" {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "query is required"}}, IsError: true}, nil, nil
	}

	results, err := h.Docs.Search(ctx, args.Provider, args.Query, args.Version)
	if err != nil {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("search failed: %v", err)}}, IsError: true}, nil, nil
	}

	if len(results) == 0 {
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "No matching documentation found."}}}, nil, nil
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search results for '%s' in %s:\n", args.Query, args.Provider))
	for _, res := range results {
		output.WriteString(fmt.Sprintf("- %s\n  %s\n", res.Path, res.Snippet))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output.String()}},
	}, nil, nil
}
