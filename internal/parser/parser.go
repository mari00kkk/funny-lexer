package parser

import (
	"fmt"

	"funny-lexer/internal/ast"
	"funny-lexer/internal/runner"
)

type Parser struct {
	tokens []runner.Token
	pos    int

	errors []ParseError
}

type ParseError struct {
	Position runner.Position
	Message  string
}

func New(tokens []runner.Token) *Parser {
	return &Parser{
		tokens: tokens,
	}
}

func (p *Parser) current() runner.Token {
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
}

func (p *Parser) checkType(tokenType string) bool {
	return p.current().Type == tokenType
}

func (p *Parser) checkValue(value string) bool {
	return p.current().Value == value
}

func (p *Parser) error(message string) {
	token := p.current()

	p.errors = append(p.errors, ParseError{
		Position: token.Position,
		Message:  message,
	})
}

func (p *Parser) expectValue(value string) bool {
	if p.checkValue(value) {
		p.advance()
		return true
	}

	p.error(fmt.Sprintf(
		"expected %q, got %q",
		value,
		p.current().Value,
	))

	return false
}

func (p *Parser) expectType(tokenType string) (runner.Token, bool) {
	token := p.current()

	if token.Type == tokenType {
		p.advance()
		return token, true
	}

	p.error(fmt.Sprintf(
		"expected token type %s, got %s",
		tokenType,
		token.Type,
	))

	return token, false
}
func (p *Parser) parseExpr() ast.Expr {
	return p.parseAdditive()
}

func (p *Parser) parseAdditive() ast.Expr {
	left := p.parseMultiplicative()

	for p.checkValue("+") || p.checkValue("-") {
		op := p.current().Value
		p.advance()

		right := p.parseMultiplicative()

		left = &ast.BinaryExpr{
			Left:  left,
			Op:    op,
			Right: right,
		}
	}

	return left
}

func (p *Parser) parseMultiplicative() ast.Expr {
	left := p.parseUnary()

	for p.checkValue("*") || p.checkValue("/") {
		op := p.current().Value
		p.advance()

		right := p.parseUnary()

		left = &ast.BinaryExpr{
			Left:  left,
			Op:    op,
			Right: right,
		}
	}

	return left
}

func (p *Parser) parseUnary() ast.Expr {
	if p.checkValue("-") {
		p.advance()

		return &ast.UnaryExpr{
			Op: "-",
			X:  p.parseUnary(),
		}
	}

	return p.parsePrimary()
}

func (p *Parser) parsePrimary() ast.Expr {
	if p.checkType("INT") {
		token := p.current()
		p.advance()

		return &ast.NumberExpr{
			Value: token.Value,
		}
	}

	if p.checkType("IDENT") || p.checkValue("length") {
		name := p.current().Value
		p.advance()

		if p.checkValue("(") {
			return p.parseCall(name)
		}

		if p.checkValue("[") {
			p.advance()

			index := p.parseExpr()

			p.expectValue("]")

			return &ast.ArrayAccessExpr{
				Array: name,
				Index: index,
			}
		}

		return &ast.VariableExpr{
			Name: name,
		}
	}

	if p.checkValue("(") {
		p.advance()

		expr := p.parseExpr()

		p.expectValue(")")

		return expr
	}

	p.error("expected expression")
	p.advance()

	return &ast.NumberExpr{
		Value: "0",
	}
}

func (p *Parser) parseCall(name string) ast.Expr {
	p.expectValue("(")

	var arguments []ast.Expr

	// Пустой список аргументов:
	// foo()
	if !p.checkValue(")") {
		arguments = append(arguments, p.parseExpr())

		// Остальные аргументы:
		// foo(a, b, c)
		for p.checkValue(",") {
			p.advance()
			arguments = append(arguments, p.parseExpr())
		}
	}

	p.expectValue(")")

	return &ast.CallExpr{
		Name: name,
		Args: arguments,
	}
}

func (p *Parser) parseComparison() ast.Condition {
	left := p.parseExpr()

	if !p.checkValue("==") &&
		!p.checkValue("!=") &&
		!p.checkValue("<") &&
		!p.checkValue("<=") &&
		!p.checkValue(">") &&
		!p.checkValue(">=") {
		p.error("expected comparison operator")
		return nil
	}

	op := p.current().Value
	p.advance()

	right := p.parseExpr()

	return &ast.Comparison{
		Left:  left,
		Op:    op,
		Right: right,
	}
}

