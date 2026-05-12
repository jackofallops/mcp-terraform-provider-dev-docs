package server

import (
	"context"
	"fmt"
	"strings"

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
}

type ReadDocArgs struct {
	Provider string `json:"provider"           jsonschema:"The provider (framework, sdk, core)"`
	Path     string `json:"path"               jsonschema:"Path to the .md file"`
}

type SearchDocsArgs struct {
	Provider string `json:"provider"           jsonschema:"The provider (framework, sdk, core)"`
	Query    string `json:"query"              jsonschema:"Search keyword"`
}

func (h *Handler) HandleListDocs(ctx context.Context, req *mcp.CallToolRequest, args ListDocsArgs) (*mcp.CallToolResult, any, error) {
	provider := args.Provider
	if provider == "" {
		provider = "framework"
	}
	files, err := h.Docs.List(provider, args.Path)
	if err != nil {
		return nil, nil, err
	}
	text := fmt.Sprintf("Files in %s/%s:\n%s", provider, args.Path, strings.Join(files, "\n"))
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

func (h *Handler) HandleReadDoc(ctx context.Context, req *mcp.CallToolRequest, args ReadDocArgs) (*mcp.CallToolResult, any, error) {
	provider := args.Provider
	if provider == "" {
		provider = "framework"
	}
	if args.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	content, err := h.Docs.Read(provider, args.Path)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: content}},
	}, nil, nil
}

func (h *Handler) HandleSearchDocs(ctx context.Context, req *mcp.CallToolRequest, args SearchDocsArgs) (*mcp.CallToolResult, any, error) {
	provider := args.Provider
	if provider == "" {
		provider = "framework"
	}
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query is required")
	}
	results, err := h.Docs.Search(provider, args.Query)
	if err != nil {
		return nil, nil, err
	}
	if len(results) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No matching documentation found."}},
		}, nil, nil
	}
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Search results for '%s' in %s:\n", args.Query, provider))
	for _, res := range results {
		output.WriteString(fmt.Sprintf("- %s: %s\n", res.Path, res.Snippet))
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output.String()}},
	}, nil, nil
}
