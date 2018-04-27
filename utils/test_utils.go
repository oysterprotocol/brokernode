package oyster_utils

import (
	"strings"
	"testing"
)

func assertStringEqual(v string, expect string, t *testing.T) {
	if v != expect {
		t.Errorf("expect '%s' does not equal to %s", expect, v)
	}
}

func assertError(e error, t *testing.T) {
	if e == nil {
		t.Error("Expect error but not error")
	}
}

func assertTrue(v bool, t *testing.T, desc string) {
	if !v {
		t.Error(desc)
	}
}

func assertContainString(v string, substr string, t *testing.T) {
	if !strings.Contains(v, substr) {
		t.Errorf("%s does not contain substring %s", v, substr)
	}
}
