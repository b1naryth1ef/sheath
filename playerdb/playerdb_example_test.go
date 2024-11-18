package playerdb_test

import (
	"fmt"
	"log"

	"github.com/b1naryth1ef/sheath/playerdb"
)

func ExampleGetMinecraftPlayer() {
	res, err := playerdb.GetMinecraftPlayer("brongle69")
	if err != nil {
		log.Panicf("Failed: %v", err)
	}

	if !res.Success {
		log.Panicf("Request failed: %+v", res)
	}

	fmt.Printf("UUID: %s\n", res.Data.Player.Id)
	// Output: UUID: 699b607b-8331-4009-b3c0-a2470a16f53d
}
