package main

import (
	"reflect"
	"testing"
)

func TestParseOption(t *testing.T) {
	args, err := parseOption(`args -opt1 --opt2 a\"c "quoted string" a/c a\ c`)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"args",
		"-opt1",
		"--opt2",
		"a\"c",
		"quoted string",
		"a/c",
		"a c",
	}
	if !reflect.DeepEqual(want, args) {
		t.Fatalf("expect %+v\n got %+v", want, args)
	}

}
