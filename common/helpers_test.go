package common

import "testing"

func TestAtomicStringSlice(t *testing.T) {
	slice := AtomicStringSlice{}
	if len(slice.List()) != 0 {
		t.Error(slice, "Slice not empty upon creation.")
	}

	slice.Add("apple")
	if len(slice.List()) != 1 {
		t.Error(slice, "Slice length not equal to one after add.")
	}

	slice.Add("banana")
	if len(slice.List()) != 2 {
		t.Error(slice, "Slice length not equal to two after 2nd add.")
	}

	if slice.Get(0) != "apple" {
		t.Error(slice, "Index at 0 is not equal to 'apple'.")
	}

	if slice.Has("orange") {
		t.Error(slice, "Slice does not have the string 'orange'.")
	}

	if !slice.Has("apple") {
		t.Error(slice, "Slice does have the string 'apple'.")
	}

	slice.Remove("apple")

	if slice.Has("apple") {
		t.Error(slice, "Slice has the string 'apple' after removal.")
	}

	slice.Add("raspberry")
	slice.Add("peach")
	slice.Add("strawberry")
	slice.Add("kiwi")

	s := slice.List()
	expected := []string{"banana", "raspberry", "peach", "strawberry", "kiwi"}
	for i := 0; i < len(s)-1; i++ {
		if s[i] != expected[i] {
			t.Log(s[i], expected[i])
			t.Error(s, "Slice result was not as expected.")
		}
	}
}
