package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gogogomoku/gomoku_v2/ai"
	"github.com/gogogomoku/gomoku_v2/arcade"
	"github.com/gogogomoku/gomoku_v2/arcade/match"
	"github.com/gogogomoku/gomoku_v2/board"
	"github.com/gogogomoku/gomoku_v2/player"
	"github.com/gorilla/mux"
)

type JsonMessage struct {
	Message string
}

type JsonMove struct {
	PlayerId int8 `json:"playerId"`
	PosX     int8 `json:"posX"`
	PosY     int8 `json:"posY"`
}

type NewGameOpts struct {
	Player1Ai int8 `json:"player1Ai"`
	Player2Ai int8 `json:"player2Ai"`
}

// GET /
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("NEW HOME REQUEST")

	_ = json.NewEncoder(w).Encode(
		JsonMessage{Message: "Welcome to Gomoku... Use /new-match to create a match"},
	)
}

// GET /match/new
func NewMatchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("NEW NEW_MATCH REQUEST")

	player1Ai, err := strconv.ParseBool(r.FormValue("p1ai"))
	if err != nil {
		player1Ai = false
	}
	player2Ai, err := strconv.ParseBool(r.FormValue("p2ai"))
	if err != nil {
		player2Ai = false
	}

	new_match := *arcade.NewMatch(player1Ai, player2Ai)
	_ = json.NewEncoder(w).Encode(
		new_match,
	)
}

// GET /match/{id}
func GetMatchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("NEW GET_MATCH REQUEST")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	match := arcade.CurrentMatches.List[id]

	if match == nil {
		http.Error(w, "Bad request: match doesn't exist", http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(
		truncateHistory(match),
	)
}

// POST /match/{id}/move
func PostMoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("NEW POST_MOVE REQUEST")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	match := arcade.CurrentMatches.List[id]

	if match == nil {
		http.Error(w, "Bad request: match doesn't exist", http.StatusBadRequest)
		return
	}
	params := JsonMove{}
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		log.Println("Error: ", err)
		http.Error(w, "Bad request: error in arguments", http.StatusBadRequest)
		return
	}
	log.Println(params)

	var player *player.Player
	if params.PlayerId == 1 {
		player = match.P1
	} else if params.PlayerId == 2 {
		player = match.P2
	} else {
		http.Error(w, "Bad request: bad player Id in arguments", http.StatusBadRequest)
		return
	}
	position := board.Position{X: params.PosX, Y: params.PosY}
	err = match.AddMove(player, &position)
	if err != nil {
		errMsg := fmt.Sprintf("Bad request: %s", err)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	match.Suggestion = ai.GetSuggestion(match.Board, match.History[len(match.History)-1], match.GetOpponent(player))

	_ = json.NewEncoder(w).Encode(
		truncateHistory(match),
	)

}

// POST /match/{id}/undo
func PostUnapplyMoveHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("NEW POST_UNAPPLY_MOVE REQUEST")

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	match := arcade.CurrentMatches.List[id]

	if match == nil {
		http.Error(w, "Bad request: match doesn't exist", http.StatusBadRequest)
		return
	}

	err := match.UnapplyLastMove()
	if err != nil {
		errMsg := fmt.Sprintf("Bad request: %s", err)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	// Recalculate suggestion for player
	if len(match.History) == 0 {
		match.Suggestion = &board.Position{X: board.SIZE / 2, Y: board.SIZE / 2}
	} else {
		lastMove := match.History[len(match.History)-1]
		match.Suggestion = ai.GetSuggestion(match.Board, lastMove, lastMove.Player)
	}

	_ = json.NewEncoder(w).Encode(
		truncateHistory(match),
	)
}

func truncateHistory(match *match.Match) *match.Match {
	if len(match.History) > 2 {
		toTruncate := *match
		toTruncate.History = toTruncate.History[len(match.History)-2:]
		return &toTruncate
	}
	return match
}
