package barrow

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestParsePerm(t *testing.T) {
	i, err := strconv.ParseInt("0755", 8, 64)
	if err != nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%d %d\n", i, 0755)
}
