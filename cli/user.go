package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/storage"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
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
		rateLimitSubCmds,
		minerSubCmds,
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
			Name: "comment",
		},
		&cli.IntFlag{
			Name: "sourceType",
		},
		&cli.IntFlag{
			Name:  "state",
			Usage: "0:disabled, 1:enabled",
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
		if ctx.IsSet("comment") {
			req.Comment = ctx.String("comment")
			req.KeySum |= 2
		}
		if ctx.IsSet("state") {
			req.State = core.UserState(ctx.Int("state"))
			req.KeySum |= 4
		}
		if ctx.IsSet("sourceType") {
			req.SourceType = ctx.Int("sourceType")
			req.KeySum |= 8
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
			Name: "skip",
		},
		&cli.UintFlag{
			Name:  "limit",
			Value: 20,
		},
		&cli.IntFlag{
			Name:  "state",
			Usage: "0:disabled, 1:enabled, not-set:[show all]",
		},
		&cli.IntFlag{
			Name: "sourceType",
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
			req.KeySum++
		}
		if ctx.IsSet("state") {
			req.KeySum += 2
		}
		users, err := client.ListUsers(req)
		if err != nil {
			return err
		}

		for k, v := range users {
			var minerStr interface{}
			if miners, err := client.ListMiners(v.Name); err != nil {
				minerStr = err.Error()
			} else if len(miners) > 0 {
				ms := make([]string, len(miners))
				for idx, m := range miners {
					ms[idx] = m.Miner
				}
				minerStr = ms
			}
			fmt.Println("number:", k+1)
			fmt.Println("name:", v.Name)
			fmt.Println("state:", v.State.String())
			if minerStr != nil {
				fmt.Println("miners:", minerStr)
			}
			if len(v.Comment) != 0 {
				fmt.Println("comment:", v.Comment)
			}
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
		fmt.Println("sourceType:", user.SourceType, "\t// miner:1")
		fmt.Println("state", user.State, "\t// 0: disable, 1: enable")
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

var rateLimitSubCmds = &cli.Command{
	Name:  "rate-limit",
	Usage: "sub cmds for managing user request limits",
	Subcommands: []*cli.Command{
		rateLimitAdd,
		rateLimitUpdate,
		rateLimitGet,
		rateLimitDel},
}

var rateLimitGet = &cli.Command{
	Name:      "get",
	Usage:     "get user request rate limit",
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

		name := ctx.Args().Get(0)
		var limits []*storage.UserRateLimit
		limits, err = client.GetUserRateLimit(name, "")
		if err != nil {
			return err
		}

		if len(limits) == 0 {
			fmt.Printf("user have no request rate limit\n")
		} else {
			for _, l := range limits {
				fmt.Printf("user:%s, limit id:%s, request limit amount:%d, duration:%.2f(h)\n",
					l.Name, l.Id, l.ReqLimit.Cap, l.ReqLimit.ResetDur.Hours())
			}
		}
		return nil
	},
}

var rateLimitAdd = &cli.Command{
	Name:  "add",
	Usage: "add user request rate limit",
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "id", Usage: "rate limit id to update"},
	},
	ArgsUsage: "user rate-limit add <name> <limitAmount> <duration(2h, 1h:20m, 2m10s)>",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		if ctx.Args().Len() < 3 {
			return cli.ShowAppHelp(ctx)
		}

		name := ctx.Args().Get(0)

		if res, _ := client.GetUserRateLimit(name, ""); len(res) > 0 {
			return fmt.Errorf("user rate limit:%s exists", res[0].Id)
		}

		var limitAmount uint64
		var resetDuration time.Duration
		if limitAmount, err = strconv.ParseUint(ctx.Args().Get(1), 10, 64); err != nil {
			return err
		}
		if resetDuration, err = time.ParseDuration(ctx.Args().Get(2)); err != nil {
			return err
		}
		if resetDuration <= 0 {
			return fmt.Errorf("reset duratoin must be positive")
		}

		userLimit := &auth.UpsertUserRateLimitReq{
			Name:     name,
			ReqLimit: storage.ReqLimit{Cap: int64(limitAmount), ResetDur: resetDuration},
		}

		if ctx.IsSet("id") {
			userLimit.Id = ctx.String("id")
		}

		if userLimit.Id, err = client.UpsertUserRateLimit(userLimit); err != nil {
			return err
		}

		fmt.Printf("upsert user rate limit success:\t%s\n", userLimit.Id)

		return nil
	},
}

var rateLimitUpdate = &cli.Command{
	Name:      "update",
	Usage:     "update user request rate limit",
	Flags:     []cli.Flag{},
	ArgsUsage: "<name> <rate-limit-id> <limitAmount> <duration(2h, 1h:20m, 2m10s)>",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		if ctx.Args().Len() != 4 {
			return cli.ShowAppHelp(ctx)
		}

		name := ctx.Args().Get(0)
		id := ctx.Args().Get(1)

		if res, err := client.GetUserRateLimit(name, id); err != nil {
			return err
		} else if len(res) == 0 {
			return fmt.Errorf("user rate limit:%s NOT exists", id)
		}

		var limitAmount uint64
		var resetDuration time.Duration
		if limitAmount, err = strconv.ParseUint(ctx.Args().Get(2), 10, 64); err != nil {
			return err
		}
		if resetDuration, err = time.ParseDuration(ctx.Args().Get(3)); err != nil {
			return err
		}
		if resetDuration <= 0 {
			return fmt.Errorf("reset duratoin must be positive")
		}

		userLimit := &auth.UpsertUserRateLimitReq{
			Id: id, Name: name,
			ReqLimit: storage.ReqLimit{Cap: int64(limitAmount), ResetDur: resetDuration},
		}

		if userLimit.Id, err = client.UpsertUserRateLimit(userLimit); err != nil {
			return err
		}

		fmt.Printf("upsert user rate limit success:\t%s\n", userLimit.Id)

		return nil
	},
}

var rateLimitDel = &cli.Command{
	Name:      "del",
	Usage:     "delete user request rate limit",
	Flags:     []cli.Flag{},
	ArgsUsage: "user rate-limit <user> <rate-limit-id> ",
	Action: func(ctx *cli.Context) error {
		client, err := GetCli(ctx)
		if err != nil {
			return err
		}

		if ctx.Args().Len() != 2 {
			return cli.ShowAppHelp(ctx)
		}

		var delReq = &auth.DelUserRateLimitReq{
			Name: ctx.Args().Get(0),
			Id:   ctx.Args().Get(1)}

		if res, err := client.GetUserRateLimit(delReq.Name, delReq.Id); err != nil {
			return err
		} else if len(res) == 0 {
			fmt.Printf("user:%s, rate-limit-id:%s Not exits\n", delReq.Name, delReq.Id)
			return nil
		}

		var id string
		if id, err = client.DelUserRateLimit(delReq); err != nil {
			return err
		}
		fmt.Printf("delete rate limit success, %s\n", id)
		return nil
	},
}
