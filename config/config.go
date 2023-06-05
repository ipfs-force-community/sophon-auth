package config

import (
	"bytes"
	"crypto/rand"
	"io"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ipfs-force-community/metrics"
	"golang.org/x/xerrors"
)

type Config struct {
	Listen       string               `json:"listen"`
	ReadTimeout  time.Duration        `json:"readTimeout"`
	WriteTimeout time.Duration        `json:"writeTimeout"`
	IdleTimeout  time.Duration        `json:"idleTimeout"`
	Log          *LogConfig           `json:"log"`
	DB           *DBConfig            `json:"db"`
	Trace        *metrics.TraceConfig `json:"traceConfig"`
}

type DBType = string

const (
	Mysql  DBType = "mysql"
	Badger DBType = "badger"
)

type DBConfig struct {
	Type         DBType        `json:"type"`
	DSN          string        `json:"dsn"`
	MaxOpenConns int           `json:"maxOpenConns"`
	MaxIdleConns int           `json:"maxIdleConns"`
	MaxLifeTime  time.Duration `json:"maxLifeTime"`
	MaxIdleTime  time.Duration `json:"maxIdleTime"`
	Debug        bool          `json:"debug"`
}

// RandSecret If the daemon does not have a secret key configured, it is automatically generated
func RandSecret() ([]byte, error) {
	sk, err := io.ReadAll(io.LimitReader(rand.Reader, 32))
	if err != nil {
		return nil, err
	}
	return sk, nil
}

func DefaultConfig() *Config {
	return &Config{
		Listen:       "127.0.0.1:8989",
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
		IdleTimeout:  time.Minute,
		Trace: &metrics.TraceConfig{
			JaegerTracingEnabled: false,
			ProbabilitySampler:   1.0,
			JaegerEndpoint:       "localhost:6831",
			ServerName:           "sophon-auth",
		},
		Log: &LogConfig{
			LogLevel:   "trace",
			HookSwitch: false,
		},
		DB: &DBConfig{
			Type:         Badger,
			MaxOpenConns: 64,
			MaxIdleConns: 128,
			MaxLifeTime:  120 * time.Second,
			MaxIdleTime:  60 * time.Second,
		},
	}
}

type LogHookType = int

const (
	LHTInfluxDB LogHookType = 1
)

type LogConfig struct {
	LogLevel   string          `json:"logLevel"`
	Type       LogHookType     `json:"type"`
	HookSwitch bool            `json:"hookSwitch"`
	InfluxDB   *InfluxDBConfig `json:"influxdb"`
}

type InfluxDBConfig struct {
	ServerURL     string        `json:"serverURL"`
	AuthToken     string        `json:"authToken"`
	Org           string        `json:"org"`
	Bucket        string        `json:"bucket"`
	FlushInterval time.Duration `json:"flushInterval"`
	BatchSize     uint          `json:"batchSize"`
}

func MakeDir(path string) error {
	err := os.Mkdir(path, 0o755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func DecodeConfig(path string) (c *Config, err error) {
	provider, err := FromConfigString(path, "toml")
	if err != nil {
		return nil, err
	}
	c = new(Config)
	err = provider.Unmarshal(c)
	if err != nil {
		return nil, err
	}
	return
}

func Cover(path string, config *Config) error {
	c, err := os.Create(path)
	if err != nil {
		return err
	}
	barr, err := config.toBytes()
	if err != nil {
		return err
	}
	_, err = c.Write(barr)
	if err != nil {
		return xerrors.Errorf("write config: %w", err)
	}
	if err := c.Close(); err != nil {
		return xerrors.Errorf("close config: %w", err)
	}
	return nil
}

func (c *Config) toBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	e := toml.NewEncoder(buf)
	if err := e.Encode(c); err != nil {
		return nil, xerrors.Errorf("encoding config: %w", err)
	}
	b := buf.Bytes()
	b = bytes.ReplaceAll(b, []byte("#["), []byte("["))
	return b, nil
}
