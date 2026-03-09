package xray

import (
	"bytes"
	"text/template"

	"github.com/net2share/nethopper/internal/config"
)

const serverTemplate = `{
  "log": {"loglevel": "warning"},
  "dns": {
    "servers": ["8.8.8.8", "1.1.1.1"],
    "queryStrategy": "UseIPv4"
  },
  "reverse": {
    "portals": [{"tag": "portal", "domain": "nethopper.internal"}]
  },
  "inbounds": [
    {
      "tag": "socks-in",
      "port": {{.SocksPort}},
      "listen": "0.0.0.0",
      "protocol": "socks",
      "settings": {"auth": "noauth"}
    },
    {
      "tag": "tunnel-in",
      "port": {{.TunnelPort}},
      "listen": "0.0.0.0",
      "protocol": "vless",
      "settings": {
        "clients": [{"id": "{{.UUID}}"}],
        "decryption": "none"
      }
    }
  ],
  "routing": {
    "domainStrategy": "UseIPv4",
    "rules": [
      {"inboundTag": ["socks-in"], "outboundTag": "portal"},
      {"inboundTag": ["tunnel-in"], "outboundTag": "portal"}
    ]
  }
}`

// GenerateServerConfig generates the Xray JSON config for the server (portal).
func GenerateServerConfig(cfg *config.ServerConfig) ([]byte, error) {
	tmpl, err := template.New("server").Parse(serverTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
