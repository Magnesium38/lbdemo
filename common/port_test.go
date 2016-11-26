package common

import "testing"

func TestGetOpenPort(t *testing.T) {
	for i := 0; i < 15; i++ {
		_, err := GetOpenPort()
		if err != nil {
			t.Error(err)
		}
	}
}
