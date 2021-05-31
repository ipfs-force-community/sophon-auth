package main

import (
	"fmt"
	"github.com/filecoin-project/venus-auth/auth"
	locli "github.com/filecoin-project/venus-auth/cli"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/core"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"net/http"
	"os"
	"path"
)

func main() {
	app := newApp()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newApp() (app *cli.App) {
	app = &cli.App{
		Action:  run,
		Version: core.Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "config dir path",
			},
			&cli.StringFlag{
				Name:    "repo",
				EnvVars: []string{"OAUTH_HOME"},
				Hidden:  true,
				Value:   "~/.venus-auth",
			},
		},
		Commands: locli.Commands,
	}
	return app

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
	app, err := auth.NewOAuthApp(cnf.Secret, dataPath, cnf.DB)
	if err != nil {
		log.Fatalf("Failed to init venus-auth: %s", err)
	}
	router := auth.InitRouter(app)
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
