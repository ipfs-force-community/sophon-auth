package cli

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/filecoin-project/venus-auth/util"
	"github.com/ipfs-force-community/metrics"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"go.opencensus.io/plugin/ochttp"
)

var runCommand = &cli.Command{
	Name:      "run",
	Usage:     "run venus-auth daemon",
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
		// todo: rm flag disable-perm-check after v1.13.0
		&cli.BoolFlag{
			Name:  "disable-perm-check",
			Usage: "disable permission check for compatible with old version",
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

		return fillConfigByFlag(cnf, cliCtx), nil
	}

	cnf, err := config.DefaultConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret : %s", err)
	}
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

	return cnf
}

func run(cliCtx *cli.Context) error {
	repoPath, err := homedir.Expand(cliCtx.String("repo"))
	if err != nil {
		return fmt.Errorf("expand home dir: %w", err)
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
	app, err := auth.NewOAuthApp(cnf.Secret, dataPath, cnf.DB)
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

	router := auth.InitRouter(app, !cliCtx.Bool("disable-perm-check"))

	if cnf.Trace != nil && cnf.Trace.JaegerTracingEnabled {
		log.Infof("register jaeger-tracing exporter to %s, with node-name:%s",
			cnf.Trace.JaegerEndpoint, cnf.Trace.ServerName)
		exporter, err := metrics.RegisterJaeger("venus-auth", cnf.Trace)
		if err != nil {
			return fmt.Errorf("RegisterJaegerExporter failed:%w", err)
		}
		defer metrics.UnregisterJaeger(exporter)
		router = &ochttp.Handler{
			Handler: router,
		}
	}

	server := &http.Server{
		Addr:         ":" + cnf.Port,
		Handler:      router,
		ReadTimeout:  cnf.ReadTimeout,
		WriteTimeout: cnf.WriteTimeout,
		IdleTimeout:  cnf.IdleTimeout,
	}
	log.Infof("server start and listen on %s", cnf.Port)
	return server.ListenAndServe()
}
