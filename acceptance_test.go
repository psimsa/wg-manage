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

func TestAcceptanceGenerateHonorsSettings(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")

	cfgPath := filepath.Join(dir, "config.yaml")
	data := readFile(t, cfgPath)
	data = strings.Replace(data, "outputDir: ./output", "outputDir: ./from-settings", 1)
	data = strings.Replace(data, "png: true", "png: false", 1)
	if err := os.WriteFile(cfgPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	run(t, dir, "generate")
	entries, err := os.ReadDir(filepath.Join(dir, "from-settings"))
	if err != nil {
		t.Fatalf("settings outputDir not honored: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".png") {
			t.Errorf("settings png=false not honored, found %s", e.Name())
		}
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 conf files in from-settings, got %d", len(entries))
	}

	run(t, dir, "generate", "-output", "./flag-wins")
	if _, err := os.Stat(filepath.Join(dir, "flag-wins")); err != nil {
		t.Errorf("flag -output should override settings: %v", err)
	}
}

func TestAcceptanceInvalidSettingsRejected(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")
	cfgPath := filepath.Join(dir, "config.yaml")
	data := readFile(t, cfgPath)
	data = strings.Replace(data, "ipPool: 10.0.2.0/24", "ipPool: 10.0.2.0/33", 1)
	if err := os.WriteFile(cfgPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binaryPath, "generate")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected failure for invalid ipPool, got success:\n%s", out)
	}
	if !strings.Contains(string(out), "ipPool") {
		t.Errorf("error should mention ipPool, got:\n%s", out)
	}
}

func TestAcceptanceOldConfigWithoutSettingsStillWorks(t *testing.T) {
	dir := t.TempDir()
	old := `peers:
- name: Server
  interfaceSection:
    privateKey: PRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRI=
  interfaceSectionWgQuick:
    address:
    - 10.0.2.1/32
  peerSection:
    publicKey: PUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPU=
    allowedIps:
    - 0.0.0.0/0
    endpoint: vpn.example.com:51820
- name: Client
  interfaceSection:
    privateKey: PRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRIVPRI2=
  interfaceSectionWgQuick:
    address:
    - 10.0.2.2/32
  peerSection:
    publicKey: PUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPUBPU2=
    allowedIps:
    - 10.0.2.2/32
`
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, "generate", "-output", "./out")
	entries, err := os.ReadDir(filepath.Join(dir, "out"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Error("old config without settings should still generate configs")
	}
	serverConf := readFile(t, filepath.Join(dir, "out", "Server.conf"))
	if strings.Contains(serverConf, "ListenPort") {
		t.Error("old config without settings must not gain ListenPort")
	}
}

