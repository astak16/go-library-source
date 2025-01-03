package multierror

import (
	"errors"
	"fmt"
	"testing"
)

func TestPrefix_Error(t *testing.T) {
	original := &Error{
		Errors: []error{errors.New("foo")},
	}

	result := Prefix(original, "bar")
	fmt.Println(result)
	if result.(*Error).Errors[0].Error() != "bar foo" {
		t.Fatalf("bad: %s", result)
	}
}

func TestPrefix_NilError(t *testing.T) {
	var err error
	result := Prefix(err, "bar")
	if result != nil {
		t.Fatalf("bad: %#v", result)
	}
}

func TestPrefix_NonError(t *testing.T) {
	original := errors.New("foo")
	result := Prefix(original, "bar")

	if result == nil {
		t.Fatal("error result was nil")
	}
	if result.Error() != "bar foo" {
		t.Fatalf("bad: %s", result)
	}
}

func TestPrefix_NilErrorArg(t *testing.T) {
	var nilErr *Error
	result := Prefix(nilErr, "bar")
	expected := `0 errors occurred:
	

`
	if result.Error() != expected {
		t.Fatalf("bad: %#v", result.Error())
	}
}
