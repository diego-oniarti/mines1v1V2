package handlers

import (
	"encoding/binary"
	"log"
	"math/rand/v2"
	"time"

	"github.com/diego-oniarti/minesV2/game"
	"github.com/diego-oniarti/minesV2/utils"
	"github.com/gorilla/websocket"
)

type handler1v1 struct {
	game    game.Game
	players []*websocket.Conn
	turn    int
	started bool
	resetTimer   chan bool
}

func new1v1(opt utils.GameOptions) *handler1v1 {
	return &handler1v1{
		game:       game.NewGame(opt),
		players:    []*websocket.Conn{},
		turn:       0,
		resetTimer: make(chan bool),
	}
}

func (h *handler1v1) Empty() bool {
	return len(h.players)==0
}

func (h *handler1v1) Join(conn *websocket.Conn) {
	isInitiator := false

	if len(h.players)==2 { // Game already full
		conn.Close()
		return
	}

	h.players = append(h.players, conn)
	Pid := 0
	for h.players[Pid]!=conn {Pid++}

	if len(h.players)==2 { // Start the game
		isInitiator = true
		updates := h.game.Start()
		startMsg := make([]byte, 10)

		// Send the game parameters
		binary.BigEndian.PutUint16(startMsg,     uint16(h.game.Width))
		binary.BigEndian.PutUint16(startMsg[2:], uint16(h.game.Height))
		binary.BigEndian.PutUint32(startMsg[4:], uint32(h.game.Time))
		startMsg[9] = 0
		if h.game.Timed { startMsg[9] = 1 }
		for i, conn := range h.players {
			startMsg[8] = byte(i+1)
			conn.WriteMessage(websocket.BinaryMessage, startMsg)
		}

		// 3 seconds and send the first move
		go func() {
			time.Sleep(3 * time.Second)
			for _, conn := range h.players {
				conn.WriteMessage(websocket.BinaryMessage, game.UpdatesToArray(updates))
			}
			h.started = true

			timer := time.NewTimer(time.Duration(h.game.Time)*time.Millisecond)
            defer timer.Stop()

			if !h.game.Timed { return }
			for {
				select {
				case <-timer.C:
					msg := []byte{0b11000000}
					msg[0] |= byte(h.turn+1)
					for _, c := range h.players {
						c.WriteMessage(websocket.BinaryMessage, msg)
						c.Close()
					}
					return
				case <-h.resetTimer:
					if !timer.Stop() {
						select {
						case <-timer.C:
						default:
						}
					}
					timer.Reset(time.Duration(h.game.Time)*time.Millisecond)
				}
			}
		}()
	}


	for {
		// Receive
		mt, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}

		// To keep connection alive
		if mt == websocket.TextMessage {
			conn.WriteMessage(mt, msg) 
			continue
		}

		// Actual Game
		if h.turn != Pid {continue}

		x := binary.BigEndian.Uint16(msg)
		y := binary.BigEndian.Uint16(msg[2:])
		f := int(msg[4])

		repl := []byte{byte(Pid+1)}

		if f==0 { // no flag
			updates, game_over := h.game.Click(int(x), int(y))
			if game_over { // lost
				repl[0] |= 0b10000000
				repl = binary.BigEndian.AppendUint16(repl, x)
				repl = binary.BigEndian.AppendUint16(repl, y)
			} else { // not lost
				if len(updates)==0 { continue }
				// valid move
				repl = append(repl, game.UpdatesToArray(updates)...)
				h.turn = (h.turn + 1) % 2

				// Game won
				if h.game.Won() {
					repl[0] |= 0b01000000
					for _, c := range h.players {
						c.WriteMessage(websocket.BinaryMessage, repl)
						c.Close()
					}
					continue
				}

				select {
				case h.resetTimer <- true:
					log.Println("Reset sent")
				default:
					log.Println("Reset NOT sent")
				}
			}
		} else { // flag
			if update, ok := h.game.Flag(int(x), int(y)); ok {
				repl = append(repl, game.UpdatesToArray([]game.Update{update})...)
			} else {
				continue
			}
		}

		for _, c := range h.players {
			c.WriteMessage(websocket.BinaryMessage, repl)
		}
	}

	if isInitiator {
		log.Println("Adding lobby back in the pool")
		h.game.GameOptions.Seed = rand.Int()
		createLobby(h.game.GameOptions)
	}
}
