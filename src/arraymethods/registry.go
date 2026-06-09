package arraymethods

import "github.com/unisc/compiladores/sol/src/ast"

// ArgKind describes an argument to an array method.
type ArgKind int

const (
	ArgNone ArgKind = iota
	ArgInt
	ArgElem // must match array element type
)

// Method describes a builtin array instance method.
type Method struct {
	Name       string
	MinArgs    int
	MaxArgs    int
	ArgKinds   []ArgKind
	ReturnType string // "void", "int", "bool", or "elem"
	Mutates    bool
}

var methods = map[string]Method{
	"push": {
		Name: "push", MinArgs: 1, MaxArgs: 1,
		ArgKinds: []ArgKind{ArgElem}, ReturnType: "void", Mutates: true,
	},
	"pop": {
		Name: "pop", MinArgs: 0, MaxArgs: 0,
		ReturnType: "elem", Mutates: true,
	},
	"remove": {
		Name: "remove", MinArgs: 1, MaxArgs: 1,
		ArgKinds: []ArgKind{ArgInt}, ReturnType: "void", Mutates: true,
	},
	"insert": {
		Name: "insert", MinArgs: 2, MaxArgs: 2,
		ArgKinds: []ArgKind{ArgInt, ArgElem}, ReturnType: "void", Mutates: true,
	},
	"contains": {
		Name: "contains", MinArgs: 1, MaxArgs: 1,
		ArgKinds: []ArgKind{ArgElem}, ReturnType: "bool", Mutates: false,
	},
	"clear": {
		Name: "clear", MinArgs: 0, MaxArgs: 0,
		ReturnType: "void", Mutates: true,
	},
	"isEmpty": {
		Name: "isEmpty", MinArgs: 0, MaxArgs: 0,
		ReturnType: "bool", Mutates: false,
	},
}

// Lookup returns the method descriptor for an array instance call.
func Lookup(name string) (*Method, bool) {
	m, ok := methods[name]
	if !ok {
		return nil, false
	}
	return &m, true
}

// ReturnTypeFor builds the AST return type for a method given element type.
func ReturnTypeFor(m *Method, elem *ast.TypeDesc) *ast.TypeDesc {
	switch m.ReturnType {
	case "void":
		return &ast.TypeDesc{Base: "void"}
	case "int":
		return &ast.TypeDesc{Base: "int"}
	case "bool":
		return &ast.TypeDesc{Base: "bool"}
	case "elem":
		if elem != nil {
			return elem.Copy()
		}
		return &ast.TypeDesc{Base: "void"}
	default:
		return &ast.TypeDesc{Base: "void"}
	}
}
