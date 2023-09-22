package hina

import (
	"fmt"
	"reflect"
)

func EvalTree(tree Object, env Environment) error {
	expression, exists := tree["expression"].(map[string]interface{})
	if !exists || len(expression) == 0 {
		return fmt.Errorf("tree has no expressions")
	}
	_, err := evalNode(expression, env)
	if err != nil {
		return err
	}
	return nil
}

func evalNode(node Object, env Environment) (Term, error) {
	kind := node["kind"]

	switch kind {
	case "Str", "Int", "Bool":
		literal, err := inspectLiteral(node)
		if err != nil {
			return nil, err
		}
		return literal, nil
	case "Print":
		print, inspectErr := inspectPrint(node)
		if inspectErr != nil {
			return nil, inspectErr
		}
		returnNode, evalErr := print.Eval(env)
		if evalErr != nil {
			return nil, evalErr
		}
		return returnNode, nil
	case "Binary":
		binary, inspectErr := inspectBinary(node)
		if inspectErr != nil {
			return nil, inspectErr
		}
		result, evalErr := binary.Eval(env)
		if evalErr != nil {
			return nil, evalErr
		}
		return result, nil
	case "Let":
		let, inspectErr := inspectLet(node)
		if inspectErr != nil {
			return nil, inspectErr
		}
		evalErr := let.Eval(env)
		if evalErr != nil {
			return nil, evalErr
		}
		return let, nil
	case "Var":
		varNode, inspectErr := inspectVar(node)
		if inspectErr != nil {
			return nil, inspectErr
		}
		value, evalErr := varNode.Eval(env)
		if evalErr != nil {
			return nil, evalErr
		}
		return value, nil
	case "Tuple":
		tupleNode, err := inspectTuple(node)
		if err != nil {
			return nil, err
		}
		tupleNode, err = tupleNode.Eval(env)
		if err != nil {
			return nil, err
		}
		return tupleNode, nil
	case "First", "Second":
		tupleFunc, inspectErr := inspectTupleFunction(node)
		if inspectErr != nil {
			return nil, inspectErr
		}
		value, evalErr := tupleFunc.Eval(env)
		if evalErr != nil {
			return nil, evalErr
		}
		return value, nil
	case "If":
		ifNode, inspectErr := inspectIf(node)
		if inspectErr != nil {
			return nil, inspectErr
		}
		result, evalErr := ifNode.Eval(env)
		if evalErr != nil {
			return nil, evalErr
		}
		return result, nil
	case "Function":
		function, err := inspectFunction(node)
		if err != nil {
			return nil, err
		}
		return function, nil
	case "Call":
		call, inspectNode := inspectCall(node)
		if inspectNode != nil {
			return nil, inspectNode
		}
		result, evalErr := call.Eval(env)
		if evalErr != nil {
			return nil, evalErr
		}
		return result, nil
	}

	return nil, fmt.Errorf("unknown term: %s", kind)
}

func (print PrintNode) Eval(env Environment) (Term, error) {
	value, err := evalNode(print.Value.(map[string]interface{}), env)
	if err != nil {
		return nil, err
	}
	fmt.Println(value)
	return value, nil
}

func (binary BinaryNode) Eval(env Environment) (Term, error) {
	lhs, lhsEvalErr := evalNode(binary.Lhs.(map[string]interface{}), env)
	if lhsEvalErr != nil {
		return nil, lhsEvalErr
	}
	rhs, rhsEvalErr := evalNode(binary.Rhs.(map[string]interface{}), env)
	if rhsEvalErr != nil {
		return nil, rhsEvalErr
	}

	switch binary.Op {
	case "Add":
		// TODO: improve this
		_, isLhsString := lhs.(StrNode)
		_, isRhsString := rhs.(StrNode)
		intLhs, isLhsInt := lhs.(IntNode)
		intRhs, isRhsInt := rhs.(IntNode)
		if isLhsString || isRhsString {
			return StrNode{Value: fmt.Sprintf("%s%s", lhs, rhs)}, nil
		} else if isLhsInt && isRhsInt {
			return IntNode{Value: intLhs.Value + intRhs.Value}, nil
		}
		return nil, fmt.Errorf("'Add' operator can only be used with Ints and/or Strs")
	case "Sub", "Mul", "Div", "Rem":
		intLhs, isLhsInt := lhs.(IntNode)
		intRhs, isRhsInt := rhs.(IntNode)
		if !isLhsInt && !isRhsInt {
			return nil, fmt.Errorf("'%s' operator can only be used with Ints", binary.Op)
		}
		switch binary.Op {
		case "Sub":
			return IntNode{Value: intLhs.Value - intRhs.Value}, nil
		case "Mul":
			return IntNode{Value: intLhs.Value * intRhs.Value}, nil
		case "Div":
			return IntNode{Value: intLhs.Value / intRhs.Value}, nil
		case "Rem":
			return IntNode{Value: intLhs.Value % intRhs.Value}, nil
		}
	case "Eq", "Neq":
		hasSameValue := lhs == rhs
		hasSameType := reflect.TypeOf(lhs) == reflect.TypeOf(rhs)
		switch binary.Op {
		case "Eq":
			return BoolNode{Value: hasSameValue && hasSameType}, nil
		case "Neq":
			return BoolNode{Value: !hasSameValue || !hasSameType}, nil
		}
	case "Lt", "Gt", "Lte", "Gte":
		intLhs, isLhsInt := lhs.(IntNode)
		intRhs, isRhsInt := rhs.(IntNode)
		if !isLhsInt && !isRhsInt {
			return nil, fmt.Errorf("'%s' comparison can only be done with Ints", binary.Op)
		}
		switch binary.Op {
		case "Lt":
			return BoolNode{Value: intLhs.Value < intRhs.Value}, nil
		case "Gt":
			return BoolNode{Value: intLhs.Value > intRhs.Value}, nil
		case "Lte":
			return BoolNode{Value: intLhs.Value <= intRhs.Value}, nil
		case "Gte":
			return BoolNode{Value: intLhs.Value >= intRhs.Value}, nil
		}
	case "And", "Or":
		boolLhs, isLhsBool := lhs.(BoolNode)
		boolRhs, isRhsBool := rhs.(BoolNode)
		if !isLhsBool && !isRhsBool {
			return nil, fmt.Errorf("'%s' operator can only be used with Bool", binary.Op)
		}
		switch binary.Op {
		case "And":
			return BoolNode{Value: boolLhs.Value && boolRhs.Value}, nil
		case "Or":
			return BoolNode{Value: boolLhs.Value || boolRhs.Value}, nil
		}
	}

	return nil, fmt.Errorf("unknown binary operator: '%s'", binary.Op)
}

