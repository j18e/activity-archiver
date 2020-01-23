package firefox

import (
	"database/sql"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3" // used by database/sql
)

// Entry is an entry in the browser page view history for a given Firefox
// profile.
type Entry struct {
	Time time.Time
	URL  string
}

// History creates and opens a temporary copy of the sqlite db file and creates
// a list of Firefox history entries, given a Firefox profile.
func (p *Profile) History() ([]*Entry, error) {
	tmpFile, err := getTmpFile(p.dbFile)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	places, err := getPlaces(db)
	if err != nil {
		return nil, err
	}

	visits, err := getVisits(db)
	if err != nil {
		return nil, err
	}

	entries := make([]*Entry, len(visits))
	for i, v := range visits {
		entries[i] = &Entry{
			Time: v.time,
			URL:  places[v.place],
		}
	}
	return entries, nil
}

func getPlaces(db *sql.DB) (map[int]string, error) {
	places := make(map[int]string)
	rows, err := db.Query("SELECT id, url FROM moz_places")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var url string
		if err := rows.Scan(&id, &url); err != nil {
			return nil, err
		}
		places[id] = url
	}
	return places, nil
}

type visit struct {
	time  time.Time
	place int
}

func getVisits(db *sql.DB) ([]*visit, error) {
	var visits []*visit
	rows, err := db.Query("SELECT visit_date, place_id FROM moz_historyvisits")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ts int64
		var id int
		if err := rows.Scan(&ts, &id); err != nil {
			return nil, err
		}
		visits = append(visits, &visit{
			time:  time.Unix(ts/1000000, 0), // timestamp is in microseconds
			place: id,
		})
	}
	return visits, nil
}
