package cli

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/ipfs-force-community/metrics"
	"github.com/ipfs-force-community/sophon-auth/auth"
	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/log"
	"github.com/ipfs-force-community/sophon-auth/util"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"go.opencensus.io/plugin/ochttp"
)

var runCommand = &cli.Command{
	Name:      "run",
	Usage:     "run sophon-auth daemon",
	ArgsUsage: "[name]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "mysql-dsn",
			Usage: "mysql connection string",
		},
		&cli.StringFlag{
			Name:  "db-type",
			Usage: "which db to use. sqlite/mysql",
		},
	},
	Action: run,
}

func configScan(path string, cliCtx *cli.Context) (*config.Config, error) {
	exist, err := util.Exist(path)
	if err != nil {
		return nil, fmt.Errorf("failed to check file exist : %s", err)
	}
	if exist {
		cnf, err := config.DecodeConfig(path)
		if err != nil {
			return nil, fmt.Errorf("failed to decode config : %s", err)
		}
		// todo: remove after v1.12
		if len(cnf.Listen) == 0 {
			cnf.Listen = cliCtx.String("listen")
			if err := config.Cover(path, cnf); err != nil {
				return nil, err
			}
		}

		return fillConfigByFlag(cnf, cliCtx), nil
	}

	cnf := config.DefaultConfig()
	cnf = fillConfigByFlag(cnf, cliCtx)
	err = config.Cover(path, cnf)
	if err != nil {
		return nil, fmt.Errorf("failed to write config to home dir : %s", err)
	}

	return cnf, nil
}

func fillConfigByFlag(cnf *config.Config, cliCtx *cli.Context) *config.Config {
	if cliCtx.IsSet("mysql-dsn") {
		cnf.DB.DSN = cliCtx.String("mysql-dsn")
	}
	if cliCtx.IsSet("db-type") {
		cnf.DB.Type = cliCtx.String("db-type")
	}
	if cliCtx.IsSet("listen") {
		cnf.Listen = cliCtx.String("listen")
	}

	return cnf
}

func run(cliCtx *cli.Context) error {
	repoPath, err := homedir.Expand(cliCtx.String("repo"))
	if err != nil {
		return fmt.Errorf("expand home dir: %w", err)
	}
	exist, err := util.Exist(repoPath)
	if err != nil {
		return fmt.Errorf("check repo exist: %w", err)
	}

	// todo: rm compatibility for repo when appropriate
	if !exist {
		deprecatedRepoPath, err := homedir.Expand("~/.venus-auth")
		if err != nil {
			return fmt.Errorf("expand deprecated home dir: %w", err)
		}

		deprecatedRepoPathExist, err := util.Exist(deprecatedRepoPath)
		if err != nil {
			return fmt.Errorf("check deprecated repo exist: %w", err)
		}
		if deprecatedRepoPathExist {
			fmt.Printf("[WARM]: repo path %s is deprecated, please transfer to %s instead\n", deprecatedRepoPath, repoPath)
			repoPath = deprecatedRepoPath
		}
	}

	repo, err := NewFsRepo(repoPath)
	if err != nil {
		return fmt.Errorf("init repo: %s", err)
	}

	cnfPath := cliCtx.String("config")
	if len(cnfPath) == 0 {
		cnfPath = filepath.Join(repoPath, DefaultConfigFile)
	}
	cnf, err := configScan(cnfPath, cliCtx)
	if err != nil {
		return err
	}

	log.InitLog(cnf.Log)

	dataPath := repo.GetDataDir()
	app, err := auth.NewOAuthApp(dataPath, cnf.DB)
	if err != nil {
		return fmt.Errorf("init oauth app: %s", err)
	}

	token, err := app.GetDefaultAdminToken()
	if err != nil {
		return fmt.Errorf("get default admin token: %s", err)
	}

	err = repo.SaveToken(token)
	if err != nil {
		return fmt.Errorf("save token: %s", err)
	}

	router := auth.InitRouter(app)

	if cnf.Trace != nil && cnf.Trace.JaegerTracingEnabled {
		log.Infof("setup jaeger-tracing exporter to %s, with node-name:%s",
			cnf.Trace.JaegerEndpoint, cnf.Trace.ServerName)
		exporter, err := metrics.SetupJaegerTracing("sophon-auth", cnf.Trace)
		if err != nil {
			return fmt.Errorf("SetupJaegerTracing failed:%w", err)
		}
		defer func() {
			if exporter != nil {
				if err := metrics.ShutdownJaeger(context.Background(), exporter); err != nil {
					log.Warnf("failed to shutdown jaeger-tracing: %s", err)
				}
			}
		}()
		router = &ochttp.Handler{
			Handler: router,
		}
	}

	server := &http.Server{
		Addr:         cnf.Listen,
		Handler:      router,
		ReadTimeout:  cnf.ReadTimeout,
		WriteTimeout: cnf.WriteTimeout,
		IdleTimeout:  cnf.IdleTimeout,
	}
	log.Infof("server start and listen on %s", cnf.Listen)
	return server.ListenAndServe()
}
