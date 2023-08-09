package test

import (
	"fmt"
	"testing"

	"github.com/elliotchance/pie/v2"
)

func TestDiff(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5, 6}
	s2 := []int{2, 5, 6}
	diff, removed := pie.Diff(s1, s2)
	fmt.Println(diff, removed)
	fmt.Println(pie.Diff(s2, s1))
}
