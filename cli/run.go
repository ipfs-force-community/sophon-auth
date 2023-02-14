package cli

import (
	"fmt"
	"net/http"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/metrics"
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

func run(cliCtx *cli.Context) error {
	gin.SetMode(gin.ReleaseMode)

	repoPath := cliCtx.String("repo")
	repo, err := NewFsRepo(repoPath)
	if err != nil {
		return fmt.Errorf("init repo: %s", err)
	}
	cnf, err := repo.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %s", err)
	}

	log.InitLog(cnf.Log)

	if cliCtx.IsSet("mysql-dsn") {
		cnf.DB.DSN = cliCtx.String("mysql-dsn")
	}
	if cliCtx.IsSet("db-type") {
		cnf.DB.Type = cliCtx.String("db-type")
	}

	dataPath, err := repo.GetDataDir()
	if err != nil {
		return fmt.Errorf("get data dir: %s", err)
	}

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
