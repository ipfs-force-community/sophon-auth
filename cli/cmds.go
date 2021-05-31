package cli

import (
	"errors"
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	cli "github.com/urfave/cli/v2"
	"time"
)

var Commands = []*cli.Command{
	genTokenCmd,
	tokensCmd,
	removeTokenCmd,
	addUserCmd,
	updateUserCmd,
	listUsersCmd,
}

var genTokenCmd = &cli.Command{
	Name:      "genToken",
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

var tokensCmd = &cli.Command{
	Name:  "listTokens",
	Usage: "list token info",
	Flags: []cli.Flag{
		&cli.UintFlag{
			Name:  "skip",
			Value: 0,
		},
		&cli.UintFlag{
			Name:  "limit",
			Value: 20,
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
	Name:      "rmToken",
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

var addUserCmd = &cli.Command{
	Name:  "addUser",
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
		miner := ctx.String("miner")
		comment := ctx.String("comment")
		sourceType := ctx.Int("sourceType")
		user := &auth.CreateUserRequest{
			Name:       name,
			Miner:      miner,
			Comment:    comment,
			State:      0,
			SourceType: sourceType,
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
	Name:  "updateUser",
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
	},
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}
		/*	type UpdateUserRequest struct {
			KeySum     KeyCode         `form:"setKeys"` // keyCode Sum
			Name       string          `form:"name"`
			Miner      string          `form:"miner"`      // keyCode:1
			Comment    string          `form:"comment"`    // keyCode:2
			State      int             `form:"state"`      // keyCode:4
			SourceType core.SourceType `form:"sourceType"` // keyCode:8
		}*/
		req := &auth.UpdateUserRequest{
			Name: ctx.String("name"),
		}
		if ctx.IsSet("miner") {
			req.Miner = ctx.String("miner")
			req.KeySum += 1
		}
		if ctx.IsSet("comment") {
			req.Comment = ctx.String("comment")
			req.KeySum += 2
		}
		if ctx.IsSet("state") {
			req.State = ctx.Int("state")
			req.KeySum += 4
		}
		if ctx.IsSet("sourceType") {
			req.SourceType = ctx.Int("sourceType")
			req.KeySum += 8
		}
		err = client.UpdateUser(req)
		if err != nil {
			return err
		}
		fmt.Println("update user success")
		return nil
	},
}

var listUsersCmd = &cli.Command{
	Name:  "listUsers",
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
			fmt.Println("comment:", v.Comment)
			fmt.Println("createTime:", time.Unix(v.CreateTime, 0).Format(time.RFC1123))
			fmt.Println("updateTime:", time.Unix(v.CreateTime, 0).Format(time.RFC1123))
			fmt.Println()
		}
		return nil
	},
}
