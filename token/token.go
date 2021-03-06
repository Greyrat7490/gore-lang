package token

import (
    "os"
    "fmt"
    "bufio"
    "strconv"
    "gorec/types"
)

var tokens []Token
var idx int = -1

type TokenType uint
const (
    Unknown TokenType = iota

    EOF             // end of file

    Name            // var/func name
    Typename        // i32, str, bool
    Str             // "string"
    Number          // 1234
    Boolean         // true/false

    Plus            // +
    Minus           // -
    Mul             // *
    Div             // /
    Mod             // %

    And             // &&
    Or              // ||

    Amp             // &

    Eql             // ==
    Neq             // !=
    Geq             // >=
    Leq             // <=
    Lss             // <
    Grt             // >
    Not             // !

    ParenL          // (
    ParenR          // )
    BrackL          // [
    BrackR          // ]
    BraceL          // {
    BraceR          // }
    UndScr          // _

    Comma           // ,
    Colon           // :
    SemiCol         // ;

    Comment         // // ..., /* ... */

    DefVar          // :=
    DefConst        // ::
    Assign          // =
    Fn              // fn
    If              // if
    Elif            // elif
    Else            // else
    While           // while
    For             // for
    Break           // break
    Continue        // continue
    Through         // through
    XSwitch         // $

    TokenTypeCount uint = iota
)

func ToTokenType(s string) TokenType {
    switch s {
    case "true", "false":
        return Boolean

    case "+":
        return Plus
    case "-":
        return Minus
    case "*":
        return Mul
    case "/":
        return Div
    case "%":
        return Mod

    case "&&":
        return And
    case "||":
        return Or

    case "&":
        return Amp

    case "==":
        return Eql
    case "!=":
        return Neq
    case ">=":
        return Geq
    case "<=":
        return Leq
    case ">":
        return Grt
    case "<":
        return Lss
    case "!":
        return Not

    case "(":
        return ParenL
    case ")":
        return ParenR
    case "[":
        return BrackL
    case "]":
        return BrackR
    case "{":
        return BraceL
    case "}":
        return BraceR
    case "_":
        return UndScr

    case ",":
        return Comma
    case ";":
        return SemiCol
    case ":":
        return Colon

    case "//", "/*", "*/":
        return Comment

    case ":=":
        return DefVar
    case "::":
        return DefConst
    case "=":
        return Assign
    case "fn":
        return Fn
    case "if":
        return If
    case "elif":
        return Elif
    case "else":
        return Else
    case "while":
        return While
    case "for":
        return For
    case "break":
        return Break
    case "continue":
        return Continue
    case "through":
        return Through
    case "$":
        return XSwitch

    default:
        if types.ToType(s, false) != nil {
            return Typename
        } else if s[0] == '"' && s[len(s) - 1] == '"' {
            return Str
        } else if _, err := strconv.Atoi(s); err == nil {
            return Number
        }

        return Name
    }
}

func (t TokenType) String() string {
    switch t {
    case EOF:
        return "EOF"

    case Plus:
        return "Plus"
    case Minus:
        return "Minus"
    case Mul:
        return "Mul"
    case Div:
        return "Div"
    case Mod:
        return "Mod"

    case And:
        return "And"
    case Or:
        return "Or"

    case Amp:
        return "Amp"

    case Eql:
        return "Eql"
    case Neq:
        return "Neq"
    case Geq:
        return "Geq"
    case Leq:
        return "Leq"
    case Grt:
        return "Grt"
    case Lss:
        return "Lss"
    case Not:
        return "Not"

    case ParenL:
        return "ParenL"
    case ParenR:
        return "ParenR"
    case BrackL:
        return "BrackL"
    case BrackR:
        return "BrackR"
    case BraceL:
        return "BraceL"
    case BraceR:
        return "BraceR"
    case UndScr:
        return "UnderS"

    case Comma:
        return "Comma"
    case SemiCol:
        return "SemiCol"
    case Colon:
        return "Colon"

    case Comment:
        return "Comment"

    case DefVar:
        return "DefVar"
    case DefConst:
        return "DefConst"
    case Assign:
        return "Assign"
    case Fn:
        return "Fn"
    case If:
        return "If"
    case Elif:
        return "Elif"
    case Else:
        return "Else"
    case While:
        return "While"
    case For:
        return "For"
    case Break:
        return "Break"
    case Continue:
        return "Continue"
    case Through:
        return "Through"
    case XSwitch:
        return "XSwitch"

    case Typename:
        return "Typename"
    case Str:
        return "Str"
    case Number:
        return "Number"
    case Boolean:
        return "Boolean"
    case Name:
        return "Name"

    default:
        return "Unknown"
    }
}

