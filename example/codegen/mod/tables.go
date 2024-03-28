/* Autogenerated file. Do not edit manually. */

package model

import (
	"github.com/concrete-eth/archetype/example/codegen/datamod"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

var (
	_ = common.Big1
)

/*
Table    KeySize  ValueSize
Config   0        9
Players  1        5
*/

const (
	TableId_Config uint8 = iota
	TableId_Players
)

var Tables = map[uint8]struct {
	Id      uint8
	Name    string
	Keys    []string
	Columns []string
}{
	TableId_Config: {
		Id:   TableId_Config,
		Name: "Config",
		Keys: []string{},
		Columns: []string{
			"startBlock",
			"maxPlayers",
		},
	},
	TableId_Players: {
		Id:   TableId_Players,
		Name: "Players",
		Keys: []string{
			"playerId",
		},
		Columns: []string{
			"x",
			"y",
			"health",
		},
	},
}

var TableIdsByMethodName = map[string]uint8{
	"getConfigRow":  TableId_Config,
	"getPlayersRow": TableId_Players,
}

type TableUpdateHandler func(tableId uint8, rowKey []interface{}, columnIndex int, value []byte)

type State struct {
	datastore  lib.Datastore
	config     *datamod.ConfigWithHooks
	players    *datamod.PlayersWithHooks
	OnSetTable TableUpdateHandler
}

func NewState(datastore lib.Datastore) *State {
	return &State{
		datastore: datastore,
	}
}

func (s *State) SetTableUpdateHandler(handler TableUpdateHandler) {
	s.OnSetTable = handler
}

func (s *State) Config() *datamod.ConfigWithHooks {
	if s.config == nil || (s.config.OnSetRow == nil && s.OnSetTable != nil) {
		s.config = datamod.NewConfigWithHooks(datamod.NewConfig(s.datastore))
		s.config.OnSetRow = func(rowKey []interface{}, columnIndex int, value []byte) {
			if s.OnSetTable != nil {
				s.OnSetTable(TableId_Config, rowKey, columnIndex, value)
			}
		}
	}
	return s.config
}

func (s *State) Players() *datamod.PlayersWithHooks {
	if s.players == nil || (s.players.OnSetRow == nil && s.OnSetTable != nil) {
		s.players = datamod.NewPlayersWithHooks(datamod.NewPlayers(s.datastore))
		s.players.OnSetRow = func(rowKey []interface{}, columnIndex int, value []byte) {
			if s.OnSetTable != nil {
				s.OnSetTable(TableId_Players, rowKey, columnIndex, value)
			}
		}
	}
	return s.players
}

func (s *State) GetConfigRow() *datamod.ConfigRow {
	return s.Config().Get()
}

func (s *State) GetPlayersRow(
	playerId uint8,
) *datamod.PlayersRow {
	return s.Players().Get(
		playerId,
	)
}

type RowData_Config struct {
	StartBlock uint64 `json:"startBlock"`
	MaxPlayers uint8  `json:"maxPlayers"`
}

type RowData_Players struct {
	X      int16 `json:"x"`
	Y      int16 `json:"y"`
	Health uint8 `json:"health"`
}

func GetData(datastore lib.Datastore, method *abi.Method, args []interface{}) (interface{}, bool) {
	switch method.Name {
	case "getConfigRow":
		row := datamod.NewConfig(datastore).Get()
		return RowData_Config{
			StartBlock: row.GetStartBlock(),
			MaxPlayers: row.GetMaxPlayers(),
		}, true

	case "getPlayersRow":
		row := datamod.NewPlayers(datastore).Get(
			args[0].(uint8),
		)
		return RowData_Players{
			X:      row.GetX(),
			Y:      row.GetY(),
			Health: row.GetHealth(),
		}, true

	}
	return nil, false
}
