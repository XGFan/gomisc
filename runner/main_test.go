package main

import (
	"reflect"
	"testing"
)

func TestRemoveEmpty(t *testing.T) {
	s := []string{"1", "", "", "3"}
	removeEmpty(&s)
	if !reflect.DeepEqual(s, []string{"1", "3"}) {
		t.Fail()
	}

	s = []string{"", "", "", ""}
	removeEmpty(&s)
	if !reflect.DeepEqual(s, []string{}) {
		t.Fail()
	}

	s = []string{"", "1", "", ""}
	removeEmpty(&s)
	if !reflect.DeepEqual(s, []string{"1"}) {
		t.Fail()
	}
}
