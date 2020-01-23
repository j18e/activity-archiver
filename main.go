package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influx "github.com/influxdata/influxdb1-client/v2"
	"github.com/j18e/activity-archiver/firefox"
	"github.com/j18e/activity-archiver/zsh"
	log "github.com/sirupsen/logrus"
)

const (
	DB_NAME        = "activity"
	ZSH_METRIC     = "zsh_history"
	FIREFOX_METRIC = "firefox_history"
)

func main() {
	// init
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("getting hostname: %v", err)
	}
	log.Infof("using hostname %s", hostname)

	idbAddr := flag.String("influx.url", "", "url to the influxdb server")
	idbUser := flag.String("influx.user", "", "user to the influxdb server")
	idbPass := flag.String("influx.password", "", "password to the influxdb server")
	flag.Parse()
	if *idbAddr == "" {
		log.Fatal("required flag -influx.url")
	}

	cli, err := NewClient(influx.HTTPConfig{Addr: *idbAddr, Username: *idbUser, Password: *idbPass}, DB_NAME)
	if err != nil {
		log.Fatalf("connecting to influxdb: %v", err)
	}

	// firefox
	profiles, err := firefox.GetProfiles()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range profiles {
		history, err := p.History()
		if err != nil {
			log.Fatal(err)
		}
		last, err := cli.LastFirefoxEntry(hostname, p.Name)
		if err != nil {
			log.Errorf("getting last firefox entry for profile %s from influxdb: %v", p.Name, err)
			continue
		}
		payload := make([]*firefox.Entry, 0, len(history))
		for _, e := range history {
			if last == nil {
				payload = history
				break
			}
			if e.Time.Before(last.Time) || *last == *e {
				continue
			}
			payload = append(payload, e)
		}

		if len(payload) < 1 {
			log.Infof("no new firefox entries from profile %s", p.Name)
			continue
		}
		sent, err := cli.PostFirefoxEntries(hostname, p.Name, payload)
		if err != nil {
			log.Errorf("posting firefox history for profile %s to influxdb: %v", p.Name, err)
			continue
		}
		log.Infof("successfully posted %d firefox history entries from profile %s to influxdb", sent, p.Name)
	}

	// zsh
	zshHistory := mustGetZshHistory(filepath.Join(os.Getenv("HOME"), ".zsh_history"))

	lastZSH, err := cli.LastZSHEntry(hostname)
	if err != nil {
		log.Fatalf("getting last zsh entry from influxdb: %v", err)
	}
	zshPayload := make([]*zsh.Entry, 0, len(zshHistory))
	for _, e := range zshHistory {
		if lastZSH != nil && e.Time.Before(lastZSH.Time) {
			continue
		}
		zshPayload = append(zshPayload, e)
	}

	if len(zshPayload) < 1 {
		log.Info("no new zsh history entries. Exiting...")
		return
	}

	sent, err := cli.PostZSHEntries(hostname, zshPayload)
	if err != nil {
		log.Fatalf("posting zsh entries to influxdb: %v", err)
	}

	log.Infof("successfully posted %d zsh history entries to influxdb", sent)
}

func mustGetZshHistory(fileName string) []*zsh.Entry {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".zsh_history"))
	if err != nil {
		log.Fatalf("opening zsh history: %v", err)
	}
	defer file.Close()

	zshHistory, err := zsh.GetHistory(file)
	if err != nil {
		log.Fatalf("getting zsh history: %v", err)
	}
	return zshHistory
}

func NewClient(conf influx.HTTPConfig, dbName string) (*Client, error) {
	const testConnStr = "SELECT last(value) FROM nonexistent"
	idb, err := influx.NewHTTPClient(conf)
	if err != nil {
		return nil, err
	}
	defer idb.Close()

	// test connection to influxdb
	if _, _, err := idb.Ping(conf.Timeout); err != nil {
		return nil, err
	}
	res, err := idb.Query(influx.NewQuery(testConnStr, dbName, "s"))
	if err != nil {
		return nil, err
	}
	if err := res.Error(); err != nil {
		return nil, err
	}
	return &Client{Client: idb, dbName: dbName}, nil
}

