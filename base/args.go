package base

// Thanks baulk

import (
	"errors"
	"strings"
)

// option argument type
const (
	REQUIRED int = iota
	NOARG
	OPTIONAL
)

// error
var (
	ErrNilArgv       = errors.New("argv is nil")
	ErrUnExpectedArg = errors.New("unexpected argument '-'")
)

// Receiver todo
type Receiver interface {
	Invoke(val int, oa, raw string) error
}

type option struct {
	name string
	ha   int
	val  int
}

// ParseArgs todo
type ParseArgs struct {
	opts       []option
	ua         []string
	index      int
	SubcmdMode bool // subcmd mode
}

// Add option
func (pa *ParseArgs) Add(name string, ha, val int) {
	pa.opts = append(pa.opts, option{name: name, ha: ha, val: val})
}

// Unresolved todo
func (pa *ParseArgs) Unresolved() []string {
	return pa.ua
}

func (pa *ParseArgs) parseInternalLong(a string, argv []string, ac Receiver) error {
	ha := OPTIONAL
	ch := -1
	var oa string
	i := strings.IndexByte(a, '=')
	if i != -1 {
		if i+1 > len(a) {
			return ErrorCat("unexpected argument '--", a, "'")
		}
		oa = a[i+1:]
		a = a[0:i]
	}
	for _, o := range pa.opts {
		if o.name == a {
			ch = o.val
			ha = o.ha
			break
		}
	}
	if ch == -1 {
		return ErrorCat("unregistered option '--", a, "'")
	}
	if len(oa) > 0 && ha == NOARG {
		return ErrorCat("option '--", a, "' unexpected parameter: ", oa)
	}
	if len(oa) == 0 && ha == REQUIRED {
		if pa.index+1 >= len(argv) {
			return ErrorCat("option '--", a, "' missing parameter")
		}
		oa = argv[pa.index+1]
		pa.index++
	}
	if err := ac.Invoke(ch, oa, a); err != nil {
		return err
	}
	return nil
}

func (pa *ParseArgs) parseInternalShort(a string, argv []string, ac Receiver) error {
	ha := OPTIONAL
	ch := -1
	if a[0] == '=' {
		return ErrorCat("unexpected argument '-", a, "'")
	}
	c := int(a[0])
	for _, o := range pa.opts {
		if o.val == c {
			ch = c
			ha = o.ha
			break
		}
	}
	if ch == -1 {
		return ErrorCat("unregistered option '-", a, "'")
	}
	var oa string
	if len(a) >= 2 {
		if a[1] == '=' {
			oa = a[2:]
		} else {
			oa = a[1:]
		}
	}
	if len(oa) > 0 && ha == NOARG {
		return ErrorCat("option '-", a[0:1], "' unexpected parameter: ", oa)
	}
	if len(oa) == 0 && ha == REQUIRED {
		if pa.index+1 >= len(argv) {
			return ErrorCat("option '-", a[0:1], "' missing parameter")
		}
		oa = argv[pa.index+1]
		pa.index++
	}
	if err := ac.Invoke(ch, oa, a); err != nil {
		return err
	}
	return nil
}

func (pa *ParseArgs) parseInternal(a string, argv []string, ac Receiver) error {
	if len(a) == 1 {
		return ErrUnExpectedArg
	}
	if a[1] == '-' {
		return pa.parseInternalLong(a[2:], argv, ac)
	}
	return pa.parseInternalShort(a[1:], argv, ac)
}

// Execute todo
func (pa *ParseArgs) Execute(argv []string, ac Receiver) error {
	if len(argv) == 0 {
		return ErrNilArgv
	}
	pa.index = 1
	for ; pa.index < len(argv); pa.index++ {
		a := argv[pa.index]
		if len(a) == 0 || a[0] != '-' {
			if pa.SubcmdMode {
				pa.ua = append(pa.ua, argv[pa.index:]...)
				return nil
			}
			pa.ua = append(pa.ua, a)
			continue
		}
		if err := pa.parseInternal(a, argv, ac); err != nil {
			return err
		}
	}
	return nil
}
