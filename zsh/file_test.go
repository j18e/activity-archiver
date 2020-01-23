package zsh

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	for i, test := range []struct {
		ts  int
		cmd string
	}{
		{3600, "ls -lah"},
		{1573742894, ": 15:0;ls"},
		{1578912345, "some\ncommand"},
	} {
		line := fmt.Sprintf(`: %d:0;%s`, test.ts, test.cmd)
		res, err := parseLine(line)
		if err != nil {
			t.Errorf("test %d: %v", i, err)
			continue
		}
		if ts := time.Unix(int64(test.ts), 0); res.Time != ts {
			t.Errorf("test %d: wrong timestamp exp %v got %v", i, ts, res.Time)
		}
		if exp := strings.ReplaceAll(test.cmd, "\n", "  "); res.Command != exp {
			t.Errorf("test %d: wrong cmd expected %v got %v", i, exp, res.Command)
		}
	}

	for i, test := range []struct {
		txt string
		err error
	}{
		{"notamatch", errNoMatch},
		{":1234:0;ls", errNoMatch},
		{": 1234:0;ls", nil},
		{": 1234:0;ls\\\n~", nil},
	} {
		if _, err := parseLine(test.txt); err != test.err {
			t.Errorf("errTest %d: expected %v, got %v", i, test.err, err)
		}
	}
}

func TestSplitFunc(t *testing.T) {
	tests := []struct {
		bs  []byte
		exp int
	}{
		{[]byte("one\n: 15:0;ls"), 4},
		{[]byte("nothing"), 0},
		{[]byte("\\\n: 15;0;ls\n: 15:0;ls -l"), 12},
		{[]byte("\\\\\n: 15;0;ls\n: 15:0;ls -l"), 13},
	}

	for i, test := range tests {
		if adv, _, _ := splitFunc(test.bs, false); adv != test.exp {
			t.Errorf("splitFunc test %d exp %d, got %d", i, test.exp, adv)
		}
	}
}