func (variable LetNode) Eval(env Environment) error {
	env.Set(variable.Identifier, variable.Value)
	_, err := evalNode(variable.Next.(map[string]interface{}), env)
	if err != nil {
		return err
	}
	return nil
}

func (varCall VarNode) Eval(env Environment) (Term, error) {
	variable, exists := env.Get(varCall.Text)
	if !exists {
		return nil, fmt.Errorf("calling an undeclared variable: %s", varCall.Text)
	}
	value, err := evalNode(variable.(map[string]interface{}), env)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (tuple TupleNode) Eval(env Environment) (TupleNode, error) {
	first, firstEvalErr := evalNode(tuple.First.(map[string]interface{}), env)
	if firstEvalErr != nil {
		return TupleNode{}, nil
	}
	second, secondEvalErr := evalNode(tuple.Second.(map[string]interface{}), env)
	if secondEvalErr != nil {
		return TupleNode{}, nil
	}
	return TupleNode{First: first, Second: second}, nil
}

func (tupleFunc TupleFunction) Eval(env Environment) (Term, error) {
	value, err := evalNode(tupleFunc.Value.(map[string]interface{}), env)
	if err != nil {
		return nil, err
	}
	tuple, isTuple := value.(TupleNode)
	if !isTuple {
		return nil, fmt.Errorf("'%s' only accepts Tuples", tupleFunc.Kind)
	}
	if tupleFunc.Kind == "Second" {
		return tuple.Second, nil
	}
	return tuple.First, nil
}

func (ifTerm IfNode) Eval(env Environment) (Term, error) {
	conditionNode, conditionEvalErr := evalNode(ifTerm.Condition.(map[string]interface{}), env)
	if conditionEvalErr != nil {
		return nil, conditionEvalErr
	}
	condition, isBool := conditionNode.(BoolNode)
	if !isBool {
		return nil, fmt.Errorf("'If' only accepts Bools as condition")
	}

	var body Term
	if condition.Value {
		body = ifTerm.Then
	} else {
		body = ifTerm.Else
	}

	result, bodyEvalErr := evalNode(body.(map[string]interface{}), env)
	if bodyEvalErr != nil {
		return nil, bodyEvalErr
	}
	return result, nil
}

func (function FunctionNode) captureEnv(env Environment) {
	for key, value := range env.SymbolTable {
		if _, exists := function.Env.Get(key); exists {
			continue
		}
		function.Env.Set(key, value)
	}
}

func (function FunctionNode) setParameters(arguments []interface{}) error {
	if len(function.Parameters) != len(arguments) {
		return fmt.Errorf("expected %d arguments, received %d", len(function.Parameters), len(arguments))
	}

	for argIndex := 0; argIndex < len(arguments); argIndex++ {
		parameter, hasParameter := function.Parameters[argIndex].(map[string]interface{})
		parameterName, parameterHasName := parameter["text"].(string)
		if !hasParameter || !parameterHasName {
			return fmt.Errorf("malformed parameter in index %d", argIndex)
		}
		argument, hasArgument := arguments[argIndex].(map[string]interface{})
		if !hasArgument {
			return fmt.Errorf("malformed argument in index %d", argIndex)
		}
		if _, exists := function.Env.Get(parameterName); exists {
			return fmt.Errorf("mixed parameter: %s", parameterName)
		}
		function.Env.Set(parameterName, argument)
	}
	return nil
}

func (call CallNode) Eval(env Environment) (Term, error) {
	calleeNode := call.Callee.(map[string]interface{})
	callee, calleeEvalErr := evalNode(calleeNode, env)
	if calleeEvalErr != nil {
		return nil, calleeEvalErr
	}
	function, isFunction := callee.(FunctionNode)
	if !isFunction {
		return nil, fmt.Errorf("'Call' can only call Functions")
	}

	parametersErr := function.setParameters(call.Arguments)
	if parametersErr != nil {
		return nil, parametersErr
	}
	function.captureEnv(env)

	result, resultEvalErr := evalNode(function.Value.(map[string]interface{}), function.Env)
	if resultEvalErr != nil {
		return nil, resultEvalErr
	}
	return result, nil
}
