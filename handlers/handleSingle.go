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

type handlerSingle struct {
	game       game.Game
	player     *websocket.Conn
	started    bool
	resetTimer chan bool
}

func newSingle(opt utils.GameOptions) *handlerSingle {
	return &handlerSingle{
		game:       game.NewGame(opt),
		player:     nil,
		resetTimer: make(chan bool),
		started:    false,
	}
}

func (h *handlerSingle) Empty() bool {
	return h.player==nil
}

func (h *handlerSingle) Join(conn *websocket.Conn) {
	isInitiator := true

	if h.player!=nil {
		conn.Close()
		return
	}

	h.player = conn

	updates := h.game.Start()
	startMsg := make([]byte, 10)

	// Send the game parameters
	binary.BigEndian.PutUint16(startMsg,     uint16(h.game.Width))
	binary.BigEndian.PutUint16(startMsg[2:], uint16(h.game.Height))
	binary.BigEndian.PutUint32(startMsg[4:], uint32(h.game.Time))
	startMsg[9] = 0
	if h.game.Timed { startMsg[9] = 1 }
	conn.WriteMessage(websocket.BinaryMessage, startMsg)

	// 3 seconds and send the first move
	go func() {
		time.Sleep(3 * time.Second)
		conn.WriteMessage(websocket.BinaryMessage, game.UpdatesToArray(updates))
		h.started = true

		timer := time.NewTimer(time.Duration(h.game.Time)*time.Millisecond)
		defer timer.Stop()

		if !h.game.Timed { return }
		for {
			select {
			case <-timer.C:
				msg := []byte{0b11000000}
				conn.WriteMessage(websocket.BinaryMessage, msg)
				conn.Close()
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
		x := binary.BigEndian.Uint16(msg)
		y := binary.BigEndian.Uint16(msg[2:])
		f := int(msg[4])

		repl := []byte{0}

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

				// Game won
				if h.game.Won() {
					repl[0] |= 0b01000000
					conn.WriteMessage(websocket.BinaryMessage, repl)
					conn.Close()
					continue
				}

				select {
				case h.resetTimer <- true:
				default:
				}
			}
		} else { // flag
			if update, ok := h.game.Flag(int(x), int(y)); ok {
				repl = append(repl, game.UpdatesToArray([]game.Update{update})...)
			} else {
				continue
			}
		}

		conn.WriteMessage(websocket.BinaryMessage, repl)
	}

	if isInitiator {
		log.Println("Adding lobby back in the pool")
		h.game.GameOptions.Seed = rand.Int()
		createLobby(h.game.GameOptions)
	}
}
