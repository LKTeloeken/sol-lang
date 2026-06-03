package ast

import "strings"

// Dump returns a simple string representation of the AST for debugging.
func Dump(node Node) string {
	var b strings.Builder
	dumpNode(&b, node, 0)
	return b.String()
}

func dumpNode(b *strings.Builder, node Node, indent int) {
	pad := strings.Repeat("  ", indent)
	switch n := node.(type) {
	case *Program:
		b.WriteString("Program\n")
		for _, d := range n.Decls {
			dumpNode(b, d, indent+1)
		}
	case *ClassDecl:
		b.WriteString(pad + "ClassDecl " + n.Name)
		if n.SuperName != "" {
			b.WriteString(" eclipse " + n.SuperName)
		}
		b.WriteString("\n")
		for _, m := range n.Members {
			dumpNode(b, m, indent+1)
		}
	case *FieldDecl:
		b.WriteString(pad + "Field " + n.Visibility + " " + n.Type.String() + " " + n.Name + "\n")
	case *GlowDecl:
		b.WriteString(pad + "Glow\n")
		dumpNode(b, n.Body, indent+1)
	case *RayDecl:
		b.WriteString(pad + "Ray " + n.Visibility + " " + n.Name + "\n")
		dumpNode(b, n.Body, indent+1)
	case *BlockStmt:
		b.WriteString(pad + "Block\n")
		for _, s := range n.Stmts {
			dumpNode(b, s, indent+1)
		}
	case *VarDeclStmt:
		b.WriteString(pad + "Var " + n.Name + " " + n.Type.String() + "\n")
	case *AssignStmt:
		b.WriteString(pad + "Assign\n")
	case *IfStmt:
		b.WriteString(pad + "If\n")
	case *WhileStmt:
		b.WriteString(pad + "While\n")
	case *ForEachStmt:
		b.WriteString(pad + "ForEach " + n.VarName + "\n")
	case *ForRangeStmt:
		b.WriteString(pad + "ForRange " + n.VarName + "\n")
	case *BreakStmt:
		b.WriteString(pad + "Break\n")
	case *ContinueStmt:
		b.WriteString(pad + "Continue\n")
	case *EmitStmt:
		b.WriteString(pad + "Emit\n")
	case *FlareStmt:
		b.WriteString(pad + "Flare\n")
	case *TryCatchStmt:
		b.WriteString(pad + "TryCatch\n")
	case *ExprStmt:
		b.WriteString(pad + "ExprStmt\n")
	default:
		b.WriteString(pad + "Node\n")
	}
}
