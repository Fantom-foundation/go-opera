package main

import (
	"gopkg.in/urfave/cli.v1"
)

var FromFlag = cli.IntFlag{
	Name:  "from",
	Usage: "start account",
	Value: 1,
}

var CountFlag = cli.IntFlag{
	Name:  "count",
	Usage: "accounts count",
	Value: 1,
}

func getFrom(ctx *cli.Context) uint {
	n := ctx.GlobalInt(FromFlag.Name)
	if n < 0 {
		panic("Count should be positive")
	}
	return uint(n)
}

func getCount(ctx *cli.Context) uint {
	n := ctx.GlobalInt(CountFlag.Name)
	if n < 0 {
		panic("Count should be positive")
	}
	return uint(n)
}
