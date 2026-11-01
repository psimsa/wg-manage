package tui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/wg"
)

type tab int

const (
	tabPeers tab = iota
	tabSettings
)

type screen int

const (
	screenList screen = iota
	screenForm
	screenConfirm
)

type formField struct {
	label string
	value string
}

type Model struct {
	cfg        *models.Configuration
	configFile string

	tab       tab
	screen    screen
	peers     []models.Peer
	cursor    int
	dirty     bool
	statusMsg string

	formTitle string
	form      []formField
	formFocus int
	formPeer  int
	formIsNew bool
	formRole  string

	confirmText   string
	confirmAction func(*Model)

	formOverrides bool

	width  int
	height int
}

func NewModel(cfg *models.Configuration, configFile string) Model {
	m := Model{cfg: cfg, configFile: configFile, tab: tabPeers, screen: screenList}
	if cfg != nil {
		m.peers = append([]models.Peer(nil), cfg.Peers...)
	}
	return m
}

func (m Model) Init() tea.Cmd { return nil }

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	serverStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	clientStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	focusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch m.screen {
	case screenForm:
		if m.formOverrides {
			return m.handleOverridesKey(key)
		}
		return m.handleFormKey(key)
	case screenConfirm:
		switch key {
		case "y", "enter":
			if m.confirmAction != nil {
				m.confirmAction(&m)
			}
			m.screen = screenList
			m.dirty = true
		case "n", "esc", "q":
			m.screen = screenList
		}
		return m, nil
	}

	switch m.tab {
	case tabPeers:
		return m.handlePeersKey(key)
	case tabSettings:
		if key == "left" || key == "right" || key == "tab" {
			m.tab = tabPeers
			return m, nil
		}
	}
	return m, nil
}

func (m Model) handlePeersKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "q", "ctrl+c":
		if m.dirty {
			m.confirmText = "Unsaved changes. Save before quitting? (y=save, n=discard)"
			m.confirmAction = func(mm *Model) { mm.save() }
			m.screen = screenConfirm
			return m, tea.Quit
		}
		return m, tea.Quit
	case "left", "right", "tab":
		m.tab = tabSettings
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.peers)-1 {
			m.cursor++
		}
	case "a":
		m.openAddForm()
	case "e":
		if len(m.peers) > 0 {
			m.openEditForm(m.cursor)
		}
	case "d":
		if len(m.peers) > 0 {
			idx := m.cursor
			name := m.peers[idx].Name
			m.confirmText = fmt.Sprintf("Delete peer %q? Its keys will be lost. (y/n)", name)
			m.confirmAction = func(mm *Model) {
				mm.peers = models.RemovePeerByName(append([]models.Peer(nil), mm.peers...), name)
				if mm.cursor >= len(mm.peers) && mm.cursor > 0 {
					mm.cursor--
				}
				mm.statusMsg = "Deleted " + name
			}
			m.screen = screenConfirm
		}
	case "r":
		if len(m.peers) > 0 && m.peers[m.cursor].IsServer() {
			m.toggleRouting(m.cursor)
		}
	case "o":
		if len(m.peers) > 0 {
			m.openOverridesForm(m.cursor)
		}
	case "s", "ctrl+s":
		m.save()
	case "g":
		m.generate()
	}
	return m, nil
}

func (m *Model) save() {
	m.cfg.Peers = m.peers
	models.SaveYaml(*m.cfg, m.configFile)
	m.dirty = false
	m.statusMsg = "Saved " + m.configFile
}

func (m *Model) generate() {
	m.cfg.Peers = m.peers
	output := m.cfg.ResolveOutputDir(false, "./output")
	png := m.cfg.ResolvePNG(false, true)
	if err := generateConfigs(m.cfg, output, png); err != nil {
		m.statusMsg = "Generate error: " + err.Error()
		return
	}
	m.statusMsg = fmt.Sprintf("Generated configs into %s (png=%v)", output, png)
}

func (m *Model) openAddForm() {
	m.formTitle = "Add peer"
	m.formIsNew = true
	m.formPeer = -1
	m.formRole = "client"
	egress := ""
	if m.cfg.Settings != nil && m.cfg.Settings.ServerDefaults != nil {
		egress = m.cfg.Settings.ServerDefaults.EgressInterface
	}
	m.form = []formField{
		{label: "Name", value: ""},
		{label: "Role (client/server)", value: "client"},
		{label: "Endpoint (servers only)", value: ""},
		{label: "Egress interface (routing)", value: egress},
		{label: "Address (empty = auto from pool)", value: ""},
	}
	m.formFocus = 0
	m.screen = screenForm
}

