/* Autogenerated file. Do not edit manually. */

package archmod

import (
	"reflect"

	"github.com/concrete-eth/archetype/arch"

	contract "github.com/concrete-eth/archetype/testutils/gogen/abigen/tables"
	"github.com/concrete-eth/archetype/testutils/gogen/datamod"
)

var TablesABIJson = contract.ContractABI

var TableSchemasJson = `{
    "counter": {
        "schema": {
            "value": "int16"
        }
    }
}`

var TableSchemas arch.TableSchemas

func init() {
	types := map[string]reflect.Type{
		"Counter": reflect.TypeOf(RowData_Counter{}),
	}
	getters := map[string]interface{}{
		"Counter": datamod.NewCounter,
	}
	var err error
	if TableSchemas, err = arch.NewTableSchemasFromRaw(TablesABIJson, TableSchemasJson, types, getters); err != nil {
		panic(err)
	}
}
