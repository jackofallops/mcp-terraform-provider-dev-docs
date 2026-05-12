package docs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SearchResult struct {
	Path    string
	Snippet string
}

var allowedExts = map[string]bool{".md": true, ".go": true, ".tf": true}

const maxFileSize = int64(2 * 1024 * 1024) // 2MB limit

func (m *Manager) Search(ctx context.Context, provider, query, version string) ([]SearchResult, error) {
	if err := m.ensureVersion(ctx, provider, version); err != nil {
		return nil, err
	}

	base, err := m.getProviderPath(provider)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err != nil || info.IsDir() {
			return err
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !allowedExts[ext] || info.Size() > maxFileSize {
			return nil
		}

		found, snippet := m.searchInFile(ctx, path, query)
		if found {
			relPath, _ := filepath.Rel(base, path)
			results = append(results, SearchResult{
				Path:    relPath,
				Snippet: snippet,
			})
		}
		return nil
	})

	return results, err
}

func (m *Manager) searchInFile(ctx context.Context, path, query string) (bool, string) {
	file, err := os.Open(path)
	if err != nil {
		return false, ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		if ctx.Err() != nil {
			return false, ""
		}
		lineNum++
		if strings.Contains(strings.ToLower(scanner.Text()), strings.ToLower(query)) {
			snippet := fmt.Sprintf("Line %d:\n%s", lineNum, scanner.Text())
			return true, snippet
		}
	}
	return false, ""
}
