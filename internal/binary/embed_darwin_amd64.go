//go:build darwin && amd64

package binary

import _ "embed"

//go:embed embedded/sing-box-darwin-amd64
var embeddedBinary []byte

var embeddedBinaryName = "sing-box-darwin-amd64"
