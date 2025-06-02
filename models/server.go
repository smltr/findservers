package models

import (
	"strings"
	"time"
)

//
// Example CS2 server returned from the Steam Web API
// {
//     "addr": "102.216.74.10:27015",
//     "gameport": 27015,
//     "steamid": "90259233304408080",
//     "name": "RapidNetworks Counter-Strike 2 Server",
//     "appid": 730,
//     "gamedir": "csgo",
//     "version": "1.40.6.7",
//     "product": "cs2",
//     "region": 1,
//     "players": 0,
//     "max_players": 32,
//     "bots": 0,
//     "map": "de_dust2",
//     "secure": true,
//     "dedicated": true,
//     "os": "l",
//     "gametype": "empty,secure"
// },
//
// Region codes:
// -1 - US this isn't specified, but upon manually testing all IPs with this code are in the US
// 0 - US East
// 1 - US West
// 2 - South America
// 3 - Europe
// 4 - Asia
// 5 - Australia
// 6 - Middle East
// 7 - Africa

type ServerRaw struct {
	Addr       string    `json:"addr"`
	GamePort   int       `json:"gameport"`
	SteamID    string    `json:"steamid"`
	Name       string    `json:"name"`
	AppID      int       `json:"appid"`
	GameDir    string    `json:"gamedir"`
	Version    string    `json:"version"`
	Product    string    `json:"product"`
	Region     int       `json:"region"`
	Players    int       `json:"players"`
	MaxPlayers int       `json:"max_players"`
	Bots       int       `json:"bots"`
	Map        string    `json:"map"`
	Secure     bool      `json:"secure"`
	Dedicated  bool      `json:"dedicated"`
	OS         string    `json:"os"`
	GameType   string    `json:"gametype"`
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
}

type Server struct {
	IP         string   `json:"ip"`
	Port       int      `json:"port"`
	Address    string   `json:"address"`
	Name       string   `json:"name"`
	Region     int      `json:"region"`
	Players    int      `json:"players"`
	MaxPlayers int      `json:"max_players"`
	Bots       int      `json:"bots"`
	Map        string   `json:"map"`
	Secure     bool     `json:"secure"`
	Dedicated  bool     `json:"dedicated"`
	Tags       []string `json:"tags"`
}

func CleanServer(s ServerRaw) Server {
	return Server{
		IP:         strings.Split(s.Addr, ":")[0],
		Port:       s.GamePort,
		Address:    s.Addr,
		Name:       s.Name,
		Region:     s.Region,
		Players:    s.Players,
		MaxPlayers: s.MaxPlayers,
		Bots:       s.Bots,
		Map:        s.Map,
		Secure:     s.Secure,
		Dedicated:  s.Dedicated,
		Tags:       strings.Split(s.GameType, ","),
	}
}
