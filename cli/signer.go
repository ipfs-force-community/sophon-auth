package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/go-address"
)

var signerSubCommand = &cli.Command{
	Name:  "signer",
	Usage: "signer command",
	Subcommands: []*cli.Command{
		signerHasCommand,
		signerDelCommand,
	},
}

var signerHasCommand = &cli.Command{
	Name:      "has",
	Usage:     "Check if signer exists",
	ArgsUsage: "<signer>",
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

		exist, err := client.HasSigner(ctx.Context, addr)
		if err != nil {
			return err
		}
		fmt.Println(exist)
		return nil
	},
}

var signerDelCommand = &cli.Command{
	Name:      "del",
	Usage:     "Delete signer of all user",
	ArgsUsage: "<signer>",
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

		_, err = client.DelSigner(ctx.Context, addr.String())
		if err != nil {
			return err
		}
		fmt.Println("delete success")
		return nil
	},
}
