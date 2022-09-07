package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/auth"
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
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "really-do-it",
			Usage: "must be specified for the action to take effect",
		},
	},
	Action: func(ctx *cli.Context) error {
		really := ctx.Bool("really-do-it")
		if !really {
			//nolint:golint
			return fmt.Errorf("--really-do-it must be specified for this action to have an effect; you have been warned")
		}

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

		exist, err := client.HasMiner(&auth.HasMinerRequest{Miner: addr.String()})
		if err != nil {
			return err
		}
		fmt.Println(exist)
		return nil
	},
}
