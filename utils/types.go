package utils

import "github.com/gorilla/websocket"

type GameMode int
const (
	Singleplayer GameMode = iota
	Vs
)
type GameOptions struct {
	Mode   GameMode
	Timed  bool
	Time   int
	Width  int
	Height int
	Bombs  int

	Seed int
	Code string
}

type GameHandler interface {
	Join(*websocket.Conn)
	Empty() bool
}

type Lobby struct {
	Opts GameOptions
	Game GameHandler
}
