/* Autogenerated file. Do not edit manually. */

package archmod

import (
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ = common.Big1
)

/*
Table  KeySize  ValueSize
Move   0        2
*/

type ActionData_Move struct {
	PlayerId  uint8 `json:"playerId"`
	Direction uint8 `json:"direction"`
}

func (row *ActionData_Move) GetPlayerId() uint8 {
	return row.PlayerId
}

func (row *ActionData_Move) GetDirection() uint8 {
	return row.Direction
}
