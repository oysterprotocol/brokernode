package oyster_utils

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/pop/columns"
	"github.com/gobuffalo/uuid"
	"github.com/pkg/errors"
)

/*SqlTimeFormat is used for time.Time.Format method */
const SqlTimeFormat = "2006-01-02 15:04:05"

type ValueT interface{}

// Interface for DB Update operation.
type dbUpdateOperation interface {
	// Get columns in string format and separated by ", "
	GetColumns() string
	// Get columns new value for INSERT in string format and separated by ", "
	GetNewInsertedValue(ValueT) string
	// Get columns value for UPDATE operation in string format and separated by ","
	GetUpdatedValue(ValueT) string
}

const (
	COLUMNS_SEPARATOR = ", "

	// The max number of retry if there is an error on SQL.
	MAX_NUMBER_OF_SQL_RETRY = 3

	//SQL_BATCH_SIZE is the maximum number of entries to update in sql at once
	SQL_BATCH_SIZE = 10
)

// Private data structure
type dbUpdateModel struct {
	columns  columns.Columns
	fieldMap map[string]string // Map from tableColumn to fieldName
}

// Expect an empty ptr struct as &MyStruct{}. Return a dbUpdateOperation interface
func CreateDbUpdateOperation(vPtr ValueT) (dbUpdateOperation, error) {
	model := pop.Model{Value: vPtr}
	if model.PrimaryKeyType() != "UUID" {
		err := errors.New("Primary key is not UUID, did not support to generate this type of key")
		raven.CaptureError(err, nil)
		return nil, err
	}

	cols := columns.ForStructWithAlias(vPtr, model.TableName(), model.As)

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

func (s *dbUpdateModel) GetNewInsertedValue(v ValueT) string {
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

func (s *dbUpdateModel) GetUpdatedValue(v ValueT) string {
	cols := getSortedColumns(s.columns)
	var columnValues []string
	stValue := reflect.Indirect(reflect.ValueOf(v))

	for _, t := range cols {
		if t == "update_at" {
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
	}

	switch v.Type().String() {
	case "time.Time":
		t := v.Interface().(time.Time)
		// This is the format SQL like.
		return fmt.Sprintf("'%s'", t.Format(SqlTimeFormat))
	case "uuid.UUID":
		id := v.Interface().(uuid.UUID)
		return fmt.Sprintf("'%s'", id.String())
	}

	err := errors.Errorf("No implemented type %v", v.String())
	raven.CaptureError(err, nil)
	panic(err)
}
