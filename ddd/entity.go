// Package ddd 提供领域驱动设计基础设施
package ddd

import (
	"strconv"
	"time"
)

// Identifier 实体标识接口
type Identifier interface {
	comparable
	String() string
	IsZero() bool
}

// StringID 字符串标识
type StringID string

func (id StringID) String() string { return string(id) }
func (id StringID) IsZero() bool   { return id == "" }

// Int64ID 64位整数标识
type Int64ID int64

func (id Int64ID) String() string { return strconv.FormatInt(int64(id), 10) }
func (id Int64ID) IsZero() bool   { return id == 0 }

// UintID 无符号整数标识
type UintID uint64

func (id UintID) String() string { return strconv.FormatUint(uint64(id), 10) }
func (id UintID) IsZero() bool   { return id == 0 }

// EntityBase 实体基类，提供通用字段和方法
type EntityBase[ID Identifier] struct {
	id        ID
	createdAt time.Time
	updatedAt time.Time
}

// NewEntityBase 创建实体基类
func NewEntityBase[ID Identifier](id ID) EntityBase[ID] {
	now := time.Now()
	return EntityBase[ID]{id: id, createdAt: now, updatedAt: now}
}

func (e *EntityBase[ID]) ID() ID                    { return e.id }
func (e *EntityBase[ID]) GetID() ID                 { return e.id }
func (e *EntityBase[ID]) SetID(id ID)               { e.id = id }
func (e *EntityBase[ID]) CreatedAt() time.Time      { return e.createdAt }
func (e *EntityBase[ID]) UpdatedAt() time.Time      { return e.updatedAt }
func (e *EntityBase[ID]) SetCreatedAt(t time.Time)  { e.createdAt = t }
func (e *EntityBase[ID]) SetUpdatedAt(t time.Time)  { e.updatedAt = t }
func (e *EntityBase[ID]) MarkUpdated()              { e.updatedAt = time.Now() }

// SameIdentityAs 判断是否同一实体
func (e *EntityBase[ID]) SameIdentityAs(other *EntityBase[ID]) bool {
	if other == nil {
		return false
	}
	return e.id == other.id
}
