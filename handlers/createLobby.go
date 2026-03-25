package handlers

import (
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/diego-oniarti/minesV2/utils"
	"github.com/gin-gonic/gin"
)

var uglyMap = map[string]utils.GameMode{
	"singleplayer": utils.Singleplayer,
	"1v1": utils.Vs,
}

func CreateLobby(c *gin.Context) {
	var form struct {
		Mode   string `form:"mode" binding:"required"`
		Timed  bool   `form:"timed" binding:"required"`
		Time   int    `form:"time"`
		Width  int    `form:"width" binding:"required"`
		Height int    `form:"height" binding:"required"`
		Bombs  int    `form:"bombs" binding:"required"`
	}
	c.ShouldBind(&form)

	code := randCode()
	opts := utils.GameOptions{
		Mode: uglyMap[form.Mode],
		Timed: form.Timed,
		Time: form.Time,
		Width: form.Width,
		Height: form.Height,
		Bombs: form.Bombs,
		Seed: rand.Int(),
		Code: code,
	}

	createLobby(opts)

	c.String(http.StatusOK, code)
}

func createLobby(opts utils.GameOptions) {
	var game utils.GameHandler
	if opts.Mode == utils.Vs {
		game = new1v1(opts)
	}else{
		game = newSingle(opts)
	}

	newLobby := utils.Lobby{
		Opts: opts,
		Game: game,
	}

	lobbyMux.Lock()
	lobbyPool[opts.Code] = newLobby
	lobbyMux.Unlock()
	log.Println("created lobby with options")
	log.Println(opts)

	go func() {
		time.Sleep(1*time.Minute)
		lobbyMux.Lock()
		if game.Empty() { delete(lobbyPool, opts.Code) }
		lobbyMux.Unlock()
	}()
}
