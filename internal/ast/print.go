package ast

import (
	"fmt"
	"strings"
)

func PrintProgram(program Program) string {
	var b strings.Builder

	b.WriteString("Program\n")

	for _, formula := range program.Formulas {
		printFormula(&b, formula, 1)
	}

	for _, function := range program.Functions {
		printFunction(&b, function, 1)
	}

	return b.String()
}

func printFormula(b *strings.Builder, formula Formula, level int) {
	indent(b, level)
	fmt.Fprintf(b, "Formula %s\n", formula.Name)

	if len(formula.Parameters) > 0 {
		indent(b, level+1)
		b.WriteString("Parameters:\n")

		for _, parameter := range formula.Parameters {
			indent(b, level+2)
			fmt.Fprintf(b, "%s: %s\n", parameter.Name, parameter.Type)
		}
	}

	indent(b, level+1)
	b.WriteString("Body:\n")
	printPredicate(b, formula.Body, level+2)
}

func printFunction(b *strings.Builder, function Function, level int) {
	indent(b, level)
	fmt.Fprintf(b, "Function %s\n", function.Name)

	if len(function.Parameters) > 0 {
		indent(b, level+1)
		b.WriteString("Parameters:\n")

		for _, parameter := range function.Parameters {
			indent(b, level+2)
			fmt.Fprintf(b, "%s: %s\n", parameter.Name, parameter.Type)
		}
	}

	if len(function.Returns) > 0 {
		indent(b, level+1)
		b.WriteString("Returns:\n")

		for _, result := range function.Returns {
			indent(b, level+2)
			fmt.Fprintf(b, "%s: %s\n", result.Name, result.Type)
		}
	}

	if function.Requires != nil {
		indent(b, level+1)
		b.WriteString("Requires:\n")
		printPredicate(b, *function.Requires, level+2)
	}

	if function.Ensures != nil {
		indent(b, level+1)
		b.WriteString("Ensures:\n")
		printPredicate(b, *function.Ensures, level+2)
	}

	if len(function.Uses) > 0 {
		indent(b, level+1)
		b.WriteString("Uses:\n")

		for _, variable := range function.Uses {
			indent(b, level+2)
			fmt.Fprintf(b, "%s: %s\n", variable.Name, variable.Type)
		}
	}

	indent(b, level+1)
	b.WriteString("Body:\n")
	printStatement(b, function.Body, level+2)
}

func printStatement(b *strings.Builder, statement Statement, level int) {
	if statement == nil {
		indent(b, level)
		b.WriteString("<nil>\n")
		return
	}

	switch statement := statement.(type) {
	case *Assignment:
		indent(b, level)
		b.WriteString("Assignment\n")

		indent(b, level+1)
		b.WriteString("Variables:\n")
		for _, variable := range statement.Variables {
			printExpr(b, variable, level+2)
		}

		indent(b, level+1)
		b.WriteString("Value:\n")
		for _, value := range statement.Value {
			printExpr(b, value, level+2)
		}

	case *Block:
		indent(b, level)
		b.WriteString("Block\n")

		for _, statement := range statement.Statements {
			printStatement(b, statement, level+1)
		}

	case *If:
		indent(b, level)
		b.WriteString("If\n")

		indent(b, level+1)
		b.WriteString("Condition:\n")
		printCondition(b, statement.Condition, level+2)

		indent(b, level+1)
		b.WriteString("Then:\n")
		printStatement(b, statement.Then, level+2)

		if statement.HasElse {
			indent(b, level+1)
			b.WriteString("Else:\n")
			printStatement(b, statement.Else, level+2)
		}

	case *While:
		indent(b, level)
		b.WriteString("While\n")

		indent(b, level+1)
		b.WriteString("Condition:\n")
		printCondition(b, statement.Condition, level+2)

		if statement.Invariant != nil {
			indent(b, level+1)
			b.WriteString("Invariant:\n")
			printPredicate(b, *statement.Invariant, level+2)
		}

		indent(b, level+1)
		b.WriteString("Body:\n")
		printStatement(b, statement.Body, level+2)

	case *Assert:
		indent(b, level)
		b.WriteString("Assert\n")
		printPredicate(b, statement.Predicate, level+1)

	case *Assume:
		indent(b, level)
		b.WriteString("Assume\n")
		printPredicate(b, statement.Predicate, level+1)

	default:
		indent(b, level)
		fmt.Fprintf(b, "<unknown statement %T>\n", statement)
	}
}

