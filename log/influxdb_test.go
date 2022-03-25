package log

import (
	"fmt"
	"os"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func TestInfluxDB(t *testing.T) {
	if os.Getenv("CI") == "test" {
		t.Skip()
	}
	// You can generate a Token from the "Tokens Tab" in the UI
	const token = "jcomkQ-dVBRoCrKSEWMuYxA4COj_EfyCvwgPW5Ql-tT-cCizIjE24rPJQNx8Kkqzz4gCW8YNFq0wcDaHJOcGMQ=="
	const bucket = "bkt2"
	const org = "venus-oauth"

	client := influxdb2.NewClient("http://192.168.1.141:8086", token)
	// always close client at the end
	defer client.Close()

	// get non-blocking write client
	writeAPI := client.WriteAPI(org, bucket)
	// write line protocol
	writeAPI.WriteRecord(fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0))
	writeAPI.WriteRecord(fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 22.5, 45.0))
	// Flush writes
	writeAPI.Flush()
}
