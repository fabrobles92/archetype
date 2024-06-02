// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

{{ range .Imports }}
import "{{ . }}";
{{- end }}

interface {{$.Name}} is {{ range $i, $v := .Interfaces }}{{ if $i }}, {{ end }}{{ $v }}{{ end }} {}
