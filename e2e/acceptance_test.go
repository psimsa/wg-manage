package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/ofcoursedude/wg-manage/models"
	"gopkg.in/yaml.v3"
)

type testFixture struct {
	repoRoot   string
	exePath    string
	cleanupFn  func()
	cleanupErr error
}

var (
	fixture     *testFixture
	fixtureOnce sync.Once
)

func getFixture(t *testing.T) *testFixture {
	t.Helper()
	fixtureOnce.Do(func() {
		root := repoRoot(t)
		exe, cleanup := buildBinary(t, root)
		fixture = &testFixture{
			repoRoot:  root,
			exePath:   exe,
			cleanupFn: cleanup,
		}
	})
	return fixture
}

func TestBootstrap(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	runCommand(t, f.repoRoot, f.exePath, "bootstrap", "-endpoint", "example.com:51820", "-output", configPath)

	cfg := models.LoadYaml(configPath)
	if len(cfg.Peers) != 3 {
		t.Fatalf("expected 3 peers after bootstrap, got %d", len(cfg.Peers))
	}
	server := findPeer(t, cfg.Peers, "Server")
	if server.Endpoint == nil || *server.Endpoint != "example.com:51820" {
		t.Fatalf("expected server endpoint to be set")
	}
}

func TestGenerateConfigs(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	runCommand(t, f.repoRoot, f.exePath, "bootstrap", "-endpoint", "example.com:51820", "-output", configPath)
	cfg := models.LoadYaml(configPath)

	outputDir := filepath.Join(tempDir, "output")
	runCommand(t, f.repoRoot, f.exePath, "generate", "-config", configPath, "-output", outputDir, "-png", "false")

	for _, peer := range cfg.Peers {
		configFile := filepath.Join(outputDir, peer.Name+".conf")
		if _, err := os.Stat(configFile); err != nil {
			t.Fatalf("expected config file for %s: %v", peer.Name, err)
		}
	}
}

func TestGeneratePNGs(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	runCommand(t, f.repoRoot, f.exePath, "bootstrap", "-endpoint", "example.com:51820", "-output", configPath)
	cfg := models.LoadYaml(configPath)

	pngOutputDir := filepath.Join(tempDir, "output-png")
	runCommand(t, f.repoRoot, f.exePath, "generate", "-config", configPath, "-output", pngOutputDir)

	for _, peer := range cfg.Peers {
		pngFile := filepath.Join(pngOutputDir, peer.Name+".png")
		if _, err := os.Stat(pngFile); err != nil {
			t.Fatalf("expected png file for %s: %v", peer.Name, err)
		}
	}
}

func TestAddPeer(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	runCommand(t, f.repoRoot, f.exePath, "bootstrap", "-endpoint", "example.com:51820", "-output", configPath)

	runCommand(t, f.repoRoot, f.exePath, "add", "-name", "Office", "-ip", "10.0.2.10/32", "-endpoint", "office.example.com:51820", "-config", configPath)

	cfg := models.LoadYaml(configPath)
	if len(cfg.Peers) != 4 {
		t.Fatalf("expected 4 peers after add, got %d", len(cfg.Peers))
	}
	office := findPeer(t, cfg.Peers, "Office")
	if office.ListenPort == nil || *office.ListenPort != 51820 {
		t.Fatalf("expected Office listen port 51820")
	}
}

func TestRemovePeer(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	runCommand(t, f.repoRoot, f.exePath, "bootstrap", "-endpoint", "example.com:51820", "-output", configPath)
	runCommand(t, f.repoRoot, f.exePath, "add", "-name", "Office", "-ip", "10.0.2.10/32", "-endpoint", "office.example.com:51820", "-config", configPath)

	runCommand(t, f.repoRoot, f.exePath, "remove", "-name", "Office", "-config", configPath)

	cfg := models.LoadYaml(configPath)
	if len(cfg.Peers) != 3 {
		t.Fatalf("expected 3 peers after remove, got %d", len(cfg.Peers))
	}
	if peerExists(cfg.Peers, "Office") {
		t.Fatalf("expected Office peer to be removed")
	}
}

func TestInit(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	initConfig := filepath.Join(tempDir, "init.yaml")

	runCommand(t, f.repoRoot, f.exePath, "init", "-peers", "2", "-output", initConfig, "-simple=true", "-preshared=false")

	initCfg := models.LoadYaml(initConfig)
	if len(initCfg.Peers) != 2 {
		t.Fatalf("expected 2 peers from init, got %d", len(initCfg.Peers))
	}
	if initCfg.PresharedKey != nil && *initCfg.PresharedKey != "" {
		t.Fatalf("expected no preshared key when disabled")
	}
}

func TestFormat(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	runCommand(t, f.repoRoot, f.exePath, "bootstrap", "-endpoint", "example.com:51820", "-output", configPath)

	formatted := runCommand(t, f.repoRoot, f.exePath, "format", "-input", configPath)

	var formattedCfg models.Configuration
	if err := yaml.Unmarshal([]byte(formatted), &formattedCfg); err != nil {
		t.Fatalf("expected format output to be valid yaml: %v", err)
	}
}

func TestRecreate(t *testing.T) {
	f := getFixture(t)
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	runCommand(t, f.repoRoot, f.exePath, "bootstrap", "-endpoint", "example.com:51820", "-output", configPath)
	cfg := models.LoadYaml(configPath)

	recreated := runCommand(t, f.repoRoot, f.exePath, "recreate", "-config", configPath)

	var recreatedCfg models.Configuration
	if err := yaml.Unmarshal([]byte(recreated), &recreatedCfg); err != nil {
		t.Fatalf("expected recreate output to be valid yaml: %v", err)
	}

	oldKeys := peerKeyMap(cfg.Peers)
	newKeys := peerKeyMap(recreatedCfg.Peers)
	for name, oldKey := range oldKeys {
		newKey, ok := newKeys[name]
		if !ok {
			t.Fatalf("expected recreate output to include %s", name)
		}
		if newKey == oldKey {
			t.Fatalf("expected new key for %s", name)
		}
	}
}

func buildBinary(t *testing.T, repoRoot string) (string, func()) {
	t.Helper()
	exeSuffix := ""
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	tmpDir := os.TempDir()
	binaryPath := filepath.Join(tmpDir, "wg-manage-test-"+fmt.Sprint(time.Now().Unix())+exeSuffix)
	command := exec.Command("go", "build", "-o", binaryPath, "./cmd/wg-manage")
	command.Dir = repoRoot
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, string(output))
	}
	cleanup := func() {
		os.Remove(binaryPath)
	}
	return binaryPath, cleanup
}

func runCommand(t *testing.T, repoRoot, exePath string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, exePath, args...)
	command.Dir = repoRoot
	output, err := command.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("command timed out: %s", fmt.Sprint(args))
	}
	if err != nil {
		t.Fatalf("command failed: %s\n%s", fmt.Sprint(args), string(output))
	}
	return string(output)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	return filepath.Dir(cwd)
}

func findPeer(t *testing.T, peers []models.Peer, name string) models.Peer {
	t.Helper()
	for _, peer := range peers {
		if peer.Name == name {
			return peer
		}
	}
	t.Fatalf("expected peer %s to exist", name)
	return models.Peer{}
}

func peerExists(peers []models.Peer, name string) bool {
	for _, peer := range peers {
		if peer.Name == name {
			return true
		}
	}
	return false
}

func peerKeyMap(peers []models.Peer) map[string]string {
	keys := make(map[string]string, len(peers))
	for _, peer := range peers {
		keys[peer.Name] = peer.PublicKey
	}
	return keys
}
