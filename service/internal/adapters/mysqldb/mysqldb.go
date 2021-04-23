package mysqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"math/rand"
	"prediction-league/service/internal/domain"
	"strings"
	"time"
)

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

// generateAlphaNumericString generates an alphanumeric string to the provided length
func generateAlphaNumericString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	source := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	var generated string

	sourceLen := len(source)

	for i := 0; i < length; i++ {
		randInt := r.Int63n(int64(sourceLen))
		randByte := []byte(source)[randInt]
		generated += string(randByte)
	}

	return generated
}

// getDBFieldsStringFromFields returns a statement-ready string of fields names
func getDBFieldsStringFromFields(fields []string) string {
	return strings.Join(fields, ", ")
}

// getDBFieldsStringFromFieldsWithTablePrefix returns a statement-ready string of fields names with table prefix
func getDBFieldsStringFromFieldsWithTablePrefix(fields []string, tablePrefix string) string {
	return fmt.Sprintf("%s.%s", tablePrefix, strings.Join(fields, fmt.Sprintf(", %s.", tablePrefix)))
}

// getDBFieldsWithEqualsPlaceholdersStringFromFields returns a statement-ready string of fields names with "equals value" placeholders
func getDBFieldsWithEqualsPlaceholdersStringFromFields(fields []string) string {
	var fieldsWithEqualsPlaceholders []string

	for _, field := range fields {
		fieldsWithEqualsPlaceholders = append(fieldsWithEqualsPlaceholders, fmt.Sprintf("%s = ?", field))
	}

	return strings.Join(fieldsWithEqualsPlaceholders, ", ")
}

// wrapDBError wraps an error from an SQL agent according to its nature as per the representations above
func wrapDBError(err error) error {
	if e, ok := err.(*mysql.MySQLError); ok {
		switch e.Number {
		case 1060:
		case 1061:
		case 1062:
			return domain.DuplicateDBRecordError{Err: err}
		}
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.MissingDBRecordError{Err: err}
	}

	return err
}
