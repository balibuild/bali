package barrow_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func mapValidChar(r rune) rune {
	if r >= 'a' && r <= 'z' ||
		r >= 'A' && r <= 'Z' ||
		r >= '0' && r <= '9' ||
		isOneOf(r, '.', '_', '+', '-') {
		return r
	}
	return -1
}

// isOneOf checks whether a rune is one of the runes in rr
func isOneOf(r rune, rr ...rune) bool {
	for _, char := range rr {
		if r == char {
			return true
		}
	}
	return false
}

var (
	validCharMap = []int{
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, '+', -1, '-', '.', -1,
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', -1, -1, -1, -1, -1, -1,
		-1, 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
		'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', -1, -1, -1, -1, '_',
		-1, 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
		'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	}
)

func TestGen(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 255; i++ {
		if i%16 == 0 {
			b.WriteString("\n")
		}
		if r := mapValidChar(rune(i)); r != -1 {
			fmt.Fprintf(&b, "'%c',", r)
			continue
		}
		fmt.Fprintf(&b, "-1,")

	}
	fmt.Fprintf(os.Stderr, "%s\n", b.String())
}

func TestRune(t *testing.T) {
	fmt.Fprintf(os.Stderr, "%c %d %d\n", validCharMap['c'], validCharMap['\n'], '\n')
}
