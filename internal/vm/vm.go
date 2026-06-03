package vm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/unisc/compiladores/sol/internal/semantic"
	"github.com/unisc/compiladores/sol/internal/tac"
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
}

// New creates a VM for the given instructions and class metadata.
func New(instrs []tac.Instr, classes map[string]*semantic.ClassInfo) *VM {
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
			val, err := vm.resolveValue(ins.Arg1)
			if err != nil {
				return err
			}
			if err := vm.store(ins.Result, val); err != nil {
				return err
			}
			vm.pc++
		case tac.OpBinOp:
			val, err := vm.evalBinOp(ins)
			if err != nil {
				return err
			}
			if err := vm.store(ins.Result, val); err != nil {
				return err
			}
			vm.pc++
		case tac.OpGoto:
			vm.pc = vm.labels[ins.Label] + 1
		case tac.OpIfFalse:
			val, err := vm.resolveValue(ins.Arg1)
			if err != nil {
				return err
			}
			if !val.AsBool() {
				vm.pc = vm.labels[ins.Label] + 1
			} else {
				vm.pc++
			}
		case tac.OpParam:
			v, err := vm.resolveValue(ins.Arg1)
			if err != nil {
				return err
			}
			vm.params = append(vm.params, v)
			vm.pc++
		case tac.OpNew:
			obj := Obj(ins.Arg1)
			if err := vm.store(ins.Result, obj); err != nil {
				return err
			}
			vm.pc++
		case tac.OpCall:
			n, _ := strconv.Atoi(ins.Arg2)
			if err := vm.startCall(ins.Arg1, n); err != nil {
				return err
			}
		case tac.OpPrint:
			n, _ := strconv.Atoi(ins.Arg2)
			if err := vm.doPrint(n); err != nil {
				return err
			}
			vm.pc++
		case tac.OpReadLine:
			val, err := vm.doReadLine(ins)
			if err != nil {
				return err
			}
			if err := vm.store(ins.Result, val); err != nil {
				return err
			}
			vm.pc++
		case tac.OpReadInt:
			val, err := vm.doReadInt()
			if err != nil {
				return err
			}
			if err := vm.store(ins.Result, val); err != nil {
				return err
			}
			vm.pc++
		case tac.OpFileRead:
			val, err := vm.doFileRead(ins.Arg1)
			if err != nil {
				return err
			}
			if err := vm.store(ins.Result, val); err != nil {
				return err
			}
			vm.pc++
		case tac.OpFileWrite:
			n, _ := strconv.Atoi(ins.Arg2)
			if err := vm.doFileWrite(n); err != nil {
				return err
			}
			vm.pc++
		case tac.OpArrayLit:
			n, _ := strconv.Atoi(ins.Arg2)
			if err := vm.doArrayLit(ins.Result, n); err != nil {
				return err
			}
			vm.pc++
		case tac.OpThrow:
			msg, err := vm.resolveValue(ins.Arg1)
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
				varName: ins.Arg1,
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
			if ins.Arg1 != "" {
				v, err := vm.resolveValue(ins.Arg1)
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
			_ = vm.store(h.varName, msg)
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
	return vm.store(result, Arr(args...))
}

func formatPrintable(v Value) string {
	return v.String()
}

func (vm *VM) doReadLine(ins tac.Instr) (Value, error) {
	n, _ := strconv.Atoi(ins.Arg2)
	if n > len(vm.params) {
		return Str(""), fmt.Errorf("readLine: expected %d params, got %d", n, len(vm.params))
	}
	if n > 0 {
		prompt := vm.params[len(vm.params)-n]
		vm.params = vm.params[:len(vm.params)-n]
		fmt.Print(prompt.String())
	}
	line, err := bufio.NewReader(vm.reader()).ReadString('\n')
	if err != nil && err != io.EOF {
		return Str(""), fmt.Errorf("readLine: %w", err)
	}
	line = strings.TrimSuffix(line, "\n")
	return Str(line), nil
}

func (vm *VM) doReadInt() (Value, error) {
	line, err := vm.doReadLine(tac.Instr{Arg2: "0"})
	if err != nil {
		return Int(0), err
	}
	i, err := strconv.ParseInt(strings.TrimSpace(line.StrVal), 10, 64)
	if err != nil {
		return Int(0), fmt.Errorf("readInt: invalid integer %q", line.StrVal)
	}
	return Int(i), nil
}

func (vm *VM) doFileRead(pathRef string) (Value, error) {
	pathVal, err := vm.resolveValue(pathRef)
	if err != nil {
		return Str(""), err
	}
	if pathVal.Kind != KindString {
		return Str(""), fmt.Errorf("fileRead: path must be string")
	}
	data, err := os.ReadFile(pathVal.StrVal)
	if err != nil {
		return Str(""), fmt.Errorf("fileRead: %w", err)
	}
	return Str(string(data)), nil
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
		locals["enlights"] = vm.superView(this, ci.Super)
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
	if ins.Arg1 == "!" {
		v, err := vm.resolveValue(ins.Arg2)
		if err != nil {
			return Null(), err
		}
		return Bool(!v.AsBool()), nil
	}
	if ins.Arg1 == "-" && ins.Operator == "" {
		v, err := vm.resolveValue(ins.Arg2)
		if err != nil {
			return Null(), err
		}
		return Float(-v.AsFloat()), nil
	}
	left, err := vm.resolveValue(ins.Arg1)
	if err != nil {
		return Null(), err
	}
	right, err := vm.resolveValue(ins.Arg2)
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
	default:
		return false
	}
}