func (p *Parser) parseConditionPrimary() ast.Condition {
	if p.checkValue("true") {
		p.advance()
		return &ast.BoolCondition{Value: true}
	}

	if p.checkValue("false") {
		p.advance()
		return &ast.BoolCondition{Value: false}
	}

	if p.checkValue("(") {
		p.advance()

		condition := p.parseCondition()

		p.expectValue(")")
		return condition
	}

	return p.parseComparison()
}

func (p *Parser) parseNot() ast.Condition {
	if p.checkValue("not") {
		p.advance()

		return &ast.NotCondition{
			X: p.parseNot(),
		}
	}

	return p.parseConditionPrimary()
}

func (p *Parser) parseAnd() ast.Condition {
	left := p.parseNot()

	for p.checkValue("and") {
		p.advance()

		right := p.parseNot()

		left = &ast.BinaryCondition{
			Left:  left,
			Op:    "and",
			Right: right,
		}
	}

	return left
}

func (p *Parser) parseOr() ast.Condition {
	left := p.parseAnd()

	for p.checkValue("or") {
		p.advance()

		right := p.parseAnd()

		left = &ast.BinaryCondition{
			Left:  left,
			Op:    "or",
			Right: right,
		}
	}

	return left
}

func (p *Parser) parseImplication() ast.Condition {
	left := p.parseOr()

	if p.checkValue("->") {
		p.advance()

		right := p.parseImplication()

		return &ast.BinaryCondition{
			Left:  left,
			Op:    "->",
			Right: right,
		}
	}

	return left
}

func (p *Parser) parseCondition() ast.Condition {
	return p.parseImplication()
}

func (p *Parser) parseBlock() ast.Statement {
	p.expectValue("{")

	var statements []ast.Statement

	for !p.checkValue("}") && !p.checkType("EOF") {
		statement := p.parseStatement()

		if statement != nil {
			statements = append(statements, statement)
		}
	}

	p.expectValue("}")

	return &ast.Block{
		Statements: statements,
	}
}

func (p *Parser) parseStatement() ast.Statement {
	if p.checkValue("{") {
		return p.parseBlock()
	}

	if p.checkValue("if") {
		return p.parseIf()
	}

	if p.checkValue("while") {
		return p.parseWhile()
	}

	if p.checkType("IDENT") {
		return p.parseAssignment()
	}

	if p.checkValue("assert") {
		return p.parseAssert()
	}

	if p.checkValue("assume") {
		return p.parseAssume()
	}

	p.error("expected statement")
	p.advance()

	return nil
}

func (p *Parser) parseAssignment() ast.Statement {
	token, ok := p.expectType("IDENT")
	if !ok {
		return nil
	}

	var variables []ast.Expr

	// Первый идентификатор
	if p.checkValue("[") {
		p.advance()

		index := p.parseExpr()

		p.expectValue("]")

		variables = append(variables, &ast.ArrayAccessExpr{
			Array: token.Value,
			Index: index,
		})
	} else {
		variables = append(variables, &ast.VariableExpr{
			Name: token.Value,
		})
	}

	for p.checkValue(",") {
		if _, ok := variables[0].(*ast.ArrayAccessExpr); ok {
			p.error("multiple assignment cannot use array access")
			break
		}

		p.advance()

		token, ok := p.expectType("IDENT")
		if !ok {
			return nil
		}

		variables = append(variables, &ast.VariableExpr{
			Name: token.Value,
		})
	}

	p.expectValue("=")

	value := p.parseExpr()

	// Если слева несколько переменных,
	// справа по грамматике должен быть functionCall
	if len(variables) > 1 {
		if _, ok := value.(*ast.CallExpr); !ok {
			p.error("multiple assignment requires function call")
		}
	}

	if !p.expectValue(";") {
		p.synchronizeStatement()
	}

	return &ast.Assignment{
		Variables: variables,
		Value:     []ast.Expr{value},
	}
}

func (p *Parser) synchronizeStatement() {
	for !p.checkType("EOF") &&
		!p.checkValue(";") &&
		!p.checkValue("}") {
		p.advance()
	}

	if p.checkValue(";") {
		p.advance()
	}
}

func (p *Parser) parseAssert() ast.Statement {
	p.expectValue("assert")
	predicate := p.parsePredicate()
	p.expectValue(";")
	return &ast.Assert{Predicate: predicate}
}

func (p *Parser) parseAssume() ast.Statement {
	p.expectValue("assume")
	predicate := p.parsePredicate()
	p.expectValue(";")
	return &ast.Assume{Predicate: predicate}
}

