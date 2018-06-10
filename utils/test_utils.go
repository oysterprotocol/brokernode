package oyster_utils

import (
	"strings"
	"testing"
)

/*AssertStringEqual fails if v != expect.*/
func AssertStringEqual(v string, expect string, t *testing.T) {
	if v != expect {
		t.Errorf("expect '%s' does not equal to %s", expect, v)
	}
}

/*AssertNoError expects not error and fails if e != nil.*/
func AssertNoError(e error, t *testing.T, desc string) {
	if e != nil {
		t.Errorf("Error unexpected: %v, error: %v", desc, e)
	}
}

/*AssertError expects e as NOT nil, fails on e == nil.*/
func AssertError(e error, t *testing.T, desc string) {
	if e == nil {
		t.Errorf("Expected an error: %v", desc)
	}
}

/*AssertTrue fails if v is false.*/
func AssertTrue(v bool, t *testing.T, desc string) {
	if !v {
		t.Error(desc)
	}
}

/*AssertContainString fails if substr is not part of v.*/
func AssertContainString(v string, substr string, t *testing.T) {
	if !strings.Contains(v, substr) {
		t.Errorf("%s does not contain substring %s", v, substr)
	}
}
