//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/ofcoursedude/wg-manage/models"
	"github.com/ofcoursedude/wg-manage/viz"
	"github.com/ofcoursedude/wg-manage/wg"
)

func result(val interface{}, err error) map[string]interface{} {
	if err != nil {
		return map[string]interface{}{"ok": false, "error": err.Error()}
	}
	return map[string]interface{}{"ok": true, "result": val}
}

func main() {
	registerAPI()
	fmt.Println("wg-manage wasm API ready")
	select {}
}

func registerAPI() {
	api := map[string]interface{}{
		"parseYaml": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return result(nil, fmt.Errorf("usage: parseYaml(yamlString)"))
			}
			cfg, err := models.ParseYaml([]byte(args[0].String()))
			if err != nil {
				return result(nil, err)
			}
			data, err := cfg.ToYaml()
			if err != nil {
				return result(nil, err)
			}
			return result(string(data), nil)
		}),
		"peerNames": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return result(nil, fmt.Errorf("usage: peerNames(yamlString)"))
			}
			cfg, err := models.ParseYaml([]byte(args[0].String()))
			if err != nil {
				return result(nil, err)
			}
			names := make([]interface{}, len(cfg.Peers))
			for i, p := range cfg.Peers {
				names[i] = map[string]interface{}{
					"name":    p.Name,
					"address": firstOrEmpty(p.Address),
					"server":  p.IsServer(),
					"endpoint": func() string {
						if p.Endpoint != nil {
							return *p.Endpoint
						}
						return ""
					}(),
				}
			}
			return result(names, nil)
		}),
		"generateConf": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 2 {
				return result(nil, fmt.Errorf("usage: generateConf(yamlString, peerName)"))
			}
			cfg, err := models.ParseYaml([]byte(args[0].String()))
			if err != nil {
				return result(nil, err)
			}
			conf, err := cfg.RenderPeerConf(args[1].String())
			if err != nil {
				return result(nil, err)
			}
			return result(conf, nil)
		}),
		"generateAll": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return result(nil, fmt.Errorf("usage: generateAll(yamlString)"))
			}
			cfg, err := models.ParseYaml([]byte(args[0].String()))
			if err != nil {
				return result(nil, err)
			}
			all := cfg.RenderAll()
			out := map[string]interface{}{}
			for k, v := range all {
				out[k] = v
			}
			return result(out, nil)
		}),
		"addPeer": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 3 {
				return result(nil, fmt.Errorf("usage: addPeer(yamlString, name, role)  role: client|server; optional args[3]=endpoint"))
			}
			cfg, err := models.ParseYaml([]byte(args[0].String()))
			if err != nil {
				return result(nil, err)
			}
			name, role := args[1].String(), args[2].String()
			priv, pub, err := wg.KeyPair()
			if err != nil {
				return result(nil, err)
			}
			peer := models.NewPeer(name, priv, pub)
			if role == "server" {
				if len(args) < 4 {
					return result(nil, fmt.Errorf("server role requires endpoint"))
				}
				ep := args[3].String()
				peer.Endpoint = &ep
			}
			if cfg.Settings != nil && cfg.Settings.IPPool != "" {
				if err := cfg.AllocateForPeer(peer); err != nil {
					return result(nil, err)
				}
			}
			cfg.ApplyDefaultsToPeer(peer)
			cfg.Peers = append(cfg.Peers, *peer)
			data, err := cfg.ToYaml()
			if err != nil {
				return result(nil, err)
			}
			return result(string(data), nil)
		}),
		"removePeer": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 2 {
				return result(nil, fmt.Errorf("usage: removePeer(yamlString, name)"))
			}
			cfg, err := models.ParseYaml([]byte(args[0].String()))
			if err != nil {
				return result(nil, err)
			}
			before := len(cfg.Peers)
			cfg.Peers = models.RemovePeerByName(cfg.Peers, args[1].String())
			if len(cfg.Peers) == before {
				return result(nil, fmt.Errorf("no peer named %q", args[1].String()))
			}
			data, err := cfg.ToYaml()
			if err != nil {
				return result(nil, err)
			}
			return result(string(data), nil)
		}),
		"bootstrap": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return result(nil, fmt.Errorf("usage: bootstrap(endpoint)"))
			}
			endpoint := args[0].String()
			cfg := models.Configuration{Settings: models.DefaultSettings()}
			cfg.Peers = make([]models.Peer, 3)
			priv, pub, err := wg.KeyPair()
			if err != nil {
				return result(nil, err)
			}
			server := models.NewPeer("Server", priv, pub)
			server.Endpoint = &endpoint
			if err := cfg.AllocateForPeer(server); err != nil {
				return result(nil, err)
			}
			server.AllowedIps = []string{"0.0.0.0/0"}
			egress := ""
			if cfg.Settings != nil && cfg.Settings.ServerDefaults != nil {
				egress = cfg.Settings.ServerDefaults.EgressInterface
			}
			if egress == "" {
				egress = "eth0"
			}
			server.PostUp = []string{
				"sysctl -q -w net.ipv4.ip_forward=1",
				"iptables -A FORWARD -i wg0 -j ACCEPT",
				"iptables -A FORWARD -o wg0 -j ACCEPT",
				"iptables -t nat -A POSTROUTING -o wg0 -j MASQUERADE",
				fmt.Sprintf("iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", egress),
			}
			server.PostDown = []string{
				"sysctl -q -w net.ipv4.ip_forward=0",
				"iptables -D FORWARD -i wg0 -j ACCEPT",
				"iptables -D FORWARD -o wg0 -j ACCEPT",
				"iptables -t nat -D POSTROUTING -o wg0 -j MASQUERADE",
				fmt.Sprintf("iptables -t nat -D POSTROUTING -o %s -j MASQUERADE", egress),
			}
			cfg.ApplyDefaultsToPeer(server)
			server.AllowedIps = []string{"0.0.0.0/0"}
			cfg.Peers[0] = *server
			for i, name := range []string{"My Laptop", "My Phone"} {
				priv, pub, err := wg.KeyPair()
				if err != nil {
					return result(nil, err)
				}
				p := models.NewPeer(name, priv, pub)
				if err := cfg.AllocateForPeer(p); err != nil {
					return result(nil, err)
				}
				cfg.ApplyDefaultsToPeer(p)
				cfg.Peers[i+1] = *p
			}
			data, err := cfg.ToYaml()
			if err != nil {
				return result(nil, err)
			}
			return result(string(data), nil)
		}),
		"viz": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			if len(args) < 1 {
				return result(nil, fmt.Errorf("usage: viz(yamlString, format?)"))
			}
			cfg, err := models.ParseYaml([]byte(args[0].String()))
			if err != nil {
				return result(nil, err)
			}
			format := "ascii"
			if len(args) > 1 {
				format = args[1].String()
			}
			out, err := viz.Render(&cfg, format)
			if err != nil {
				return result(nil, err)
			}
			return result(out, nil)
		}),
		"version": js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			return result("wg-manage wasm dev", nil)
		}),
	}
	js.Global().Set("wgManage", api)
}

func firstOrEmpty(s []string) string {
	if len(s) > 0 {
		return s[0]
	}
	return ""
}
