package oyster_utils

import (
	"fmt"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/columns"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type ValueT interface{}

// Interface for DB Update operation.
type dbUpdateOperation interface {
	// Get columns in string format and separated by ", "
	GetColumns() string
	// Get columns new value in string format and separated by ", "
	GetNewUpdateValue(ValueT) string
}

const COLUMNS_SEPARATOR = ", "

// Private data structure
type dbUpdateModel struct {
	columns  columns.Columns
	fieldMap map[string]string // Map from tableColumn to fieldName
}

// Expect an empty ptr struct as &MyStruct{}. Return a dbUpdateOperation interface
func CreateDbUpdateOperation(vPtr ValueT) (dbUpdateOperation, error) {
	model := pop.Model{Value: vPtr}
	if model.PrimaryKeyType() != "UUID" {
		return nil, errors.New("Primary key is not UUID, did not support to generate this type of key")
	}

	cols := columns.ColumnsForStructWithAlias(vPtr, model.TableName(), model.As)

	f := make(map[string]string)

	st := reflect.TypeOf(vPtr)
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)

		popTags := columns.TagsFor(field)
		tag := popTags.Find("db")

		if tag.Ignored() || tag.Empty() {
			continue
		}
		columnName := tag.Value
		if col := cols.Cols[columnName]; col != nil {
			f[col.Name] = field.Name
		}
	}
	return &dbUpdateModel{columns: cols, fieldMap: f}, nil
}

func (s *dbUpdateModel) GetColumns() string {
	return strings.Join(getSortedColumns(s.columns), COLUMNS_SEPARATOR)
}

func (s *dbUpdateModel) GetNewUpdateValue(v ValueT) string {
	cols := getSortedColumns(s.columns)
	var columnValues []string
	stValue := reflect.Indirect(reflect.ValueOf(v))

	for _, t := range cols {
		// Generated UUID for 'id' column
		if t == "id" {
			u, _ := uuid.NewV4()
			columnValues = append(columnValues, fmt.Sprintf("'%s'", u.String()))
			continue
		}
		// Use Sql NOW() method for 'updated_at' and 'created_at' column
		if t == "updated_at" || t == "created_at" {
			columnValues = append(columnValues, "NOW()")
			continue
		}

		if f := s.fieldMap[t]; len(f) > 0 {
			columnValues = append(columnValues, getStringPresentation(stValue.FieldByName(f)))
		}
	}
	return strings.Join(columnValues, COLUMNS_SEPARATOR)
}

// Returns a sorted column name list.
func getSortedColumns(cols columns.Columns) []string {
	var columnNames []string
	for _, t := range cols.Cols {
		columnNames = append(columnNames, t.Name)
	}
	sort.Strings(columnNames)
	return columnNames
}

// Returns string presentation of underlying value for both int and string. String will include single quote (')
func getStringPresentation(v reflect.Value) string {
	switch v.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.String:
		return fmt.Sprintf("'%v'", v.String())
	default:
		panic(errors.Errorf("No implemented type %v", v.String()))
	}
}
