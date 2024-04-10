package types

// TODO: rename package? how to not have to name imports?
// TODO: error messages in types.go

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod"
	"github.com/ethereum/go-ethereum/concrete/lib"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/concrete-eth/archetype/params"
)

var (
	ErrInvalidAction   = errors.New("invalid action")
	ErrInvalidActionId = errors.New("invalid action ID")
)

var (
	ActionExecutedEvent *abi.Event
)

type RawIdType = [4]byte

type validId struct {
	id    RawIdType
	valid bool
}

func (v validId) Raw() RawIdType {
	if !v.valid {
		panic("Invalid id")
	}
	return v.id
}

/*

Schema: Spec for a single action or table
Schemas: Specs for either all actions or all tables

*/

type archSchema struct {
	datamod.TableSchema
	Method *abi.Method
	Type   reflect.Type
}

type archSchemas struct {
	abi     *abi.ABI
	schemas map[RawIdType]archSchema
}

func newArchSchemas(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
	methodNameFn func(string) string,
) (archSchemas, error) {
	s := archSchemas{abi: abi, schemas: make(map[RawIdType]archSchema, len(schemas))}
	for _, schema := range schemas {
		methodName := methodNameFn(schema.Name)
		method := abi.Methods[methodName]
		var id [4]byte
		copy(id[:], method.ID)

		actionType, ok := types[schema.Name]
		if !ok {
			return archSchemas{}, errors.New("missing type")
		}

		s.schemas[id] = archSchema{
			TableSchema: datamod.TableSchema{Name: schema.Name},
			Method:      &method,
			Type:        actionType,
		}
	}
	return s, nil
}

func (a archSchemas) newValidId(id RawIdType) (validId, bool) {
	if _, ok := a.schemas[id]; ok {
		return validId{id: id, valid: true}, true
	}
	return validId{}, false
}

func (a archSchemas) getSchema(actionId validId) archSchema {
	return a.schemas[actionId.Raw()]
}

// ABI returns the ABI of the interface.
func (a archSchemas) ABI() *abi.ABI {
	return a.abi
}

type ValidActionId struct {
	validId
}

type ActionSchema struct {
	archSchema
}

type ActionSpecs struct {
	archSchemas
}

// TODO: Action schemas are table schemas without keys. Is there a better way to portray this [?]

// NewActionSpecs creates a new ActionSpecs instance.
func NewActionSpecs(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
) (ActionSpecs, error) {
	s, err := newArchSchemas(abi, schemas, types, params.ActionMethodName)
	if err != nil {
		return ActionSpecs{}, err
	}
	return ActionSpecs{archSchemas: s}, nil
}

// NewActionSpecsFromRaw creates a new ActionSpecs instance from raw JSON strings.
func NewActionSpecsFromRaw(
	abiJson string,
	schemasJson string,
	types map[string]reflect.Type,
) (ActionSpecs, error) {
	// Load the contract ABI
	ABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return ActionSpecs{}, err
	}
	// Load the table schemas
	schemas, err := datamod.UnmarshalTableSchemas([]byte(schemasJson), false)
	if err != nil {
		return ActionSpecs{}, err
	}
	return NewActionSpecs(&ABI, schemas, types)
}

// ActionIdFromAction returns the action ID of the given action.
func (a ActionSpecs) ActionIdFromAction(action Action) (ValidActionId, bool) {
	actionType := reflect.TypeOf(action)
	for id, schema := range a.schemas {
		if actionType == schema.Type {
			return ValidActionId{validId{id: id, valid: true}}, true
		}
	}
	return ValidActionId{}, false
}

// NewValidId wraps a valid ID in a ValidActionId.
func (a ActionSpecs) NewValidId(id RawIdType) (ValidActionId, bool) {
	validId, ok := a.newValidId(id)
	return ValidActionId{validId}, ok
}

