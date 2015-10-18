package main

import (
	"testing"
)

func TestParseMatches(t *testing.T) {
	aMatch := `

Match #76:

../3rdparty/libprocess/3rdparty/protobuf-2.5.0/src/google/protobuf/stubs/common.h:1218:1: note: "root" binds here
using namespace std;  // Don't do this at home, kids.
^~~~~~~~~~~~~~~~~~~

Match #77:

../3rdparty/libprocess/3rdparty/protobuf-2.5.0/src/google/protobuf/stubs/common.h:1218:1: note: "root" binds here
using namespace std;  // Don't do this at home, kids.
^~~~~~~~~~~~~~~~~~~

`

	loc := "../3rdparty/libprocess/3rdparty/protobuf-2.5.0/src/google/protobuf/stubs/common.h:1218:1: note: \"root\" binds here"
	line := "using namespace std;  // Don't do this at home, kids."
	annotation := "^~~~~~~~~~~~~~~~~~~"

	result := ParseMatches(aMatch)

	if len(result) != 2 {
		t.Errorf("Expected 2 result, got %d", len(result))
	}

	for _, r := range result {
		if r.loc != loc {
			t.Errorf("Didn't get expected result, but instead %s", r.loc)
		}
		if r.line != line {
			t.Errorf("Didn't get expected result, but instead %s", r.line)
		}
		if r.annotation != annotation {
			t.Errorf("Didn't get expected result, but instead %s", r.annotation)
		}
	}

}
