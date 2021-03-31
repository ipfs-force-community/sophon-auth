package log

import (
	"github.com/ipfs-force-community/venus-auth/config"
	"github.com/sirupsen/logrus"
	"os"
)

func InitLog(c *config.LogConfig) {
	localLog = logrus.New()
	localLog.SetFormatter(&logrus.TextFormatter{ForceColors: true, FullTimestamp: true})
	localLog.SetOutput(os.Stdout)
	lvl, err := logrus.ParseLevel(c.LogLevel)
	if err != nil {
		lvl = logrus.TraceLevel
	}
	localLog.SetLevel(lvl)
	if c.HookSwitch {
		err = WithInflux(c.InfluxDB)
		if err != nil {
			panic(err)
		}
	}
}
func WithInflux(c *config.InfluxDBConfig) error {
	hook := NewInfluxHook(c)
	localLog.AddHook(hook)
	return nil
}
