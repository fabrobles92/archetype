package types

// TODO: rename package? how to not have to name imports?

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
	ErrInvalidTableId  = errors.New("invalid table ID")
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
			return archSchemas{}, fmt.Errorf("no type found for schema %s", schema.Name)
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
	if !isStructPtr(actionType) {
		return ValidActionId{}, false
	}
	actionElem := actionType.Elem()
	for id, schema := range a.schemas {
		if actionElem == schema.Type {
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

// EncodeAction encodes an action into a byte slice.
func (a *ActionSpecs) EncodeAction(action Action) (ValidActionId, []byte, error) {
	actionId, ok := a.ActionIdFromAction(action)
	if !ok {
		return ValidActionId{}, nil, fmt.Errorf("action of type %T does not match any canonical action type", action)
	}
	schema := a.GetActionSchema(actionId)
	data, err := packActionMethodInput(schema.Method, action)
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
	// Create a canonically typed action from the unpacked data
	// i.e., anonymous struct{...} -> archmod.ActionData_<action name>{...}
	// All methods are autogenerated to have a single argument, so we can safely assume len(args) == 1
	action := reflect.New(schema.Type).Interface()
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
		return nil, fmt.Errorf("action of type %T does not match any canonical action type", action)
	}
	schema := a.GetActionSchema(actionId)
	data, err := packActionMethodInput(schema.Method, action)
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
		return nil, errors.New("invalid calldata (length < 4)")
	}
	var methodId [4]byte
	copy(methodId[:], calldata[:4])
	actionId, ok := a.NewValidId(methodId)
	if !ok {
		return nil, errors.New("method signature does not match any action")
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
		Topics: []common.Hash{params.ActionExecutedEventID},
		Data:   data,
	}
	return log, nil
}

// LogToAction converts a log to an action.
func (a *ActionSpecs) LogToAction(log types.Log) (Action, error) {
	if len(log.Topics) != 1 || log.Topics[0] != params.ActionExecutedEventID {
		return nil, errors.New("log topics do not match action executed event")
	}
	return a.CalldataToAction(log.Data)
}

func packActionMethodInput(method *abi.Method, arg interface{}) ([]byte, error) {
	switch len(method.Inputs) {
	case 0:
		return method.Inputs.Pack()
	case 1:
		return method.Inputs.Pack(arg)
	default:
		panic("unreachable")
	}
}

type ValidTableId struct {
	validId
}

type tableGetter struct {
	constructor   reflect.Value
	rowGetterType reflect.Type
}

func newTableGetter(constructor interface{}, rowType reflect.Type) (tableGetter, error) {
	// Constructor(Datastore) -> Table
	// Table.Get(Keys) -> Row

	if constructor == nil {
		return tableGetter{}, errors.New("table constructor is nil")
	}

	constructorVal := reflect.ValueOf(constructor)
	if !constructorVal.IsValid() {
		return tableGetter{}, errors.New("table constructor is invalid")
	}
	if constructorVal.Kind() != reflect.Func {
		return tableGetter{}, errors.New("table constructor is not a function")
	}
	if constructorVal.Type().NumIn() != 1 {
		return tableGetter{}, errors.New("table constructor should take a single argument")
	}
	if constructorVal.Type().In(0) != reflect.TypeOf((*lib.Datastore)(nil)).Elem() {
		return tableGetter{}, errors.New("table constructor should take a lib.Datastore argument")
	}
	if constructorVal.Type().NumOut() != 1 {
		return tableGetter{}, errors.New("table constructor should return a single value")
	}

	tblType := constructorVal.Type().Out(0)
	getMth, ok := tblType.MethodByName("Get")
	if !ok {
		return tableGetter{}, errors.New("table missing Get method")
	}
	if getMth.Type.NumOut() != 1 {
		return tableGetter{}, errors.New("table Get method should return a single value")
	}

	retType := getMth.Type.Out(0)
	if err := canPopulateStruct(retType, reflect.PtrTo(rowType)); err != nil {
		return tableGetter{}, fmt.Errorf("datamod row type cannot populate table row type: %v", err)
	}

	return tableGetter{
		constructor:   constructorVal,
		rowGetterType: getMth.Type,
	}, nil
}

