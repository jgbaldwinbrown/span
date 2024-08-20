package span

import (
	"testing"
	"fmt"
	"cmp"
)

func TestSet(t *testing.T) {
	s := NewSet[int](cmp.Compare)
	s.Add(Span[int]{0, 1})
	s.Add(Span[int]{-5, 22})
	s.Add(Span[int]{3, 4})
	fmt.Println(s)
}
