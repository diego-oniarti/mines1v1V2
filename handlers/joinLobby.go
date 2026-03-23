package handlers

import (
	"log"
	"net/http"

	"github.com/diego-oniarti/minesV2/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

func JoinLobby(c *gin.Context) {
	code := c.Query("lobby")
	log.Printf("Joining %s\n", code)
	if len(code)==0 {
		log.Println("No lobby id provided")
		c.String(http.StatusBadRequest, "No code")
		return
	}
	lobbyMux.Lock()
	var (lobby utils.Lobby; ok bool)
	if lobby, ok = lobbyPool[code]; !ok {
		lobbyMux.Unlock()
		log.Printf("Lobby [%s] does not exist\n", code)
		c.String(http.StatusNotFound, "Bad lobby id")
		return
	}
	lobbyMux.Unlock()

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print(err.Error())
		c.String(http.StatusInternalServerError, "")
		return
	}

	lobby.Game.Join(conn)
}

