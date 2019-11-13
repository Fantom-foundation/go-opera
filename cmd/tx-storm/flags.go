package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"
)

var DonorFlag = cli.IntFlag{
	Name:  "donor",
	Usage: "fake account number to take balance from",
}

func getDonor(ctx *cli.Context) uint {
	return uint(ctx.GlobalInt(DonorFlag.Name))
}

var TxnsRateFlag = cli.IntFlag{
	Name:  "rate",
	Usage: "transactions per second (max sum of all generators)",
}

func getTxnsRate(ctx *cli.Context) uint {
	return uint(ctx.GlobalInt(TxnsRateFlag.Name))
}

var BlockPeriodFlag = cli.IntFlag{
	Name:  "period",
	Usage: "seconds beetwen blocks (estimation)",
}

func getBlockPeriod(ctx *cli.Context) time.Duration {
	return time.Second * time.Duration(ctx.GlobalInt(BlockPeriodFlag.Name))
}

var NumberFlag = cli.StringFlag{
	Name:  "num",
	Usage: "'N/X' - it is a N-th generator of X",
}

func getNumber(ctx *cli.Context) (num, total uint) {
	var err error
	num, total, err = parseNumber(ctx.GlobalString(NumberFlag.Name))
	if err != nil {
		panic(err)
	}
	return
}

func parseNumber(s string) (num, total uint, err error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("use %%d/%%d format")
		return
	}

	num64, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return
	}
	num = uint(num64)

	total64, err := strconv.ParseUint(parts[1], 10, 64)
	total = uint(total64)

	if num64 < 1 || num64 > total64 {
		err = fmt.Errorf("key-num should be in range from 1 to total : <key-num>/<total>")
	}

	return
}
