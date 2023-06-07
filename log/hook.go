package log

import (
	"fmt"
	"strconv"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/ipfs-force-community/sophon-auth/config"
	"github.com/ipfs-force-community/sophon-auth/core"
	"github.com/sirupsen/logrus"
)

const (
	empty = ""
	msg   = "message"
)

type InfluxHook struct {
	client   influxdb2.Client
	writeAPI api.WriteAPI
	tags     []string
}

func NewInfluxHook(c *config.InfluxDBConfig) *InfluxHook {
	cli := influxdb2.NewClientWithOptions(
		c.ServerURL,
		c.AuthToken,
		influxdb2.DefaultOptions().
			SetBatchSize(c.BatchSize).
			SetFlushInterval(uint(c.FlushInterval.Milliseconds())),
	)
	writeAPI := cli.WriteAPI(c.Org, c.Bucket)
	return &InfluxHook{
		client:   cli,
		writeAPI: writeAPI,
		tags:     core.TagFields,
	}
}

func (h *InfluxHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *InfluxHook) Fire(entry *logrus.Entry) error {
	tags := map[string]string{
		core.FieldLevel: entry.Level.String(),
	}
	measurement, ok := entry.Data[core.MTMethod]
	if !ok {
		return nil
	}
	delete(entry.Data, core.MTMethod)
	for _, tag := range h.tags {
		if tagValue, ok := getTag(entry.Data, tag); ok {
			tags[tag] = tagValue
		}
	}
	fields := make(map[string]interface{}, len(entry.Data)+1)
	fields[msg] = entry.Message
	for k, v := range entry.Data {
		fields[k] = v
	}
	for _, tag := range h.tags {
		delete(fields, tag)
	}

	pt := influxdb2.NewPoint(measurement.(string), tags, fields, entry.Time)
	h.writeAPI.WritePoint(pt)
	return nil
}

func getTag(fields logrus.Fields, tag string) (string, bool) {
	value, ok := fields[tag]
	if !ok {
		return empty, ok
	}
	switch vs := value.(type) {
	case fmt.Stringer:
		return vs.String(), ok
	case string:
		return vs, ok
	case byte:
		return string(vs), ok
	case int:
		return strconv.FormatInt(int64(vs), 10), ok
	case int32:
		return strconv.FormatInt(int64(vs), 10), ok
	case int64:
		return strconv.FormatInt(vs, 10), ok
	case uint:
		return strconv.FormatUint(uint64(vs), 10), ok
	case uint32:
		return strconv.FormatUint(uint64(vs), 10), ok
	case uint64:
		return strconv.FormatUint(vs, 10), ok
	default:
		return empty, false
	}
}