type Pos struct {
    Line int
    Col int
    // later filename
}

func (p Pos) At() string {
    return fmt.Sprintf("at line: %d, col: %d", p.Line, p.Col)
}

type Token struct {
    Type TokenType
    Str string
    Pos Pos
}

func (t Token) String() string {
    return fmt.Sprintf("\"%s\"(%v)", t.Str, t.Type)
}

func (t Token) At() string {
    return t.Pos.At()
}

func split(s string, start int, end int, line int) {
    if start != end {
        s := s[start:end]
        t := ToTokenType(s)

        tokens = append(tokens, Token{t, s, Pos{line, start+1} })
    }
}

func Tokenize(path string) {
    fmt.Println("[INFO] tokenizing...")

    src, err := os.Open(path)
    if err != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]", err)
        os.Exit(1)
    }
    scanner := bufio.NewScanner(src)

    comment := false
    mlComment := false
    strLit := false
    escape := false

    for lineNum := 1; scanner.Scan(); lineNum++ {
        line := scanner.Text()

        start := 0
        comment = false
        for i := 0; i < len(line); i++ {
            // in single line comment
            if comment {
                break
            }

            // in multiline comment
            if mlComment {
                if i+2 < len(line) && line[i:i+2] == "*/" {
                    mlComment = false
                    start = i+2
                    i++
                }

                continue
            }

            // in string literal
            if strLit {
                if escape {
                    escape = false
                } else {
                    if line[i] == '"' {
                        strLit = false
                    } else if line[i] == '\\' {
                        escape = true
                    }
                }

                continue
            }

            switch line[i] {
            // start string literal
            case '"':
                strLit = true

            // split at space
            case ' ', '\t':
                split(line, start, i, lineNum)
                start = i+1

            // split at //, /*, :=, ::, <=, >=, ==, !=, &&
            case '/', ':', '<', '>', '=', '!', '&':
                if i+2 <= len(line) {
                    s := line[i:i+2]
                    switch s {
                    // start single line comment
                    case "//":
                        split(line, start, i, lineNum)
                        comment = true
                        i++
                        continue
                    // start multiline comment
                    case "/*":
                        split(line, start, i, lineNum)
                        mlComment = true
                        start = i+1
                        i++
                        continue

                    case "&&", ":=", "::", "!=", "==", "<=", ">=":
                        split(line, start, i, lineNum)
                        tokens = append(tokens, Token{ ToTokenType(s), s, Pos{lineNum, i+1} })
                        start = i+3
                        i += 2
                        continue
                    }
                }

                fallthrough

            // split at non space char (and keep char)
            case '(', ')', '{', '}', '+', '-', '*', '%', ',', ';', '$':
                split(line, start, i, lineNum)

                tokens = append(tokens, Token{ ToTokenType(string(line[i])), string(line[i]), Pos{lineNum, i+1} })
                start = i+1
            }
        }

        if !comment && !mlComment && len(line) > start {
            split(line, start, len(line), lineNum)
        }
    }

    pos := tokens[len(tokens)-1].Pos
    pos.Col += len(tokens[len(tokens)-1].Str)
    tokens = append(tokens, Token{EOF, "EOF", pos})

    if strLit {
        fmt.Fprintln(os.Stderr, "string literal not terminated (missing '\"')")
        os.Exit(1)
    }
    if mlComment {
        fmt.Fprintln(os.Stderr, "comment not terminated (missing \"*/\")")
        os.Exit(1)
    }
}

func Cur() Token {
    return tokens[idx]
}

func Next() Token {
    idx++

    if idx >= len(tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return tokens[idx]
}

func Peek() Token {
    if idx+1 >= len(tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return tokens[idx+1]
}

func Peek2() Token {
    if idx+2 >= len(tokens) {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected end of file")
        os.Exit(1)
    }

    return tokens[idx+2]
}

func Last() Token {
    if idx < 1 {
        fmt.Fprintln(os.Stderr, "[ERROR] unexpected beginning of file (expected 1 word more at the start of the file)")
        os.Exit(1)
    }

    return tokens[idx-1]
}

func Last2() Token {
    if idx < 2 {
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected beginning of file (expected %d words more at the start of the file)\n", 2-idx)
        os.Exit(1)
    }

    return tokens[idx-2]
}

// returns Pos{ -1, -1 } if not found
func FindNext (t TokenType) Pos {
    for i := idx+1; i < len(tokens); i++ {
        if tokens[i].Type == t {
            return tokens[i].Pos
        }
    }

    return Pos{ -1, -1 }
}
