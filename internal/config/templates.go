package config

import (
	"bytes"
	"encoding/json"
	"text/template"
)

// Server config template based on sing-box config structure
const serverConfigTemplate = `{
  "log": {
    {{if eq .LogLevel "disabled"}}"disabled": true{{else}}"level": "{{.LogLevel}}"{{end}}
  },
  "endpoints": [
    {
      "type": "reverse",
      "tag": "reverse-ep",
      "listen": "::",
      "listen_port": {{.ListenPort}}
    }
  ],
  "inbounds": [
    {
      "type": "mtproto",
      "tag": "mtproto",
      "listen": "0.0.0.0",
      "listen_port": {{.MTProtoPort}},
      "users": [
        {
          "name": "user1",
          "secret": "{{.MTProtoSecret}}"
        }
      ],
      "multiplex_per_connection": {{.MultiplexPerConnection}},
      "fallback_host": "{{.FallbackHost}}",
      "detour": "freenet"
    }
  ],
  "outbounds": [
    {
      "type": "vmess",
      "tag": "freenet",
      "uuid": "{{.VMESSUuid}}",
      "server": "127.0.0.1",
      "server_port": {{.VMESSPort}},
      "detour": "reverse-ep"
    },
    {
      "type": "block",
      "tag": "block"
    }
  ]
}`

// Freenet config template based on sing-box config structure
const freenetConfigTemplate = `{
  "log": {
    {{if eq .LogLevel "disabled"}}"disabled": true{{else}}"level": "{{.LogLevel}}"{{end}}
  },
  "endpoints": [
    {
      "type": "reverse",
      "tag": "reverse-ep",
      "server": "{{.ServerIP}}",
      "server_port": {{.ServerPort}},
      "detour": "filternet",
      "multiplex": {
        "max_connections": {{.MaxConnections}}
      }
    }
  ],
  "inbounds": [
    {
      "type": "vmess",
      "tag": "vmess-in",
      "listen": "::",
      "listen_port": {{.VMESSPort}},
      "users": [
        {
          "uuid": "{{.VMESSUuid}}"
        }
      ]
    }
  ],
  "outbounds": [
    {
      "type": "direct",
      "bind_interface": "{{.FreeInterface}}",
      "tag": "free"
    },
    {
      "type": "direct",
      "bind_interface": "{{.RestrictedInterface}}",
      "tag": "filternet"
    }
  ]
}`

// GenerateServerConfig generates a sing-box configuration for server mode
func GenerateServerConfig(config *ServerConfig) ([]byte, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	tmpl, err := template.New("server").Parse(serverConfigTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return nil, err
	}

	// Validate JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonCheck); err != nil {
		return nil, err
	}

	// Pretty print JSON
	return json.MarshalIndent(jsonCheck, "", "  ")
}

// GenerateFreenetConfig generates a sing-box configuration for freenet mode
func GenerateFreenetConfig(config *FreenetConfig) ([]byte, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	tmpl, err := template.New("freenet").Parse(freenetConfigTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, config); err != nil {
		return nil, err
	}

	// Validate JSON
	var jsonCheck map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonCheck); err != nil {
		return nil, err
	}

	// Pretty print JSON
	return json.MarshalIndent(jsonCheck, "", "  ")
}
