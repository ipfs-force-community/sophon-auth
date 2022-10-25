package log

import (
	"os"
	"testing"
	"time"

	"github.com/filecoin-project/venus-auth/config"
	"github.com/sirupsen/logrus"
)

func TestLog(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}
	err := WithInflux(&config.InfluxDBConfig{
		ServerURL:     "http://192.168.1.141:8086",
		AuthToken:     "jcomkQ-dVBRoCrKSEWMuYxA4COj_EfyCvwgPW5Ql-tT-cCizIjE24rPJQNx8Kkqzz4gCW8YNFq0wcDaHJOcGMQ==",
		Org:           "venus-oauth",
		Bucket:        "bkt2",
		FlushInterval: time.Second,
		BatchSize:     20,
	})
	if err != nil {
		t.Fatal(err)
	}
	localLog.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true})
	localLog.SetOutput(os.Stdout)
	localLog.SetLevel(logrus.TraceLevel)
	localLog.WithFields(logrus.Fields{
		"method": "verify",
		"name":   "Rennbon",
		"ip":     "192.168.1.1",
		"level":  "1",
		"custom": "This field will not be tag",
	}).Trace("This is a trace id: 12345")

	localLog.WithFields(logrus.Fields{
		"method": "verify",
		"name":   "YangJian",
		"ip":     "192.168.1.2",
		"level":  "2",
		"from":   "RennbonTest",
	}).Error("This is an error")
	time.Sleep(time.Second * 2)
}
