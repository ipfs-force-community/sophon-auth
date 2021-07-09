package cli

import (
	"fmt"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"time"
)

var accountSubCommand = &cli.Command{
	Name:  "account",
	Usage: "account command",
	Subcommands: []*cli.Command{
		addAccountCmd,
		updateAccountCmd,
		listAccountsCmd,
		activeAccountCmd,
		getAccountCmd,
		hasMinerCmd,
	},
}

var addAccountCmd = &cli.Command{
	Name:  "add",
	Usage: "add account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Required: true,
			Usage:    "required",
		},
		&cli.StringFlag{
			Name: "miner",
		},
		&cli.StringFlag{
			Name: "comment",
		},
		&cli.IntFlag{
			Name:  "sourceType",
			Value: 0,
		},
		&cli.IntFlag{Name: "reqLimitAmount"},
		&cli.StringFlag{Name: "reqLimitResetDuration", Value: "24h", Usage: "10h, 5m, 10h5m, 2h5m20s"},
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		name := ctx.String("name")
		comment := ctx.String("comment")
		sourceType := ctx.Int("sourceType")
		account := &auth.CreateAccountRequest{
			Name:       name,
			Comment:    comment,
			State:      0,
			SourceType: sourceType,
		}
		if ctx.IsSet("miner") {
			mAddr, err := address.NewFromString(ctx.String("miner"))
			if err != nil {
				return err
			}
			account.Miner = mAddr.String()
		}

		if ctx.IsSet("reqLimitAmount") {
			account.ReqLimit.Cap = ctx.Int64("reqLimitAmount")
			if account.ReqLimit.ResetDur, err = time.ParseDuration(ctx.String("reqLimitResetDuration")); err != nil {
				return err
			}
			if account.ReqLimit.ResetDur <= time.Second {
				return fmt.Errorf("request limit reset duration must larger than 1(second)")
			}
		}

		res, err := client.CreateAccount(account)
		if err != nil {
			return err
		}
		fmt.Printf("add account success: %s\n", res.Id)
		return nil
	},
}

var updateAccountCmd = &cli.Command{
	Name:  "update",
	Usage: "update account",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "name",
			Required: true,
		},
		&cli.StringFlag{
			Name: "miner",
		},
		&cli.StringFlag{
			Name: "comment",
		},
		&cli.IntFlag{
			Name: "sourceType",
		},
		&cli.IntFlag{
			Name: "state",
		},
		&cli.IntFlag{Name: "reqLimitAmount"},
		&cli.StringFlag{Name: "reqLimitResetDuration"},
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		req := &auth.UpdateAccountRequest{
			Name: ctx.String("name"),
		}
		if ctx.IsSet("miner") {
			addr, err := address.NewFromString(ctx.String("miner"))
			if err != nil {
				return err
			}
			req.Miner = addr.String()
			req.KeySum |= 1
		}
		if ctx.IsSet("comment") {
			req.Comment = ctx.String("comment")
			req.KeySum |= 2
		}
		if ctx.IsSet("state") {
			req.State = ctx.Int("state")
			req.KeySum |= 4
		}
		if ctx.IsSet("sourceType") {
			req.SourceType = ctx.Int("sourceType")
			req.KeySum |= 8
		}
		if ctx.IsSet("reqLimitAmount") {
			req.ReqLimit.Cap = ctx.Int64("reqLimitAmount")
			if req.ReqLimit.ResetDur, err = time.ParseDuration(ctx.String("reqLimitResetDuration")); err != nil {
				return err
			}
			if req.ReqLimit.ResetDur <= time.Second {
				return fmt.Errorf("request limit reset duration must larger than 1(second)")
			}
			req.KeySum |= 0x10
		}
		err = client.UpdateAccount(req)
		if err != nil {
			return err
		}
		fmt.Println("update account success")
		return nil
	},
}

