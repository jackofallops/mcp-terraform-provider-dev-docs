package docs

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type SearchResult struct {
	Path    string
	Snippet string
}

func (m *Manager) Search(ctx context.Context, provider, query, version string) ([]SearchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	repoDir := filepath.Join(m.CacheDir, "web-unified-docs")
	base, err := m.getVersionDir(ctx, provider, version)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(repoDir, base)
	if err != nil {
		return nil, fmt.Errorf("failed to compute relative search path: %w", err)
	}

	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "grep", "-n", "-i", "-F", "--no-color", "-e", query, "--", relPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// git grep exits with status 1 if no matches are found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil
		}
		return nil, fmt.Errorf("git grep failed: %v, output: %s", err, string(out))
	}

	var results []SearchResult
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 3 {
			continue
		}

		filePath := parts[0]
		lineNumStr := parts[1]
		content := parts[2]

		relFilePath, err := filepath.Rel(relPath, filePath)
		if err != nil {
			relFilePath = filePath
		}

		snippet := fmt.Sprintf("Line %s:\n%s", lineNumStr, content)
		results = append(results, SearchResult{
			Path:    relFilePath,
			Snippet: snippet,
		})

		// Cap search results to 100 to avoid overwhelming LLM contexts
		if len(results) >= 100 {
			break
		}
	}

	return results, nil
}
