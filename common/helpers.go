package common

import "sync"

// An AtomicStringSlice is a string slice with a lock so it is atomic.
type AtomicStringSlice struct {
	sync.Mutex
	s []string
}

// Add adds a string to the AtomicStringSlice
func (slice *AtomicStringSlice) Add(str string) {
	slice.Lock()
	defer slice.Unlock()

	slice.s = append(slice.s, str)
}

// Get returns a string from the AtomicStringSlice
func (slice *AtomicStringSlice) Get(index int) string {
	slice.Lock()
	defer slice.Unlock()

	return slice.s[index]
}

// Has returns whether a string already exists in the AtomicStringSlice
func (slice *AtomicStringSlice) Has(str string) bool {
	slice.Lock()
	defer slice.Unlock()

	for _, value := range slice.s {
		if str == value {
			return true
		}
	}

	return false
}

// List returns a copy of every element in the AtomicStringSlice
func (slice *AtomicStringSlice) List() []string {
	duplicate := make([]string, len(slice.s))
	copy(duplicate, slice.s)
	return duplicate
}

// Remove removes a string from the AtomicStringSlice
func (slice *AtomicStringSlice) Remove(str string) {
	slice.Lock()
	defer slice.Unlock()

	for i := len(slice.s) - 1; i >= 0; i-- {
		if slice.s[i] == str {
			slice.s = append(slice.s[:i], slice.s[i+1:]...)
			return
		}
	}
}
