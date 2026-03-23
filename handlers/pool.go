package handlers

import (
	"math/rand/v2"
	"strings"
	"sync"

	"github.com/diego-oniarti/minesV2/utils"
)

var (
	lobbyPool = make(map[string]utils.Lobby)
	lobbyMux  = new(sync.Mutex)
)

func randCode() string {
	const letters = "abcdefghijklmnopqrstuvwxysABCDEFGHIJKLMNOPQRSTUVWXYS0123456789"
	var sb strings.Builder
	for {
		for range 6 {
			sb.WriteByte( letters[rand.IntN(len(letters))] )
		}
		lobbyMux.Lock()
		if _, ok := lobbyPool[sb.String()]; !ok {
			lobbyMux.Unlock()
			return sb.String()
		}
		lobbyMux.Unlock()
		sb.Reset()
	}
}
