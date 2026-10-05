package appdir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirMigratesLegacyFolder(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	base, err := os.UserConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(base, "cerebrum")
	if err := os.MkdirAll(legacy, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "session.json"), []byte(`{"token":"t"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	path, err := File("session.json")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(filepath.Dir(path)) != "tridennote" {
		t.Fatalf("unexpected dir %s", path)
	}
	if data, err := os.ReadFile(path); err != nil || string(data) != `{"token":"t"}` {
		t.Fatalf("session not carried over: %q %v", data, err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatalf("legacy folder should be moved")
	}
}
