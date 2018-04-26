package oyster_utils

import (
	"github.com/gobuffalo/uuid"
	"testing"
	"time"
)

type testingA struct {
	ID         uuid.UUID `db:"id"`
	created_at time.Time `db:"created_at"`
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

func Test_GetNewUpdateValue_testingA(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingA{})

	assertContainString(v.GetNewUpdateValue(testingA{}), "NOW(), '", t)
}

func Test_GetNewUpdateValue_testingC(t *testing.T) {
	v, _ := CreateDbUpdateOperation(&testingC{})

	st := testingC{a: "a", b: "b", c: "c", e: 123}
	assertContainString(v.GetNewUpdateValue(st), "'a', 'b', 'c', '', 123", t)
}
