package xui

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Inbound represents an x-ui inbound as returned by the API.
type Inbound struct {
	ID       int    `json:"id"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
	Tag      string `json:"tag"`
	Remark   string `json:"remark"`
	Enable   bool   `json:"enable"`
}

// AddSocksInbound adds a SOCKS5 inbound via x-ui API.
// Returns the inbound ID and auto-generated tag.
func (c *Client) AddSocksInbound(port int) (int, string, error) {
	settings := `{"auth":"noauth","accounts":[],"udp":true,"ip":"0.0.0.0"}`
	streamSettings := `{"network":"tcp","security":"none","tcpSettings":{"header":{"type":"none"}}}`
	sniffing := `{"enabled":true,"destOverride":["http","tls"]}`

	return c.addInbound("NH-SOCKS5", port, "socks", settings, streamSettings, sniffing)
}

// AddVLESSInbound adds a VLESS tunnel inbound via x-ui API.
// Returns the inbound ID and auto-generated tag.
func (c *Client) AddVLESSInbound(port int, uuid string) (int, string, error) {
	settings := fmt.Sprintf(`{"clients":[{"id":"%s","flow":"","email":"nethopper","limitIp":0,"totalGB":0,"expiryTime":0,"enable":true}],"decryption":"none","fallbacks":[]}`, uuid)
	streamSettings := `{"network":"tcp","security":"none","tcpSettings":{"header":{"type":"none"}}}`
	sniffing := `{"enabled":true,"destOverride":["http","tls"]}`

	return c.addInbound("NH-Tunnel", port, "vless", settings, streamSettings, sniffing)
}

// DeleteInbound deletes an inbound by its database ID.
func (c *Client) DeleteInbound(id int) error {
	resp, err := c.post(fmt.Sprintf("panel/api/inbounds/del/%d", id))
	if err != nil {
		return fmt.Errorf("failed to delete inbound %d: %w", id, err)
	}
	if !resp.Success {
		return fmt.Errorf("failed to delete inbound %d: %s", id, resp.Msg)
	}
	return nil
}

func (c *Client) addInbound(remark string, port int, protocol, settings, streamSettings, sniffing string) (int, string, error) {
	data := url.Values{
		"up":             {"0"},
		"down":           {"0"},
		"total":          {"0"},
		"remark":         {remark},
		"enable":         {"true"},
		"expiryTime":     {"0"},
		"listen":         {""},
		"port":           {fmt.Sprintf("%d", port)},
		"protocol":       {protocol},
		"settings":       {settings},
		"streamSettings": {streamSettings},
		"sniffing":       {sniffing},
	}

	resp, err := c.postForm("panel/api/inbounds/add", data)
	if err != nil {
		return 0, "", fmt.Errorf("failed to add %s inbound: %w", protocol, err)
	}
	if !resp.Success {
		return 0, "", fmt.Errorf("failed to add %s inbound: %s", protocol, resp.Msg)
	}

	var inbound Inbound
	if err := json.Unmarshal(resp.Obj, &inbound); err != nil {
		return 0, "", fmt.Errorf("failed to parse inbound response: %w", err)
	}

	return inbound.ID, inbound.Tag, nil
}
