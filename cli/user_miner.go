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

var minerSubCmds = &cli.Command{
	Name:  "miner",
	Usage: "Sub commands for managing user miners",
	Subcommands: []*cli.Command{
		minerAddCmd,
		minerExistCmd,
		minerListCmd,
		minerDeleteCmd,
	},
}

var minerAddCmd = &cli.Command{
	Name:  "add",
	Usage: "Add miner for specified user",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "openMining",
			Usage: "false/true",
			Value: true,
		},
	},
	ArgsUsage: "<user> <miner>",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() != 2 {
			cli.ShowSubcommandHelpAndExit(ctx, 1)
			return nil
		}
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		user, miner := ctx.Args().Get(0), ctx.Args().Get(1)
		openMining := ctx.Bool("openMining")

		var isCreate bool
		if isCreate, err = client.UpsertMiner(user, miner, openMining); err != nil {
			return err
		}
		var opStr string
		if isCreate {
			opStr = "create"
		} else {
			opStr = "update"
		}

		fmt.Printf("%s user:%s miner:%s success.\n", opStr, user, miner)
		return nil
	},
}

var minerExistCmd = &cli.Command{
	Name:      "exist",
	Usage:     "Check if miner exist in the user",
	ArgsUsage: "<miner>",
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
		miner := ctx.Args().Get(0)
		addr, err := address.NewFromString(miner)
		if err != nil {
			return err
		}

		exist, err := client.MinerExistInUser(user, addr.String())
		if err != nil {
			return err
		}
		fmt.Println(exist)
		return nil
	},
}

var minerListCmd = &cli.Command{
	Name:      "list",
	Usage:     "List of miners for the specified user",
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
			return xerrors.Errorf("list user:%s miner failed: %w", user, err)
		}

		miners, err := client.ListMiners(user)
		if err != nil {
			return err
		}
		fmt.Printf("user: %s, miner count:%d\n", user, len(miners))

		if len(miners) == 0 {
			return nil
		}

		const padding = 2
		w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "idx\tminer\topenMining\tcreate-time\t")
		for idx, miner := range miners {
			fmt.Fprintf(w, "%d\t%s\t%v\t%s\t\n", idx, miner.Miner, miner.OpenMining, miner.CreatedAt.Format(time.RFC1123))
		}
		_ = w.Flush()
		return nil
	},
}

var minerDeleteCmd = &cli.Command{
	Name:      "delete",
	Usage:     "Delete miner",
	ArgsUsage: "<miner>",
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

		miner := args.First()
		exists, err := client.DelMiner(miner)
		if err != nil {
			return xerrors.Errorf("delete miner:%s failed: %w", miner, err)
		}

		if exists {
			fmt.Printf("remove miner:%s success.\n", miner)
		} else {
			fmt.Printf("miner:%s not exists.\n", miner)
		}
		return nil
	},
}