func (t *tableGetter) get(datastore lib.Datastore, args ...interface{}) (interface{}, error) {
	// Construct the table
	constructorArgs := []reflect.Value{reflect.ValueOf(datastore)}
	table := t.constructor.Call(constructorArgs)[0]
	// Call the Get method
	rowGetter := table.MethodByName("Get")
	// Call the Get method
	rowArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		argVal := reflect.ValueOf(arg)
		if argVal.Type() != t.rowGetterType.In(i) {
			return nil, fmt.Errorf("argument %d has wrong type", i)
		}
		rowArgs[i] = argVal
	}
	result := rowGetter.Call(rowArgs)[0]
	// Return the result
	return result.Interface(), nil
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
	getters map[string]interface{},
) (TableSpecs, error) {
	s, err := newArchSchemas(abi, schemas, types, params.TableMethodName)
	if err != nil {
		return TableSpecs{}, err
	}
	tableGetters := make(map[RawIdType]tableGetter, len(getters))
	for id, schema := range s.schemas {
		getterFn, ok := getters[schema.Name]
		if !ok {
			return TableSpecs{}, fmt.Errorf("no table getter found for schema %s", schema.Name)
		}
		tableGetters[id], err = newTableGetter(getterFn, schema.Type)
		if err != nil {
			return TableSpecs{}, err
		}
	}
	return TableSpecs{archSchemas: s, tableGetters: tableGetters}, nil
}

