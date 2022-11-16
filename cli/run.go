package cli

import (
	"net/http"
	"path"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/gin-gonic/gin"
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
	},
	Action: run,
}

func MakeDir(path string) {
	exist, err := config.Exist(path)
	if err != nil {
		log.Fatalf("Failed to check file exist : %s", err)
	}
	if !exist {
		err = config.MakeDir(path)
		if err != nil {
			log.Fatalf("Failed to crate dir : %s", err)
		}
	}
}

func configScan(path string) *config.Config {
	exist, err := config.Exist(path)
	if err != nil {
		log.Fatalf("Failed to check file exist : %s", err)
	}
	if exist {
		cnf, err := config.DecodeConfig(path)
		if err != nil {
			log.Fatalf("Failed to decode config : %s", err)
		}
		return cnf
	}
	cnf, err := config.DefaultConfig()
	if err != nil {
		log.Fatalf("Failed to generate secret : %s", err)
	}
	err = config.Cover(path, cnf)
	if err != nil {
		log.Fatalf("Failed to write config to home dir : %s", err)
	}
	return cnf
}

func run(cliCtx *cli.Context) error {
	gin.SetMode(gin.ReleaseMode)
	cnfPath := cliCtx.String("config")
	repo := cliCtx.String("repo")
	repo, err := homedir.Expand(repo)
	if err != nil {
		log.Fatal(err)
	}
	if cnfPath == "" {
		cnfPath = path.Join(repo, "config.toml")
	}
	MakeDir(repo)
	dataPath := path.Join(repo, "data")
	MakeDir(dataPath)
	cnf := configScan(cnfPath)
	log.InitLog(cnf.Log)

	if cliCtx.IsSet("mysql-dsn") {
		cnf.DB.DSN = cliCtx.String("mysql-dsn")
	}
	if cliCtx.IsSet("db-type") {
		cnf.DB.Type = cliCtx.String("db-type")
	}

	err = config.Cover(cnfPath, cnf)
	if err != nil {
		log.Fatal(err)
	}

	app, err := auth.NewOAuthApp(cnf.Secret, dataPath, cnf.DB)
	if err != nil {
		log.Fatalf("Failed to init venus-auth: %s", err)
	}
	router := auth.InitRouter(app)

	if cnf.Trace != nil && cnf.Trace.JaegerTracingEnabled {
		log.Infof("register jaeger-tracing exporter to %s, with node-name:%s",
			cnf.Trace.JaegerEndpoint, cnf.Trace.ServerName)
		if exporter, err := metrics.RegisterJaeger("venus-auth", cnf.Trace); err != nil {
			log.Fatalf("RegisterJaegerExporter failed:%s", err.Error())
		} else {
			defer metrics.UnregisterJaeger(exporter)
		}
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
