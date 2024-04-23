// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

{{ range $schema := .Schemas }}
{{- if $schema.Values }}
struct {{SolidityActionStructNameFn $schema.Name}} {
    {{- range $value := $schema.Values }}
    {{$value.Type.SolType}} {{$value.Name}};
    {{- end }}
}
{{ end }}
{{- end }}

interface {{.Name}} {
    event {{$.ArchParams.ActionExecutedEventName}}(bytes4 actionId, bytes data);

    function {{SolidityActionMethodNameFn $.ArchParams.TickActionName}}() external;

{{ range $schema := .Schemas }}
    {{- if $schema.Values }}
    function {{SolidityActionMethodNameFn $schema.Name}}({{SolidityActionStructNameFn $schema.Name}} memory action) external;
    {{- else }}
    function {{SolidityActionMethodNameFn $schema.Name}}() external;
    {{- end }}
{{- end }}
}

