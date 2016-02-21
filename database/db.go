package database

import (
	"database/sql"
	"log"
	"net/url"

	"github.com/TF2Stadium/Helen/models"
	_ "github.com/lib/pq"
)

var (
	db *sql.DB
)

func Connect(dburl, database, username, password string) {
	DBUrl := url.URL{
		Scheme:   "postgres",
		Host:     dburl,
		Path:     database,
		RawQuery: "sslmode=disable",
	}

	log.Printf("Connecting to DB on %s", DBUrl.String())

	DBUrl.User = url.UserPassword(username, password)
	var err error

	db, err = sql.Open("postgres", DBUrl.String())
	if err != nil {
		log.Fatal(err)
	}
}

func IsAllowed(userid uint32, lobbyid uint) bool {
	var state int
	db.QueryRow("SELECT state FROM player_slots WHERE id = $1 AND lobby_id = $2", userid, lobbyid).Scan(&state)

	return state != 0
}

func IsLobbyClosed(lobbyid uint) bool {
	var state int
	db.QueryRow("SELECT state FROM lobbies where id = $1", lobbyid).Scan(&state)
	return state != int(models.LobbyStateEnded)
}

func GetSteamID(userid uint32) string {
	var steamid string
	db.QueryRow("SELECT steam_id FROM players WHERE id = $1", userid).Scan(&steamid)
	return steamid
}