func (m *Model) openEditForm(idx int) {
	p := m.peers[idx]
	m.formTitle = "Edit peer: " + p.Name
	m.formIsNew = false
	m.formPeer = idx
	m.formRole = "client"
	if p.IsServer() {
		m.formRole = "server"
	}
	m.form = peerFormFields(p)
	m.formFocus = 0
	m.screen = screenForm
}

func peerFormFields(p models.Peer) []formField {
	str := func(sp *string) string {
		if sp == nil {
			return ""
		}
		return *sp
	}
	intStr := func(ip *int) string {
		if ip == nil {
			return ""
		}
		return strconv.Itoa(*ip)
	}
	join := func(s []string) string { return strings.Join(s, ", ") }
	return []formField{
		{label: "Name", value: p.Name},
		{label: "Description", value: str(p.Description)},
		{label: "Role (client/server)", value: map[bool]string{true: "server", false: "client"}[p.IsServer()]},
		{label: "Endpoint", value: str(p.Endpoint)},
		{label: "ListenPort", value: intStr(p.ListenPort)},
		{label: "Address (comma separated)", value: join(p.Address)},
		{label: "AllowedIPs (comma separated)", value: join(p.AllowedIps)},
		{label: "DNS (comma separated)", value: join(p.DNS)},
		{label: "PersistentKeepalive", value: intStr(p.PersistentKeepalive)},
		{label: "MTU", value: str(p.MTU)},
		{label: "Table", value: str(p.Table)},
		{label: "PostUp (semicolon separated)", value: strings.Join(p.PostUp, "; ")},
		{label: "PostDown (semicolon separated)", value: strings.Join(p.PostDown, "; ")},
	}
}

func (m *Model) applyForm() {
	get := func(label string) string {
		for _, f := range m.form {
			if f.label == label {
				return strings.TrimSpace(f.value)
			}
		}
		return ""
	}
	split := func(s string) []string {
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	splitSemi := func(s string) []string {
		if s == "" {
			return nil
		}
		parts := strings.Split(s, ";")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}
	optStr := func(s string) *string {
		if s == "" {
			return nil
		}
		return &s
	}
	optInt := func(s string) *int {
		if s == "" {
			return nil
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}
		return &v
	}

	var p models.Peer
	if m.formIsNew {
		priv, pub := wg.GetKeyPair()
		p = *models.NewPeer(get("Name"), priv, pub)
	} else {
		p = m.peers[m.formPeer]
		p.Name = get("Name")
	}
	p.Description = optStr(get("Description"))
	if m.formIsNew {
		role := strings.ToLower(get("Role (client/server)"))
		p.Endpoint = nil
		if role == "server" || strings.HasPrefix(role, "s") {
			p.Endpoint = optStr(get("Endpoint (servers only)"))
			if p.Endpoint != nil {
				if parts := strings.Split(*p.Endpoint, ":"); len(parts) == 2 {
					if v, err := strconv.Atoi(parts[1]); err == nil {
						p.ListenPort = &v
					}
				}
			}
		}
	} else {
		role := strings.ToLower(get("Role (client/server)"))
		p.Endpoint = nil
		if role == "server" || strings.HasPrefix(role, "s") {
			p.Endpoint = optStr(get("Endpoint"))
			if p.Endpoint != nil {
				if parts := strings.Split(*p.Endpoint, ":"); len(parts) == 2 {
					if v, err := strconv.Atoi(parts[1]); err == nil && p.ListenPort == nil {
						p.ListenPort = &v
					}
				}
			}
		}
		p.ListenPort = optInt(get("ListenPort"))
	}
	p.Address = split(get("Address (comma separated)"))
	p.AllowedIps = split(get("AllowedIPs (comma separated)"))
	p.DNS = split(get("DNS (comma separated)"))
	p.PersistentKeepalive = optInt(get("PersistentKeepalive"))
	p.MTU = optStr(get("MTU"))
	p.Table = optStr(get("Table"))
	p.PostUp = splitSemi(get("PostUp (semicolon separated)"))
	p.PostDown = splitSemi(get("PostDown (semicolon separated)"))

	if m.formIsNew {
		addr := get("Address (empty = auto from pool)")
		if addr != "" {
			if err := m.cfg.ValidateAddressInPool(addr); err != nil {
				m.statusMsg = "Error: " + err.Error()
				m.screen = screenList
				return
			}
			p.Address = []string{addr}
			if len(p.AllowedIps) == 0 {
				p.AllowedIps = []string{addr}
			}
		} else if m.cfg.Settings != nil && m.cfg.Settings.IPPool != "" {
			if err := m.cfg.AllocateForPeer(&p); err != nil {
				m.statusMsg = "Error: " + err.Error()
				m.screen = screenList
				return
			}
		}
		m.cfg.ApplyDefaultsToPeer(&p)
		m.peers = append(m.peers, p)
		m.cursor = len(m.peers) - 1
		m.statusMsg = "Added " + p.Name
	} else {
		m.peers[m.formPeer] = p
		m.statusMsg = "Updated " + p.Name
	}
	m.dirty = true
	m.screen = screenList
}

func (m *Model) handleOverridesKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenList
		m.formOverrides = false
	case "enter":
		if m.formFocus == len(m.form)-1 {
			m.applyOverridesForm()
			m.formOverrides = false
			return m, nil
		}
		m.formFocus++
		return m, nil
	case "shift+tab", "up":
		if m.formFocus > 0 {
			m.formFocus--
		}
	case "tab", "down":
		if m.formFocus < len(m.form)-1 {
			m.formFocus++
		}
	case "backspace":
		f := &m.form[m.formFocus]
		if len(f.value) > 0 {
			f.value = f.value[:len(f.value)-1]
		}
	default:
		if len(key) == 1 {
			m.form[m.formFocus].value += key
		}
	}
	return m, nil
}

