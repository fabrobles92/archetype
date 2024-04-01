/* Autogenerated file. Do not edit manually. */

package datamod

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod/codec"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

// Reference imports to suppress errors if they are not used.
var (
	_ = crypto.Keccak256
	_ = big.NewInt
	_ = common.Big1
	_ = codec.EncodeAddress
)

func NewConfigRowWithHooks(row *ConfigRow) *ConfigRow {
	return &ConfigRow{lib.NewDatastoreStructWithHooks(row.IDatastoreStruct)}
}

type ConfigWithHooks struct {
	Config
	OnSetRow func(rowKey []interface{}, column int, value []byte)
}

func NewConfigWithHooks(table *Config) *ConfigWithHooks {
	return &ConfigWithHooks{
		Config: *table,
	}
}

func (m *ConfigWithHooks) Get() *ConfigRow {
	row := m.Config.Get()
	row = NewConfigRowWithHooks(row)
	row.IDatastoreStruct.(*lib.DatastoreStructWithHooks).OnSetField = func(column int, value []byte) {
		if m.OnSetRow != nil {
			m.OnSetRow([]interface{}{}, column, value)
		}
	}
	return row
}
