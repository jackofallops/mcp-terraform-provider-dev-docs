package docs

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type SearchResult struct {
	Path    string
	Snippet string
}

// Search scans the provider's directory for the query string
func (m *Manager) Search(provider, query string) ([]SearchResult, error) {
	base, err := m.getProviderPath(provider)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			found, snippet := m.searchInFile(path, query)
			if found {
				// Make path relative to the provider root for the agent
				relPath, _ := filepath.Rel(base, path)
				results = append(results, SearchResult{
					Path:    relPath,
					Snippet: snippet,
				})
			}
		}
		return nil
	})

	return results, err
}

func (m *Manager) searchInFile(path, query string) (bool, string) {
	file, err := os.Open(path)
	if err != nil {
		return false, ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(strings.ToLower(line), strings.ToLower(query)) {
			return true, strings.TrimSpace(line)
		}
	}
	return false, ""
}
