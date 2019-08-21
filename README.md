# Backerupper

This app is both server and client, and is geared towards archiving data from
workstations and mobile phones, to a server with a database. The archived data
can be viewed in time series using Grafana. Currently ZSH command history is the
only supported data source, but Firefox browsing history and SMS messages from
Android phones are in the works.

## Usage
To start the server:
```
go run cmd/server/main.go -merchant.name=osl -http.address=localhost \
    -storage.type=mysql -mysql.server=localhost -mysql.db=testmock \
    -mysql.user=root -mysql.password=supersecret \
```

To run the client:
```
go run cmd/client-zsh/main.go -server.address=http://localhost:8080
```
