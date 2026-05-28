package docs

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type Manager struct {
	mu        sync.RWMutex
	RepoURL   string
	CacheDir  string
	Providers map[string]string // Maps preferred verbose keys to relative paths in the repo
	Aliases   map[string]string // Maps short/alternative keys to preferred keys
}

func NewManager(repoURL, cacheDir string, versions map[string]string) *Manager {
	return &Manager{
		RepoURL:  repoURL,
		CacheDir: cacheDir,
		Providers: map[string]string{
			"plugin-framework": "content/terraform-plugin-framework",
			"plugin-sdk-v2":    "content/terraform-plugin-sdk",
			"terraform-core":   "content/terraform",
			"plugin-testing":   "content/terraform-plugin-testing",
			"plugin-go":        "content/terraform-plugin-go",
			"plugin-log":       "content/terraform-plugin-log",
			"plugin-mux":       "content/terraform-plugin-mux",
		},
		Aliases: map[string]string{
			"framework": "plugin-framework",
			"sdk":       "plugin-sdk-v2",
			"sdkv2":     "plugin-sdk-v2",
			"core":      "terraform-core",
			"testing":   "plugin-testing",
		},
	}
}

// Sync clones the unified documentation mono-repository if missing,
// or updates it to the latest main branch.
func (m *Manager) Sync(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	repoDir := filepath.Join(m.CacheDir, "web-unified-docs")
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		if err := os.MkdirAll(m.CacheDir, 0755); err != nil {
			return fmt.Errorf("failed to create cache directory: %w", err)
		}
		log.Printf("Cloning unified documentation repository to %s...", repoDir)
		cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", m.RepoURL, repoDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone failed: %v, output: %s", err, string(output))
		}
	} else {
		log.Printf("Updating unified documentation repository in %s...", repoDir)
		cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "pull", "origin", "main")
		if output, err := cmd.CombinedOutput(); err != nil {
			log.Printf("git pull failed: %v, output: %s. attempting hard reset to origin/main...", err, string(output))
			fetchCmd := exec.CommandContext(ctx, "git", "-C", repoDir, "fetch", "origin", "main")
			if fOut, fErr := fetchCmd.CombinedOutput(); fErr != nil {
				return fmt.Errorf("git fetch failed: %v, output: %s", fErr, string(fOut))
			}
			resetCmd := exec.CommandContext(ctx, "git", "-C", repoDir, "reset", "--hard", "origin/main")
			if rOut, rErr := resetCmd.CombinedOutput(); rErr != nil {
				return fmt.Errorf("git reset failed: %v, output: %s", rErr, string(rOut))
			}
		}
	}
	return nil
}

type semver struct {
	raw        string
	major      int
	minor      int
	patch      int
	isWildcard bool
}

func parseSemver(s string) (semver, bool) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return semver{}, false
	}

	var v semver
	v.raw = "v" + s

	if _, err := fmt.Sscanf(parts[0], "%d", &v.major); err != nil {
		return semver{}, false
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &v.minor); err != nil {
		return semver{}, false
	}

	if len(parts) >= 3 {
		if parts[2] == "x" || parts[2] == "X" || parts[2] == "*" {
			v.patch = 999999
			v.isWildcard = true
		} else {
			if _, err := fmt.Sscanf(parts[2], "%d", &v.patch); err != nil {
				v.patch = 0
			}
		}
	} else {
		v.patch = 999999
		v.isWildcard = true
	}

	return v, true
}

func (m *Manager) resolveProvider(provider string) (string, string, error) {
	if verbose, exists := m.Aliases[provider]; exists {
		provider = verbose
	}
	relPath, exists := m.Providers[provider]
	if !exists {
		return "", "", fmt.Errorf("unsupported provider: %s", provider)
	}
	return provider, relPath, nil
}

// getVersionDir resolves the absolute path to the directory for a provider version.
// If requestedVersion is empty, it returns the latest semantic version directory.
func (m *Manager) getVersionDir(ctx context.Context, provider, requestedVersion string) (string, error) {
	verboseProvider, relPath, err := m.resolveProvider(provider)
	if err != nil {
		return "", err
	}

	repoDir := filepath.Join(m.CacheDir, "web-unified-docs")
	providerDir := filepath.Join(repoDir, relPath)

	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		if err := m.Sync(ctx); err != nil {
			return "", fmt.Errorf("repository not initialized: %w", err)
		}
	}

	entries, err := os.ReadDir(providerDir)
	if err != nil {
		return "", fmt.Errorf("failed to read provider directory: %w", err)
	}

	var versions []semver
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "v") {
			continue
		}
		if v, ok := parseSemver(name); ok {
			versions = append(versions, v)
		}
	}

	if len(versions) == 0 {
		return "", fmt.Errorf("no version directories found for provider %s", verboseProvider)
	}

	sort.Slice(versions, func(i, j int) bool {
		if versions[i].major != versions[j].major {
			return versions[i].major > versions[j].major
		}
		if versions[i].minor != versions[j].minor {
			return versions[i].minor > versions[j].minor
		}
		return versions[i].patch > versions[j].patch
	})

	if requestedVersion == "" {
		return filepath.Join(providerDir, versions[0].raw), nil
	}

	normRequested := requestedVersion
	if !strings.HasPrefix(normRequested, "v") {
		normRequested = "v" + normRequested
	}

	for _, v := range versions {
		if v.raw == normRequested {
			return filepath.Join(providerDir, v.raw), nil
		}
	}

	reqV, ok := parseSemver(normRequested)
	if ok {
		for _, v := range versions {
			if v.major == reqV.major && v.minor == reqV.minor {
				return filepath.Join(providerDir, v.raw), nil
			}
		}
	}

	log.Printf("warning: version %s not found for provider %s, falling back to latest %s", requestedVersion, verboseProvider, versions[0].raw)
	return filepath.Join(providerDir, versions[0].raw), nil
}

func (m *Manager) List(ctx context.Context, provider, subPath, version string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	base, err := m.getVersionDir(ctx, provider, version)
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
		if entry.IsDir() {
			result = append(result, entry.Name()+"/")
		} else {
			result = append(result, entry.Name())
		}
	}
	return result, nil
}

func (m *Manager) Read(ctx context.Context, provider, path, version string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	base, err := m.getVersionDir(ctx, provider, version)
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
