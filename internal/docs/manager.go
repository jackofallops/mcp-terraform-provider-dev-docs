package docs

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	RepoURL   string
	CacheDir  string
	Providers map[string]string // framework, sdk, core -> relative paths
	Versions  map[string]string // provider -> git tag/branch (e.g., "v1.6.0", "main")
}

func NewManager(repoURL, cacheDir string, versions map[string]string) *Manager {
	return &Manager{
		RepoURL:   repoURL,
		CacheDir:  cacheDir,
		Providers: map[string]string{
			"framework": "content/terraform-plugin-framework",
			"sdk":       "content/terraform-plugin-sdk",
			"core":      "content/terraform",
		},
		Versions: versions,
	}
}

func (m *Manager) Sync(ctx context.Context) error {
	if _, err := os.Stat(m.CacheDir); os.IsNotExist(err) {
		fmt.Println("Cloning documentation repository...")
		for provider, tag := range m.Versions {
			targetDir := filepath.Join(m.CacheDir, m.Providers[provider])
			cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--single-branch", "--branch", tag, m.RepoURL, targetDir)
			if output, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("git clone failed for %s: %v, output: %s", provider, err, string(output))
			}
		}
	} else {
		fmt.Println("Updating documentation repository...")
		for provider, tag := range m.Versions {
			targetDir := filepath.Join(m.CacheDir, m.Providers[provider])
			
			fetchCmd := exec.CommandContext(ctx, "git", "-C", targetDir, "fetch", "origin", tag)
			if output, err := fetchCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("git fetch failed for %s: %v, output: %s", provider, err, string(output))
			}
			
			checkoutCmd := exec.CommandContext(ctx, "git", "-C", targetDir, "checkout", tag)
			if output, err := checkoutCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("git checkout failed for %s: %v, output: %s", provider, err, string(output))
			}
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

func (m *Manager) ensureVersion(ctx context.Context, provider, version string) error {
	if version == "" {
		return nil // Use currently checked out state
	}

	targetDir := filepath.Join(m.CacheDir, m.Providers[provider])
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", "--single-branch", "--branch", version, m.RepoURL, targetDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone failed for %s: %v, output: %s", provider, err, string(output))
		}
		return nil
	}

	out, err := exec.CommandContext(ctx, "git", "-C", targetDir, "rev-parse", "--abbrev-ref", "HEAD").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get current ref for %s: %v", provider, err)
	}
	currentRef := strings.TrimSpace(string(out))

	if currentRef == version || (strings.HasPrefix(version, "refs/tags/") && currentRef == version) {
		return nil
	}

	fetchCmd := exec.CommandContext(ctx, "git", "-C", targetDir, "fetch", "origin", version)
	if output, err := fetchCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git fetch failed for %s: %v, output: %s", provider, err, string(output))
	}

	checkoutCmd := exec.CommandContext(ctx, "git", "-C", targetDir, "checkout", version)
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout failed for %s: %v, output: %s", provider, err, string(output))
	}
	return nil
}

func (m *Manager) List(ctx context.Context, provider, subPath, version string) ([]string, error) {
	if err := m.ensureVersion(ctx, provider, version); err != nil {
		return nil, err
	}

	base, err := m.getProviderPath(provider)
	if err != nil {
		return nil, err
	}

	cleanedPath := filepath.Clean(subPath)
	fullPath := filepath.Join(base, cleanedPath)
	if !strings.HasPrefix(fullPath, base+string(filepath.Separator)) && fullPath != base {
		return nil, fmt.Errorf("invalid path: directory traversal detected")
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

func (m *Manager) Read(ctx context.Context, provider, path, version string) (string, error) {
	if err := m.ensureVersion(ctx, provider, version); err != nil {
		return "", err
	}

	base, err := m.getProviderPath(provider)
	if err != nil {
		return "", err
	}

	cleanedPath := filepath.Clean(path)
	fullPath := filepath.Join(base, cleanedPath)
	if !strings.HasPrefix(fullPath, base+string(filepath.Separator)) && fullPath != base {
		return "", fmt.Errorf("invalid path: directory traversal detected")
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
