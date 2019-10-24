package waitinglist

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/briscola-as-a-service/game"
	"github.com/briscola-as-a-service/game/player"
)

var mux sync.Mutex

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type WaitingLists map[string]*WaitingList

var waitingLists WaitingLists

type WaitingList struct {
	listname      string
	deckerID      string
	playersNumber int
	players       []player.Player
}

func init() {
	waitingLists = make(map[string]*WaitingList)
}

func New() *WaitingLists {
	return &waitingLists
}

func (wls *WaitingLists) AddList(listName string, playersNumber int) error {
	if _, exists := (*wls)[listName]; exists == true {
		return errors.New("waiting List Exists")
	}

	// deckerID will be reinitializated after the beginning of the match
	deckerID := initDeckerID(32)

	players := make([]player.Player, 0)
	wl := WaitingList{listName, deckerID, playersNumber, players}
	(*wls)[listName] = &wl
	return nil
}

func (wls *WaitingLists) AddPlayer(listName string, playerName string, playerID string) error {
	// Protect from new player race conditions! No new players can be added until
	// is verified if a new game can start (StartGame())
	mux.Lock()

	if _, exists := (*wls)[listName]; exists == false {
		return errors.New("waiting list does not exists")
	}
	if len((*wls)[listName].players) > (*wls)[listName].playersNumber-1 {
		return errors.New("too many players")
	}
	player := player.New(playerName, playerID)

	waitingListPtr := (*wls)[listName]
	players := (*waitingListPtr).players

	for _, p := range players {
		if p.Is(player) {
			return errors.New("player is already in waiting list")
		}
	}

	players = append(players, player)
	(*waitingListPtr).players = players

	return nil
}

func (wls *WaitingLists) StartGame(listName string) (deckerID string, decker *game.Decker, err error) {
	// Protect from new player race conditions! No new players can be added until
	// mux created in AddPlayer
	defer mux.Unlock()

	deckerID = (*wls)[listName].deckerID

	if _, exists := (*wls)[listName]; exists == false {
		err = errors.New("waiting list does not exists")
		return
	}
	if len((*wls)[listName].players) < (*wls)[listName].playersNumber {
		err = errors.New("waiting for players")
		return
	}

	waitingListPtr := (*wls)[listName]
	players := (*waitingListPtr).players

	d := game.New(players, (*waitingListPtr).deckerID)
	decker = &d

	// Reset the waiting list and reinitialize the deckerID
	emptyPlayers := make([]player.Player, 0)
	newDeckerID := initDeckerID(32)
	(*waitingListPtr).players = emptyPlayers
	(*waitingListPtr).deckerID = newDeckerID

	return
}

func initDeckerID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
