package promhelper

import (
	"errors"
	"testing"
)

var randomError = errors.New("random error")

func TestPromHelperCustomError(t *testing.T) {
	f := func() error {
		return randomError
	}

	err := f()
	var promError PromError
	if errors.As(err, &promError) {
		t.Fatal("this error is not a prom error")
	}

	if !errors.Is(err, randomError) {
		t.Fatal("this error must be the random error")
	}

	f2 := func() error {
		return NewPromError(StatusNotFound, randomError)
	}

	err = f2()
	if !errors.As(err, &promError) {
		t.Fatal("this must be a prom error")
	}
	if !errors.Is(promError.error, randomError) {
		t.Fatal("this error must be the random error")
	}
	if promError.status != StatusNotFound {
		t.Fatal("the status must be status not found")
	}

}