func TestAcceptanceCustomYamlKeysSurviveCommands(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")
	cfgPath := filepath.Join(dir, "config.yaml")
	data := readFile(t, cfgPath)
	data = strings.Replace(data, "peers:", "projectmeta: my homelab\npeers:", 1)
	data = strings.Replace(data, "- name: Server", "- name: Server\n  location: Berlin\n  owner: alice\n  tags:\n  - infra\n  - vpn", 1)
	if err := os.WriteFile(cfgPath, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	run(t, dir, "add", "-name", "Test Pi", "-ip", "10.0.2.4/32")
	after := readFile(t, cfgPath)
	for _, want := range []string{"projectmeta: my homelab", "location: Berlin", "owner: alice", "- infra", "- vpn"} {
		if !strings.Contains(after, want) {
			t.Errorf("after add: %q missing:\n%s", want, after)
		}
	}

	formatted := run(t, dir, "format", "-input", "config.yaml")
	for _, want := range []string{"projectmeta: my homelab", "location: Berlin", "owner: alice"} {
		if !strings.Contains(formatted, want) {
			t.Errorf("after format: %q missing:\n%s", want, formatted)
		}
	}

	run(t, dir, "remove", "-name", "Test Pi")
	after = readFile(t, cfgPath)
	for _, want := range []string{"projectmeta: my homelab", "location: Berlin", "owner: alice", "- infra"} {
		if !strings.Contains(after, want) {
			t.Errorf("after remove: %q missing:\n%s", want, after)
		}
	}

	run(t, dir, "recreate", "-config", "config.yaml")
	run(t, dir, "generate", "-output", "./out")
	if _, err := os.Stat(filepath.Join(dir, "out", "Server.conf")); err != nil {
		t.Errorf("generate after mutations failed: %v", err)
	}
}

func TestAcceptanceIPAMAutoAllocate(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")

	run(t, dir, "add", "-name", "Auto Pi")
	data := readFile(t, filepath.Join(dir, "config.yaml"))
	if !strings.Contains(data, "- 10.0.2.4/32") {
		t.Errorf("expected auto-allocated 10.0.2.4/32 for Auto Pi:\n%s", data)
	}

	run(t, dir, "add", "-server", "-endpoint", "vpn2.example.com:51821", "-name", "Second Server")
	data = readFile(t, filepath.Join(dir, "config.yaml"))
	if !strings.Contains(data, "- 10.0.2.2/32") {
		t.Errorf("expected server allocated 10.0.2.2/32 from servers subnet:\n%s", data)
	}
	if !strings.Contains(data, "listenPort: 51821") {
		t.Errorf("expected listenPort 51821 from endpoint:\n%s", data)
	}

	run(t, dir, "remove", "-name", "Auto Pi")
	run(t, dir, "add", "-name", "Reused Pi")
	data = readFile(t, filepath.Join(dir, "config.yaml"))
	if !strings.Contains(data, "- 10.0.2.4/32") {
		t.Errorf("expected freed address 10.0.2.4/32 to be reused:\n%s", data)
	}
}

func TestAcceptanceIPAMRejectsOutOfPool(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")
	cmd := exec.Command(binaryPath, "add", "-name", "Bad", "-ip", "192.168.5.5/32")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err == nil {
		t.Fatalf("expected out-of-pool -ip to fail, got success:\n%s", out)
	}
	run(t, dir, "add", "-name", "Forced", "-ip", "192.168.5.5/32", "-force")
	data := readFile(t, filepath.Join(dir, "config.yaml"))
	if !strings.Contains(data, "192.168.5.5/32") {
		t.Errorf("-force should allow out-of-pool address:\n%s", data)
	}
}

func TestAcceptancePoolCommand(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")
	out := run(t, dir, "pool")
	for _, want := range []string{"IP pool: 10.0.2.0/24", "servers:", "10.0.2.0/26", "Server", "My Laptop", "My Phone", "free"} {
		if !strings.Contains(out, want) {
			t.Errorf("pool output missing %q:\n%s", want, out)
		}
	}
	assertGolden(t, "pool_output.txt.golden", out)
}

func TestAcceptanceBootstrapPortAndPreshared(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com", "-port", "443", "-preshared", "per-peer")
	data := readFile(t, filepath.Join(dir, "config.yaml"))
	if !strings.Contains(data, "endpoint: vpn.example.com:443") {
		t.Errorf("endpoint should gain port 443:\n%s", data)
	}
	if !strings.Contains(data, "listenPort: 443") {
		t.Errorf("server should have listenPort 443:\n%s", data)
	}
	pskCount := strings.Count(data, "presharedKey:")
	if pskCount != 3 {
		t.Errorf("per-peer mode should write 3 presharedKey entries, got %d:\n%s", pskCount, data)
	}
	if regexp.MustCompile(`(?m)^presharedKey:`).MatchString(data) {
		t.Errorf("per-peer mode must not set a top-level global key:\n%s", data)
	}

	run(t, dir, "generate", "-output", "./out", "-png=false")
	serverConf := readFile(t, filepath.Join(dir, "out", "Server.conf"))
	if !strings.Contains(serverConf, "ListenPort=443") {
		t.Errorf("generated server conf should contain ListenPort=443:\n%s", serverConf)
	}
	if !strings.Contains(serverConf, "PresharedKey=") {
		t.Errorf("generated conf should contain per-peer PresharedKey:\n%s", serverConf)
	}

	dir2 := t.TempDir()
	run(t, dir2, "bootstrap", "-endpoint", "vpn.example.com:51820", "-preshared", "single")
	data2 := readFile(t, filepath.Join(dir2, "config.yaml"))
	if !regexp.MustCompile(`(?m)^presharedKey:`).MatchString(data2) || strings.Contains(data2, "    presharedKey:") {
		t.Errorf("single mode should set exactly one top-level global key:\n%s", data2)
	}
}

func TestAcceptancePskCommand(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")

	out := run(t, dir, "psk", "-mode", "status")
	if !strings.Contains(out, "global preshared key: false") || !strings.Contains(out, "none") {
		t.Errorf("status without psk:\n%s", out)
	}

	run(t, dir, "psk", "-mode", "per-peer")
	data := readFile(t, filepath.Join(dir, "config.yaml"))
	if strings.Count(data, "presharedKey:") != 3 {
		t.Errorf("per-peer should create 3 keys:\n%s", data)
	}
	out = run(t, dir, "psk", "-mode", "status")
	if strings.Count(out, "own per-peer key") != 3 {
		t.Errorf("status should show 3 own keys:\n%s", out)
	}

	run(t, dir, "psk", "-mode", "single")
	data = readFile(t, filepath.Join(dir, "config.yaml"))
	if strings.Count(data, "presharedKey:") != 1 || strings.Contains(data, "    presharedKey:") {
		t.Errorf("single should leave one global key only:\n%s", data)
	}

	run(t, dir, "psk", "-mode", "remove")
	data = readFile(t, filepath.Join(dir, "config.yaml"))
	if strings.Contains(data, "presharedKey") {
		t.Errorf("remove should clear all keys:\n%s", data)
	}
}

func TestAcceptanceAddServerWithPort(t *testing.T) {
	dir := t.TempDir()
	run(t, dir, "bootstrap", "-endpoint", "vpn.example.com:51820")
	run(t, dir, "add", "-server", "-endpoint", "vpn2.example.com:443", "-port", "8443", "-name", "Srv2")
	data := readFile(t, filepath.Join(dir, "config.yaml"))
	if !strings.Contains(data, "listenPort: 8443") {
		t.Errorf("-port should win over endpoint port:\n%s", data)
	}
	if !strings.Contains(data, "endpoint: vpn2.example.com:443") {
		t.Errorf("endpoint should keep its own port:\n%s", data)
	}
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
