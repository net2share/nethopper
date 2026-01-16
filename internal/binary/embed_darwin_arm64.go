//go:build darwin && arm64

package binary

import _ "embed"

//go:embed embedded/sing-box-darwin-arm64
var embeddedBinary []byte

var embeddedBinaryName = "sing-box-darwin-arm64"
