/* Autogenerated file. Do not edit manually. */

package {{.Package}}

import (
    "reflect"

	"github.com/concrete-eth/archetype/arch"

	{{ range .Imports }}
	{{- if .Name }}{{.Name}} "{{.Path}}"
    {{- else }}"{{.Path}}"{{ end }}
	{{ end }}
)

var TablesABIJson = contract.ContractABI

var TableSchemasJson = `{{.Json}}`

var TableSchemas arch.TableSchemas

func init() {
    types := map[string]reflect.Type{
        {{- range .Schemas }}
        "{{.Name}}": reflect.TypeOf({{ GoTableStructNameFn .Name }}{}),
        {{- end }}
    }
    getters := map[string]interface{}{
        {{- range .Schemas }}
        "{{.Name}}": datamod.New{{.Name}},
        {{- end }}
    }
    var err error
    if TableSchemas, err = arch.NewTableSchemasFromRaw(TablesABIJson, TableSchemasJson, types, getters); err != nil {
        panic(err)
    }
}
