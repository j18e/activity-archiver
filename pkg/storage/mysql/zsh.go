package mysql

import (
	"crypto/md5"
	"fmt"

	"github.com/j18e/backerupper/zsh"
)

const (
	zshCreateTable = `CREATE TABLE IF NOT EXISTS zsh (
		md5sum VARCHAR(32) NOT NULL, time BIGINT NOT NULL,
		command VARCHAR(256) NOT NULL, PRIMARY KEY (md5sum))`
	zshLastCMDQuery   = `SELECT time, command FROM zsh WHERE time=(SELECT max(time) FROM zsh) LIMIT 0, 1`
	zshWriteCMDsQuery = `INSERT INTO zsh (md5sum, time, command) VALUES(?,?,?)`
)

func (s *Storage) LastCMD() (zsh.Command, error) {

	var c zsh.Command
	var cmd string
	var ts int64

	row := s.db.QueryRow(zshLastCMDQuery)
	err := row.Scan(&ts, &cmd)

	if err != nil && err.Error() == "sql: no rows in result set" {
		return c, nil
	} else if err != nil {
		return c, err
	}

	c = zsh.Command{
		Command: cmd,
		Time:    ts,
	}

	return c, nil
}

// WriteCMDs writes a given slice of zsh.Command to the database, provided the
// list is below the limit
func (s *Storage) WriteCMDs(cx []zsh.Command) error {
	ins, err := s.db.Prepare(zshWriteCMDsQuery)
	if err != nil {
		return fmt.Errorf("db.Prepare: %v", err)
	}

	for _, c := range cx {
		ins.Exec(md5sum(c), c.Time, c.Command)
	}
	return nil
}

// md5sum returns a stringified md5 hash of the zsh.Command. This is to assign
// a unique id for storing in the database, so that no duplicate entries will
// be stored.
func md5sum(c zsh.Command) string {
	s := fmt.Sprintf("%d%s", c.Time, c.Command)
	b := []byte(s)
	return fmt.Sprintf("%x", md5.Sum(b))
}
