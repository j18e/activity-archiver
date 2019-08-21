package mysql

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) Close() {
	s.db.Close()
}

type Config struct {
	User     string
	Password string
	Server   string
	Port     int
	DB       string
}

func NewStorage(c Config) (Storage, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", c.User, c.Password, c.Server, c.Port, c.DB)
	fmt.Printf("Connecting to database %s:%d as %s... ", c.Server, c.Port, c.User)

	for i := 1; i <= 5; i++ {
		if i > 1 {
			time.Sleep(time.Second * time.Duration(i*i))
		}
		db, err := sql.Open("mysql", connStr)
		if err != nil {
			fmt.Printf("attempt %d failed: %v\n", i, err)
			continue
		}

		_, err = db.Exec(zshCreateTable)
		if err != nil {
			fmt.Printf("attempt %d failed: %v\n", i, err)
			continue
		}

		fmt.Println("success")
		return Storage{db: db}, nil
	}
	return Storage{}, fmt.Errorf("couldn't connect to database - out of retries")
}
