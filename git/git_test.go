package git

import (
	"testing"
	"os"
	"path/filepath"
)

func TestClone(t *testing.T) {
	base := os.TempDir()
	os.MkdirAll(base, 0666)
	dir := filepath.Join(base, "goboot-starter")
	url := "https://github.com/gostones/goboot-starter.git"
	Clone(url, dir)

	t.Logf("dir: %v", dir)
}

