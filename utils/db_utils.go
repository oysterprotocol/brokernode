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
	m := pop.Model{Value: vPtr}
	if m.PrimaryKeyType() != "UUID" {
		return nil, errors.New("Primary key is not UUID, did not support to generate this type of key")
	}

	c := columns.ColumnsForStructWithAlias(vPtr, m.TableName(), m.As)

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
		col := tag.Value
		if cols := c.Cols[col]; cols != nil {
			f[cols.Name] = field.Name
		}
	}
	return &dbUpdateModel{columns: c, fieldMap: f}, nil
}

func (s *dbUpdateModel) GetColumns() string {
	return strings.Join(s.getSortedColumns(), COLUMNS_SEPARATOR)
}

func (s *dbUpdateModel) GetNewUpdateValue(v ValueT) string {
	cols := s.getSortedColumns()
	var xs []string
	r := reflect.Indirect(reflect.ValueOf(v))

	for _, t := range cols {
		if t == "id" {
			u, _ := uuid.NewV4()
			xs = append(xs, fmt.Sprintf("'%s'", u.String()))
			continue
		}
		if t == "updated_at" || t == "created_at" {
			xs = append(xs, "NOW()")
			continue
		}

		if f := s.fieldMap[t]; len(f) > 0 {
			xs = append(xs, getStringPresentation(r.FieldByName(f)))
		}
	}
	return strings.Join(xs, COLUMNS_SEPARATOR)
}

func (s *dbUpdateModel) getSortedColumns() []string {
	var xs []string
	for _, t := range s.columns.Cols {
		xs = append(xs, t.Name)
	}
	sort.Strings(xs)
	return xs
}

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
