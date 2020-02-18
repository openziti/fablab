package fablib

import (
	"testing"
)

func TestSummarizeSar(t *testing.T) {
	_, err := SummarizeSar(sarTestData[:])
	if err != nil {
		t.Fatalf("unexpected error (%w)", err)
		t.Failed()
	}
}