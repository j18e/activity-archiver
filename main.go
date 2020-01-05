package main

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influx "github.com/influxdata/influxdb1-client/v2"
)

func main() {
}

type Client influx.Client

func influxStuff() error {
	addr := "http://localhost:8086"
	timeout := time.Second * 5
	cli, err := influx.NewHTTPClient(influx.HTTPConfig{Addr: addr, Timeout: timeout})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	// test connection to influxdb
	if _, _, err := cli.Ping(timeout); err != nil {
		log.Fatal(err)
	}

	bp, err := influx.NewBatchPoints(client.BatchPointsConfig{Database: DB_NAME, Precision: "s"})
	if err != nil {
		log.Fatal(err)
	}

	// define tags
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	tags := map[string]string{"device": "cpu-total"}
	fields := map[string]interface{}{
		"idle":   10.1,
		"system": 53.3,
		"user":   46.6,
	}
	pt, err := influx.NewPoint("cpu_usage", tags, fields, time.Now())
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := cli.Write(bp); err != nil {
		return err
	}

	return nil
}