func (vm *VM) resolveValue(name string) (Value, error) {
	name = strings.TrimSpace(name)
	if name == "call_result" {
		return vm.callResult, nil
	}
	if strings.HasPrefix(name, `"`) {
		return parseLiteral(name)
	}
	if idx := strings.Index(name, "["); idx > 0 && strings.HasSuffix(name, "]") {
		base := name[:idx]
		idxStr := name[idx+1 : len(name)-1]
		arr, err := vm.resolveValue(base)
		if err != nil {
			return Null(), err
		}
		iVal, err := vm.resolveValue(idxStr)
		if err != nil {
			return Null(), err
		}
		i := int(iVal.AsFloat())
		if arr.Kind != KindArray || i < 0 || i >= len(arr.Array) {
			return Null(), fmt.Errorf("array index out of range: %s", name)
		}
		return arr.Array[i], nil
	}
	if strings.Contains(name, ".") {
		return vm.resolveFieldPath(name)
	}
	if v, ok := vm.lookup(name); ok {
		return v, nil
	}
	return parseLiteral(name)
}

func (vm *VM) resolveFieldPath(path string) (Value, error) {
	parts := strings.Split(path, ".")
	root, err := vm.lookupOrLiteral(parts[0])
	if err != nil {
		return Null(), err
	}
	cur := root
	for _, field := range parts[1:] {
		if cur.Kind == KindArray && field == "length" {
			return Int(int64(len(cur.Array))), nil
		}
		if cur.Kind != KindObject {
			return Null(), fmt.Errorf("field access on non-object: %s", path)
		}
		v, ok := cur.Object.Fields[field]
		if !ok {
			return Null(), fmt.Errorf("unknown field %q", field)
		}
		cur = v
	}
	return cur, nil
}

func (vm *VM) lookupOrLiteral(name string) (Value, error) {
	if v, ok := vm.lookup(name); ok {
		return v, nil
	}
	return parseLiteral(name)
}

func (vm *VM) store(target string, val Value) error {
	target = strings.TrimSpace(target)
	if idx := strings.Index(target, "["); idx > 0 && strings.HasSuffix(target, "]") {
		base := target[:idx]
		idxStr := target[idx+1 : len(target)-1]
		arr, err := vm.resolveValue(base)
		if err != nil {
			return err
		}
		iVal, err := vm.resolveValue(idxStr)
		if err != nil {
			return err
		}
		i := int(iVal.AsFloat())
		if arr.Kind != KindArray || i < 0 || i >= len(arr.Array) {
			return fmt.Errorf("array index out of range: %s", target)
		}
		arr.Array[i] = val
		return vm.store(base, arr)
	}
	if strings.Contains(target, ".") {
		parts := strings.Split(target, ".")
		rootName := parts[0]
		field := parts[len(parts)-1]
		obj, err := vm.lookupOrLiteral(rootName)
		if err != nil {
			return err
		}
		if obj.Kind != KindObject {
			return fmt.Errorf("cannot assign field on non-object")
		}
		obj.Object.Fields[field] = val
		if rootName == "this" || rootName == "enlights" {
			if len(vm.frames) > 0 {
				vm.frames[len(vm.frames)-1].locals[rootName] = obj
			}
		} else {
			vm.globals[rootName] = obj
		}
		return nil
	}
	if len(vm.frames) > 0 {
		vm.frames[len(vm.frames)-1].locals[target] = val
	} else {
		vm.globals[target] = val
	}
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