// GetActionSchema returns the schema of the action with the given ID.
func (a ActionSpecs) GetActionSchema(actionId ValidActionId) ActionSchema {
	return ActionSchema{a.archSchemas.getSchema(actionId.validId)}
}

// TODO: error capitalization

// EncodeAction encodes an action into a byte slice.
func (a *ActionSpecs) EncodeAction(action Action) (ValidActionId, []byte, error) {
	actionId, ok := a.ActionIdFromAction(action)
	if !ok {
		return ValidActionId{}, nil, errors.New("invalid action")
	}
	schema := a.GetActionSchema(actionId)
	data, err := schema.Method.Inputs.Pack(action)
	if err != nil {
		return ValidActionId{}, nil, err
	}
	return actionId, data, nil
}

// DecodeAction decodes the given calldata into an action.
func (a *ActionSpecs) DecodeAction(actionId ValidActionId, data []byte) (Action, error) {
	schema := a.GetActionSchema(actionId)
	args, err := schema.Method.Inputs.Unpack(data)
	if err != nil {
		return nil, err
	}
	// TODO: check if args is a single element
	// Create a canonically typed action from the unpacked data
	// i.e., anonymous struct{...} -> archmod.ActionData_<action name>{...}
	action := reflect.New(schema.Type)
	if err := convertStruct(args[0], action); err != nil {
		return nil, err
	}
	return action, nil
}

// ActionToCalldata converts an action to calldata.
// The same encoding is used for log data.
func (a *ActionSpecs) ActionToCalldata(action Action) ([]byte, error) {
	actionId, ok := a.ActionIdFromAction(action)
	if !ok {
		return nil, errors.New("invalid action")
	}
	schema := a.GetActionSchema(actionId)
	data, err := schema.Method.Inputs.Pack(action)
	if err != nil {
		return nil, err
	}
	calldata := make([]byte, 4+len(data))
	copy(calldata[:4], schema.Method.ID[:])
	copy(calldata[4:], data)
	return calldata, nil
}

// CalldataToAction converts calldata to an action.
func (a *ActionSpecs) CalldataToAction(calldata []byte) (Action, error) {
	if len(calldata) < 4 {
		return nil, errors.New("invalid calldata")
	}
	var methodId [4]byte
	copy(methodId[:], calldata[:4])
	actionId, ok := a.NewValidId(methodId)
	if !ok {
		return nil, errors.New("invalid method signature")
	}
	return a.DecodeAction(actionId, calldata[4:])
}

// ActionToLog converts an action to a log.
func (a *ActionSpecs) ActionToLog(action Action) (types.Log, error) {
	data, err := a.ActionToCalldata(action)
	if err != nil {
		return types.Log{}, err
	}
	log := types.Log{
		Topics: []common.Hash{ActionExecutedEvent.ID},
		Data:   data,
	}
	return log, nil
}

// LogToAction converts a log to an action.
func (a *ActionSpecs) LogToAction(log types.Log) (Action, error) {
	if len(log.Topics) != 1 || log.Topics[0] != ActionExecutedEvent.ID {
		return nil, errors.New("invalid log")
	}
	return a.CalldataToAction(log.Data)
}

type ValidTableId struct {
	validId
}

type GetterFn = func(lib.Datastore) interface{}

type tableGetter struct {
	dsInstantiateTable GetterFn
	dsTable            interface{}
}

func newTableGetter(dsInstantiateTable GetterFn) tableGetter {
	return tableGetter{dsInstantiateTable: dsInstantiateTable}
}

func (t *tableGetter) get(datastore lib.Datastore, args ...interface{}) (interface{}, error) {
	if t.dsTable == nil {
		t.dsTable = t.dsInstantiateTable(datastore)
	}
	tblVal := reflect.ValueOf(t.dsTable)
	mthVal := tblVal.MethodByName("Get")
	if !mthVal.IsValid() {
		return nil, errors.New("get method not found")
	}

	// Prepare arguments for the call
	callArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		callArgs[i] = reflect.ValueOf(arg)
	}

	// Call the method
	results := mthVal.Call(callArgs)

	// Assuming your Get method returns a single value,
	// you could return the first result directly.
	// Ensure there's at least one result to avoid panicking.
	if len(results) > 0 {
		return results[0].Interface(), nil
	}

	// Return nil or an appropriate zero value if no results were returned
	// This part depends on your function's expected behavior
	return nil, errors.New("no results returned")
}

