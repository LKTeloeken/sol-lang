package vm

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/unisc/compiladores/sol/src/semantic"
	"github.com/unisc/compiladores/sol/src/tac"
)

type frame struct {
	locals map[string]Value
	retPC  int
}

type catchHandler struct {
	varName string
	pc      int
}

// VM executes TAC instructions.
type VM struct {
	instrs     []tac.Instr
	labels     map[string]int
	globals    map[string]Value
	classes    map[string]*semantic.ClassInfo
	frames     []frame
	params     []Value
	pc         int
	callResult Value
	catchStack []catchHandler
	inCatch    bool
	stdin      io.Reader
	scriptArgs []string
}

// New creates a VM for the given instructions and class metadata.
func New(instrs []tac.Instr, classes map[string]*semantic.ClassInfo) *VM {
	rand.Seed(time.Now().UnixNano())
	vm := &VM{
		instrs:  instrs,
		labels:  make(map[string]int),
		globals: make(map[string]Value),
		classes: classes,
	}
	for i, ins := range instrs {
		if ins.Op == tac.OpLabel {
			vm.labels[ins.Label] = i
		}
	}
	return vm
}

// SetStdin configures the input reader for Console.readLine/readInt (defaults to os.Stdin).
func (vm *VM) SetStdin(r io.Reader) {
	vm.stdin = r
}

func (vm *VM) reader() io.Reader {
	if vm.stdin != nil {
		return vm.stdin
	}
	return os.Stdin
}

// Run executes from __program until __end.
func (vm *VM) Run() error {
	start, ok := vm.labels["__program"]
	if !ok {
		return fmt.Errorf("missing __program label")
	}
	vm.pc = start + 1
	endPC := len(vm.instrs)
	if end, ok := vm.labels["__end"]; ok {
		endPC = end
	}

	for vm.pc < endPC {
		ins := vm.instrs[vm.pc]
		switch ins.Op {
		case tac.OpComment, tac.OpLabel:
			vm.pc++
		case tac.OpAssign:
			val, err := vm.resolveOperand(ins.A)
			if err != nil {
				return err
			}
			vm.storeName(ins.Dst, val)
			vm.pc++
		case tac.OpBinOp:
			val, err := vm.evalBinOp(ins)
			if err != nil {
				return err
			}
			vm.storeName(ins.Dst, val)
			vm.pc++
		case tac.OpUnary:
			val, err := vm.evalUnary(ins)
			if err != nil {
				return err
			}
			vm.storeName(ins.Dst, val)
			vm.pc++
		case tac.OpFieldGet:
			val, err := vm.loadField(ins.Obj, ins.Field)
			if err != nil {
				return err
			}
			vm.storeName(ins.Dst, val)
			vm.pc++
		case tac.OpFieldSet:
			val, err := vm.resolveOperand(ins.A)
			if err != nil {
				return err
			}
			if err := vm.storeField(ins.Obj, ins.Field, val); err != nil {
				return err
			}
			vm.pc++
		case tac.OpIndexGet:
			val, err := vm.loadIndex(ins.Obj, ins.Idx)
			if err != nil {
				return err
			}
			vm.storeName(ins.Dst, val)
			vm.pc++
		case tac.OpIndexSet:
			val, err := vm.resolveOperand(ins.A)
			if err != nil {
				return err
			}
			if err := vm.storeIndex(ins.Obj, ins.Idx, val); err != nil {
				return err
			}
			vm.pc++
		case tac.OpLen:
			arr, err := vm.resolveOperand(ins.Obj)
			if err != nil {
				return err
			}
			if arr.Kind != KindArray {
				return fmt.Errorf("len: operand is not an array")
			}
			vm.storeName(ins.Dst, Int(int64(len(arr.Array))))
			vm.pc++
		case tac.OpGoto:
			vm.pc = vm.labels[ins.Label] + 1
		case tac.OpIfFalse:
			val, err := vm.resolveOperand(ins.A)
			if err != nil {
				return err
			}
			if !val.AsBool() {
				vm.pc = vm.labels[ins.Label] + 1
			} else {
				vm.pc++
			}
		case tac.OpParam:
			v, err := vm.resolveOperand(ins.A)
			if err != nil {
				return err
			}
			vm.params = append(vm.params, v)
			vm.pc++
		case tac.OpNew:
			vm.storeName(ins.Dst, Obj(ins.Sym))
			vm.pc++
		case tac.OpCall:
			if err := vm.startCall(ins.Sym, ins.NArgs); err != nil {
				return err
			}
		case tac.OpBuiltin:
			if err := vm.doBuiltin(ins); err != nil {
				return err
			}
			vm.pc++
		case tac.OpArrayLit:
			if err := vm.doArrayLit(ins.Dst, ins.NArgs); err != nil {
				return err
			}
			vm.pc++
		case tac.OpArrayCall:
			if err := vm.doArrayCall(ins); err != nil {
				return err
			}
			vm.pc++
		case tac.OpThrow:
			msg, err := vm.resolveOperand(ins.A)
			if err != nil {
				return err
			}
			if vm.handleThrow(msg) {
				continue
			}
			text := msg.StrVal
			if msg.Kind != KindString {
				text = msg.String()
			}
			return fmt.Errorf("flare: %s", text)
		case tac.OpBeginTry:
			vm.catchStack = append(vm.catchStack, catchHandler{
				varName: ins.Sym,
				pc:      vm.labels[ins.Label] + 1,
			})
			vm.inCatch = false
			vm.pc++
		case tac.OpEndTry:
			if len(vm.catchStack) > 0 {
				vm.catchStack = vm.catchStack[:len(vm.catchStack)-1]
			}
			vm.inCatch = false
			vm.pc = vm.labels[ins.Label] + 1
		case tac.OpReturn:
			if !ins.A.IsZero() {
				v, err := vm.resolveOperand(ins.A)
				if err != nil {
					return err
				}
				vm.callResult = v
			} else {
				vm.callResult = Null()
			}
			if len(vm.frames) == 0 {
				vm.pc++
			} else {
				f := vm.frames[len(vm.frames)-1]
				vm.frames = vm.frames[:len(vm.frames)-1]
				vm.pc = f.retPC
			}
		default:
			vm.pc++
		}
	}
	return nil
}

