package vm

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/unisc/compiladores/sol/src/tac"
)

// SetScriptArgs sets CLI arguments available to Args.count/at.
func (vm *VM) SetScriptArgs(args []string) {
	vm.scriptArgs = append([]string(nil), args...)
}

func (vm *VM) doBuiltin(ins tac.Instr) error {
	val, err := vm.dispatchBuiltin(ins.Sym, ins.NArgs)
	if err != nil {
		return err
	}
	if ins.Dst != "" {
		vm.storeName(ins.Dst, val)
	}
	return nil
}

func (vm *VM) dispatchBuiltin(id string, n int) (Value, error) {
	switch id {
	case "Console.print":
		if err := vm.doPrint(n); err != nil {
			return Null(), err
		}
		return Null(), nil
	case "Console.readLine":
		return vm.doReadLine(n)
	case "Console.readInt":
		return vm.doReadInt()
	case "File.read":
		return vm.doFileRead(n)
	case "File.write":
		if err := vm.doFileWrite(n); err != nil {
			return Null(), err
		}
		return Null(), nil
	case "File.append":
		if err := vm.doFileAppend(n); err != nil {
			return Null(), err
		}
		return Null(), nil
	case "File.exists":
		return vm.doFileExists(n)
	case "Time.now":
		return Int(time.Now().Unix()), nil
	case "Time.sleepMillis":
		return vm.doTimeSleep(n)
	case "Time.format":
		return vm.doTimeFormat(n)
	case "String.length":
		return vm.doStringLength(n)
	case "String.trim":
		return vm.doStringTrim(n)
	case "String.split":
		return vm.doStringSplit(n)
	case "String.contains":
		return vm.doStringContains(n)
	case "String.substring":
		return vm.doStringSubstring(n)
	case "Math.abs":
		return vm.doMathAbs(n)
	case "Math.min":
		return vm.doMathMin(n)
	case "Math.max":
		return vm.doMathMax(n)
	case "Math.floor":
		return vm.doMathFloor(n)
	case "Math.random":
		return Float(rand.Float64()), nil
	case "Args.count":
		return Int(int64(len(vm.scriptArgs))), nil
	case "Args.at":
		return vm.doArgsAt(n)
	default:
		return Null(), fmt.Errorf("unknown builtin %s", id)
	}
}

func (vm *VM) popParams(n int) ([]Value, error) {
	if n > len(vm.params) {
		return nil, fmt.Errorf("builtin: expected %d params, got %d", n, len(vm.params))
	}
	args := vm.params[len(vm.params)-n:]
	vm.params = vm.params[:len(vm.params)-n]
	return args, nil
}

func (vm *VM) requireString(v Value, what string) (string, error) {
	if v.Kind != KindString {
		return "", fmt.Errorf("%s: expected string", what)
	}
	return v.StrVal, nil
}

func (vm *VM) requireInt(v Value, what string) (int64, error) {
	switch v.Kind {
	case KindInt:
		return v.Int, nil
	case KindFloat:
		return int64(v.Float), nil
	default:
		return 0, fmt.Errorf("%s: expected int", what)
	}
}

func (vm *VM) requireFloat(v Value, what string) (float64, error) {
	switch v.Kind {
	case KindFloat:
		return v.Float, nil
	case KindInt:
		return float64(v.Int), nil
	default:
		return 0, fmt.Errorf("%s: expected float", what)
	}
}

func (vm *VM) doReadLine(n int) (Value, error) {
	if n > 0 {
		args, err := vm.popParams(n)
		if err != nil {
			return Str(""), err
		}
		fmt.Print(args[0].String())
	}
	line, err := bufio.NewReader(vm.reader()).ReadString('\n')
	if err != nil && err != io.EOF {
		return Str(""), fmt.Errorf("readLine: %w", err)
	}
	line = strings.TrimSuffix(line, "\n")
	return Str(line), nil
}

func (vm *VM) doFileRead(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Str(""), err
	}
	path, err := vm.requireString(args[0], "fileRead")
	if err != nil {
		return Str(""), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Str(""), fmt.Errorf("fileRead: %w", err)
	}
	return Str(string(data)), nil
}

func (vm *VM) doFileAppend(n int) error {
	args, err := vm.popParams(n)
	if err != nil {
		return err
	}
	path, err := vm.requireString(args[0], "fileAppend")
	if err != nil {
		return err
	}
	content, err := vm.requireString(args[1], "fileAppend")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("fileAppend: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return fmt.Errorf("fileAppend: %w", err)
	}
	return nil
}

