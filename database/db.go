package database

import (
	"database/sql"
	"log"
	"net/url"
	"strings"

	"github.com/TF2Stadium/Helen/helpers"
	"github.com/TF2Stadium/Helen/helpers/authority"
	"github.com/TF2Stadium/Helen/models/lobby"
	"github.com/TF2Stadium/Helen/models/lobby/format"
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
	team, _, err := format.GetSlotTeamClass(format.Format(lobbyType), slot)
	if err != nil {
		log.Println(err)
	}

	if team != strings.ToLower(channelname) {
		return false, "You're in team " + strings.ToUpper(team) + ", not " + channelname
	}

	return true, ""
}

func GetSteamID(userid uint32) string {
	var steamid string
	db.QueryRow("SELECT steam_id FROM players WHERE id = $1", userid).Scan(&steamid)
	return steamid
}

func IsAdmin(userid uint32) bool {
	var role authority.AuthRole
	err := db.QueryRow("SELECT role FROM players WHERE id = $1", userid).Scan(&role)
	if err != nil {
		log.Println(err)
		return false
	}
	return role == helpers.RoleAdmin || role == helpers.RoleMod
}

func GetCurrentLobby(userid uint32) (uint, string) {
	var lobbyID uint
	var slot int
	var lobbyFormat format.Format

	err := db.QueryRow("SELECT lobby_slots.lobby_id, lobby_slots.slot, lobbies.type FROM lobbies INNER JOIN lobby_slots ON lobbies.id = lobby_slots.lobby_id WHERE lobby_slots.player_id = $1 AND lobbies.state <> $2 AND lobbies.state <> $3", userid, lobby.Ended, lobby.Initializing).Scan(&lobbyID, &slot, &lobbyFormat)
	if err != nil && err != sql.ErrNoRows { // if err == ErrNoRows, player isn't in any active lobby
		log.Println(err)
	}

	team, _, _ := format.GetSlotTeamClass(lobbyFormat, slot)

	return lobbyID, team
}
