package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "wg-manage-bin")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	binaryPath = filepath.Join(dir, "wg-manage")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		panic("could not build binary: " + err.Error())
	}
	os.Exit(m.Run())
}

var keyPattern = regexp.MustCompile(`[A-Za-z0-9+/]{43}=`)

func normalize(s string) string {
	return keyPattern.ReplaceAllString(s, "<KEY>")
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command %v failed: %v\n%s", args, err, out)
	}
	return string(out)
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	got = normalize(got)
	path := filepath.Join("testdata", "acceptance", name)
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("golden %s missing (run with UPDATE_GOLDEN=1): %v", name, err)
	}
	if string(want) != got {
		t.Errorf("output does not match golden %s\n--- want ---\n%s\n--- got ---\n%s", name, want, got)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestAcceptanceHelp(t *testing.T) {
	dir := t.TempDir()
	assertGolden(t, "help.txt.golden", run(t, dir))
}

func TestAcceptanceBootstrapGenerate(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")
	assertGolden(t, "bootstrap_config.yaml.golden", readFile(t, filepath.Join(dir, "config.yaml")))

	run(t, dir, "generate")
	entries, err := os.ReadDir(filepath.Join(dir, "output"))
	if err != nil {
		t.Fatal(err)
	}
	var confs []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".conf") {
			confs = append(confs, e.Name())
		}
	}
	sort.Strings(confs)
	if len(confs) != 3 {
		t.Fatalf("expected 3 .conf files, got %v", confs)
	}
	for _, c := range confs {
		assertGolden(t, "generate_"+c+".golden", readFile(t, filepath.Join(dir, "output", c)))
	}
	for _, c := range confs {
		png := strings.TrimSuffix(c, ".conf") + ".png"
		if _, err := os.Stat(filepath.Join(dir, "output", png)); err != nil {
			t.Errorf("expected QR png %s: %v", png, err)
		}
	}
}

func TestAcceptanceAddRemoveFormatRecreate(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")

	run(t, dir, "add", "-name", "Test Pi", "-ip", "10.0.2.4/32")
	assertGolden(t, "after_add_config.yaml.golden", readFile(t, filepath.Join(dir, "config.yaml")))

	assertGolden(t, "format_output.golden", run(t, dir, "format", "-input", "config.yaml"))

	run(t, dir, "remove", "-name", "Test Pi")
	assertGolden(t, "after_remove_config.yaml.golden", readFile(t, filepath.Join(dir, "config.yaml")))

	assertGolden(t, "recreate_output.golden", run(t, dir, "recreate", "-config", "config.yaml"))
}

func TestAcceptanceInit(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "init", "-peers", "2", "-output", "config.yaml")
	assertGolden(t, "init_config.yaml.golden", readFile(t, filepath.Join(dir, "config.yaml")))

	dir2 := t.TempDir()
	run(t, dir2, "init", "-peers", "1", "-simple", "-preshared=false", "-output", "simple.yaml")
	assertGolden(t, "init_simple_config.yaml.golden", readFile(t, filepath.Join(dir2, "simple.yaml")))
}

func TestAcceptanceGenerateCustomOutput(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")
	run(t, dir, "generate", "-output", "./custom-out", "-png=false")
	entries, err := os.ReadDir(filepath.Join(dir, "custom-out"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".png") {
			t.Errorf("unexpected png with -png=false: %s", e.Name())
		}
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 files in custom-out, got %d", len(entries))
	}
}
