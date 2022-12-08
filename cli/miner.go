package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/go-address"
)

var minerSubCommand = &cli.Command{
	Name:  "miner",
	Usage: "miner command",
	Subcommands: []*cli.Command{
		minerHasCommand,
	},
}

var minerHasCommand = &cli.Command{
	Name:      "has",
	Usage:     "Check if miner exists",
	ArgsUsage: "<miner>",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() != 1 {
			cli.ShowSubcommandHelpAndExit(ctx, 1)
			return nil
		}

		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		addr, err := address.NewFromString(ctx.Args().Get(0))
		if err != nil {
			return err
		}

		exist, err := client.HasMiner(ctx.Context, addr)
		if err != nil {
			return err
		}
		fmt.Println(exist)
		return nil
	},
}
