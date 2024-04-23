/* Autogenerated file. Do not edit manually. */

package datamod

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/concrete/codegen/datamod/codec"
	"github.com/ethereum/go-ethereum/concrete/crypto"
	"github.com/ethereum/go-ethereum/concrete/lib"
)

// Reference imports to suppress errors if they are not used.
var (
	_ = crypto.Keccak256
	_ = big.NewInt
	_ = common.Big1
	_ = codec.EncodeAddress
)

// var (
//	BodiesDefaultKey = crypto.Keccak256([]byte("datamod.v1.Bodies"))
// )

func BodiesDefaultKey() []byte {
	return crypto.Keccak256([]byte("datamod.v1.Bodies"))
}

type BodiesRow struct {
	lib.IDatastoreStruct
}

func NewBodiesRow(dsSlot lib.DatastoreSlot) *BodiesRow {
	sizes := []int{4, 4, 4, 4, 4, 4, 4}
	return &BodiesRow{lib.NewDatastoreStruct(dsSlot, sizes)}
}

func (v *BodiesRow) Get() (
	x int32,
	y int32,
	r uint32,
	vx int32,
	vy int32,
	ax int32,
	ay int32,
) {
	return codec.DecodeSmallInt32(4, v.GetField(0)),
		codec.DecodeSmallInt32(4, v.GetField(1)),
		codec.DecodeSmallUint32(4, v.GetField(2)),
		codec.DecodeSmallInt32(4, v.GetField(3)),
		codec.DecodeSmallInt32(4, v.GetField(4)),
		codec.DecodeSmallInt32(4, v.GetField(5)),
		codec.DecodeSmallInt32(4, v.GetField(6))
}

func (v *BodiesRow) Set(
	x int32,
	y int32,
	r uint32,
	vx int32,
	vy int32,
	ax int32,
	ay int32,
) {
	v.SetField(0, codec.EncodeSmallInt32(4, x))
	v.SetField(1, codec.EncodeSmallInt32(4, y))
	v.SetField(2, codec.EncodeSmallUint32(4, r))
	v.SetField(3, codec.EncodeSmallInt32(4, vx))
	v.SetField(4, codec.EncodeSmallInt32(4, vy))
	v.SetField(5, codec.EncodeSmallInt32(4, ax))
	v.SetField(6, codec.EncodeSmallInt32(4, ay))
}

func (v *BodiesRow) GetX() int32 {
	data := v.GetField(0)
	return codec.DecodeSmallInt32(4, data)
}

func (v *BodiesRow) SetX(value int32) {
	data := codec.EncodeSmallInt32(4, value)
	v.SetField(0, data)
}

func (v *BodiesRow) GetY() int32 {
	data := v.GetField(1)
	return codec.DecodeSmallInt32(4, data)
}

func (v *BodiesRow) SetY(value int32) {
	data := codec.EncodeSmallInt32(4, value)
	v.SetField(1, data)
}

func (v *BodiesRow) GetR() uint32 {
	data := v.GetField(2)
	return codec.DecodeSmallUint32(4, data)
}

func (v *BodiesRow) SetR(value uint32) {
	data := codec.EncodeSmallUint32(4, value)
	v.SetField(2, data)
}

func (v *BodiesRow) GetVx() int32 {
	data := v.GetField(3)
	return codec.DecodeSmallInt32(4, data)
}

func (v *BodiesRow) SetVx(value int32) {
	data := codec.EncodeSmallInt32(4, value)
	v.SetField(3, data)
}

func (v *BodiesRow) GetVy() int32 {
	data := v.GetField(4)
	return codec.DecodeSmallInt32(4, data)
}

func (v *BodiesRow) SetVy(value int32) {
	data := codec.EncodeSmallInt32(4, value)
	v.SetField(4, data)
}

func (v *BodiesRow) GetAx() int32 {
	data := v.GetField(5)
	return codec.DecodeSmallInt32(4, data)
}

func (v *BodiesRow) SetAx(value int32) {
	data := codec.EncodeSmallInt32(4, value)
	v.SetField(5, data)
}

func (v *BodiesRow) GetAy() int32 {
	data := v.GetField(6)
	return codec.DecodeSmallInt32(4, data)
}

func (v *BodiesRow) SetAy(value int32) {
	data := codec.EncodeSmallInt32(4, value)
	v.SetField(6, data)
}

type Bodies struct {
	dsSlot lib.DatastoreSlot
}

func NewBodies(ds lib.Datastore) *Bodies {
	dsSlot := ds.Get(BodiesDefaultKey())
	return &Bodies{dsSlot}
}

func NewBodiesFromSlot(dsSlot lib.DatastoreSlot) *Bodies {
	return &Bodies{dsSlot}
}
func (m *Bodies) Get(
	bodyId uint8,
) *BodiesRow {
	dsSlot := m.dsSlot.Mapping().GetNested(
		codec.EncodeSmallUint8(1, bodyId),
	)
	return NewBodiesRow(dsSlot)
}
