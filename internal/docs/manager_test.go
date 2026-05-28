package docs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input     string
		wantMajor int
		wantMinor int
		wantPatch int
		wantWild  bool
		wantOk    bool
	}{
		{"v1.18.x", 1, 18, 999999, true, true},
		{"v1.18.0", 1, 18, 0, false, true},
		{"v2.5.3", 2, 5, 3, false, true},
		{"1.0.0", 1, 0, 0, false, true},
		{"v1.18", 1, 18, 999999, true, true},
		{"vinvalid", 0, 0, 0, false, false},
	}

	for _, tt := range tests {
		v, ok := parseSemver(tt.input)
		if ok != tt.wantOk {
			t.Errorf("parseSemver(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
			continue
		}
		if ok {
			if v.major != tt.wantMajor || v.minor != tt.wantMinor || v.patch != tt.wantPatch || v.isWildcard != tt.wantWild {
				t.Errorf("parseSemver(%q) = %+v, want major=%d minor=%d patch=%d wild=%v", tt.input, v, tt.wantMajor, tt.wantMinor, tt.wantPatch, tt.wantWild)
			}
		}
	}
}

func TestGetVersionDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock web-unified-docs directories
	frameworkDir := filepath.Join(tempDir, "web-unified-docs", "content", "terraform-plugin-framework")
	err := os.MkdirAll(frameworkDir, 0755)
	if err != nil {
		t.Fatalf("failed to create mock directory: %v", err)
	}

	// Create version subdirectories
	versions := []string{"v1.18.x", "v1.17.x", "v1.18.0", "v2.0.0"}
	for _, v := range versions {
		err := os.Mkdir(filepath.Join(frameworkDir, v), 0755)
		if err != nil {
			t.Fatalf("failed to create version directory %s: %v", v, err)
		}
	}

	// Create manager
	m := NewManager("https://github.com/mock/repo.git", tempDir, nil)

	// Test case 1: Empty version yields latest semantic version (v2.0.0)
	dir, err := m.getVersionDir(context.Background(), "plugin-framework", "")
	if err != nil {
		t.Fatalf("getVersionDir failed: %v", err)
	}
	expectedDir := filepath.Join(frameworkDir, "v2.0.0")
	if dir != expectedDir {
		t.Errorf("getVersionDir(empty) = %q, want %q", dir, expectedDir)
	}

	// Test case 2: Exact match for existing version
	dir, err = m.getVersionDir(context.Background(), "plugin-framework", "v1.17.x")
	if err != nil {
		t.Fatalf("getVersionDir failed: %v", err)
	}
	expectedDir = filepath.Join(frameworkDir, "v1.17.x")
	if dir != expectedDir {
		t.Errorf("getVersionDir(v1.17.x) = %q, want %q", dir, expectedDir)
	}

	// Test case 3: Short form exact alias mapping
	dir, err = m.getVersionDir(context.Background(), "framework", "v1.18.0")
	if err != nil {
		t.Fatalf("getVersionDir failed: %v", err)
	}
	expectedDir = filepath.Join(frameworkDir, "v1.18.0")
	if dir != expectedDir {
		t.Errorf("getVersionDir(framework, v1.18.0) = %q, want %q", dir, expectedDir)
	}

	// Test case 4: Non-existent version triggers fallback to latest (v2.0.0)
	dir, err = m.getVersionDir(context.Background(), "framework", "v99.99.99")
	if err != nil {
		t.Fatalf("getVersionDir failed: %v", err)
	}
	expectedDir = filepath.Join(frameworkDir, "v2.0.0")
	if dir != expectedDir {
		t.Errorf("getVersionDir(v99.99.99) = %q, want %q (fallback to latest)", dir, expectedDir)
	}
}
