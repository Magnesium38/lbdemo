package common

import (
	"errors"
	"testing"
	"time"
)

func TestWorkGroup(t *testing.T) {
	jobs := NewWorkGroup()

	f := func() error {
		time.Sleep(5 * time.Second)
		return errors.New("unexpected")
	}

	for i := 0; i < 10; i++ {
		jobs.Add(f)
	}

	jobs.Add(func() error {
		return errors.New("expected")
	})

	for i := 0; i < 20; i++ {
		jobs.Add(f)
	}

	jobs.Start()

	fatalErr := jobs.Wait()
	if fatalErr.Error() != "expected" {
		t.Error("Unexpected error.")
	}
}
