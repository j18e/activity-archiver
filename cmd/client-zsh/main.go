package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/j18e/backerupper/zsh"
)

func main() {
	serverAddr := flag.String("server.address", "", "address of the backup server. Example: http://localhost:8080")
	flag.Parse()

	startTime := time.Now()
	histFileName := os.Getenv("HOME") + `/.zsh_history`
	batchLimit := 500

	serverURL, err := testConn(serverAddr)
	if err != nil {
		fatalErr(err)
	}

	history, errs, err := getHistory(histFileName)
	if err != nil {
		fatalErr(err)
	}

	if len(errs) > 0 {
		fmt.Println(len(errs), "errors while reading history")
	}

	fmt.Printf("processing %d events...\n", len(history))
	for len(history) > batchLimit {
		if err := postList(serverURL, history[:batchLimit-1]); err != nil {
			fatalErr(err)
		}
		history = history[batchLimit:]
		fmt.Printf("%d remaining...\n", len(history))
	}
	if err := postList(serverURL, history); err != nil {
		fatalErr(err)
	}

	fmt.Printf("finished in %v\n", time.Since(startTime))
}

func fatalErr(err error) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	os.Exit(1)
}

func postList(serverURL url.URL, history []zsh.Command) error {
	cli := &http.Client{Timeout: time.Second * 5}

	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(history)

	uri := fmt.Sprintf(`%s/zsh/list`, serverURL.String())

	req, err := http.NewRequest(http.MethodPost, uri, buf)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "github.com/j18e/backerupper client-zsh")

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		msg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("%d - reading response: %v", resp.StatusCode, err)
		}
		return fmt.Errorf("%d - %s", resp.StatusCode, msg)
	}
	return nil
}

// testConn ensures the availability of the server
func testConn(addr *string) (url.URL, error) {
	var serverURL url.URL
	timeout := time.Second * 5

	// parse server address
	if *addr == "" {
		fmt.Println("in testconn")
		return serverURL, fmt.Errorf("flag -server.address not set")
	}

	// parse address
	ptr, err := url.Parse(*addr)
	if err != nil {
		return serverURL, fmt.Errorf("parsing server address: %v", err)
	}
	serverURL = *ptr

	// test TCP connection to server
	conn, err := net.DialTimeout("tcp", serverURL.Host, timeout)
	if err != nil {
		return serverURL, fmt.Errorf("could not connect to %s: %v", serverURL.Host, err)
	}
	conn.Close()
	return serverURL, nil
}

func getHistory(fileName string) ([]zsh.Command, []string, error) {
	var history []zsh.Command
	var errs []string

	file, err := os.Open(fileName)
	if err != nil {
		return history, errs, fmt.Errorf("opening %s: %v", fileName, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(splitFunc)

	for i := 1; scanner.Scan(); i++ {
		line, err := parseLine(scanner.Text())
		if err != nil {
			errs = append(errs, fmt.Sprintf("%d: %v", i, err))
			continue
		}
		history = append(history, line)
	}

	return history, errs, nil
}

// parseLine parses the zsh history line. It assumes a particular format, eg:
// ": 1566311592:0;ls -l", the first number being a Unix timestamp
func parseLine(l string) (zsh.Command, error) {
	var cmd zsh.Command

	slice := strings.Split(l, ":")
	if len(slice) < 2 {
		return cmd, fmt.Errorf("line improperly formatted: %s", l)
	}

	dateInt, err := strconv.Atoi(strings.TrimSpace(slice[1]))
	if err != nil {
		return cmd, fmt.Errorf("parsing date: %v", err)
	}

	cmd.Time = int64(dateInt)
	cmd.Command = (l[strings.Index(l, ";")+1:])
	return cmd, nil
}

// splitFunc does the same thing as bufio.ScanLines except that it splits on
// "\n: " rather than just newlines, and that it does not filter out carriage
// returns.
func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte("\n: ")); i >= 0 {
		return i + 1, data[0:i], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
