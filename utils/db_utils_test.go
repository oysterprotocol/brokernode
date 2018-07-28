package oyster_utils_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gobuffalo/uuid"
	"github.com/oysterprotocol/brokernode/utils"
)

type testingA struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

type testingB struct {
	noId string `db:no_id`
}

type testingC struct {
	ID uuid.UUID `db:"id"`
	e  int       `db:e`
	d  string    `db:"d"`
	c  string    `db:"c"`
	b  string    `db:"b"`
	a  string    `db:"a"`
}

type testingD struct {
	ID       uuid.UUID `db:"id"`
	UpdateAt time.Time `db:"updated_at"`
}

func Test_GetColumns_testingA(t *testing.T) {
	v, _ := oyster_utils.CreateDbUpdateOperation(&testingA{})

	oyster_utils.AssertStringEqual(v.GetColumns(), "created_at, id", t)
}

func Test_GetColumns_testingB(t *testing.T) {
	_, e := oyster_utils.CreateDbUpdateOperation(&testingB{})

	oyster_utils.AssertError(e, t, "")
}

func Test_GetColumns_testingC(t *testing.T) {
	v, _ := oyster_utils.CreateDbUpdateOperation(&testingC{})

	oyster_utils.AssertStringEqual(v.GetColumns(), "a, b, c, d, e, id", t)
}

func Test_GetNewInsertedValue_testingA(t *testing.T) {
	v, _ := oyster_utils.CreateDbUpdateOperation(&testingA{})

	oyster_utils.AssertContainString(v.GetNewInsertedValue(testingA{}), "NOW(), '", t)
}

func Test_GetNewInsertedValue_testingC(t *testing.T) {
	v, _ := oyster_utils.CreateDbUpdateOperation(&testingC{})

	st := testingC{a: "a", b: "b", c: "c", e: 123}
	oyster_utils.AssertContainString(v.GetNewInsertedValue(st), "'a', 'b', 'c', '', 123", t)
}

func Test_GetUpdatedValue_tesingA(t *testing.T) {
	v, _ := oyster_utils.CreateDbUpdateOperation(&testingA{})

	id, _ := uuid.NewV4()
	updated := testingA{ID: id, CreatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}

	oyster_utils.AssertStringEqual(v.GetUpdatedValue(updated), fmt.Sprintf("'2009-11-17 20:34:58', '%s'", id), t)
}

func Test_GetUpdatedValue_tesingD(t *testing.T) {
	v, _ := oyster_utils.CreateDbUpdateOperation(&testingD{})

	id, _ := uuid.NewV4()
	updated := testingD{ID: id}

	oyster_utils.AssertStringEqual(v.GetUpdatedValue(updated), fmt.Sprintf("'%s', NOW()", id), t)
}
