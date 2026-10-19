package semantic

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"funny-lexer/internal/ast"
)

type annotatedDocument struct {
	AST    any              `json:"ast"`
	Errors []annotatedError `json:"errors"`
}

type annotatedError struct {
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Message string `json:"message"`
}

func MarshalAnnotatedJSON(program *ast.Program, result *Result) ([]byte, error) {
	if program == nil {
		return nil, fmt.Errorf("cannot serialize a nil AST")
	}
	if result == nil {
		return nil, fmt.Errorf("cannot serialize a nil semantic result")
	}

	doc := annotatedDocument{
		AST:    convertAnnotatedValue(reflect.ValueOf(program), result),
		Errors: make([]annotatedError, 0, len(result.Errors)),
	}
	for _, err := range result.Errors {
		doc.Errors = append(doc.Errors, annotatedError{
			Line: err.Line, Column: err.Column, Message: err.Message,
		})
	}

	return json.MarshalIndent(doc, "", "  ")
}

func convertAnnotatedValue(value reflect.Value, result *Result) any {
	if !value.IsValid() {
		return nil
	}

	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil
		}
		return convertAnnotatedValue(value.Elem(), result)
	}

	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil
		}

		converted := convertAnnotatedValue(value.Elem(), result)
		if expr, ok := value.Interface().(ast.Expr); ok {
			if node, ok := converted.(map[string]any); ok {
				if inferredType, found := lookupExpressionType(expr, result); found {
					node["inferredType"] = fmt.Sprint(inferredType)
				} else {
					node["inferredType"] = "<unknown>"
				}
			}
		}
		return converted
	}

	switch value.Kind() {
	case reflect.Struct:
		out := make(map[string]any)
		out["nodeType"] = value.Type().Name()
		for i := 0; i < value.NumField(); i++ {
			fieldInfo := value.Type().Field(i)
			if !fieldInfo.IsExported() {
				continue
			}
			out[lowerFirst(fieldInfo.Name)] = convertAnnotatedValue(value.Field(i), result)
		}
		// This also supports AST implementations that use value expressions.
		if value.CanInterface() {
			if expr, ok := value.Interface().(ast.Expr); ok {
				if inferredType, found := lookupExpressionType(expr, result); found {
					out["inferredType"] = fmt.Sprint(inferredType)
				} else {
					out["inferredType"] = "<unknown>"
				}
			}
		}
		return out

	case reflect.Slice, reflect.Array:
		out := make([]any, value.Len())
		for i := 0; i < value.Len(); i++ {
			out[i] = convertAnnotatedValue(value.Index(i), result)
		}
		return out

	case reflect.Map:
		out := make(map[string]any)
		iter := value.MapRange()
		for iter.Next() {
			out[fmt.Sprint(iter.Key().Interface())] = convertAnnotatedValue(iter.Value(), result)
		}
		return out

	case reflect.String:
		return value.String()
	case reflect.Bool:
		return value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint()
	case reflect.Float32, reflect.Float64:
		return value.Float()
	default:
		return nil
	}
}

func lookupExpressionType(expr ast.Expr, result *Result) (Type, bool) {
	if expr == nil || result == nil || result.Types == nil {
		return Type("<unknown>"), false
	}

	t := reflect.TypeOf(expr)
	if t == nil || !t.Comparable() {
		return Type("<unknown>"), false
	}
	typ, found := result.Types[expr]
	return typ, found
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