func (vm *VM) doFileExists(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Bool(false), err
	}
	path, err := vm.requireString(args[0], "fileExists")
	if err != nil {
		return Bool(false), err
	}
	_, err = os.Stat(path)
	return Bool(err == nil), nil
}

func (vm *VM) doTimeSleep(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Null(), err
	}
	ms, err := vm.requireInt(args[0], "sleepMillis")
	if err != nil {
		return Null(), err
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return Null(), nil
}

func (vm *VM) doTimeFormat(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Str(""), err
	}
	unix, err := vm.requireInt(args[0], "format")
	if err != nil {
		return Str(""), err
	}
	layout, err := vm.requireString(args[1], "format")
	if err != nil {
		return Str(""), err
	}
	return Str(time.Unix(unix, 0).Format(layout)), nil
}

func (vm *VM) doStringLength(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Int(0), err
	}
	s, err := vm.requireString(args[0], "length")
	if err != nil {
		return Int(0), err
	}
	return Int(int64(len(s))), nil
}

func (vm *VM) doStringTrim(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Str(""), err
	}
	s, err := vm.requireString(args[0], "trim")
	if err != nil {
		return Str(""), err
	}
	return Str(strings.TrimSpace(s)), nil
}

func (vm *VM) doStringSplit(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Arr(), err
	}
	s, err := vm.requireString(args[0], "split")
	if err != nil {
		return Arr(), err
	}
	sep, err := vm.requireString(args[1], "split")
	if err != nil {
		return Arr(), err
	}
	parts := strings.Split(s, sep)
	out := make([]Value, len(parts))
	for i, p := range parts {
		out[i] = Str(p)
	}
	return Arr(out...), nil
}

func (vm *VM) doStringContains(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Bool(false), err
	}
	s, err := vm.requireString(args[0], "contains")
	if err != nil {
		return Bool(false), err
	}
	sub, err := vm.requireString(args[1], "contains")
	if err != nil {
		return Bool(false), err
	}
	return Bool(strings.Contains(s, sub)), nil
}

func (vm *VM) doStringSubstring(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Str(""), err
	}
	s, err := vm.requireString(args[0], "substring")
	if err != nil {
		return Str(""), err
	}
	start, err := vm.requireInt(args[1], "substring")
	if err != nil {
		return Str(""), err
	}
	end, err := vm.requireInt(args[2], "substring")
	if err != nil {
		return Str(""), err
	}
	if start < 0 || end < start || int(end) > len(s) {
		return Str(""), fmt.Errorf("substring: invalid range [%d:%d] for length %d", start, end, len(s))
	}
	return Str(s[start:end]), nil
}

func (vm *VM) doMathAbs(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Float(0), err
	}
	x, err := vm.requireFloat(args[0], "abs")
	if err != nil {
		return Float(0), err
	}
	return Float(math.Abs(x)), nil
}

func (vm *VM) doMathMin(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Float(0), err
	}
	a, err := vm.requireFloat(args[0], "min")
	if err != nil {
		return Float(0), err
	}
	b, err := vm.requireFloat(args[1], "min")
	if err != nil {
		return Float(0), err
	}
	return Float(math.Min(a, b)), nil
}

func (vm *VM) doMathMax(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Float(0), err
	}
	a, err := vm.requireFloat(args[0], "max")
	if err != nil {
		return Float(0), err
	}
	b, err := vm.requireFloat(args[1], "max")
	if err != nil {
		return Float(0), err
	}
	return Float(math.Max(a, b)), nil
}

func (vm *VM) doMathFloor(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Int(0), err
	}
	x, err := vm.requireFloat(args[0], "floor")
	if err != nil {
		return Int(0), err
	}
	return Int(int64(math.Floor(x))), nil
}

func (vm *VM) doArgsAt(n int) (Value, error) {
	args, err := vm.popParams(n)
	if err != nil {
		return Str(""), err
	}
	i, err := vm.requireInt(args[0], "at")
	if err != nil {
		return Str(""), err
	}
	if i < 0 || int(i) >= len(vm.scriptArgs) {
		return Str(""), fmt.Errorf("flare: Args.at index out of range")
	}
	return Str(vm.scriptArgs[i]), nil
}
