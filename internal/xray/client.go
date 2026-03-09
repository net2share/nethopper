package xray

import (
	"bytes"
	"text/template"

	"github.com/net2share/nethopper/internal/config"
)

const clientTemplate = `{
  "log": {"loglevel": "warning"},
  "dns": {
    "servers": ["8.8.8.8", "1.1.1.1"],
    "queryStrategy": "UseIPv4"
  },
  "reverse": {
    "bridges": [{"tag": "bridge", "domain": "nethopper.internal"}]
  },
  "outbounds": [
    {
      "tag": "free-internet",
      "protocol": "freedom",
      "settings": {"domainStrategy": "UseIPv4"},
      "streamSettings": {
        "sockopt": {"interface": "{{.FreeInterface}}"}
      }
    },
    {
      "tag": "tunnel-out",
      "protocol": "vless",
      "settings": {
        "vnext": [{
          "address": "{{.ServerIP}}",
          "port": {{.TunnelPort}},
          "users": [{"id": "{{.UUID}}", "encryption": "none"}]
        }]
      },
      "streamSettings": {
        "sockopt": {"interface": "{{.RestrictedInterface}}"}
      }
    }
  ],
  "routing": {
    "domainStrategy": "UseIPv4",
    "rules": [
      {"inboundTag": ["bridge"], "domain": ["full:nethopper.internal"], "outboundTag": "tunnel-out"},
      {"inboundTag": ["bridge"], "outboundTag": "free-internet"}
    ]
  }
}`

// ClientTemplateData holds the data for client xray config generation.
type ClientTemplateData struct {
	FreeInterface       string
	RestrictedInterface string
	ServerIP            string
	TunnelPort          int
	UUID                string
}

// GenerateClientConfig generates the Xray JSON config for the client (bridge).
func GenerateClientConfig(cfg *config.ClientConfig) ([]byte, error) {
	data := ClientTemplateData{
		FreeInterface:       cfg.FreeInterface,
		RestrictedInterface: cfg.RestrictedInterface,
		ServerIP:            cfg.ServerIP,
		TunnelPort:          cfg.TunnelPort,
		UUID:                cfg.UUID,
	}

	tmpl, err := template.New("client").Parse(clientTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
