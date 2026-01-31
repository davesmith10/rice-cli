//go:build linux

package convert

import _ "embed"

//go:embed lame-binaries/lame-linux-amd64
var lameBinary []byte