// NewTableSpecsFromRaw creates a new TableSpecs instance from raw JSON strings.
func NewTableSpecsFromRaw(
	abiJson string,
	schemasJson string,
	types map[string]reflect.Type,
	getters map[string]interface{},
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
	getter := t.tableGetters[tableId.Raw()]
	dsRow, err := getter.get(datastore, args...)
	if err != nil {
		return nil, err
	}
	schema := t.GetTableSchema(tableId)
	row := reflect.New(schema.Type).Interface()
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
		return ValidTableId{}, nil, errors.New("calldata does not correspond to a table read operation")
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

func (a ActionBatch) Len() int {
	return len(a.Actions)
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
		return fmt.Errorf("expected src to be a struct, got %v", srcVal.Type())
	}

	destVal := reflect.ValueOf(dest)
	if !isStructPtr(destVal.Type()) {
		return fmt.Errorf("expected dest to be a pointer to a struct, got %v", destVal.Type())
	}

	destElem := destVal.Elem()
	destType := destElem.Type()

	for i := 0; i < destElem.NumField(); i++ {
		destField := destElem.Field(i)
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

func canPopulateStruct(srcType reflect.Type, destType reflect.Type) error {
	if !isStruct(srcType) && !isStructPtr(srcType) {
		return errors.New("src is not a struct or a pointer to a struct")
	}
	if !isStructPtr(destType) {
		return errors.New("dest is not a pointer to a struct")
	}

	destElemType := destType.Elem()

	// TODO: checks to avoid panics
	// TODO: dest vs dst

	for i := 0; i < destElemType.NumField(); i++ {
		destField := destElemType.Field(i)
		destFieldType := destElemType.Field(i)
		getMethodName := "Get" + destFieldType.Name
		srcGetMethod, ok := srcType.MethodByName(getMethodName)
		if !ok {
			return fmt.Errorf("method %s not found", getMethodName)
		}
		if srcGetMethod.Type.NumOut() != 1 {
			return errors.New("method should return a single value")
		}
		if srcGetMethod.Type.Out(0) != destField.Type {
			return fmt.Errorf("field %s has different type", destFieldType.Name)
		}
	}

	return nil

}

// populateStruct sets all the fields in dest to the values returned by the Get<field name> methods in src.
func populateStruct(src interface{}, dest interface{}) error {
	if err := canPopulateStruct(reflect.TypeOf(src), reflect.TypeOf(dest)); err != nil {
		return err
	}
	var (
		srcVal       = reflect.ValueOf(src)
		destVal      = reflect.ValueOf(dest)
		destElem     = destVal.Elem()
		destElemType = destElem.Type()
	)
	for i := 0; i < destVal.NumField(); i++ {
		var (
			destField     = destElem.Field(i)
			destTypeField = destElemType.Field(i)
			getMethodName = "Get" + destTypeField.Name
			srcGetMethod  = srcVal.MethodByName(getMethodName)
			values        = srcGetMethod.Call(nil)
			value         = values[0]
		)
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
	KV() lib.KeyValueStore      // Get the key-value store
	// ExecuteAction(Action) error // Execute the given action
	SetBlockNumber(uint64) // Set the block number
	BlockNumber() uint64   // Get the block number
	// RunSingleTick()             // Run a single tick
	// RunBlockTicks()             // Run all ticks in a block
	Tick()                    // Run a single tick
	TicksPerBlock() uint      // Get the number of ticks per block
	ExpectTick() bool         // Check if a tick is expected
	SetInBlockTickIndex(uint) // Set the in-block tick index
	InBlockTickIndex() uint   // Get the in-block tick index
}

// TODO: force actions to be valid like ids [?]

type BaseCore struct {
	kv               lib.KeyValueStore
	ds               lib.Datastore
	blockNumber      uint64
	inBlockTickIndex uint
	ticksPerBlock    uint
}

var _ Core = &BaseCore{}

func (b *BaseCore) SetKV(kv lib.KeyValueStore) {
	b.kv = kv
	b.ds = lib.NewKVDatastore(kv)
}

func (b *BaseCore) KV() lib.KeyValueStore {
	return b.kv
}

func (b *BaseCore) Datastore() lib.Datastore {
	return b.ds
}

func (b *BaseCore) SetBlockNumber(blockNumber uint64) {
	b.blockNumber = blockNumber
}

func (b *BaseCore) BlockNumber() uint64 {
	return b.blockNumber
}

func (b *BaseCore) SetInBlockTickIndex(index uint) {
	b.inBlockTickIndex = index
}

func (b *BaseCore) InBlockTickIndex() uint {
	return b.inBlockTickIndex
}

func (b *BaseCore) TicksPerBlock() uint {
	return b.ticksPerBlock
}

func (b *BaseCore) ExpectTick() bool {
	return true
}

func (b *BaseCore) Tick() {}

func incrementBlockTickIndex(c Core) {
	c.SetInBlockTickIndex(c.InBlockTickIndex() + 1)
}

func RunSingleTick(c Core) {
	c.Tick()
}

func RunBlockTicks(c Core) {
	for i := uint(0); i < c.TicksPerBlock(); i++ {
		RunSingleTick(c)
		incrementBlockTickIndex(c)
	}
}

// TODO: have this as a method of action specs [?]

// ExecuteAction executes the method in the target matching the action name with the given action as argument.
// The action must either be a canonical actions (i.e. Tick) or be in the action specs.
func ExecuteAction(spec ActionSpecs, action Action, target interface{}) error {
	if _, ok := action.(*CanonicalTickAction); ok {
		RunBlockTicks(target.(Core))
		return nil
	}
	actionId, ok := spec.ActionIdFromAction(action)
	if !ok {
		return ErrInvalidAction
	}
	schema := spec.GetActionSchema(actionId)
	actionName := schema.Name
	methodName := actionName
	targetVal := reflect.ValueOf(target)
	if !targetVal.IsValid() {
		return fmt.Errorf("target is invalid")
	}
	method := targetVal.MethodByName(methodName)
	if !method.IsValid() {
		return fmt.Errorf("method %s not found", methodName)
	}
	args := []reflect.Value{reflect.ValueOf(action)}
	result := method.Call(args)
	if len(result) == 0 {
		return nil
	}
	errVal := result[len(result)-1]
	if !errVal.IsNil() {
		return errVal.Interface().(error)
	}
	return nil
}