func (vm *VM) handleThrow(msg Value) bool {
	for len(vm.catchStack) > 0 {
		i := len(vm.catchStack) - 1
		h := vm.catchStack[i]
		vm.catchStack = vm.catchStack[:i]
		if h.varName != "" {
			vm.storeName(h.varName, msg)
		}
		vm.pc = h.pc
		return true
	}
	return false
}

func (vm *VM) doPrint(n int) error {
	if n > len(vm.params) {
		return fmt.Errorf("print: expected %d params, got %d", n, len(vm.params))
	}
	args := vm.params[len(vm.params)-n:]
	vm.params = vm.params[:len(vm.params)-n]
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatPrintable(arg)
	}
	fmt.Println(strings.Join(parts, " "))
	return nil
}

func (vm *VM) doArrayLit(result string, n int) error {
	if n > len(vm.params) {
		return fmt.Errorf("arrayLit: expected %d params, got %d", n, len(vm.params))
	}
	args := vm.params[len(vm.params)-n:]
	vm.params = vm.params[:len(vm.params)-n]
	vm.storeName(result, Arr(args...))
	return nil
}

func formatPrintable(v Value) string {
	return v.String()
}

func (vm *VM) doReadInt() (Value, error) {
	line, err := vm.doReadLine(0)
	if err != nil {
		return Int(0), err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(line.StrVal), 10, 64)
	if err != nil {
		return Int(0), fmt.Errorf("readInt: invalid integer %q", line.StrVal)
	}
	return Int(i), nil
}

func (vm *VM) doFileWrite(n int) error {
	if n > len(vm.params) {
		return fmt.Errorf("fileWrite: expected %d params, got %d", n, len(vm.params))
	}
	args := vm.params[len(vm.params)-n:]
	vm.params = vm.params[:len(vm.params)-n]
	if len(args) != 2 {
		return fmt.Errorf("fileWrite: expected 2 params")
	}
	path := args[0].StrVal
	content := args[1].StrVal
	if args[0].Kind != KindString || args[1].Kind != KindString {
		return fmt.Errorf("fileWrite: path and content must be strings")
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("fileWrite: %w", err)
	}
	return nil
}

// Global returns a top-level variable value.
func (vm *VM) Global(name string) (Value, bool) {
	v, ok := vm.globals[name]
	return v, ok
}

