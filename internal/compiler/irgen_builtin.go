package compiler

import (
	"fmt"
	"strings"

	"github.com/unisc/compiladores/sol/internal/ast"
	"github.com/unisc/compiladores/sol/internal/stdlib"
)

func (g *IRGen) writeRuntimeDecls(header *strings.Builder) {
	header.WriteString("declare void @sol_args_init(i32, i8**)\n")
	header.WriteString("declare i64 @sol_args_count()\n")
	header.WriteString("declare i8* @sol_args_at(i64)\n")
	header.WriteString("declare i64 @sol_time_now()\n")
	header.WriteString("declare void @sol_sleep_ms(i64)\n")
	header.WriteString("declare i8* @sol_time_format(i64, i8*)\n")
	header.WriteString("declare i64 @sol_str_len(i8*)\n")
	header.WriteString("declare i8* @sol_str_trim(i8*)\n")
	header.WriteString("declare i1 @sol_str_contains(i8*, i8*)\n")
	header.WriteString("declare i8* @sol_str_substring(i8*, i64, i64)\n")
	header.WriteString("declare double @sol_math_abs(double)\n")
	header.WriteString("declare double @sol_math_min(double, double)\n")
	header.WriteString("declare double @sol_math_max(double, double)\n")
	header.WriteString("declare i64 @sol_math_floor(double)\n")
	header.WriteString("declare double @sol_math_random()\n")
	header.WriteString("declare i8* @sol_file_read(i8*)\n")
	header.WriteString("declare void @sol_file_write(i8*, i8*)\n")
	header.WriteString("declare void @sol_file_append(i8*, i8*)\n")
	header.WriteString("declare i1 @sol_file_exists(i8*)\n")
}

func (g *IRGen) genBuiltinCall(class, method string, args []ast.Expr) string {
	id := stdlib.BuiltinID(class, method)
	switch id {
	case "Console.print":
		g.genConsolePrint(args)
		return ""
	case "String.split":
		ptr := g.stringConst(&ast.StringLit{Value: "String.split not supported in native build"})
		fmt.Fprintf(&g.buf, "  call void @sol_panic(i8* %s)\n", ptr)
		return g.freshTemp()
	default:
		return g.genBuiltinByID(id, args)
	}
}

func (g *IRGen) genBuiltinByID(id string, args []ast.Expr) string {
	switch id {
	case "Console.readLine":
		// Native: no stdin read; panic
		ptr := g.stringConst(&ast.StringLit{Value: "Console.readLine not supported in native build"})
		fmt.Fprintf(&g.buf, "  call void @sol_panic(i8* %s)\n", ptr)
		return g.freshTemp()
	case "Console.readInt":
		ptr := g.stringConst(&ast.StringLit{Value: "Console.readInt not supported in native build"})
		fmt.Fprintf(&g.buf, "  call void @sol_panic(i8* %s)\n", ptr)
		return g.freshTemp()
	case "Time.now":
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i64 @sol_time_now()\n", t)
		return t
	case "Time.sleepMillis":
		ms := g.genI64Arg(args, 0)
		fmt.Fprintf(&g.buf, "  call void @sol_sleep_ms(i64 %s)\n", ms)
		return ""
	case "Time.format":
		unix := g.genI64Arg(args, 0)
		layout := g.genStringArg(args, 1)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_time_format(i64 %s, i8* %s)\n", t, unix, layout)
		return t
	case "String.length":
		s := g.genStringArg(args, 0)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i64 @sol_str_len(i8* %s)\n", t, s)
		return t
	case "String.trim":
		s := g.genStringArg(args, 0)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_str_trim(i8* %s)\n", t, s)
		return t
	case "String.contains":
		s := g.genStringArg(args, 0)
		sub := g.genStringArg(args, 1)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i1 @sol_str_contains(i8* %s, i8* %s)\n", t, s, sub)
		return t
	case "String.substring":
		s := g.genStringArg(args, 0)
		start := g.genI64Arg(args, 1)
		end := g.genI64Arg(args, 2)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_str_substring(i8* %s, i64 %s, i64 %s)\n", t, s, start, end)
		return t
	case "Math.abs":
		x := g.genFloatArg(args, 0)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call double @sol_math_abs(double %s)\n", t, x)
		return t
	case "Math.min", "Math.max":
		a := g.genFloatArg(args, 0)
		b := g.genFloatArg(args, 1)
		fn := "@sol_math_min"
		if id == "Math.max" {
			fn = "@sol_math_max"
		}
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call double %s(double %s, double %s)\n", t, fn, a, b)
		return t
	case "Math.floor":
		x := g.genFloatArg(args, 0)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i64 @sol_math_floor(double %s)\n", t, x)
		return t
	case "Math.random":
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call double @sol_math_random()\n", t)
		return t
	case "Args.count":
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i64 @sol_args_count()\n", t)
		return t
	case "Args.at":
		idx := g.genI64Arg(args, 0)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_args_at(i64 %s)\n", t, idx)
		return t
	case "File.read":
		path := g.genStringArg(args, 0)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i8* @sol_file_read(i8* %s)\n", t, path)
		return t
	case "File.write":
		path := g.genStringArg(args, 0)
		content := g.genStringArg(args, 1)
		fmt.Fprintf(&g.buf, "  call void @sol_file_write(i8* %s, i8* %s)\n", path, content)
		return ""
	case "File.append":
		path := g.genStringArg(args, 0)
		content := g.genStringArg(args, 1)
		fmt.Fprintf(&g.buf, "  call void @sol_file_append(i8* %s, i8* %s)\n", path, content)
		return ""
	case "File.exists":
		path := g.genStringArg(args, 0)
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = call i1 @sol_file_exists(i8* %s)\n", t, path)
		return t
	default:
		ptr := g.stringConst(&ast.StringLit{Value: "unknown builtin in native build"})
		fmt.Fprintf(&g.buf, "  call void @sol_panic(i8* %s)\n", ptr)
		return g.freshTemp()
	}
}

func (g *IRGen) genStringArg(args []ast.Expr, i int) string {
	if i >= len(args) {
		return g.stringConst(&ast.StringLit{Value: ""})
	}
	return g.genStringValue(args[i])
}

func (g *IRGen) genI64Arg(args []ast.Expr, i int) string {
	if i >= len(args) {
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = add i64 0, 0\n", t)
		return t
	}
	ex := args[i]
	if lit, ok := ex.(*ast.IntLit); ok {
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = add i64 0, %d\n", t, lit.Value)
		return t
	}
	v := g.genExpr(ex)
	return v
}

func (g *IRGen) genFloatArg(args []ast.Expr, i int) string {
	if i >= len(args) {
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = fadd double 0.0, 0.0\n", t)
		return t
	}
	ex := args[i]
	if lit, ok := ex.(*ast.FloatLit); ok {
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = fadd double 0.0, %g\n", t, lit.Value)
		return t
	}
	if lit, ok := ex.(*ast.IntLit); ok {
		t := g.freshTemp()
		fmt.Fprintf(&g.buf, "  %s = sitofp i64 %d to double\n", t, lit.Value)
		return t
	}
	v := g.genExpr(ex)
	t := g.freshTemp()
	fmt.Fprintf(&g.buf, "  %s = sitofp i64 %s to double\n", t, v)
	return t
}
