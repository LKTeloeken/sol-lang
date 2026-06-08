package stdlib

import "github.com/unisc/compiladores/sol/src/ast"

// Class names (reserved).
const (
	Console = "Console"
	File    = "File"
	Time    = "Time"
	String  = "String"
	Math    = "Math"
	Args    = "Args"
)

// ArgSpec describes one argument position for a builtin method.
type ArgSpec struct {
	Base     string // int, float, bool, string
	IsArray  bool
	ElemBase string // for [string] etc.
}

// Method describes a static stdlib method.
type Method struct {
	Name          string
	ReturnType    *ast.TypeDesc
	MinArgs       int
	MaxArgs       int       // -1 means unbounded (only Console.print)
	Args          []ArgSpec // used when MinArgs == MaxArgs and len(Args) > 0
	PrintableArgs bool      // Console.print: int/float/bool/string
}

// Class describes a builtin stdlib class.
type Class struct {
	Name    string
	Methods map[string]Method
}

func typ(base string) *ast.TypeDesc {
	return &ast.TypeDesc{Base: base}
}

func arrTyp(elem string) *ast.TypeDesc {
	return &ast.TypeDesc{IsArray: true, ElemType: &ast.TypeDesc{Base: elem}}
}

// Classes is the canonical stdlib registry.
var Classes = []Class{
	{
		Name: Console,
		Methods: map[string]Method{
			"print": {
				Name:          "print",
				ReturnType:    typ("void"),
				MinArgs:       1,
				MaxArgs:       -1,
				PrintableArgs: true,
			},
			"readLine": {
				Name:       "readLine",
				ReturnType: typ("string"),
				MinArgs:    0,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "string"}},
			},
			"readInt": {
				Name:       "readInt",
				ReturnType: typ("int"),
				MinArgs:    0,
				MaxArgs:    0,
			},
		},
	},
	{
		Name: File,
		Methods: map[string]Method{
			"read": {
				Name:       "read",
				ReturnType: typ("string"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "string"}},
			},
			"write": {
				Name:       "write",
				ReturnType: typ("void"),
				MinArgs:    2,
				MaxArgs:    2,
				Args:       []ArgSpec{{Base: "string"}, {Base: "string"}},
			},
			"append": {
				Name:       "append",
				ReturnType: typ("void"),
				MinArgs:    2,
				MaxArgs:    2,
				Args:       []ArgSpec{{Base: "string"}, {Base: "string"}},
			},
			"exists": {
				Name:       "exists",
				ReturnType: typ("bool"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "string"}},
			},
		},
	},
	{
		Name: Time,
		Methods: map[string]Method{
			"now": {
				Name:       "now",
				ReturnType: typ("int"),
				MinArgs:    0,
				MaxArgs:    0,
			},
			"sleepMillis": {
				Name:       "sleepMillis",
				ReturnType: typ("void"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "int"}},
			},
			"format": {
				Name:       "format",
				ReturnType: typ("string"),
				MinArgs:    2,
				MaxArgs:    2,
				Args:       []ArgSpec{{Base: "int"}, {Base: "string"}},
			},
		},
	},
	{
		Name: String,
		Methods: map[string]Method{
			"length": {
				Name:       "length",
				ReturnType: typ("int"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "string"}},
			},
			"trim": {
				Name:       "trim",
				ReturnType: typ("string"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "string"}},
			},
			"split": {
				Name:       "split",
				ReturnType: arrTyp("string"),
				MinArgs:    2,
				MaxArgs:    2,
				Args:       []ArgSpec{{Base: "string"}, {Base: "string"}},
			},
			"contains": {
				Name:       "contains",
				ReturnType: typ("bool"),
				MinArgs:    2,
				MaxArgs:    2,
				Args:       []ArgSpec{{Base: "string"}, {Base: "string"}},
			},
			"substring": {
				Name:       "substring",
				ReturnType: typ("string"),
				MinArgs:    3,
				MaxArgs:    3,
				Args:       []ArgSpec{{Base: "string"}, {Base: "int"}, {Base: "int"}},
			},
		},
	},
	{
		Name: Math,
		Methods: map[string]Method{
			"abs": {
				Name:       "abs",
				ReturnType: typ("float"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "float"}},
			},
			"min": {
				Name:       "min",
				ReturnType: typ("float"),
				MinArgs:    2,
				MaxArgs:    2,
				Args:       []ArgSpec{{Base: "float"}, {Base: "float"}},
			},
			"max": {
				Name:       "max",
				ReturnType: typ("float"),
				MinArgs:    2,
				MaxArgs:    2,
				Args:       []ArgSpec{{Base: "float"}, {Base: "float"}},
			},
			"floor": {
				Name:       "floor",
				ReturnType: typ("int"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "float"}},
			},
			"random": {
				Name:       "random",
				ReturnType: typ("float"),
				MinArgs:    0,
				MaxArgs:    0,
			},
		},
	},
	{
		Name: Args,
		Methods: map[string]Method{
			"count": {
				Name:       "count",
				ReturnType: typ("int"),
				MinArgs:    0,
				MaxArgs:    0,
			},
			"at": {
				Name:       "at",
				ReturnType: typ("string"),
				MinArgs:    1,
				MaxArgs:    1,
				Args:       []ArgSpec{{Base: "int"}},
			},
		},
	},
}

// BuiltinID returns the dispatch key "Class.method".
func BuiltinID(class, method string) string {
	return class + "." + method
}

// ReservedNames returns all stdlib class names.
func ReservedNames() []string {
	names := make([]string, len(Classes))
	for i, c := range Classes {
		names[i] = c.Name
	}
	return names
}

// IsBuiltin reports whether name is a reserved stdlib class.
func IsBuiltin(name string) bool {
	for _, c := range Classes {
		if c.Name == name {
			return true
		}
	}
	return false
}

// Lookup returns the method descriptor for a static builtin call.
func Lookup(class, method string) (*Method, bool) {
	for _, c := range Classes {
		if c.Name != class {
			continue
		}
		m, ok := c.Methods[method]
		if ok {
			return &m, true
		}
		return nil, false
	}
	return nil, false
}

// ReturnsVoid reports whether a builtin call has void result.
func ReturnsVoid(class, method string) bool {
	m, ok := Lookup(class, method)
	if !ok || m.ReturnType == nil {
		return true
	}
	return m.ReturnType.Base == "void"
}