// GetField reads obj.field from a global object variable.
func (vm *VM) GetField(objName, field string) (Value, error) {
	obj, ok := vm.globals[objName]
	if !ok || obj.Kind != KindObject {
		return Null(), fmt.Errorf("global %q is not an object", objName)
	}
	v, ok := obj.Object.Fields[field]
	if !ok {
		return Null(), fmt.Errorf("field %q not found", field)
	}
	return v, nil
}

func (vm *VM) startCall(label string, nArgs int) error {
	if nArgs > len(vm.params) {
		return fmt.Errorf("call %s: expected %d params, got %d", label, nArgs, len(vm.params))
	}
	args := vm.params[len(vm.params)-nArgs:]
	vm.params = vm.params[:len(vm.params)-nArgs]

	parts := strings.SplitN(label, ".", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid call label %q", label)
	}
	className, methodName := parts[0], parts[1]
	ci := vm.classes[className]
	if ci == nil {
		return fmt.Errorf("unknown class %q", className)
	}

	this := args[0]
	locals := map[string]Value{"this": this}
	var paramNames []string
	if methodName == "glow" && ci.Constructor != nil {
		for _, p := range ci.Constructor.Params {
			paramNames = append(paramNames, p.Name)
		}
	} else if mi := findMethod(ci, methodName); mi != nil {
		for _, p := range mi.Params {
			paramNames = append(paramNames, p.Name)
		}
	}
	for i, name := range paramNames {
		if i+1 < len(args) {
			locals[name] = args[i+1]
		}
	}
	if ci.Super != nil {
		locals["radiate"] = vm.superView(this, ci.Super)
	}

	target, ok := vm.labels[label]
	if !ok {
		return fmt.Errorf("unknown method %q", label)
	}
	vm.frames = append(vm.frames, frame{locals: locals, retPC: vm.pc + 1})
	vm.pc = target + 1
	return nil
}

func findMethod(ci *semantic.ClassInfo, name string) *semantic.MethodInfo {
	for ci != nil {
		if mi, ok := ci.Methods[name]; ok {
			return mi
		}
		ci = ci.Super
	}
	return nil
}

func (vm *VM) superView(this Value, super *semantic.ClassInfo) Value {
	parent := Obj(super.Name)
	if this.Kind != KindObject {
		return parent
	}
	for fname := range super.Fields {
		if v, ok := this.Object.Fields[fname]; ok {
			parent.Object.Fields[fname] = v
		}
	}
	return parent
}

func (vm *VM) evalBinOp(ins tac.Instr) (Value, error) {
	left, err := vm.resolveOperand(ins.A)
	if err != nil {
		return Null(), err
	}
	right, err := vm.resolveOperand(ins.B)
	if err != nil {
		return Null(), err
	}
	if ins.Operator == "+" && (left.Kind == KindString || right.Kind == KindString) {
		return Str(formatPrintable(left) + formatPrintable(right)), nil
	}
	lf, rf := left.AsFloat(), right.AsFloat()
	switch ins.Operator {
	case "+":
		return Float(lf + rf), nil
	case "-":
		return Float(lf - rf), nil
	case "*":
		return Float(lf * rf), nil
	case "/":
		return Float(lf / rf), nil
	case "%":
		return Int(int64(lf) % int64(rf)), nil
	case "<":
		return Bool(lf < rf), nil
	case ">":
		return Bool(lf > rf), nil
	case "<=":
		return Bool(lf <= rf), nil
	case ">=":
		return Bool(lf >= rf), nil
	case "==":
		return Bool(valuesEqual(left, right)), nil
	case "!=":
		return Bool(!valuesEqual(left, right)), nil
	case "&&":
		return Bool(left.AsBool() && right.AsBool()), nil
	case "||":
		return Bool(left.AsBool() || right.AsBool()), nil
	default:
		return Null(), fmt.Errorf("unknown operator %q", ins.Operator)
	}
}

func (vm *VM) evalUnary(ins tac.Instr) (Value, error) {
	v, err := vm.resolveOperand(ins.A)
	if err != nil {
		return Null(), err
	}
	switch ins.Operator {
	case "!":
		return Bool(!v.AsBool()), nil
	case "-":
		return Float(-v.AsFloat()), nil
	default:
		return Null(), fmt.Errorf("unknown unary operator %q", ins.Operator)
	}
}

