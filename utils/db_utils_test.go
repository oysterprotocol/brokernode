package oyster_utils

import (
	"fmt"
	"github.com/gobuffalo/uuid"
	"testing"
	"time"
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
	UpdateAt time.Time `db:"update_at"`
}

func Test_GetColumns_testingA(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingA{})

	assertStringEqual(v.GetColumns(), "created_at, id", t)
}

func Test_GetColumns_testingB(t *testing.T) {
	_, e := CreateDbUpdateOperation(&testingB{})

	assertError(e, t)
}

func Test_GetColumns_testingC(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingC{})

	assertStringEqual(v.GetColumns(), "a, b, c, d, e, id", t)
}

func Test_GetNewInsertedValue_testingA(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingA{})

	assertContainString(v.GetNewInsertedValue(testingA{}), "NOW(), '", t)
}

func Test_GetNewInsertedValue_testingC(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingC{})

	st := testingC{a: "a", b: "b", c: "c", e: 123}
	assertContainString(v.GetNewInsertedValue(st), "'a', 'b', 'c', '', 123", t)
}

func Test_GetUpdatedValue_tesingA(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingA{})

	id, _ := uuid.NewV4()
	updated := testingA{ID: id, CreatedAt: time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)}

	assertStringEqual(v.GetUpdatedValue(updated), fmt.Sprintf("'2009-11-17 20:34:58', '%s'", id), t)
}

func Test_GetUpdatedValue_tesingD(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingD{})

	id, _ := uuid.NewV4()
	updated := testingD{ID: id}

	assertStringEqual(v.GetUpdatedValue(updated), fmt.Sprintf("'%s', NOW()", id), t)
}