func (p *Parser) parsePredicate() ast.Predicate {
	return p.parsePredicateOr()
}

func (p *Parser) parsePredicateOr() ast.Predicate {
	left := p.parsePredicateAnd()

	for p.checkValue("or") {
		p.advance()

		right := p.parsePredicateAnd()

		left = &ast.BinaryPredicate{
			Left:  left,
			Op:    "or",
			Right: right,
		}
	}

	return left
}

func (p *Parser) parsePredicateAnd() ast.Predicate {
	left := p.parsePredicateNot()

	for p.checkValue("and") {
		p.advance()

		right := p.parsePredicateNot()

		left = &ast.BinaryPredicate{
			Left:  left,
			Op:    "and",
			Right: right,
		}
	}

	return left
}

func (p *Parser) parsePredicateNot() ast.Predicate {
	if p.checkValue("not") {
		p.advance()

		return &ast.NotPredicate{
			X: p.parsePredicateNot(),
		}
	}

	return p.parsePredicatePrimary()
}

func (p *Parser) parsePredicatePrimary() ast.Predicate {
	// true
	if p.checkValue("true") {
		p.advance()

		return &ast.BoolPredicate{
			Value: true,
		}
	}

	// false
	if p.checkValue("false") {
		p.advance()

		return &ast.BoolPredicate{
			Value: false,
		}
	}

	// forall / exists
	if p.checkValue("forall") || p.checkValue("exists") {
		return p.parseQuantifier()
	}

	// predicate 
	if p.checkValue("(") {
		p.advance()

		predicate := p.parsePredicate()

		p.expectValue(")")

		return predicate
	}

	expr := p.parseExpr()

	if p.isComparisonOperator() {
		op := p.current().Value
		p.advance()

		right := p.parseExpr()

		return &ast.PredicateComparison{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}

	// Если выражение оказалось вызовом функции,
	// то в predicate это formulaRef
	if call, ok := expr.(*ast.CallExpr); ok {
		return &ast.FormulaRef{
			Name: call.Name,
			Args: call.Args,
		}
	}

	p.error("expected predicate")

	return nil
}

func (p *Parser) parseQuantifier() ast.Predicate {
	kind := p.current().Value
	p.advance()

	p.expectValue("(")

	variable := p.parseVariableDef()

	p.expectValue("|")

	predicate := p.parsePredicate()

	p.expectValue(")")

	return &ast.Quantifier{
		Kind:      kind,
		Variable:  variable,
		Predicate: predicate,
	}
}

func (p *Parser) parseVariableDef() ast.Variable {
	token, ok := p.expectType("IDENT")
	if !ok {
		return ast.Variable{}
	}

	p.expectValue(":")

	varType := p.parseVariableType()

	return ast.Variable{
		Name: token.Value,
		Type: varType,
	}
}

func (p *Parser) parseVariableType() ast.Type {
	if p.checkValue("int[]") {
		p.advance()
		return ast.IntArray
	}

	if p.checkValue("int") {
		p.advance()

		// Вариант, если lexer выдаёт:
		// int [ ]
		if p.checkValue("[") {
			p.advance()
			p.expectValue("]")

			return ast.IntArray
		}

		return ast.IntType
	}

	p.error("expected variable type")

	return ast.IntType
}

func (p *Parser) peek(offset int) runner.Token {
	index := p.pos + offset

	if index >= len(p.tokens) {
		return p.tokens[len(p.tokens)-1]
	}

	return p.tokens[index]
}

func (p *Parser) isComparisonOperator() bool {
	switch p.current().Value {
	case "==", "!=", "<", "<=", ">", ">=":
		return true
	default:
		return false
	}
}

func (p *Parser) parseWhile() ast.Statement {
	p.expectValue("while")
	p.expectValue("(")

	condition := p.parseCondition()

	p.expectValue(")")

	var invariant *ast.Predicate

	if p.checkValue("invariant") {
		p.advance()

		predicate := p.parsePredicate()
		invariant = &predicate
	}

	body := p.parseStatement()

	return &ast.While{
		Condition: condition,
		Invariant: invariant,
		Body:      body,
	}
}

func (p *Parser) parseIf() ast.Statement {
	p.expectValue("if")
	p.expectValue("(")

	condition := p.parseCondition()

	p.expectValue(")")

	thenBranch := p.parseStatement()

	var elseBranch ast.Statement
	hasElse := false

	if p.checkValue("else") {
		p.advance()
		elseBranch = p.parseStatement()
		hasElse = true
	}

	return &ast.If{
		Condition: condition,
		Then:      thenBranch,
		Else:      elseBranch,
		HasElse:   hasElse,
	}
}

func (p *Parser) parseLocalVariableDef() ast.Variable {
	token, ok := p.expectType("IDENT")
	if !ok {
		return ast.Variable{}
	}

	var variableType ast.Type

	if p.checkValue(":") {
		p.advance()
		variableType = p.parseVariableType()
	}

	return ast.Variable{
		Name: token.Value,
		Type: variableType,
	}
}

func (p *Parser) parseLocalVariableList() []ast.Variable {
	var variables []ast.Variable

	variables = append(variables, p.parseLocalVariableDef())

	for p.checkValue(",") {
		p.advance()
		variables = append(variables, p.parseLocalVariableDef())
	}

	return variables
}

func (p *Parser) parseVariableList() []ast.Variable {
	var variables []ast.Variable

	variables = append(variables, p.parseVariableDef())

	for p.checkValue(",") {
		p.advance()
		variables = append(variables, p.parseVariableDef())
	}

	return variables
}

func (p *Parser) parseFunction() ast.Function {
	nameToken, ok := p.expectType("IDENT")
	if !ok {
		return ast.Function{}
	}

	p.expectValue("(")

	var parameters []ast.Variable

	if !p.checkValue(")") {
		parameters = p.parseVariableList()
	}

	p.expectValue(")")

	var requires *ast.Predicate

	if p.checkValue("requires") {
		p.advance()

		predicate := p.parsePredicate()
		requires = &predicate
	}

	p.expectValue("returns")

	returns := p.parseVariableList()

	var ensures *ast.Predicate

	if p.checkValue("ensures") {
		p.advance()

		predicate := p.parsePredicate()
		ensures = &predicate
	}

	var uses []ast.Variable

	if p.checkValue("uses") {
		p.advance()

		uses = p.parseLocalVariableList()
	}

	body := p.parseStatement()

	return ast.Function{
		Name:       nameToken.Value,
		Parameters: parameters,
		Returns:    returns,
		Requires:   requires,
		Ensures:    ensures,
		Uses:       uses,
		Body:       body,
	}
}

func (p *Parser) parseFormula() ast.Formula {
	nameToken, ok := p.expectType("IDENT")
	if !ok {
		return ast.Formula{}
	}

	p.expectValue("(")

	var parameters []ast.Variable
	if !p.checkValue(")") {
		parameters = p.parseVariableList()
	}

	p.expectValue(")")

	p.expectValue("=>")

	body := p.parsePredicate()

	return ast.Formula{
		Name:       nameToken.Value,
		Parameters: parameters,
		Body:       body,
	}
}

func (p *Parser) parseProgram() ast.Program {
	var functions []ast.Function
	var formulas []ast.Formula

	if p.checkType("EOF") {
		p.error("expected function or formula, got end of input")
		return ast.Program{
			Functions: functions,
			Formulas:  formulas,
		}
	}

	for !p.checkType("EOF") {
		if p.checkType("ERROR") {
			p.error(fmt.Sprintf("invalid token %q", p.current().Value))
			p.advance()
			continue
		}

		if !p.checkType("IDENT") {
			p.error(fmt.Sprintf(
				"expected function or formula, got %q",
				p.current().Value,
			))
			p.advance()
			continue
		}

		if p.looksLikeFormula() {
			formulas = append(formulas, p.parseFormula())
		} else {
			functions = append(functions, p.parseFunction())
		}
	}

	return ast.Program{
		Functions: functions,
		Formulas:  formulas,
	}
}

func (p *Parser) looksLikeFormula() bool {
	offset := 0

	if p.peek(offset).Type != "IDENT" {
		return false
	}
	offset++

	if p.peek(offset).Value != "(" {
		return false
	}
	offset++

	depth := 1

	for {
		token := p.peek(offset)

		if token.Type == "EOF" {
			return false
		}

		if token.Value == "(" {
			depth++
		}

		if token.Value == ")" {
			depth--

			if depth == 0 {
				offset++
				break
			}
		}

		offset++
	}

	return p.peek(offset).Value == "=>"
}

func (p *Parser) Parse() ast.Program {
	return p.parseProgram()
}

func (p *Parser) Errors() []ParseError {
	return p.errors
}
