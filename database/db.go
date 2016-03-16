package database

import (
	"database/sql"
	"log"
	"net/url"
	"strings"

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

func IsAllowed(userid uint32, lobbyid uint, channelname string) (bool, string) {
	var lobbyType, slot int
	db.QueryRow("SELECT type FROM lobbies WHERE id = $1", lobbyid).Scan(&lobbyType)
	err := db.QueryRow("SELECT slot FROM lobby_slots WHERE player_id = $1 AND lobby_id = $2", userid, lobbyid).Scan(&slot)
	if err == sql.ErrNoRows {
		return false, "You're not in this lobby"
	} else if err != nil {
		log.Println(err)
		return false, "Internal fumble error"
	}

	if channelname[0] == 'L' { // channel name is "Lobby..."
		return true, ""
	}

	//channel name is either "RED" or "BLU"
	team, _, err := models.LobbyGetSlotInfoString(models.LobbyType(lobbyType), slot)
	if err != nil {
		log.Println(err)
	}

	if team != strings.ToLower(channelname) {
		return false, "You're in team " + strings.ToUpper(team) + ", not " + channelname
	}

	return true, ""
}

func IsLobbyClosed(lobbyid uint) bool {
	var state int
	db.QueryRow("SELECT state FROM lobbies where id = $1", lobbyid).Scan(&state)
	return state != 5
}

func GetSteamID(userid uint32) string {
	var steamid string
	db.QueryRow("SELECT steam_id FROM players WHERE id = $1", userid).Scan(&steamid)
	return steamid
}