type TableSchema struct {
	archSchema
}

type TableSpecs struct {
	archSchemas
	tableGetters map[RawIdType]tableGetter
}

// NewTableSpecs creates a new TableSpecs instance.
func NewTableSpecs(
	abi *abi.ABI,
	schemas []datamod.TableSchema,
	types map[string]reflect.Type,
	getters map[string]GetterFn,
) (TableSpecs, error) {
	s, err := newArchSchemas(abi, schemas, types, params.TableMethodName)
	if err != nil {
		return TableSpecs{}, err
	}
	tableGetters := make(map[RawIdType]tableGetter, len(getters))
	for id, schema := range s.schemas {
		getterFn, ok := getters[schema.Name]
		if !ok {
			return TableSpecs{}, errors.New("missing getter")
		}
		tableGetters[id] = newTableGetter(getterFn)
	}
	return TableSpecs{archSchemas: s, tableGetters: tableGetters}, nil
}

// NewTableSpecsFromRaw creates a new TableSpecs instance from raw JSON strings.
func NewTableSpecsFromRaw(
	abiJson string,
	schemasJson string,
	types map[string]reflect.Type,
	getters map[string]GetterFn,
) (TableSpecs, error) {
	// Load the contract ABI
	ABI, err := abi.JSON(strings.NewReader(abiJson))
	if err != nil {
		return TableSpecs{}, err
	}
	// Load the table schemas
	schemas, err := datamod.UnmarshalTableSchemas([]byte(schemasJson), false)
	if err != nil {
		return TableSpecs{}, err
	}
	return NewTableSpecs(&ABI, schemas, types, getters)
}

// read reads a row from the datastore.
func (t TableSpecs) read(datastore lib.Datastore, tableId ValidTableId, args ...interface{}) (interface{}, error) {
	getter, ok := t.tableGetters[tableId.Raw()]
	if !ok {
		return nil, errors.New("invalid table ID")
	}
	dsRow, err := getter.get(datastore, args...)
	if err != nil {
		return nil, err
	}
	schema := t.GetTableSchema(tableId)
	row := reflect.New(schema.Type)
	if err := populateStruct(dsRow, row); err != nil {
		return nil, err
	}
	return row, nil
}

// NewValidId wraps a valid ID in a ValidTableId.
func (t TableSpecs) NewValidId(id RawIdType) (ValidTableId, bool) {
	validId, ok := t.newValidId(id)
	return ValidTableId{validId}, ok
}

// GetTableSchema returns the schema of the table with the given ID.
func (t TableSpecs) GetTableSchema(tableId ValidTableId) TableSchema {
	return TableSchema{t.archSchemas.getSchema(tableId.validId)}
}

// TableIdFromCalldata returns the table ID of the table targeted by the given calldata.
// If the calldata does not encode a table read, the second return value is false.
func (t *TableSpecs) TargetTableId(calldata []byte) (ValidTableId, bool) {
	if len(calldata) < 4 {
		return ValidTableId{}, false
	}
	var methodId [4]byte
	copy(methodId[:], calldata[:4])
	tableId, ok := t.NewValidId(methodId)
	return tableId, ok
}

// Read reads a row from the datastore if the calldata corresponds to a table read operation.
func (t *TableSpecs) Read(datastore lib.Datastore, calldata []byte) (ValidTableId, interface{}, error) {
	tableId, ok := t.TargetTableId(calldata)
	if !ok {
		return ValidTableId{}, nil, errors.New("invalid calldata")
	}
	row, err := t.read(datastore, tableId)
	if err != nil {
		return tableId, nil, err
	}
	return tableId, row, nil
}

