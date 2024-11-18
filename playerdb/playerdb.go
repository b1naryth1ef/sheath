/*
Package playerdb is a Go client for the PlayerDB API.

For information about the API itself and to inspect the terms and conditions of
its use please visit https://playerdb.co/.

# Usage

To lookup a minecraft user by name:

	package main

	import (
		"log"

		"github.com/b1naryth1ef/sheath/playerdb"
	)

	func main() {
		res, err := playerdb.GetMinecraftPlayer("brongle69")
		if err != nil {
			log.Panicf("Failed: %v", err)
		}

		if !res.Success {
			log.Panicf("Request failed: %+v", res)
		}

		log.Printf("UUID: %s", res.Data.Player.Id)
	}
*/
package playerdb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Response wraps a PlayerDB response payload of type T
type Response[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Data    T      `json:"data"`
}

// MinecraftPlayer represents a minecraft player as defined by the playerdb API
type MinecraftPlayer struct {
	Username    string `json:"username"`
	Id          string `json:"id"`
	RawId       string `json:"raw_id"`
	Avatar      string `json:"avatar"`
	SkinTexture string `json:"skin_texture"`
}

// MinecraftPlayerData wraps a MinecraftPlayer as it is returned from the API
type MinecraftPlayerData struct {
	Player MinecraftPlayer `json:"player"`
}

var client = &http.Client{}

// GetMinecraftPlayer fetches a minecraft player by name or ID, returning nil if
// no result is found for the given parameter.
func GetMinecraftPlayer(query string) (*Response[MinecraftPlayerData], error) {
	var result Response[MinecraftPlayerData]

	res, err := client.Get(fmt.Sprintf("https://playerdb.co/api/player/minecraft/%s", query))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return nil, nil
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 status code from playerdb: %v", res.StatusCode)
	}

	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
