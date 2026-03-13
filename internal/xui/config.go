package xui

import (
	"encoding/json"
	"fmt"
	"net/url"
)

const (
	NHPortalTag    = "nh-portal"
	NHPortalDomain = "nethopper.internal"
)

// GetXrayConfig fetches the current xray template config from x-ui.
func (c *Client) GetXrayConfig() (map[string]interface{}, error) {
	resp, err := c.post("panel/xray/")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch xray config: %w", err)
	}
	if !resp.Success {
		return nil, fmt.Errorf("failed to fetch xray config: %s", resp.Msg)
	}

	// Response obj is a JSON string (double-encoded)
	var objStr string
	if err := json.Unmarshal(resp.Obj, &objStr); err != nil {
		return nil, fmt.Errorf("failed to parse xray config response: %w", err)
	}

	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal([]byte(objStr), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse xray config wrapper: %w", err)
	}

	xraySettingRaw, ok := wrapper["xraySetting"]
	if !ok {
		return nil, fmt.Errorf("xraySetting not found in response")
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(xraySettingRaw, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse xray config: %w", err)
	}

	return cfg, nil
}

// SaveXrayConfig saves a modified xray template config to x-ui.
func (c *Client) SaveXrayConfig(cfg map[string]interface{}) error {
	cfgJSON, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal xray config: %w", err)
	}

	data := url.Values{
		"xraySetting": {string(cfgJSON)},
	}

	resp, err := c.postForm("panel/xray/update", data)
	if err != nil {
		return fmt.Errorf("failed to save xray config: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("failed to save xray config: %s", resp.Msg)
	}
	return nil
}

// RestartXray restarts the xray service through x-ui.
func (c *Client) RestartXray() error {
	resp, err := c.post("panel/api/server/restartXrayService")
	if err != nil {
		return fmt.Errorf("failed to restart xray: %w", err)
	}
	if !resp.Success {
		return fmt.Errorf("failed to restart xray: %s", resp.Msg)
	}
	return nil
}

// AddPortalAndRouting adds the nh-portal entry and routing rules for the given inbound tags.
func AddPortalAndRouting(cfg map[string]interface{}, socksTag, tunnelTag string) {
	// Add reverse.portals
	reverse, _ := cfg["reverse"].(map[string]interface{})
	if reverse == nil {
		reverse = map[string]interface{}{}
		cfg["reverse"] = reverse
	}

	portals, _ := reverse["portals"].([]interface{})
	portals = append(portals, map[string]interface{}{
		"tag":    NHPortalTag,
		"domain": NHPortalDomain,
	})
	reverse["portals"] = portals

	// Add routing rules
	routing, _ := cfg["routing"].(map[string]interface{})
	if routing == nil {
		routing = map[string]interface{}{}
		cfg["routing"] = routing
	}

	rules, _ := routing["rules"].([]interface{})

	// Rule: socks inbound -> portal
	rules = append(rules, map[string]interface{}{
		"type":        "field",
		"inboundTag":  []interface{}{socksTag},
		"outboundTag": NHPortalTag,
	})

	// Rule: tunnel inbound -> portal
	rules = append(rules, map[string]interface{}{
		"type":        "field",
		"inboundTag":  []interface{}{tunnelTag},
		"outboundTag": NHPortalTag,
	})

	routing["rules"] = rules
}

// RemovePortalAndRouting removes the nh-portal entry and its associated routing rules.
func RemovePortalAndRouting(cfg map[string]interface{}) {
	// Remove portal
	if reverse, ok := cfg["reverse"].(map[string]interface{}); ok {
		if portals, ok := reverse["portals"].([]interface{}); ok {
			var filtered []interface{}
			for _, p := range portals {
				if pm, ok := p.(map[string]interface{}); ok {
					if pm["tag"] == NHPortalTag {
						continue
					}
				}
				filtered = append(filtered, p)
			}
			reverse["portals"] = filtered
		}
	}

	// Remove routing rules that reference nh-portal
	if routing, ok := cfg["routing"].(map[string]interface{}); ok {
		if rules, ok := routing["rules"].([]interface{}); ok {
			var filtered []interface{}
			for _, r := range rules {
				if rm, ok := r.(map[string]interface{}); ok {
					if rm["outboundTag"] == NHPortalTag {
						continue
					}
				}
				filtered = append(filtered, r)
			}
			routing["rules"] = filtered
		}
	}
}

// UpdateRoutingTags updates routing rules: removes rules with old tags and adds new ones.
func UpdateRoutingTags(cfg map[string]interface{}, oldSocksTag, oldTunnelTag, newSocksTag, newTunnelTag string) {
	// Remove old routing rules for nh-portal
	RemovePortalAndRouting(cfg)

	// Re-add portal (it was removed above)
	reverse, _ := cfg["reverse"].(map[string]interface{})
	if reverse == nil {
		reverse = map[string]interface{}{}
		cfg["reverse"] = reverse
	}
	portals, _ := reverse["portals"].([]interface{})
	portals = append(portals, map[string]interface{}{
		"tag":    NHPortalTag,
		"domain": NHPortalDomain,
	})
	reverse["portals"] = portals

	// Add new routing rules
	routing, _ := cfg["routing"].(map[string]interface{})
	if routing == nil {
		routing = map[string]interface{}{}
		cfg["routing"] = routing
	}
	rules, _ := routing["rules"].([]interface{})
	rules = append(rules, map[string]interface{}{
		"type":        "field",
		"inboundTag":  []interface{}{newSocksTag},
		"outboundTag": NHPortalTag,
	})
	rules = append(rules, map[string]interface{}{
		"type":        "field",
		"inboundTag":  []interface{}{newTunnelTag},
		"outboundTag": NHPortalTag,
	})
	routing["rules"] = rules
}
