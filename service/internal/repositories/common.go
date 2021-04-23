package repositories

import (
	"fmt"
	"prediction-league/service/internal/domain"
	"strings"
)

// getDBFieldsStringFromFields returns a statement-ready string of fields names
func getDBFieldsStringFromFields(fields []string) string {
	return strings.Join(fields, ", ")
}

// getDBFieldsWithEqualsPlaceholdersStringFromFields returns a statement-ready string of fields names with "equals value" placeholders
func getDBFieldsWithEqualsPlaceholdersStringFromFields(fields []string) string {
	var fieldsWithEqualsPlaceholders []string

	for _, field := range fields {
		fieldsWithEqualsPlaceholders = append(fieldsWithEqualsPlaceholders, fmt.Sprintf("%s = ?", field))
	}

	return strings.Join(fieldsWithEqualsPlaceholders, ", ")
}

// dbWhereStmt returns the WHERE clause portion of an SQL statement as a string, plus the parameters to
// pass to the operation, from a given map of criteria to query on
func dbWhereStmt(criteria map[string]interface{}, matchAny bool) (stmt string, params []interface{}) {
	var conditions []string

	for field, value := range criteria {
		var condition domain.DBQueryCondition
		switch value.(type) {
		case domain.DBQueryCondition:
			condition = value.(domain.DBQueryCondition)
		default:
			condition.Operator = "="
			condition.Operand = value
		}

		if condition.Operand == nil {
			conditions = append(conditions, fmt.Sprintf("%s %s", field, condition.Operator))
			continue
		}

		conditions = append(conditions, fmt.Sprintf("%s %s ?", field, condition.Operator))
		params = append(params, condition.Operand)
	}

	comparison := " AND "
	if matchAny {
		comparison = " OR "
	}

	if len(conditions) > 0 {
		stmt = `WHERE ` + strings.Join(conditions, comparison)
	}

	return stmt, params
}

