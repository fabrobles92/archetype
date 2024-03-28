// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

enum TableId {
{{- $length := len .Schemas }}
{{- range $index, $element := .Schemas }}
    {{.Name}}{{if lt $index (_sub $length 1)}},{{end}}
{{- end }}
}
{{ range .Schemas }}
{{- if .Values }}
struct RowData_{{.Name}} {
    {{- range .Values }}
    {{.Type.SolType}} {{.Name}};
    {{- end }}
}
{{ end }}
{{- end }}
interface {{.Name}} {
{{- range .Schemas }}
    function get{{.Name}}Row(
        {{- $length := len .Keys }}
        {{- range $index, $element := .Keys }}
        {{.Type.SolType}} {{.Name}}{{if lt $index (_sub $length 1)}},{{end}}
        {{- end }}
    ) external view returns (RowData_{{.Name}} memory);
{{ end }}
}
