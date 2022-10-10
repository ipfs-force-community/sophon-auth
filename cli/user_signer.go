package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/filecoin-project/go-address"

	"github.com/filecoin-project/venus-auth/auth"
)

var signerSubCmds = &cli.Command{
	Name:  "signer",
	Usage: "Sub commands for managing user signed accounts",
	Subcommands: []*cli.Command{
		signerRegisterCmd,
		signerExistCmd,
		signerListCmd,
		signerUnregisterCmd,
	},
}

var signerRegisterCmd = &cli.Command{
	Name:      "register",
	Usage:     "Add signer address for specified user",
	ArgsUsage: "<user> <signer address>",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() != 2 {
			cli.ShowSubcommandHelpAndExit(ctx, 1)
			return nil
		}
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		user, addr := ctx.Args().Get(0), ctx.Args().Get(1)
		if err = client.RegisterSigners(user, []string{addr}); err != nil {
			return err
		}

		fmt.Printf("register signer address:%s success for %s.\n", addr, user)
		return nil
	},
}

var signerExistCmd = &cli.Command{
	Name:      "exist",
	Usage:     "Check if signer address exists",
	ArgsUsage: "<signer address>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "user",
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() != 1 {
			cli.ShowSubcommandHelpAndExit(ctx, 1)
			return nil
		}

		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		user := ctx.String("user")
		addrStr := ctx.Args().Get(0)
		addr, err := address.NewFromString(addrStr)
		if err != nil {
			return err
		}

		has, err := client.SignerExistInUser(user, addr.String())
		if err != nil {
			return err
		}
		fmt.Println(has)
		return nil
	},
}

var signerListCmd = &cli.Command{
	Name:      "list",
	Usage:     "List of signer address for the specified user",
	ArgsUsage: "<user>",
	Action: func(ctx *cli.Context) error {
		args := ctx.Args()
		if args.Len() != 1 {
			cli.ShowSubcommandHelpAndExit(ctx, 1)
			return nil
		}

		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		user := args.First()
		if _, err := client.GetUser(&auth.GetUserRequest{Name: user}); err != nil {
			return xerrors.Errorf("list user:%s signer failed: %w", user, err)
		}

		signers, err := client.ListSigners(user)
		if err != nil {
			return err
		}
		fmt.Printf("user: %s, signer count:%d\n", user, len(signers))

		if len(signers) == 0 {
			return nil
		}

		const padding = 2
		w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "idx\tsigner\tcreate-time\t")
		for idx, signer := range signers {
			fmt.Fprintf(w, "%d\t%s\t%s\t\n", idx, signer.Signer, signer.CreatedAt.Format(time.RFC1123))
		}
		_ = w.Flush()
		return nil
	},
}

var signerUnregisterCmd = &cli.Command{
	Name:      "unregister",
	Usage:     "Unregister signer",
	ArgsUsage: "<signer address>",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "user",
			Required: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		args := ctx.Args()
		if args.Len() != 1 {
			cli.ShowSubcommandHelpAndExit(ctx, 1)
			return nil
		}

		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		signer := args.First()
		user := ctx.String("user")
		err = client.UnregisterSigners(user, []string{signer})
		if err != nil {
			return xerrors.Errorf("unregister signer:%s failed: %w", signer, err)
		}

		fmt.Printf("unregister signers: %v of %s success.\n", signer, user)
		return nil
	},
}
