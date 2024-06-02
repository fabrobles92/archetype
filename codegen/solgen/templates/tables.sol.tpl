// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

{{ range $schema := .Schemas }}
{{- if $schema.Values }}
struct {{ SolidityTableStructNameFn .Name }} {
    {{- range $value := $schema.Values }}
    {{$value.Type.SolType}} {{$value.Name}};
    {{- end }}
}
{{ end }}
{{- end }}

interface {{$.Name}} {
{{- range $schema := .Schemas }}
    function {{ SolidityTableMethodNameFn $schema.Name }}(
        {{- $length := len $schema.Keys -}}
        {{- range $index, $key := $schema.Keys -}}
        {{- $key.Type.SolType }} {{$key.Name}}{{if lt $index (_sub $length 1)}},{{ end -}}
        {{- end -}}
    ) external view returns ({{ SolidityTableStructNameFn .Name }} memory);
{{- end }}
}
