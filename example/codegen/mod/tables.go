/* Autogenerated file. Do not edit manually. */

package model

import (
    "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/concrete/lib"
    "github.com/ethereum/go-ethereum/common"
	"github.com/concrete-eth/archetype/example/datamod"
)

var (
	_ = common.Big1
)

/*
Table  KeySize  ValueSize
Tick   0        0
Move   0        2
*/


const (
    TableId_Tick uint8 = iota
    TableId_Move
)
var Tables = map[uint8]struct{
    Id uint8
    Name string
    Keys []string
    Columns []string
}{
    TableId_Tick: {
        Id: TableId_Tick,
        Name: "Tick",
        Keys: []string{
        },
        Columns: []string{
        },
    },
    TableId_Move: {
        Id: TableId_Move,
        Name: "Move",
        Keys: []string{
        },
        Columns: []string{
            "playerId",
            "direction",
        },
    },
}

var TableIdsByMethodName = map[string]uint8{
    "getTickRow": TableId_Tick,
    "getMoveRow": TableId_Move,
}

type TableUpdateHandler func(tableId uint8, rowKey []interface{}, columnIndex int, value []byte)

type State struct {
	datastore  lib.Datastore
    tick *datamod.TickWithHooks
    move *datamod.MoveWithHooks
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

func (s *State) Tick() *datamod.TickWithHooks {
    if s.tick == nil || (s.tick.OnSetRow == nil && s.OnSetTable != nil) {
        s.tick = datamod.NewTickWithHooks(datamod.NewTick(s.datastore))
        s.tick.OnSetRow = func(rowKey []interface{}, columnIndex int, value []byte) {
            if s.OnSetTable != nil {
                s.OnSetTable(TableId_Tick, rowKey, columnIndex, value)
            }
        }
    }
    return s.tick
}

func (s *State) Move() *datamod.MoveWithHooks {
    if s.move == nil || (s.move.OnSetRow == nil && s.OnSetTable != nil) {
        s.move = datamod.NewMoveWithHooks(datamod.NewMove(s.datastore))
        s.move.OnSetRow = func(rowKey []interface{}, columnIndex int, value []byte) {
            if s.OnSetTable != nil {
                s.OnSetTable(TableId_Move, rowKey, columnIndex, value)
            }
        }
    }
    return s.move
}



func (s *State) GetTickRow(
) *datamod.TickRow {
    return s.Tick().Get(
    )
}

func (s *State) GetMoveRow(
) *datamod.MoveRow {
    return s.Move().Get(
    )
}


type RowData_Tick struct{
}

type RowData_Move struct{
    PlayerId uint8 `json:"playerId"`
    Direction uint8 `json:"direction"`
}

func GetData(datastore lib.Datastore, method *abi.Method, args []interface{}) (interface{}, bool) {
	switch method.Name {
	case "getTickRow":
		row := datamod.NewTick(datastore).Get(
		)
		return RowData_Tick{
		}, true
	
	case "getMoveRow":
		row := datamod.NewMove(datastore).Get(
		)
		return RowData_Move{
			PlayerId: row.GetPlayerId(),
			Direction: row.GetDirection(),
		}, true
	
	}
	return nil, false
}