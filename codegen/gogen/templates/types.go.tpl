/* Autogenerated file. Do not edit manually. */

package {{.Package}}

import (
    "github.com/ethereum/go-ethereum/common"
    	
    {{ range .Imports }}
	"{{.}}"
	{{ end }}
)

var (
	_ = common.Big1
)

{{ if .Comment }}
/*
{{ .Comment }}
*/
{{ end }}

{{ range $schema := .Schemas }}
type {{$.TypePrefix}}{{$schema.Name}} struct{
    {{- range $value := $schema.Values }}
    {{$value.PascalCase}} {{$value.Type.GoType}} `json:"{{$value.Name}}"`
    {{- end }}
}
{{ range $value := $schema.Values }}
func (row *{{$.TypePrefix}}{{$schema.Name}}) Get{{$value.PascalCase}}() {{$value.Type.GoType}} {
    return row.{{$value.PascalCase}}
}
{{ end }}
{{ end }}
