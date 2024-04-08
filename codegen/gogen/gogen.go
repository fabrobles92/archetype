package gogen

import (
	_ "embed"
	"errors"
	"html/template"
	"path/filepath"

	"github.com/concrete-eth/archetype/codegen"
	"github.com/concrete-eth/archetype/params"
)

//go:embed templates/types.go.tpl
var typesTpl string

//go:embed templates/actions.go.tpl
var actionsTpl string

//go:embed templates/tables.go.tpl
var tablesTpl string

type importSpecs struct {
	Name string
	Path string
}

type Config struct {
	codegen.Config
	Contracts    string
	Package      string
	Datamod      string
	Experimental bool
}

func (c Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if c.Package == "" {
		return errors.New("package is required")
	}
	if c.Datamod == "" {
		return errors.New("datamod is required")
	}
	return nil
}

func GenerateActionTypes(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.Package
	funcMap := make(template.FuncMap)
	funcMap["StructNameFn"] = params.ActionStructName
	outPath := filepath.Join(config.Out, "action_types.go")
	return codegen.ExecuteTemplate(typesTpl, config.Actions, outPath, data, funcMap)
}

func GenerateActions(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.Package
	data["Imports"] = []importSpecs{
		{"contract", filepath.Join(config.Contracts, params.IActionsContract.PackageName)},
	}
	data["Experimental"] = config.Experimental
	outPath := filepath.Join(config.Out, "actions.go")
	return codegen.ExecuteTemplate(actionsTpl, config.Actions, outPath, data, nil)
}

func GenerateTableTypes(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.Package
	funcMap := make(template.FuncMap)
	funcMap["StructNameFn"] = params.TableStructName
	outPath := filepath.Join(config.Out, "table_types.go")
	return codegen.ExecuteTemplate(typesTpl, config.Tables, outPath, data, funcMap)
}

func GenerateTables(config Config) error {
	data := make(map[string]interface{})
	data["Package"] = config.Package
	data["Imports"] = []importSpecs{
		{"mod", config.Datamod},
		{"contract", filepath.Join(config.Contracts, params.ITablesContract.PackageName)},
	}
	data["Experimental"] = config.Experimental
	outPath := filepath.Join(config.Out, "tables.go")
	return codegen.ExecuteTemplate(tablesTpl, config.Tables, outPath, data, nil)
}

func Codegen(config Config) error {
	if err := config.Validate(); err != nil {
		return errors.New("error validating config for go code generation: " + err.Error())
	}
	if err := GenerateActionTypes(config); err != nil {
		return errors.New("error generating go action types binding: " + err.Error())
	}
	if err := GenerateActions(config); err != nil {
		return errors.New("error generating go actions binding: " + err.Error())
	}
	if err := GenerateTableTypes(config); err != nil {
		return errors.New("error generating go table types binding: " + err.Error())
	}
	// TODO: error messages
	if err := GenerateTables(config); err != nil {
		return errors.New("error generating go tables binding: " + err.Error())
	}
	return nil
}
