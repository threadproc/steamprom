package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type steamResponse struct {
	Response struct {
		GameCount int          `json:"game_count"`
		Games     []*steamGame `json:"games"`
	}
}

type steamGame struct {
	AppID           int    `json:"appid"`
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtime_forever"`
	PlaytimeWindows int    `json:"playtime_windows_forever"`
	PlaytimeMac     int    `json:"playtime_mac_forever"`
	PlaytimeLinux   int    `json:"playtime_linux_forever"`
}

func (g *steamGame) playtimeWindows() int {
	return g.PlaytimeForever - g.PlaytimeLinux - g.PlaytimeMac
}

func (g *steamGame) writeMetric(w io.Writer, platform string, amount int, steamid string) {
	w.Write([]byte("steam_game_playtime{appid=\"" + strconv.Itoa(g.AppID) + "\", platform=\"" + platform + "\", steamid=\"" + steamid + "\", name=\"" + strings.ReplaceAll(g.Name, "\"", "'") + "\"} " + strconv.Itoa(amount) + "\n"))
}

func main() {
	listen := os.Getenv("PORT")
	if listen == "" {
		listen = "8080"
	}
	listen = ":" + listen

	r := mux.NewRouter()
	r.HandleFunc("/id/{apikey}/{steamid}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		steamID := vars["steamid"]
		apiKey := vars["apikey"]

		_, err := strconv.ParseInt(steamID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		url := "https://api.steampowered.com/IPlayerService/GetOwnedGames/v1/?key=" + apiKey + "&steamid=" + steamID + "&include_appinfo=true&include_played_free_games=true&include_free_sub=false&include_extended_appinfo=false"
		resp, err := http.Get(url)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if resp.StatusCode != 200 {
			w.WriteHeader(resp.StatusCode)
		}

		bdbs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			if resp.StatusCode == 200 {
				// We can only set the status code once, we if we previously set
				// it to something bad and then failed to read the body, we just
				// lose that error to the ether
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write([]byte(err.Error()))
			return
		}

		stResp := steamResponse{}
		if err := json.Unmarshal(bdbs, &stResp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		sort.Slice(stResp.Response.Games, func(i, j int) bool {
			return stResp.Response.Games[i].PlaytimeForever > stResp.Response.Games[j].PlaytimeForever
		})

		for _, game := range stResp.Response.Games {
			w.Write([]byte("# " + game.Name + "\n"))

			game.writeMetric(w, "windows", game.playtimeWindows(), steamID)
			game.writeMetric(w, "mac", game.PlaytimeMac, steamID)
			game.writeMetric(w, "linux", game.PlaytimeLinux, steamID)

			w.Write([]byte("\n"))
		}
	})

	fmt.Println("Listening on " + listen)
	http.ListenAndServe(listen, r)
}