func printExpr(b *strings.Builder, expr Expr, level int) {
	if expr == nil {
		indent(b, level)
		b.WriteString("<nil>\n")
		return
	}

	switch expr := expr.(type) {
	case *NumberExpr:
		indent(b, level)
		fmt.Fprintf(b, "Number %s\n", expr.Value)

	case *VariableExpr:
		indent(b, level)
		fmt.Fprintf(b, "Variable %s\n", expr.Name)

	case *UnaryExpr:
		indent(b, level)
		fmt.Fprintf(b, "Unary %s\n", expr.Op)
		printExpr(b, expr.X, level+1)

	case *BinaryExpr:
		indent(b, level)
		fmt.Fprintf(b, "Binary %s\n", expr.Op)

		indent(b, level+1)
		b.WriteString("Left:\n")
		printExpr(b, expr.Left, level+2)

		indent(b, level+1)
		b.WriteString("Right:\n")
		printExpr(b, expr.Right, level+2)

	case *CallExpr:
		indent(b, level)
		fmt.Fprintf(b, "Call %s\n", expr.Name)

		for _, arg := range expr.Args {
			printExpr(b, arg, level+1)
		}

	case *ArrayAccessExpr:
		indent(b, level)
		fmt.Fprintf(b, "ArrayAccess %s\n", expr.Array)

		indent(b, level+1)
		b.WriteString("Index:\n")
		printExpr(b, expr.Index, level+2)

	default:
		indent(b, level)
		fmt.Fprintf(b, "<unknown expression %T>\n", expr)
	}
}

func printCondition(b *strings.Builder, condition Condition, level int) {
	if condition == nil {
		indent(b, level)
		b.WriteString("<nil>\n")
		return
	}

	switch condition := condition.(type) {
	case *BoolCondition:
		indent(b, level)
		fmt.Fprintf(b, "Bool %t\n", condition.Value)

	case *Comparison:
		indent(b, level)
		fmt.Fprintf(b, "Comparison %s\n", condition.Op)

		printExpr(b, condition.Left, level+1)
		printExpr(b, condition.Right, level+1)

	case *NotCondition:
		indent(b, level)
		b.WriteString("Not\n")
		printCondition(b, condition.X, level+1)

	case *BinaryCondition:
		indent(b, level)
		fmt.Fprintf(b, "BinaryCondition %s\n", condition.Op)

		printCondition(b, condition.Left, level+1)
		printCondition(b, condition.Right, level+1)

	default:
		indent(b, level)
		fmt.Fprintf(b, "<unknown condition %T>\n", condition)
	}
}

func printPredicate(b *strings.Builder, predicate Predicate, level int) {
	if predicate == nil {
		indent(b, level)
		b.WriteString("<nil>\n")
		return
	}

	switch predicate := predicate.(type) {
	case *BoolPredicate:
		indent(b, level)
		fmt.Fprintf(b, "Bool %t\n", predicate.Value)

	case *PredicateComparison:
		indent(b, level)
		fmt.Fprintf(b, "Comparison %s\n", predicate.Op)

		printExpr(b, predicate.Left, level+1)
		printExpr(b, predicate.Right, level+1)

	case *NotPredicate:
		indent(b, level)
		b.WriteString("Not\n")
		printPredicate(b, predicate.X, level+1)

	case *BinaryPredicate:
		indent(b, level)
		fmt.Fprintf(b, "BinaryPredicate %s\n", predicate.Op)

		printPredicate(b, predicate.Left, level+1)
		printPredicate(b, predicate.Right, level+1)

	case *Quantifier:
		indent(b, level)
		fmt.Fprintf(b, "Quantifier %s\n", predicate.Kind)

		indent(b, level+1)
		fmt.Fprintf(b, "Variable: %s: %s\n",
			predicate.Variable.Name,
			predicate.Variable.Type,
		)

		indent(b, level+1)
		b.WriteString("Predicate:\n")
		printPredicate(b, predicate.Predicate, level+2)

	case *FormulaRef:
		indent(b, level)
		fmt.Fprintf(b, "FormulaRef %s\n", predicate.Name)

		for _, arg := range predicate.Args {
			printExpr(b, arg, level+1)
		}

	default:
		indent(b, level)
		fmt.Fprintf(b, "<unknown predicate %T>\n", predicate)
	}
}

func indent(b *strings.Builder, level int) {
	b.WriteString(strings.Repeat("  ", level))
}
