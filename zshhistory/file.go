package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	history, err := getHistory("/Users/jamie/.zsh_history")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(len(history))
	fmt.Println(history[len(history)-100])
}

type HistoryEntry struct {
	Timestamp time.Time
	Command   string
}

func getHistory(fileName string) ([]*HistoryEntry, error) {
	var history []*HistoryEntry

	file, err := os.Open(fileName)
	if err != nil {
		return history, fmt.Errorf("opening %s: %v", fileName, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(splitFunc)

	for scanner.Scan() {
		line, err := parseLine(scanner.Bytes())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		history = append(history, line)
	}

	return history, nil
}

// parseLine parses the zsh history line. It assumes a particular format, eg:
// ": 1566311592:0;ls -l", the first number being a Unix timestamp
func parseLine(line []byte) (*HistoryEntry, error) {
	var res HistoryEntry

	timeX := strings.Split(line, ':')
	if len(timeX) < 2 {
		return &res, fmt.Errorf("line improperly formatted: %s", l)
	}

	time, err := strconv.Atoi(strings.TrimSpace(timeX[1]))
	if err != nil {
		return &res, fmt.Errorf("parsing date: %v", err)
	}

	res.Timestamp = time.Unix(int64(time), 0)
	res.Command = (l[strings.Index(l, ";")+1:])
	return &res, nil
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