var activeAccountCmd = &cli.Command{
	Name:      "active",
	Usage:     "update account",
	Flags:     []cli.Flag{},
	ArgsUsage: "name",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		if ctx.NArg() != 1 {
			return xerrors.New("expect name")
		}

		req := &auth.UpdateAccountRequest{
			Name: ctx.Args().Get(0),
		}

		req.State = 1
		req.KeySum += 4

		err = client.UpdateAccount(req)
		if err != nil {
			return err
		}
		fmt.Println("active account success")
		return nil
	},
}

var listAccountsCmd = &cli.Command{
	Name:  "list",
	Usage: "list accounts",
	Flags: []cli.Flag{
		&cli.UintFlag{
			Name:  "skip",
			Value: 0,
		},
		&cli.UintFlag{
			Name:  "limit",
			Value: 20,
		},
		&cli.IntFlag{
			Name:  "state",
			Value: 0,
		},
		&cli.IntFlag{
			Name:  "sourceType",
			Value: 0,
		},
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		req := &auth.ListAccountsRequest{
			Page: &core.Page{
				Limit: ctx.Int64("limit"),
				Skip:  ctx.Int64("skip"),
			},
			SourceType: ctx.Int("sourceType"),
			State:      ctx.Int("state"),
		}
		if ctx.IsSet("sourceType") {
			req.KeySum += 1
		}
		if ctx.IsSet("state") {
			req.KeySum += 2
		}
		accounts, err := client.ListAccounts(req)
		if err != nil {
			return err
		}

		for k, v := range accounts {
			fmt.Println("number:", k+1)
			fmt.Println("name:", v.Name)
			fmt.Println("miner:", v.Miner)
			fmt.Println("sourceType:", v.SourceType, "\t// miner:1")
			fmt.Println("state", v.State, "\t// 0: disable, 1: enable")
			fmt.Printf("reqLimit:\tamount:%d, resetDuration:%s\n", v.ReqLimit.Cap, v.ReqLimit.ResetDur.String())
			fmt.Println("comment:", v.Comment)
			fmt.Println("createTime:", time.Unix(v.CreateTime, 0).Format(time.RFC1123))
			fmt.Println("updateTime:", time.Unix(v.CreateTime, 0).Format(time.RFC1123))
			fmt.Println()
		}
		return nil
	},
}

var getAccountCmd = &cli.Command{
	Name:      "get",
	Usage:     "get account by name",
	ArgsUsage: "<name>",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		if ctx.NArg() != 1 {
			return xerrors.Errorf("specify account name")
		}
		name := ctx.Args().Get(0)
		account, err := client.GetAccount(&auth.GetAccountRequest{Name: name})
		if err != nil {
			return err
		}

		fmt.Println("name:", account.Name)
		fmt.Println("miner:", account.Miner)
		fmt.Println("sourceType:", account.SourceType, "\t// miner:1")
		fmt.Println("state", account.State, "\t// 0: disable, 1: enable")
		fmt.Printf("reqLimit:\tamount:%d, resetDuration:%s\n", account.ReqLimit.Cap, account.ReqLimit.ResetDur.String())
		fmt.Println("comment:", account.Comment)
		fmt.Println("createTime:", time.Unix(account.CreateTime, 0).Format(time.RFC1123))
		fmt.Println("updateTime:", time.Unix(account.CreateTime, 0).Format(time.RFC1123))
		fmt.Println()
		return nil
	},
}

var hasMinerCmd = &cli.Command{
	Name:      "has",
	Usage:     "check miner exit",
	ArgsUsage: "<miner>",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		if ctx.NArg() != 1 {
			return xerrors.Errorf("specify miner address")
		}
		miner := ctx.Args().Get(0)
		addr, err := address.NewFromString(miner)
		if err != nil {
			return err
		}

		has, err := client.HasMiner(&auth.HasMinerRequest{Miner: addr.String()})
		if err != nil {
			return err
		}
		fmt.Println(has)
		return nil
	},
}
