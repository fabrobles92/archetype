/* Autogenerated file. Do not edit manually. */

package archmod

import (
	"reflect"
	"strings"

	archtypes "github.com/concrete-eth/archetype/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/ethereum/go-ethereum/concrete/lib"

	contract "github.com/concrete-eth/archetype/example/gogen/abigen/tables"
	mod "github.com/concrete-eth/archetype/example/gogen/datamod"
)

var TablesABIJson = contract.ContractABI

var TablesSchemaJson = `{
    "config": {
        "schema": {
            "startBlock": "uint64",
            "maxPlayers": "uint8"
        }
    },
    "players": {
        "keySchema": {
            "playerId": "uint8"
        },
        "schema": {
            "x": "int16",
            "y": "int16",
            "health": "uint8"
        }
    }
}`

var TableSpecs archtypes.TableSpecs

func init() {
	types := map[string]reflect.Type{
		"Config":  reflect.TypeOf(RowData_Config{}),
		"Players": reflect.TypeOf(RowData_Players{}),
	}
	getters := map[string]archtypes.GetterFn{
		"Config": func(ds lib.Datastore) interface{} {
			return mod.NewConfig(ds)
		},
		"Players": func(ds lib.Datastore) interface{} {
			return mod.NewPlayers(ds)
		},
	}
	var (
		ABI     abi.ABI
		schemas []datamod.TableSchema
		err     error
	)
	// Load the contract ABI
	if ABI, err = abi.JSON(strings.NewReader(TablesABIJson)); err != nil {
		panic(err)
	}
	// Load the table schemas
	if schemas, err = datamod.UnmarshalTableSchemas([]byte(TablesSchemaJson), false); err != nil {
		panic(err)
	}
	// Create the specs
	if TableSpecs, err = archtypes.NewTableSpecs(&ABI, schemas, types, getters); err != nil {
		panic(err)
	}
}
