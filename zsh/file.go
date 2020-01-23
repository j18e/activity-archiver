package zsh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Entry struct {
	Time    time.Time
	Command string
}

func GetHistory(file io.Reader) ([]*Entry, error) {
	var history []*Entry

	scanner := bufio.NewScanner(file)
	scanner.Split(splitFunc)
	for scanner.Scan() {
		res, err := parseLine(scanner.Text())
		if err != nil {
			fmt.Fprint(os.Stderr, err)
		}
		history = append(history, res)
	}
	return history, nil
}

var (
	errNoMatch = errors.New("no regexp match found")
	lineRE     = regexp.MustCompile(`^: (\d+):\d;([\s\S]+)`)
	splitRE    = regexp.MustCompile(`[^\\](\n): \d`)
)

func parseLine(txt string) (*Entry, error) {
	matches := lineRE.FindStringSubmatch(txt)
	if len(matches) != 3 {
		return nil, errNoMatch
	}
	ts, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("parsing timestamp: %w", err)
	}
	return &Entry{
		Time:    time.Unix(int64(ts), 0),
		Command: strings.ReplaceAll(matches[2], "\n", "  "),
	}, nil
}

// splitFunc does the same thing as bufio.ScanLines except that it splits on
// splitRE rather than just newlines, and that it does not filter out carriage
// returns.
func splitFunc(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if m := splitRE.FindSubmatchIndex(data); len(m) == 4 {
		return m[3], data[0:m[3]], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
