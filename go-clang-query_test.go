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

	info := `../3rdparty/libprocess/3rdparty/protobuf-2.5.0/src/google/protobuf/stubs/common.h:1218:1: note: "root" binds here
using namespace std;  // Don't do this at home, kids.
^~~~~~~~~~~~~~~~~~~`

	result := ParseMatches(aMatch)

	if len(result) != 2 {
		t.Errorf("Expected 2 result, got %d", len(result))
	}

	for i, r := range result {
		if r.info != info {
			t.Errorf("Didn't get expected result for record %d,\n%s\n, but instead\n %s\n", i, info, r.info)
		}
	}

}

func TestParseMatches2(t *testing.T) {
	aMatch := `
Match #1:

some_source_line.cpp:123:456: note: "root" binds here
bla;
^~~

Match #2:

some_source_line.cpp:123:456: note: "root" binds here
bla;
^~~

`

	result := ParseMatches(aMatch)

	if len(result) != 2 {
		t.Errorf("Expected 2 result, got %d", len(result))
	}
}