func (m *Model) handleFormKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.screen = screenList
	case "enter":
		if m.formFocus == len(m.form)-1 {
			m.applyForm()
		} else {
			m.formFocus++
		}
	case "shift+tab", "up":
		if m.formFocus > 0 {
			m.formFocus--
		}
	case "tab", "down":
		if m.formFocus < len(m.form)-1 {
			m.formFocus++
		}
	case "backspace":
		f := &m.form[m.formFocus]
		if len(f.value) > 0 {
			f.value = f.value[:len(f.value)-1]
		}
	default:
		if len(key) == 1 {
			m.form[m.formFocus].value += key
		}
	}
	return m, nil
}

func routingLines(egress string, up bool) []string {
	sign := "-A"
	verb := "add"
	if !up {
		sign = "-D"
		verb = "del"
	}
	_ = verb
	lines := []string{
		fmt.Sprintf("sysctl -q -w net.ipv4.ip_forward=%d", boolInt(up)),
		fmt.Sprintf("iptables %s FORWARD -i wg0 -j ACCEPT", sign),
		fmt.Sprintf("iptables %s FORWARD -o wg0 -j ACCEPT", sign),
		fmt.Sprintf("iptables -t nat %s POSTROUTING -o wg0 -j MASQUERADE", sign),
	}
	if egress != "" {
		lines = append(lines, fmt.Sprintf("iptables -t nat %s POSTROUTING -o %s -j MASQUERADE", sign, egress))
	}
	return lines
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func hasRouting(p models.Peer) bool {
	for _, l := range p.PostUp {
		if strings.Contains(l, "ip_forward") {
			return true
		}
	}
	return false
}

func (m *Model) toggleRouting(idx int) {
	p := &m.peers[idx]
	if hasRouting(*p) {
		p.PostUp = nil
		p.PostDown = nil
		m.statusMsg = "Routing removed from " + p.Name
	} else {
		egress := ""
		if m.cfg.Settings != nil && m.cfg.Settings.ServerDefaults != nil {
			egress = m.cfg.Settings.ServerDefaults.EgressInterface
		}
		p.PostUp = routingLines(egress, true)
		p.PostDown = routingLines(egress, false)
		m.statusMsg = "Routing/NAT enabled on " + p.Name + " (egress: " + egress + ")"
	}
	m.dirty = true
}

func (m *Model) openOverridesForm(idx int) {
	p := &m.peers[idx]
	m.formTitle = "Overrides for " + p.Name + " (extra [Peer] lines per target public key)"
	m.formIsNew = false
	m.formPeer = idx
	m.form = []formField{{label: "Target peer (name)", value: ""}, {label: "Extra lines (semicolon separated)", value: ""}}
	m.formFocus = 0
	m.screen = screenForm
	m.formOverrides = true
}

func (m *Model) applyOverridesForm() {
	if m.formPeer < 0 || m.formPeer >= len(m.peers) {
		m.screen = screenList
		return
	}
	targetName := strings.TrimSpace(m.form[0].value)
	linesRaw := strings.TrimSpace(m.form[1].value)
	var lines []string
	if linesRaw != "" {
		for _, l := range strings.Split(linesRaw, ";") {
			if t := strings.TrimSpace(l); t != "" {
				lines = append(lines, t)
			}
		}
	}
	p := &m.peers[m.formPeer]
	for i := range m.peers {
		if m.peers[i].Name == targetName && i != m.formPeer {
			if p.PeerOverrides == nil {
				p.PeerOverrides = map[string][]string{}
			}
			if len(lines) == 0 {
				delete(p.PeerOverrides, m.peers[i].PublicKey)
			} else {
				p.PeerOverrides[m.peers[i].PublicKey] = lines
			}
			m.statusMsg = fmt.Sprintf("Overrides for %s -> %s updated", p.Name, targetName)
			m.dirty = true
			m.screen = screenList
			return
		}
	}
	m.statusMsg = "No peer named " + targetName
	m.screen = screenList
}

func (m Model) View() string {
	var b strings.Builder
	tabs := fmt.Sprintf("[%s] [%s]", tabName(m.tab == tabPeers), tabName(m.tab == tabSettings))
	b.WriteString(titleStyle.Render("wg-manage — "+m.configFile) + "  " + tabs + "\n\n")
	switch m.screen {
	case screenForm:
		b.WriteString(m.formView())
	case screenConfirm:
		b.WriteString(m.confirmText + "\n")
	default:
		switch m.tab {
		case tabPeers:
			b.WriteString(m.peersView())
		case tabSettings:
			b.WriteString(m.settingsView())
		}
	}
	b.WriteString("\n" + helpStyle.Render(m.helpLine()))
	if m.statusMsg != "" {
		b.WriteString("\n" + m.statusMsg)
	}
	return b.String()
}

func tabName(active bool) string {
	if active {
		return focusStyle.Render("Peers/Settings")
	}
	return ""
}

func (m Model) peersView() string {
	var b strings.Builder
	if len(m.peers) == 0 {
		b.WriteString("No peers yet. Press 'a' to add one.\n")
	}
	for i, p := range m.peers {
		addr := ""
		if len(p.Address) > 0 {
			addr = p.Address[0]
		}
		ep := ""
		if p.Endpoint != nil {
			ep = *p.Endpoint
		}
		role := clientStyle.Render("client")
		if p.IsServer() {
			role = serverStyle.Render("server")
		}
		line := fmt.Sprintf("%-20s %-18s %-9s %s", p.Name, addr, role, ep)
		if i == m.cursor {
			line = focusStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}

func (m Model) settingsView() string {
	s := m.cfg.Settings
	if s == nil {
		return "No settings section. Saving will keep it absent; run 'init -settings' or add one manually.\n"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "ipPool: %s\n", s.IPPool)
	for name, sub := range s.Subnets {
		fmt.Fprintf(&b, "subnet %s: %s\n", name, sub.CIDR)
	}
	if s.ServerDefaults != nil {
		fmt.Fprintf(&b, "serverDefaults: listenPort=%s egressInterface=%s\n", intPtrStr(s.ServerDefaults.ListenPort), s.ServerDefaults.EgressInterface)
	}
	if s.ClientDefaults != nil && s.ClientDefaults.PersistentKeepalive != nil {
		fmt.Fprintf(&b, "clientDefaults: persistentKeepalive=%d\n", *s.ClientDefaults.PersistentKeepalive)
	}
	if s.Generate != nil {
		fmt.Fprintf(&b, "generate: outputDir=%s png=%s\n", s.Generate.OutputDir, boolPtrStr(s.Generate.PNG))
	}
	return b.String()
}

func intPtrStr(p *int) string {
	if p == nil {
		return "-"
	}
	return strconv.Itoa(*p)
}

func boolPtrStr(p *bool) string {
	if p == nil {
		return "-"
	}
	return strconv.FormatBool(*p)
}

func (m Model) formView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.formTitle) + "\n\n")
	for i, f := range m.form {
		marker := "  "
		if i == m.formFocus {
			marker = "> "
		}
		fmt.Fprintf(&b, "%s%-38s %s\n", marker, f.label+":", f.value)
	}
	if m.formOverrides {
		b.WriteString("\n(enter on last field applies overrides)\n")
	} else {
		b.WriteString("\n(enter on last field applies; esc cancels)\n")
	}
	return b.String()
}

func (m Model) helpLine() string {
	if m.screen == screenForm {
		return "tab/down: next field  shift+tab/up: prev  enter: next/apply  esc: cancel"
	}
	if m.screen == screenConfirm {
		return "y/enter: confirm  n/esc: cancel"
	}
	if m.tab == tabSettings {
		return "tab/left/right: back to peers  q: quit"
	}
	return "a: add  e: edit  d: delete  r: routing on/off (servers)  o: overrides  g: generate  s: save  tab: settings  q: quit"
}

func Run(configFile string) error {
	cfg := models.LoadYaml(configFile)
	m := NewModel(&cfg, configFile)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
