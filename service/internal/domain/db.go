package domain

// DBQueryCondition represents an operator/operand pair
type DBQueryCondition struct {
	Operator string
	Operand  interface{}
}
