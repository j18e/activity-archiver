package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/j18e/backerupper/pkg/handlers"
	"github.com/j18e/backerupper/pkg/storage/mysql"
	"github.com/j18e/backerupper/pkg/zsh"
)

func main() {
	c, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	stor, err := mysql.NewStorage(mysql.Config{
		Server:   c.mysqlServer,
		Port:     c.mysqlPort,
		DB:       c.mysqlDB,
		User:     c.mysqlUser,
		Password: c.mysqlPassword,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stor.Close()

	z := zsh.Service(&stor)

	router := gin.Default()

	router.GET("/zsh/last", handlers.GetLastCMD(z))
	router.POST("/zsh/list", handlers.WriteCMDs(z))

	router.Run(c.listenAddr)
}

type config struct {
	listenAddr string

	mysqlUser     string
	mysqlPassword string
	mysqlServer   string
	mysqlPort     int
	mysqlDB       string
}

// getConfig parses command line arguments for mysql and http configuration
// parameters
func getConfig() (config, error) {
	var c config

	httpPort := flag.Int("http.port", 8080, "port on which to run the HTTP server")
	httpAddr := flag.String("http.address", "", "IP or hostname on which to listen for HTTP requests")
	mysqlUser := flag.String("mysql.user", "", "MYSQL user to connect as")
	mysqlPassword := flag.String("mysql.password", "", "password for the MYSQL user")
	mysqlServer := flag.String("mysql.server", "", "MYSQL server to connect to")
	mysqlPort := flag.Int("mysql.port", 3306, "port on the MYSQL server to connect to")
	mysqlDB := flag.String("mysql.db", "", "MYSQL database to use")

	flag.Parse()

	notSet := "flag %s not set"

	if *mysqlUser == "" {
		return c, fmt.Errorf(notSet, "-mysql.user")
	} else if *mysqlPassword == "" {
		return c, fmt.Errorf(notSet, "-mysql.password")
	} else if *mysqlServer == "" {
		return c, fmt.Errorf(notSet, "-mysql.server")
	} else if *mysqlDB == "" {
		return c, fmt.Errorf(notSet, "-mysql.db")
	}

	c = config{
		listenAddr:    fmt.Sprintf("%s:%d", *httpAddr, *httpPort),
		mysqlServer:   *mysqlServer,
		mysqlPort:     *mysqlPort,
		mysqlDB:       *mysqlDB,
		mysqlUser:     *mysqlUser,
		mysqlPassword: *mysqlPassword,
	}

	return c, nil
}
