package main

import (
	"github.com/diego-oniarti/minesV2/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.POST("/lobby/create", handlers.CreateLobby)
    r.GET("/lobby/join", handlers.JoinLobby)

    r.NoRoute(func(c *gin.Context) {
        c.File("./static" + c.Request.URL.Path)
    })

    r.GET("/", func(c *gin.Context) {
        c.File("./static/index.html")
    })

    r.Run("0.0.0.0:2357")
}
