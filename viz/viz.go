package viz

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ofcoursedude/wg-manage/models"
)

func Render(cfg *models.Configuration, format string) (string, error) {
	switch format {
	case "ascii", "":
		return renderASCII(cfg), nil
	case "dot":
		return renderDOT(cfg), nil
	case "mermaid":
		return renderMermaid(cfg), nil
	}
	return "", fmt.Errorf("unknown format %q (use ascii, dot or mermaid)", format)
}

func keyFingerprint(pub string) string {
	if len(pub) > 8 {
		return pub[:8]
	}
	return pub
}

func peerLine(p *models.Peer) string {
	addr := ""
	if len(p.Address) > 0 {
		addr = p.Address[0]
	}
	ep := ""
	if p.Endpoint != nil {
		ep = "  " + *p.Endpoint
	}
	return fmt.Sprintf("%s  %s  [%s]%s", p.Name, addr, keyFingerprint(p.PublicKey), ep)
}

func serversAndClients(cfg *models.Configuration) (servers, clients []models.Peer) {
	for _, p := range cfg.Peers {
		if p.IsServer() {
			servers = append(servers, p)
		} else {
			clients = append(clients, p)
		}
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })
	sort.Slice(clients, func(i, j int) bool { return clients[i].Name < clients[j].Name })
	return
}

func overrideCount(cfg *models.Configuration, p *models.Peer) int {
	n := 0
	for _, lines := range p.PeerOverrides {
		n += len(lines)
	}
	return n
}

func renderASCII(cfg *models.Configuration) string {
	var b strings.Builder
	servers, clients := serversAndClients(cfg)

	pool := ""
	if cfg.Settings != nil && cfg.Settings.IPPool != "" {
		pool = fmt.Sprintf("  (pool %s", cfg.Settings.IPPool)
		for name, sub := range cfg.Settings.Subnets {
			pool += fmt.Sprintf(", %s %s", name, sub.CIDR)
		}
		pool += ")"
	}
	fmt.Fprintf(&b, "WireGuard network%s\n\n", pool)

	if len(servers) == 0 {
		b.WriteString("  (no servers - peers with an endpoint)\n")
	}
	for _, s := range servers {
		fmt.Fprintf(&b, "  [SERVER] %s\n", peerLine(&s))
	}
	if len(servers) > 0 {
		b.WriteString("     |\n")
		for range clients {
			b.WriteString("     |\n")
		}
	}
	for _, c := range clients {
		extra := ""
		if n := overrideCount(cfg, &c); n > 0 {
			extra = fmt.Sprintf("  (+%d override lines)", n)
		}
		fmt.Fprintf(&b, "  [client] %s%s\n", peerLine(&c), extra)
	}

	if cfg.Settings != nil && cfg.Settings.IPPool != "" {
		if entries, _, err := cfg.PoolStatus(); err == nil {
			free := 0
			for _, e := range entries {
				if e.Free {
					free++
				}
			}
			fmt.Fprintf(&b, "\n  free addresses: %d of %d\n", free, len(entries))
		}
	}
	return b.String()
}

func sanitizeID(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		}
		return '_'
	}, s)
}

func renderDOT(cfg *models.Configuration) string {
	var b strings.Builder
	servers, clients := serversAndClients(cfg)
	b.WriteString("digraph wg {\n  rankdir=TB;\n  node [shape=box];\n")
	for _, s := range servers {
		fmt.Fprintf(&b, "  %q [label=\"%s\\n%s\", style=filled, fillcolor=lightblue];\n", sanitizeID(s.Name), s.Name, peerLine(&s))
	}
	for _, c := range clients {
		label := peerLine(&c)
		if n := overrideCount(cfg, &c); n > 0 {
			label += fmt.Sprintf("\\n(+%d overrides)", n)
		}
		fmt.Fprintf(&b, "  %q [label=\"%s\"];\n", sanitizeID(c.Name), label)
	}
	for _, c := range clients {
		for _, s := range servers {
			fmt.Fprintf(&b, "  %q -> %q;\n", sanitizeID(c.Name), sanitizeID(s.Name))
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func renderMermaid(cfg *models.Configuration) string {
	var b strings.Builder
	servers, clients := serversAndClients(cfg)
	b.WriteString("graph TD\n")
	for _, s := range servers {
		fmt.Fprintf(&b, "  %s[\"%s<br/>%s\"]:::server\n", sanitizeID(s.Name), s.Name, peerLine(&s))
	}
	for _, c := range clients {
		fmt.Fprintf(&b, "  %s[\"%s<br/>%s\"]\n", sanitizeID(c.Name), c.Name, peerLine(&c))
	}
	for _, c := range clients {
		for _, s := range servers {
			fmt.Fprintf(&b, "  %s --> %s\n", sanitizeID(c.Name), sanitizeID(s.Name))
		}
	}
	b.WriteString("  classDef server fill:#cce5ff,stroke:#333;\n")
	return b.String()
}
