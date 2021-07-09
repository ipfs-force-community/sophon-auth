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

var userSubCommand = &cli.Command{
	Name:  "user",
	Usage: "user command",
	Subcommands: []*cli.Command{
		addUserCmd,
		updateUserCmd,
		listUsersCmd,
		activeUserCmd,
		getUserCmd,
		hasMinerCmd,
	},
}

var addUserCmd = &cli.Command{
	Name:  "add",
	Usage: "add user",
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
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		name := ctx.String("name")
		comment := ctx.String("comment")
		sourceType := ctx.Int("sourceType")
		user := &auth.CreateUserRequest{
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
			user.Miner = mAddr.String()
		}
		res, err := client.CreateUser(user)
		if err != nil {
			return err
		}
		fmt.Printf("add user success: %s\n", res.Id)
		return nil
	},
}

var updateUserCmd = &cli.Command{
	Name:  "update",
	Usage: "update user",
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
		&cli.IntFlag{
			Name: "burst",
		},
		&cli.IntFlag{
			Name: "rate",
		},
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		req := &auth.UpdateUserRequest{
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
		if ctx.IsSet("burst") {
			req.Burst = ctx.Int("burst")
			req.KeySum |= 0x10
		}
		if ctx.IsSet("rate") {
			req.Rate = ctx.Int("rate")
			req.KeySum |= 0x20
		}
		err = client.UpdateUser(req)
		if err != nil {
			return err
		}
		fmt.Println("update user success")
		return nil
	},
}

var activeUserCmd = &cli.Command{
	Name:      "active",
	Usage:     "update user",
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

		req := &auth.UpdateUserRequest{
			Name: ctx.Args().Get(0),
		}

		req.State = 1
		req.KeySum += 4

		err = client.UpdateUser(req)
		if err != nil {
			return err
		}
		fmt.Println("active user success")
		return nil
	},
}

var listUsersCmd = &cli.Command{
	Name:  "list",
	Usage: "list users",
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
		req := &auth.ListUsersRequest{
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
		users, err := client.ListUsers(req)
		if err != nil {
			return err
		}

		for k, v := range users {
			fmt.Println("number:", k+1)
			fmt.Println("name:", v.Name)
			fmt.Println("miner:", v.Miner)
			fmt.Println("sourceType:", v.SourceType, "\t// miner:1")
			fmt.Println("state", v.State, "\t// 0: disable, 1: enable")
			fmt.Printf("rate-limit burst:%d, rate:%d\n", v.Burst, v.Rate)
			fmt.Println("comment:", v.Comment)
			fmt.Println("createTime:", time.Unix(v.CreateTime, 0).Format(time.RFC1123))
			fmt.Println("updateTime:", time.Unix(v.CreateTime, 0).Format(time.RFC1123))
			fmt.Println()
		}
		return nil
	},
}

var getUserCmd = &cli.Command{
	Name:      "get",
	Usage:     "get user by name",
	ArgsUsage: "<name>",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		if ctx.NArg() != 1 {
			return xerrors.Errorf("specify user name")
		}
		name := ctx.Args().Get(0)
		user, err := client.GetUser(&auth.GetUserRequest{Name: name})
		if err != nil {
			return err
		}

		fmt.Println("name:", user.Name)
		fmt.Println("miner:", user.Miner)
		fmt.Println("sourceType:", user.SourceType, "\t// miner:1")
		fmt.Println("state", user.State, "\t// 0: disable, 1: enable")
		fmt.Printf("rate-limit burst:%d, rate:%d\n", user.Burst, user.Rate)
		fmt.Println("comment:", user.Comment)
		fmt.Println("createTime:", time.Unix(user.CreateTime, 0).Format(time.RFC1123))
		fmt.Println("updateTime:", time.Unix(user.CreateTime, 0).Format(time.RFC1123))
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
