package utils

import (
	"testing"
)

func TestShuffle(t *testing.T) {
	arr := []string{"a", "b", "c", "d", "e"}
	Shuffle(arr)

	// check arr was shuffled
	if arr[0] == "a" && arr[1] == "b" && arr[2] == "c" && arr[3] == "d" && arr[4] == "e" {
		t.Errorf("Array was not shuffled")
	}
}
