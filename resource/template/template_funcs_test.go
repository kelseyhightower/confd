package template

import (
	"testing"
)

func TestParseBool(t *testing.T) {
	parseBool := newFuncMap()["parseBool"].(func(string) bool)

	if !parseBool("true") {
		t.Errorf("true is not true")
	}
	if parseBool("false") {
		t.Errorf("false is not false")
	}
	if parseBool("SOME UNKNOWN VALUE") {
		t.Errorf("errors are not false")
	}
}
