package ast

// вся программа
type Program struct {
	Functions []Function
	Formulas  []Formula
}

// объявление функции
type Function struct {
	Name       string
	Parameters []Variable
	Returns    []Variable

	Requires *Predicate
	Ensures  *Predicate
	Uses     []Variable

	Body Statement
}

type Formula struct {
	Name       string
	Parameters []Variable
	Body       Predicate
}

type Variable struct {
	Name string
	Type Type
}

type Type string

const (
	IntType  Type = "int"
	IntArray Type = "int[]"
)

// Statements

type Statement interface {
	statementNode()
}

// присваивание
type Assignment struct {
	Variables []Expr
	Value     []Expr
}

func (Assignment) statementNode() {}

// Block — блок операторов

type Block struct {
	Statements []Statement
}

func (Block) statementNode() {}

// условный оператор
type If struct {
	Condition Condition
	Then      Statement
	Else      Statement
	HasElse   bool
}

func (If) statementNode() {}

// цикл
type While struct {
	Condition Condition
	Invariant *Predicate
	Body      Statement
}

func (While) statementNode() {}

// проверка предиката
type Assert struct {
	Predicate Predicate
}

func (Assert) statementNode() {}

// предположение
type Assume struct {
	Predicate Predicate
}

func (Assume) statementNode() {}

// Expressions

type Expr interface {
	exprNode()
	//GetType() Type
	//SetType(Type)
}

// целочисленная константа
type NumberExpr struct {
	Value string
}

func (NumberExpr) exprNode() {}

// обращение к переменной
type VariableExpr struct {
	Name string
}

func (VariableExpr) exprNode() {}

// унарная операция

type UnaryExpr struct {
	Op string
	X  Expr
}

func (UnaryExpr) exprNode() {}

//бинарная операция

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (BinaryExpr) exprNode() {}

// вызов функции

// foo(x, 10)
type CallExpr struct {
	Name string
	Args []Expr
}

func (CallExpr) exprNode() {}

// обращение к элементу массива
type ArrayAccessExpr struct {
	Array string
	Index Expr
}

func (ArrayAccessExpr) exprNode() {}

// Condition — интерфейс условий
type Condition interface {
	conditionNode()
}

// true / false
type BoolCondition struct {
	Value bool
}

func (BoolCondition) conditionNode() {}

type Comparison struct {
	Left  Expr
	Op    string
	Right Expr
}

func (Comparison) conditionNode() {}

type NotCondition struct {
	X Condition
}

func (NotCondition) conditionNode() {}

// and / or / ->
type BinaryCondition struct {
	Left  Condition
	Op    string
	Right Condition
}

func (BinaryCondition) conditionNode() {}

// Predicate — интерфейс предикатов
type Predicate interface {
	predicateNode()
}

type BoolPredicate struct {
	Value bool
}

func (BoolPredicate) predicateNode() {}

// сравнение в предикате
type PredicateComparison struct {
	Left  Expr
	Op    string
	Right Expr
}

func (PredicateComparison) predicateNode() {}

type NotPredicate struct {
	X Predicate
}

func (NotPredicate) predicateNode() {}

// and / or
type BinaryPredicate struct {
	Left  Predicate
	Op    string
	Right Predicate
}

func (BinaryPredicate) predicateNode() {}

// Quantifier — forall / exists
// пример:
//
//	forall(x: int | x > 0)
type Quantifier struct {
	Kind      string
	Variable  Variable
	Predicate Predicate
}

func (Quantifier) predicateNode() {}

// FormulaRef — ссылка на формулу

type FormulaRef struct {
	Name string
	Args []Expr
}

func (FormulaRef) predicateNode() {}