func valuesEqual(a, b Value) bool {
	if a.Kind != b.Kind {
		if (a.Kind == KindInt || a.Kind == KindFloat) && (b.Kind == KindInt || b.Kind == KindFloat) {
			return a.AsFloat() == b.AsFloat()
		}
		return false
	}
	switch a.Kind {
	case KindInt:
		return a.Int == b.Int
	case KindFloat:
		return a.Float == b.Float
	case KindBool:
		return a.Bool == b.Bool
	case KindString:
		return a.StrVal == b.StrVal
	case KindNull:
		return true
	case KindObject:
		return a.Object == b.Object
	default:
		return false
	}
}

// resolveOperand turns a typed TAC operand into a runtime Value. There is no
// string parsing: a constant carries its literal directly and a name is a
// straight variable/temporary lookup.
func (vm *VM) resolveOperand(o tac.Operand) (Value, error) {
	switch o.Kind {
	case tac.OpndConst:
		return constValue(o), nil
	case tac.OpndCallResult:
		return vm.callResult, nil
	case tac.OpndName:
		if v, ok := vm.lookup(o.Name); ok {
			return v, nil
		}
		// An unset variable reads as null (e.g. `var x int;` with no initializer).
		return Null(), nil
	default:
		return Null(), nil
	}
}

func constValue(o tac.Operand) Value {
	switch o.Lit {
	case tac.LitInt:
		return Int(o.Int)
	case tac.LitFloat:
		return Float(o.Float)
	case tac.LitBool:
		return Bool(o.Bool)
	case tac.LitStr:
		return Str(o.Str)
	default:
		return Null()
	}
}

// storeName assigns to a plain variable/temporary, in the current frame if any.
func (vm *VM) storeName(name string, val Value) {
	if len(vm.frames) > 0 {
		vm.frames[len(vm.frames)-1].locals[name] = val
	} else {
		vm.globals[name] = val
	}
}

// loadField reads obj.field (objects only; array length is handled by OpLen).
func (vm *VM) loadField(objOp tac.Operand, field string) (Value, error) {
	obj, err := vm.resolveOperand(objOp)
	if err != nil {
		return Null(), err
	}
	if obj.Kind != KindObject {
		return Null(), fmt.Errorf("field access %q on non-object", field)
	}
	v, ok := obj.Object.Fields[field]
	if !ok {
		return Null(), fmt.Errorf("unknown field %q", field)
	}
	return v, nil
}

// storeField writes obj.field. The object is a pointer, so the mutation is
// visible through every alias without re-storing the object.
func (vm *VM) storeField(objOp tac.Operand, field string, val Value) error {
	obj, err := vm.resolveOperand(objOp)
	if err != nil {
		return err
	}
	if obj.Kind != KindObject {
		return fmt.Errorf("cannot set field %q on non-object", field)
	}
	obj.Object.Fields[field] = val
	return nil
}

func (vm *VM) loadIndex(arrOp, idxOp tac.Operand) (Value, error) {
	arr, err := vm.resolveOperand(arrOp)
	if err != nil {
		return Null(), err
	}
	idxVal, err := vm.resolveOperand(idxOp)
	if err != nil {
		return Null(), err
	}
	i := int(idxVal.AsFloat())
	if arr.Kind != KindArray || i < 0 || i >= len(arr.Array) {
		return Null(), fmt.Errorf("array index out of range")
	}
	return arr.Array[i], nil
}

func (vm *VM) storeIndex(arrOp, idxOp tac.Operand, val Value) error {
	arr, err := vm.resolveOperand(arrOp)
	if err != nil {
		return err
	}
	idxVal, err := vm.resolveOperand(idxOp)
	if err != nil {
		return err
	}
	i := int(idxVal.AsFloat())
	if arr.Kind != KindArray || i < 0 || i >= len(arr.Array) {
		return fmt.Errorf("array index out of range")
	}
	arr.Array[i] = val
	return nil
}

func (vm *VM) lookup(name string) (Value, bool) {
	for i := len(vm.frames) - 1; i >= 0; i-- {
		if v, ok := vm.frames[i].locals[name]; ok {
			return v, true
		}
	}
	v, ok := vm.globals[name]
	return v, ok
}
