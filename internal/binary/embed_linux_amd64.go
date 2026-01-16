//go:build linux && amd64

package binary

import _ "embed"

//go:embed embedded/sing-box-linux-amd64
var embeddedBinary []byte

var embeddedBinaryName = "sing-box-linux-amd64"
