package vm

import (
	"fmt"
	"strconv"
)

// Kind identifies a runtime value type.
type Kind int

const (
	KindNull Kind = iota
	KindInt
	KindFloat
	KindBool
	KindString
	KindObject
	KindArray
)

// Value is a runtime SOL value.
type Value struct {
	Kind   Kind
	Int    int64
	Float  float64
	Bool   bool
	StrVal string
	Object *Object
	Array  []Value
}

type Object struct {
	Class  string
	Fields map[string]Value
}

func Null() Value           { return Value{Kind: KindNull} }
func Int(v int64) Value     { return Value{Kind: KindInt, Int: v} }
func Float(v float64) Value { return Value{Kind: KindFloat, Float: v} }
func Bool(v bool) Value     { return Value{Kind: KindBool, Bool: v} }
func Str(v string) Value    { return Value{Kind: KindString, StrVal: v} }
func Obj(class string) Value {
	return Value{Kind: KindObject, Object: &Object{Class: class, Fields: make(map[string]Value)}}
}
func Arr(items ...Value) Value { return Value{Kind: KindArray, Array: items} }

func (v Value) AsFloat() float64 {
	switch v.Kind {
	case KindInt:
		return float64(v.Int)
	case KindFloat:
		return v.Float
	case KindBool:
		if v.Bool {
			return 1
		}
		return 0
	default:
		return 0
	}
}

func (v Value) AsBool() bool {
	switch v.Kind {
	case KindBool:
		return v.Bool
	case KindInt:
		return v.Int != 0
	case KindFloat:
		return v.Float != 0
	case KindNull:
		return false
	case KindString:
		return v.StrVal != ""
	default:
		return true
	}
}

func (v Value) String() string {
	switch v.Kind {
	case KindNull:
		return "null"
	case KindInt:
		return strconv.FormatInt(v.Int, 10)
	case KindFloat:
		return strconv.FormatFloat(v.Float, 'g', -1, 64)
	case KindBool:
		if v.Bool {
			return "true"
		}
		return "false"
	case KindString:
		return v.StrVal
	case KindObject:
		return v.Object.Class + "{}"
	case KindArray:
		return fmt.Sprintf("array[%d]", len(v.Array))
	default:
		return "?"
	}
}
