package utilities

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

// ArgvParser todo
type ArgvParser struct {
	opts       []option
	ua         []string
	index      int
	SubcmdMode bool // subcmd mode
}

// Add option
func (ae *ArgvParser) Add(name string, ha, val int) {
	ae.opts = append(ae.opts, option{name: name, ha: ha, val: val})
}

// Unresolved todo
func (ae *ArgvParser) Unresolved() []string {
	return ae.ua
}

func (ae *ArgvParser) parseInternalLong(a string, argv []string, ac Receiver) error {
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
	for _, o := range ae.opts {
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
		if ae.index+1 >= len(argv) {
			return ErrorCat("option '--", a, "' missing parameter")
		}
		oa = argv[ae.index+1]
		ae.index++
	}
	if err := ac.Invoke(ch, oa, a); err != nil {
		return err
	}
	return nil
}

func (ae *ArgvParser) parseInternalShort(a string, argv []string, ac Receiver) error {
	ha := OPTIONAL
	ch := -1
	if a[0] == '=' {
		return ErrorCat("unexpected argument '-", a, "'")
	}
	c := int(a[0])
	for _, o := range ae.opts {
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
	if len(a) > 2 {
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
		if ae.index+1 >= len(argv) {
			return ErrorCat("option '-", a[0:1], "' missing parameter")
		}
		oa = argv[ae.index+1]
		ae.index++
	}
	if err := ac.Invoke(ch, oa, a); err != nil {
		return err
	}
	return nil
}

func (ae *ArgvParser) parseInternal(a string, argv []string, ac Receiver) error {
	if len(a) == 1 {
		return ErrUnExpectedArg
	}
	if a[1] == '-' {
		return ae.parseInternalLong(a[2:], argv, ac)
	}
	return ae.parseInternalShort(a[1:], argv, ac)
}

// Execute todo
func (ae *ArgvParser) Execute(argv []string, ac Receiver) error {
	if len(argv) == 0 {
		return ErrNilArgv
	}
	ae.index = 1
	for ; ae.index < len(argv); ae.index++ {
		a := argv[ae.index]
		if len(a) == 0 || a[0] != '-' {
			if ae.SubcmdMode {
				ae.ua = append(ae.ua, argv[ae.index:]...)
				return nil
			}
			ae.ua = append(ae.ua, a)
			continue
		}
		if err := ae.parseInternal(a, argv, ac); err != nil {
			return err
		}
	}
	return nil
}
