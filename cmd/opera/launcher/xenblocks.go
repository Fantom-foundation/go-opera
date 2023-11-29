package launcher

import (
	"fmt"
	cli "gopkg.in/urfave/cli.v1"
	"strings"
)

var XenBlocksEndpointFlag = cli.StringFlag{
	Name:  "xenblocks-endpoint",
	Usage: "Sets the Xenblocks websocket endpoint.",
}

func parseXenBlocksEndpoint(s string) (url string, err error) {
	if !strings.HasPrefix(s, "ws://") && !strings.HasPrefix(s, "wss://") {
		err = fmt.Errorf("use ws:// or wss:// prefix")
	}
	return s, err
}
