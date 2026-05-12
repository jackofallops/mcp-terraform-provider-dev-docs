package docs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	RepoURL   string
	CacheDir  string
	Providers map[string]string
}

func NewManager(repoURL, cacheDir string) *Manager {
	return &Manager{
		RepoURL:  repoURL,
		CacheDir: cacheDir,
		Providers: map[string]string{
			"framework": "content/terraform-plugin-framework",
			"sdk":       "content/terraform-plugin-sdk",
			"core":      "content/terraform",
		},
	}
}

// Sync ensures the documentation is cloned locally
func (m *Manager) Sync() error {
	if _, err := os.Stat(m.CacheDir); os.IsNotExist(err) {
		fmt.Println("Cloning documentation repository...")
		cmd := exec.Command("git", "clone", "--depth", "1", m.RepoURL, m.CacheDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone failed: %v, output: %s", err, string(output))
		}
	} else {
		fmt.Println("Updating documentation repository...")
		cmd := exec.Command("git", "-C", m.CacheDir, "pull")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git pull failed: %v, output: %s", err, string(output))
		}
	}
	return nil
}

func (m *Manager) getProviderPath(provider string) (string, error) {
	relPath, ok := m.Providers[provider]
	if !ok {
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
	return filepath.Join(m.CacheDir, relPath), nil
}

// List returns files and directories in the given provider path
func (m *Manager) List(provider, subPath string) ([]string, error) {
	base, err := m.getProviderPath(provider)
	if err != nil {
		return nil, err
	}

	fullPath := filepath.Join(base, subPath)
	// Prevent directory traversal
	if !strings.HasPrefix(fullPath, base) {
		return nil, fmt.Errorf("invalid path")
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, entry := range entries {
		result = append(result, entry.Name())
	}
	return result, nil
}

// Read returns the content of a specific markdown file
func (m *Manager) Read(provider, path string) (string, error) {
	base, err := m.getProviderPath(provider)
	if err != nil {
		return "", err
	}

	fullPath := filepath.Join(base, path)
	if !strings.HasPrefix(fullPath, base) {
		return "", fmt.Errorf("invalid path")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