// ReadPacked reads a row from the datastore and packs it into an ABI-encoded byte slice.
func (t *TableSpecs) ReadPacked(datastore lib.Datastore, calldata []byte) ([]byte, error) {
	tableId, data, err := t.Read(datastore, calldata)
	if err != nil {
		return nil, err
	}
	schema := t.GetTableSchema(tableId)
	return schema.Method.Outputs.Pack(data)
}

type ArchSpecs struct {
	Actions ActionSpecs
	Tables  TableSpecs
}

type Action interface{}

type CanonicalTickAction struct{}

// Holds all the actions included to a specific core in a specific block
type ActionBatch struct {
	BlockNumber uint64
	Actions     []Action
}

// NewActionBatch creates a new ActionBatch instance.
func NewActionBatch(blockNumber uint64, actions []Action) ActionBatch {
	return ActionBatch{BlockNumber: blockNumber, Actions: actions}
}

// ConvertStruct copies the fields from src to dest if they have the same name and type.
// All fields in dest must be set.
func convertStruct(src interface{}, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	if !isStruct(srcVal.Type()) {
		return errors.New("src is not a struct")
	}

	destVal := reflect.ValueOf(dest)
	if !isStructPtr(destVal.Type()) {
		return errors.New("dest is not a pointer to a struct")
	}

	destElem := destVal.Elem()
	destType := destElem.Type()

	for i := 0; i < destVal.NumField(); i++ {
		destField := destVal.Field(i)
		destFieldType := destType.Field(i)
		if !destField.CanSet() {
			return fmt.Errorf("field %s is not settable", destFieldType.Name)
		}
		srcField := srcVal.FieldByName(destFieldType.Name)
		if !srcField.IsValid() {
			return fmt.Errorf("field %s not found", destFieldType.Name)
		}
		if srcField.Type() != destField.Type() {
			return fmt.Errorf("field %s has different type", destFieldType.Name)
		}
		destField.Set(srcField)
	}

	return nil
}

// populateStruct sets all the fields in dest to the values returned by the Get<field name> methods in src.
func populateStruct(src interface{}, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	if !isStruct(srcVal.Type()) && !isStructPtr(srcVal.Type()) {
		return errors.New("src is not a struct or a pointer to a struct")
	}

	destVal := reflect.ValueOf(dest)
	if !isStructPtr(destVal.Type()) {
		return errors.New("dest is not a pointer to a struct")
	}

	destElem := destVal.Elem()
	destType := destElem.Type()

	for i := 0; i < destVal.NumField(); i++ {
		destField := destElem.Field(i)
		destTypeField := destType.Field(i)
		if destField.CanSet() {
			return fmt.Errorf("field %s is not settable", destTypeField.Name)
		}
		getMethodName := "Get" + destTypeField.Name
		srcGetMethod := srcVal.MethodByName(getMethodName)
		if !srcGetMethod.IsValid() {
			return fmt.Errorf("method %s not found", getMethodName)
		}
		values := srcGetMethod.Call(nil)
		if len(values) != 1 {
			return errors.New("method should return a single value")
		}
		value := values[0]
		if value.Type() != destField.Type() {
			return fmt.Errorf("field %s has different type", destTypeField.Name)
		}
		destField.Set(value)
	}

	return nil
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}

type Core interface {
	SetKV(kv lib.KeyValueStore) // Set the key-value store
	ExecuteAction(Action) error // Execute the given action
	SetBlockNumber(uint64)      // Set the block number
	BlockNumber() uint64        // Get the block number
	RunSingleTick()             // Run a single tick
	RunBlockTicks()             // Run all ticks in a block
	TicksPerBlock() uint        // Get the number of ticks per block
	ExpectTick() bool           // Check if a tick is expected
	SetInBlockTickIndex(uint)   // Set the in-block tick index
	InBlockTickIndex() uint     // Get the in-block tick index
}
