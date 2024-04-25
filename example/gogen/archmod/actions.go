/* Autogenerated file. Do not edit manually. */

package archmod

import (
	"reflect"

	"github.com/concrete-eth/archetype/arch"

	contract "github.com/concrete-eth/archetype/example/gogen/abigen/actions"
)

var ActionsABIJson = contract.ContractABI

var ActionSchemasJson = `{
    "addBody": {
        "schema": {
            "x": "int32",
            "y": "int32",
            "r": "uint32",
            "vx": "int32",
            "vy": "int32"
        }
    }
}`

var ActionSchemas arch.ActionSchemas

func init() {
	types := map[string]reflect.Type{
		"AddBody": reflect.TypeOf(ActionData_AddBody{}),
		// "Tick": reflect.TypeOf(arch.CanonicalTickAction{}),
	}
	var err error
	if ActionSchemas, err = arch.NewActionSchemasFromRaw(ActionsABIJson, ActionSchemasJson, types); err != nil {
		panic(err)
	}
}

type IActions interface {
	AddBody(action *ActionData_AddBody) error
	Tick()
}
