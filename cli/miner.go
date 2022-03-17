package cli

import (
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"os"
	"text/tabwriter"
	"time"
)

var minerSubCmds = &cli.Command{
	Name:  "miner",
	Usage: "sub cmds for managing user miners",
	Subcommands: []*cli.Command{
		cmdAddMiner,
		cmdListMiners,
		cmdDelMiner,
	},
}

var cmdAddMiner = &cli.Command{
	Name:      "add",
	Usage:     "add user miner",
	ArgsUsage: "add <user> <miner>",
	Action: func(ctx *cli.Context) error {
		if ctx.Args().Len() != 2 {
			return cli.ShowAppHelp(ctx)
		}
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		user, miner := ctx.Args().Get(0), ctx.Args().Get(1)

		var isCreate bool
		if isCreate, err = client.UpsertMiner(user, miner); err != nil {
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

var cmdListMiners = &cli.Command{
	Name:      "list",
	Usage:     "list miners by user",
	ArgsUsage: "list <user>",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		args := ctx.Args()
		if args.Len() != 1 {
			fmt.Println("input <user> to list it's miners")
			return nil
		}
		user := args.First()
		if _, err := client.GetUser(&auth.GetUserRequest{Name: user}); err != nil {
			return xerrors.Errorf("list user:%s miner failed:%w", user, err)
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
		fmt.Fprintln(w, "idx\tminer\tcreate-time\t")
		for idx, miner := range miners {
			fmt.Fprintf(w, "%d\t%s\t%s\t\n", idx, miner.Miner, miner.CreatedAt.Format(time.RFC1123))
		}
		w.Flush()
		return nil
	},
}

var cmdDelMiner = &cli.Command{
	Name:      "del",
	Usage:     "delete miner",
	ArgsUsage: "del <miner>",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		args := ctx.Args()
		if args.Len() != 1 {
			fmt.Println("input <miner> to delete")
		}
		miner := args.First()

		exists, err := client.DelMiner(miner)
		if err != nil {
			return xerrors.Errorf("delete miner:%s failed:%w", miner, err)
		}

		if exists {
			fmt.Printf("delete miner:%s success.\n", miner)
		} else {
			fmt.Printf("miner:%s not exists.\n", miner)
		}
		return nil
	},
}
