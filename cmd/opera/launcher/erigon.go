package launcher

import (
	cli "gopkg.in/urfave/cli.v1"
)

// ErigonDBIdFlag is used for demo only
var ErigonDBIdFlag = cli.IntFlag{
	Name:  "erigonid",
	Usage: "'n' - sets erigon id to differenciate erigon db path for demo network ",
	Value: 0,
}
