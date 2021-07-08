package cli

import (
	"errors"
	"fmt"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/urfave/cli/v2"
)

var tokenSubCommand = &cli.Command{
	Name:  "token",
	Usage: "token command",
	Subcommands: []*cli.Command{
		genTokenCmd,
		listTokensCmd,
		removeTokenCmd,
	},
}

var genTokenCmd = &cli.Command{
	Name:      "gen",
	Usage:     "generate token",
	ArgsUsage: "[name]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "perm",
			Usage: "permission for API auth (read, write, sign, admin)",
			Value: core.PermRead,
		},
		&cli.StringFlag{
			Name:  "extra",
			Usage: "custom string in JWT payload",
			Value: "",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		if ctx.NArg() < 1 {
			return errors.New("usage: genToken name")
		}
		name := ctx.Args().Get(0)
		perm := ctx.String("perm")
		if err = core.ContainsPerm(perm); err != nil {
			return err
		}

		extra := ctx.String("extra")
		tk, err := client.GenerateToken(name, perm, extra)
		if err != nil {
			return err
		}
		fmt.Printf("generate token success: %s\n", tk)
		return nil
	},
}

var listTokensCmd = &cli.Command{
	Name:  "list",
	Usage: "list token info",
	Flags: []cli.Flag{
		&cli.UintFlag{
			Name:  "skip",
			Value: 0,
		},
		&cli.UintFlag{
			Name:  "limit",
			Value: 20,
			Usage: "max value:100 (default: 20)",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		skip := int64(ctx.Uint("skip"))
		limit := int64(ctx.Uint("limit"))
		tks, err := client.Tokens(skip, limit)
		if err != nil {
			return err
		}
		//	Token     string    `json:"token"`
		//	Name      string    `json:"name"`
		//	CreatTime time.Time `json:"createTime"`
		fmt.Println("num\tname\t\tperm\t\tcreateTime\t\ttoken")
		for k, v := range tks {
			name := v.Name
			if len(name) < 8 {
				name = name + "\t"
			}
			fmt.Printf("%d\t%s\t%s\t%s\t%s\n", k+1, name, v.Perm, v.CreateTime.Format("2006-01-02 15:04:05"), v.Token)
		}
		return nil
	},
}

var removeTokenCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove token",
	ArgsUsage: "[token]",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return errors.New("usage: rmToken [token]")
		}
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		tk := ctx.Args().First()
		err = client.RemoveToken(tk)
		if err != nil {
			return err
		}
		fmt.Printf("remove token success: %s\n", tk)
		return nil
	},
}
