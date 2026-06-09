package vm

import (
	"fmt"
	"strconv"

	"github.com/unisc/compiladores/sol/src/arraymethods"
	"github.com/unisc/compiladores/sol/src/tac"
)

func (vm *VM) doArrayCall(ins tac.Instr) error {
	n, _ := strconv.Atoi(ins.Arg3)
	if n > len(vm.params) {
		return fmt.Errorf("arrayCall: expected %d params, got %d", n, len(vm.params))
	}
	args := vm.params[len(vm.params)-n:]
	vm.params = vm.params[:len(vm.params)-n]

	arrVal, err := vm.resolveValue(ins.Arg1)
	if err != nil {
		return err
	}
	if arrVal.Kind != KindArray {
		return fmt.Errorf("arrayCall: receiver is not an array")
	}

	m, ok := arraymethods.Lookup(ins.Arg2)
	if !ok {
		return fmt.Errorf("arrayCall: unknown method %q", ins.Arg2)
	}

	items := append([]Value(nil), arrVal.Array...)
	var result Value

	switch ins.Arg2 {
	case "push":
		if len(args) != 1 {
			return fmt.Errorf("arrayCall push: expected 1 argument")
		}
		items = append(items, args[0])
	case "pop":
		if len(items) == 0 {
			return fmt.Errorf("flare: pop on empty array")
		}
		result = items[len(items)-1]
		items = items[:len(items)-1]
	case "remove":
		if len(args) != 1 {
			return fmt.Errorf("arrayCall remove: expected 1 argument")
		}
		idx := int(args[0].AsFloat())
		if idx < 0 || idx >= len(items) {
			return fmt.Errorf("flare: array index out of range")
		}
		items = append(items[:idx], items[idx+1:]...)
	case "insert":
		if len(args) != 2 {
			return fmt.Errorf("arrayCall insert: expected 2 arguments")
		}
		idx := int(args[0].AsFloat())
		if idx < 0 || idx > len(items) {
			return fmt.Errorf("flare: array index out of range")
		}
		items = append(items, Value{})
		copy(items[idx+1:], items[idx:])
		items[idx] = args[1]
	case "contains":
		if len(args) != 1 {
			return fmt.Errorf("arrayCall contains: expected 1 argument")
		}
		found := false
		for _, item := range items {
			if valuesEqual(item, args[0]) {
				found = true
				break
			}
		}
		result = Bool(found)
	case "clear":
		items = nil
	case "isEmpty":
		result = Bool(len(items) == 0)
	default:
		return fmt.Errorf("arrayCall: unimplemented method %q", ins.Arg2)
	}

	if m.Mutates {
		if err := vm.store(ins.Arg1, Arr(items...)); err != nil {
			return err
		}
	}

	if ins.Result != "" {
		if m.ReturnType == "void" {
			return vm.store(ins.Result, Null())
		}
		return vm.store(ins.Result, result)
	}
	return nil
}