type Client struct {
	influx.Client
	dbName string
}

func (c *Client) LastFirefoxEntry(hostname, profile string) (*firefox.Entry, error) {
	qs := fmt.Sprintf("SELECT last(url) FROM %s WHERE device = '%s' AND profile = '%s'",
		FIREFOX_METRIC, hostname, profile)
	res, err := c.Query(influx.NewQuery(qs, c.dbName, "s"))
	if err != nil {
		return nil, err
	}
	if err := res.Error(); err != nil {
		return nil, err
	}
	if len(res.Results) < 1 {
		return nil, errors.New("no results found")
	}
	if len(res.Results[0].Series) < 1 {
		return nil, nil
	}
	vals := res.Results[0].Series[0].Values
	if len(vals) < 1 || len(vals[0]) < 2 {
		return nil, fmt.Errorf("format not recognized: %v", vals)
	}

	ts, err := vals[0][0].(json.Number).Int64()
	if err != nil {
		return nil, fmt.Errorf("getting timestamp: %v", err)
	}

	url, ok := vals[0][1].(string)
	if !ok {
		return nil, fmt.Errorf("converting interface{} to string: %v", vals[0][1])
	}

	return &firefox.Entry{Time: time.Unix(ts, 0), URL: url}, nil
}

func (c *Client) PostFirefoxEntries(hostname, profile string, entries []*firefox.Entry) (int, error) {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{Database: c.dbName, Precision: "s"})
	if err != nil {
		return 0, err
	}

	tags := map[string]string{"device": hostname, "profile": profile}
	count := 0
	for _, e := range entries {
		pt, err := influx.NewPoint(FIREFOX_METRIC, tags,
			map[string]interface{}{"url": e.URL},
			e.Time,
		)
		if err != nil {
			log.Errorf("creating datapoint: %v", err)
			continue
		}
		bp.AddPoint(pt)
		count++
	}

	// Write the batch
	if err := c.Write(bp); err != nil {
		return 0, err
	}
	return count, nil
}

func (c *Client) LastZSHEntry(hostname string) (*zsh.Entry, error) {
	res, err := c.Query(influx.NewQuery(
		fmt.Sprintf("SELECT last(command) FROM %s WHERE device = '%s'", ZSH_METRIC, hostname),
		c.dbName, "s"),
	)
	if err != nil {
		return nil, err
	}
	if err := res.Error(); err != nil {
		return nil, err
	}
	if len(res.Results) < 1 {
		return nil, errors.New("no results found")
	}
	if len(res.Results[0].Series) < 1 {
		return nil, nil
	}
	vals := res.Results[0].Series[0].Values
	if len(vals) < 1 || len(vals[0]) < 2 {
		return nil, fmt.Errorf("format not recognized: %v", vals)
	}

	ts, err := vals[0][0].(json.Number).Int64()
	if err != nil {
		return nil, fmt.Errorf("getting timestamp: %v", err)
	}

	cmd, ok := vals[0][1].(string)
	if !ok {
		return nil, fmt.Errorf("converting interface{} to string: %v", vals[0][1])
	}

	return &zsh.Entry{Time: time.Unix(ts, 0), Command: cmd}, nil
}

func (c *Client) PostZSHEntries(hostname string, entries []*zsh.Entry) (int, error) {
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{Database: c.dbName, Precision: "s"})
	if err != nil {
		return 0, err
	}

	tags := map[string]string{"device": hostname}
	count := 0
	for _, e := range entries {
		pt, err := influx.NewPoint(ZSH_METRIC, tags,
			map[string]interface{}{"command": e.Command},
			e.Time,
		)
		if err != nil {
			log.Errorf("creating datapoint: %v", err)
			continue
		}
		bp.AddPoint(pt)
		count++
	}

	// Write the batch
	if err := c.Write(bp); err != nil {
		return 0, err
	}
	return count, nil
}
