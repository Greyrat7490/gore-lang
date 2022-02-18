package types

import "strconv"

type Type int
const (
    I32 Type = iota
    Str Type = iota
    TypesCount uint = iota
)

func (t Type) Readable() string {
    // compile time reminder to add cases when Types are added
    const _ uint = 2 - TypesCount

    switch t {
    case Str:
        return "str"
    case I32:
        return "i32"
    default:
        return ""
    }
}

// -1 if string does not contain a valid type
func ToType(s string) Type {
    const _ uint = 2 - TypesCount

    switch s {
    case "str":
        return Str
    case "i32":
        return I32
    default:
        return -1
    }
}

// -1 if neigther Str nor i32
func TypeOfVal(val string) Type {
    if val[0] == '"' && val[len(val) - 1] == '"' {
        return Str
    } else if _, err := strconv.Atoi(val); err == nil {
        return I32
    } else {
        return -1
    }
}