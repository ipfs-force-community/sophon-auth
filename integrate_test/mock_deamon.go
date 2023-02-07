package integrate

import (
	"io/ioutil"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/filecoin-project/venus-auth/auth"
	"github.com/filecoin-project/venus-auth/config"
	"github.com/filecoin-project/venus-auth/log"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-homedir"
)

func setup(t *testing.T) (server *httptest.Server, dir string, token string) {
	tempDir, err := ioutil.TempDir("/var/tmp", "venus-auth")
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("create storage temp dir: %s", tempDir)

	cnf, err := config.DefaultConfig()
	if err != nil {
		t.Fatal(err)
	}
	cnf.DB.DSN = tempDir

	dir, err = homedir.Expand(tempDir)
	if err != nil {
		t.Fatalf("could not expand repo location error:%s", err)
	} else {
		log.Infof("venus repo: %s", dir)
	}
	gin.SetMode(gin.DebugMode)
	dataPath := path.Join(dir, "data")

	app, err := auth.NewOAuthApp(cnf.Secret, dataPath, cnf.DB)
	if err != nil {
		t.Fatalf("Failed to init venus-auth: %s", err)
	}
	token, err = app.GetDefaultAdminToken()
	if err != nil {
		t.Fatalf("Failed to get default admin token: %s", err)
	}
	cnf.Token = token

	router := auth.InitRouter(app, true )
	srv := httptest.NewServer(router)
	return srv, tempDir, token
}

func shutdown(t *testing.T, tempDir string) {
	log.Infof("shutdown, remove dir %s", tempDir)
	err := os.RemoveAll(tempDir)
	if err != nil {
		t.Fatal(err)
	}
}
