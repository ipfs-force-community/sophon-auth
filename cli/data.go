package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/ipfs-force-community/sophon-auth/storage"
)

// DataDump 用于序列化所有的核心存储数据
type DataDump struct {
	Users      []*storage.User          `json:"users"`
	Tokens     []*storage.KeyPair       `json:"tokens"`
	Miners     []*storage.Miner         `json:"miners"`
	Signers    []*storage.Signer        `json:"signers"`
	RateLimits []*storage.UserRateLimit `json:"rate_limits"`
}

var ExportCmd = &cli.Command{
	Name:  "export",
	Usage: "Export all data from the database to a JSON file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "out",
			Usage:    "Output JSON file path",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "skip-users",
			Usage: "List of users to skip during export",
		},
	},
	Action: func(cctx *cli.Context) error {
		// 1. 初始化配置与 Store 接口（支持 MySQL 和 Badger）
		repoPath := cctx.String("repo")
		repo, err := NewFsRepo(repoPath)
		if err != nil {
			return fmt.Errorf("new repo failed: %w", err)
		}

		cnf, err := repo.GetConfig()
		if err != nil {
			return fmt.Errorf("decode config failed: %w", err)
		}

		store, err := storage.NewStore(cnf.DB, repoPath)
		if err != nil {
			return fmt.Errorf("new store failed: %w", err)
		}

		dump := DataDump{}

		skipUsersMap := make(map[string]bool)
		for _, name := range cctx.StringSlice("skip-users") {
			skipUsersMap[name] = true
		}

		// 2. 导出所有的 Users (传递比较大的限额 99999999 以获取全部)
		allUsers, err := store.ListUsers(0, 99999999, core.UserStateUndefined)
		if err != nil {
			return fmt.Errorf("list users failed: %w", err)
		}

		var users []*storage.User
		for _, u := range allUsers {
			if !skipUsersMap[u.Name] {
				users = append(users, u)
			}
		}
		dump.Users = users

		// 3. 导出所有的 Tokens
		allTokens, err := store.List(0, 99999999)
		if err != nil {
			return fmt.Errorf("list tokens failed: %w", err)
		}

		var tokens []*storage.KeyPair
		for _, t := range allTokens {
			if !skipUsersMap[t.Name] {
				tokens = append(tokens, t)
			}
		}
		dump.Tokens = tokens

		// 4. 用户衍生的 Miners, Signers, RateLimits
		for _, u := range users {
			miners, _ := store.ListMiners(u.Name)
			if len(miners) > 0 {
				dump.Miners = append(dump.Miners, miners...)
			}

			signers, _ := store.ListSigner(u.Name)
			if len(signers) > 0 {
				dump.Signers = append(dump.Signers, signers...)
			}

			// id 传空字符串意味着拉取该 User 的所有限流规则
			limits, _ := store.GetRateLimits(u.Name, "")
			if len(limits) > 0 {
				dump.RateLimits = append(dump.RateLimits, limits...)
			}
		}

		// 5. 序列化并写入文件
		data, err := json.MarshalIndent(dump, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal data failed: %w", err)
		}

		outPath := cctx.String("out")
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("write file failed: %w", err)
		}

		fmt.Printf("Successfully exported %d users, %d tokens, %d miners, %d signers, %d rate limits to %s\n",
			len(dump.Users), len(dump.Tokens), len(dump.Miners), len(dump.Signers), len(dump.RateLimits), outPath)
		return nil
	},
}

var ImportCmd = &cli.Command{
	Name:  "import",
	Usage: "Import data into the database from a JSON file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "in",
			Usage:    "Input JSON file path",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		repoPath := cctx.String("repo")
		repo, err := NewFsRepo(repoPath)
		if err != nil {
			return fmt.Errorf("new repo failed: %w", err)
		}

		cnf, err := repo.GetConfig()
		if err != nil {
			return fmt.Errorf("decode config failed: %w", err)
		}

		store, err := storage.NewStore(cnf.DB, repoPath)
		if err != nil {
			return fmt.Errorf("new store failed: %w", err)
		}

		inPath := cctx.String("in")
		data, err := os.ReadFile(inPath)
		if err != nil {
			return fmt.Errorf("read file failed: %w", err)
		}

		var dump DataDump
		if err := json.Unmarshal(data, &dump); err != nil {
			return fmt.Errorf("unmarshal data failed: %w", err)
		}

		// 1. 恢复 Users
		for _, u := range dump.Users {
			if err := store.PutUser(u); err != nil {
				fmt.Printf("Warning: failed to import user %s: %v\n", u.Name, err)
			}
		}

		// 2. 恢复 Tokens (包含 Secret)
		for _, t := range dump.Tokens {
			if err := store.Put(t); err != nil {
				fmt.Printf("Warning: failed to import token %s: %v\n", t.Name, err)
			}
		}

		// 3. 恢复 Miners
		for _, m := range dump.Miners {
			if _, err := store.UpsertMiner(m.Miner.Address(), m.User, m.OpenMining); err != nil {
				fmt.Printf("Warning: failed to import miner %s: %v\n", m.Miner.Address().String(), err)
			}
		}

		// 4. 恢复 Signers
		for _, s := range dump.Signers {
			if err := store.RegisterSigner(s.Signer.Address(), s.User); err != nil {
				fmt.Printf("Warning: failed to import signer %s: %v\n", s.Signer.Address().String(), err)
			}
		}

		// 5. 恢复 RateLimits
		for _, rl := range dump.RateLimits {
			if _, err := store.PutRateLimit(rl); err != nil {
				fmt.Printf("Warning: failed to import rate limit %s: %v\n", rl.Id, err)
			}
		}

		fmt.Printf("Successfully imported data from %s\n", inPath)
		return nil
	},
}
