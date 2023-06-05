package cli

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/sophon-auth/core"
)

var tokenSubCommand = &cli.Command{
	Name:  "token",
	Usage: "token command",
	Subcommands: []*cli.Command{
		genTokenCmd,
		getTokenCmd,
		listTokensCmd,
		removeTokenCmd,
		recoverTokenCmd,
	},
}

var genTokenCmd = &cli.Command{
	Name:      "gen",
	Usage:     "generate token",
	ArgsUsage: "[name]",
	UsageText: "./sophon-auth token gen --perm=<auth> [name]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "perm",
			Usage: "permission for API auth (read, write, sign, admin)",
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
			return fmt.Errorf("usage: genToken name")
		}
		name := ctx.Args().Get(0)

		if !ctx.IsSet("perm") {
			return fmt.Errorf("`perm` flag not set")
		}

		perm := ctx.String("perm")
		if !core.IsValid(perm) {
			return fmt.Errorf("`perm` flag invalid")
		}

		extra := ctx.String("extra")
		tk, err := client.GenerateToken(ctx.Context, name, perm, extra)
		if err != nil {
			return err
		}

		fmt.Printf("generate token success: %s\n", tk)
		return nil
	},
}

var getTokenCmd = &cli.Command{
	Name:  "get",
	Usage: "get token",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "name",
		},
		&cli.StringFlag{
			Name: "token",
		},
	},
	Action: func(ctx *cli.Context) error {
		name := ctx.String("name")
		token := ctx.String("token")
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		tokens, err := client.GetToken(ctx.Context, name, token)
		if err != nil {
			return err
		}
		for _, token := range tokens {
			fmt.Println("name:       ", token.Name)
			fmt.Println("perm:       ", token.Perm)
			fmt.Println("create time:", token.CreateTime)
			fmt.Println("token:      ", token.Token)
			fmt.Println()
		}

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
		tks, err := client.Tokens(ctx.Context, skip, limit)
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
			return fmt.Errorf("usage: rmToken [token]")
		}
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		tk := ctx.Args().First()
		err = client.RemoveToken(ctx.Context, tk)
		if err != nil {
			return err
		}
		fmt.Printf("remove token success: %s\n", tk)
		return nil
	},
}

var recoverTokenCmd = &cli.Command{
	Name:      "recover",
	Usage:     "recover deleted token",
	ArgsUsage: "[token]",
	Action: func(ctx *cli.Context) error {
		if ctx.NArg() < 1 {
			return fmt.Errorf("usage: recover [token]")
		}
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		tk := ctx.Args().First()
		err = client.RecoverToken(ctx.Context, tk)
		if err != nil {
			return err
		}
		fmt.Printf("recover token success: %s\n", tk)
		return nil
	},
}
