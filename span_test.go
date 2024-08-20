package span

import (
	"testing"
	"fmt"
)

func TestSet(t *testing.T) {
	s := NewSet[int]()
	s.Add(Span[int]{0, 1})
	s.Add(Span[int]{-5, 22})
	s.Add(Span[int]{3, 4})
	fmt.Println(s)
}
