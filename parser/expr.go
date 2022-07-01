package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

type precedence int
const (
    LOGICAL_PRECEDENCE precedence = iota // &&, ||
    COMPARE_PRECEDENCE            = iota // ==, !=, <, <=, >, >=
    XSWITCH_PRECEDENCE            = iota // $ ... { ... }
    ADD_SUB_PRECEDENCE            = iota // +, -
    MUL_DIV_PRECEDENCE            = iota // *, /, %
    EXP_PRECEDENCE                = iota // **(TODO)
    PAREN_PRECEDENCE              = iota // ()
)

func prsExpr() ast.OpExpr {
    var expr ast.OpExpr
    switch token.Cur().Type {
    case token.Number, token.Str, token.Boolean:
        expr = prsLitExpr()

    case token.Name:
        if token.Peek().Type == token.ParenL {
            return prsCallFn()  // only tmp because binary ops are not supported with func calls yet
        } else {
            expr = prsIdentExpr()
        }

    case token.XSwitch:
        expr = prsXSwitch()

    case token.UndScr:
        expr = prsIdentExpr()

    case token.ParenL:
        expr = prsParenExpr()

    case token.Plus, token.Minus, token.Mul, token.Amp:
        expr = prsUnaryExpr()

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] no valid expression (got type %s)\n", token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)

        return &ast.BadExpr{}
    }

    if isBinaryExpr() {
        expr = prsBinary(expr, 0)
    }

    return expr
}

func isUnaryExpr() bool {
    return  token.Cur().Type == token.Plus || token.Cur().Type == token.Minus ||
            token.Cur().Type == token.Mul  || token.Cur().Type == token.Amp
}

func isParenExpr() bool {
    return  token.Cur().Type == token.ParenL
}

func isBinaryExpr() bool {
    if token.Peek().Pos.Line > token.Cur().Pos.Line {
        return false
    }

    return  token.Peek().Type == token.Plus || token.Peek().Type == token.Minus ||
            token.Peek().Type == token.Mul  || token.Peek().Type == token.Div   ||
            token.Peek().Type == token.Mod  ||
            token.Peek().Type == token.And  || token.Peek().Type == token.Or    ||
            isComparison()
}

func isComparison() bool {
    return  token.Peek().Type == token.Eql || token.Peek().Type == token.Neq ||
            token.Peek().Type == token.Grt || token.Peek().Type == token.Lss ||
            token.Peek().Type == token.Geq || token.Peek().Type == token.Leq
}

func getPrecedence() precedence {
    switch {
    case token.Peek().Type == token.And || token.Peek().Type == token.Or:
        return LOGICAL_PRECEDENCE
    case isComparison():
        return COMPARE_PRECEDENCE
    case token.Peek().Type == token.Plus || token.Peek().Type == token.Minus:
        return ADD_SUB_PRECEDENCE
    case token.Peek().Type == token.Mul || token.Peek().Type == token.Div || token.Peek().Type == token.Mod:
        return MUL_DIV_PRECEDENCE
    case isParenExpr():
        return PAREN_PRECEDENCE
    default:
        return precedence(0)
    }
}

func prsIdentExpr() *ast.IdentExpr {
    return &ast.IdentExpr{ Ident: token.Cur() }
}

func prsLitExpr() *ast.LitExpr {
    val := token.Cur()
    t := types.TypeOfVal(val.Str)

    return &ast.LitExpr{ Val: val, Type: t }
}

func prsValue() ast.OpExpr {
    if token.Cur().Type == token.Name {
        return prsIdentExpr()
    } else {
        return prsLitExpr()
    }
}

func prsParenExpr() *ast.ParenExpr {
    expr := ast.ParenExpr{ ParenLPos: token.Cur().Pos }

    token.Next()
    expr.Expr = prsExpr()

    expr.ParenRPos = token.Next().Pos

    if token.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got \"%s\"(%s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return &expr
}

func prsUnaryExpr() *ast.UnaryExpr {
    expr := ast.UnaryExpr{ Operator: token.Cur() }

    switch expr.Operator.Type {
    case token.Mul:
        if token.Next().Type == token.ParenL {
            expr.Operand = prsParenExpr()
        } else {
            expr.Operand = prsIdentExpr()
        }
    case token.Amp:
        token.Next()
        expr.Operand = prsIdentExpr()
    default:
        token.Next()
        expr.Operand = prsValue()
    }

    return &expr
}

func prsCaseExpr(condBase ast.OpExpr, placeholder *ast.OpExpr, lastCaseEnd token.Pos) (caseExpr ast.CaseExpr) {
    if token.Cur().Type == token.Colon {
        if token.Last().Pos.Line == token.Cur().Pos.Line {
            fmt.Fprintln(os.Stderr, "[ERROR] missing case body(expr) for this case")
            fmt.Fprintln(os.Stderr, "\t" + lastCaseEnd.At())
        } else {
            fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        }
        os.Exit(1)
    }
    if token.Cur().Type == token.Comma {
        fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \",\"")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    if token.Last().Pos.Line == token.Cur().Pos.Line && token.Last().Type != token.SemiCol && token.Last().Type != token.BraceL {
        fmt.Fprintln(os.Stderr, "[ERROR] cases should always start in a new line or after a \";\"")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    // parse case cond(s) ----------------
    expr := prsExpr()
    var conds ast.OpExpr = nil
    for token.Next().Type == token.Comma {
        conds = completeCond(placeholder, condBase, expr, conds)

        if token.Peek().Type == token.Colon || token.Peek().Type == token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: no expr after \",\"")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }

        token.Next()
        expr = prsExpr()
    }

    caseExpr.ColonPos = token.Cur().Pos
    caseExpr.Cond = completeCond(placeholder, condBase, expr, conds)

    if token.Cur().Type != token.Colon {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \":\" but got \"%s\"(%s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    if nextColon := token.FindNext(token.Colon); token.Cur().Pos.Line == nextColon.Line {
        nextSemiCol := token.FindNext(token.SemiCol)

        if nextSemiCol.Line == -1 || (nextSemiCol.Line == nextColon.Line && nextSemiCol.Col > nextColon.Col) {
            fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \";\"")
            fmt.Fprintln(os.Stderr, "\t" + nextColon.At())
            os.Exit(1)
        }
    }


    // parse case body -------------------
    if token.Peek().Type == token.SemiCol {
        fmt.Fprintln(os.Stderr, "[ERROR] missing case body(expr) for this case")
        fmt.Fprintln(os.Stderr, "\t" + token.Last().At())
        os.Exit(1)
    }

    token.Next()
    caseExpr.Expr = prsExpr()

    if token.Peek().Type == token.SemiCol { token.Next() }

    return
}

func prsXSwitch() *ast.SwitchExpr {
    switchExpr := ast.SwitchExpr{ Pos: token.Cur().Pos }
    var condBase ast.OpExpr = nil
    var placeholder *ast.OpExpr = nil

    // set condBase -----------------------
    if token.Next().Type != token.BraceL {
        condBase = prsExpr()
        placeholder = getPlaceholder(condBase)
    }

    // parse cases ------------------------
    if token.Cur().Type != token.BraceL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"{\" at the beginning of the xswitch " +
            "but got \"%s\"(%v)\n", token.Cur().Str, token.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
    switchExpr.BraceLPos = token.Cur().Pos

    lastCaseEnd := token.Pos{}
    for token.Next().Type != token.BraceR {
        expr := prsCaseExpr(condBase, placeholder, lastCaseEnd)
        lastCaseEnd = expr.ColonPos
        switchExpr.Cases = append(switchExpr.Cases, expr)
    }

    switchExpr.BraceRPos = token.Cur().Pos


    // catch some syntax errors -----------
    if len(switchExpr.Cases) == 0 {
        fmt.Fprintln(os.Stderr, "[ERROR] empty xswitch")
        fmt.Fprintln(os.Stderr, "\t" + switchExpr.BraceLPos.At())
        os.Exit(1)
    }
    for i,c := range switchExpr.Cases {
        if c.Cond == nil && i != len(switchExpr.Cases)-1 {
            i = len(switchExpr.Cases)-1 - i
            if i == 1 {
                fmt.Fprintln(os.Stderr, "[ERROR] one case after the default case (unreachable code)")
            } else {
                fmt.Fprintf(os.Stderr, "[ERROR] %d cases after the default case (unreachable code)\n", i)
            }
            fmt.Fprintln(os.Stderr, "\t" + c.ColonPos.At())
            os.Exit(1)
        }
    }
    if switchExpr.Cases[len(switchExpr.Cases)-1].Cond != nil {
        fmt.Fprintln(os.Stderr, "[ERROR] every xswitch requires a default case")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return &switchExpr
}

func prsBinary(expr ast.OpExpr, min_precedence precedence) ast.OpExpr {
    for isBinaryExpr() && getPrecedence() >= min_precedence {
        var b ast.BinaryExpr
        b.OperandL = expr

        precedenceL := getPrecedence()
        b.Operator = token.Next()

        token.Next()
        precedenceR := getPrecedence()

        // switch/xswitch
        if token.Cur().Type == token.BraceL {
            return &b
        }

        switch {
        case isParenExpr():
            b.OperandR = prsParenExpr()
        case isUnaryExpr():
            b.OperandR = prsUnaryExpr()
        default:
            b.OperandR = prsValue()
        }

        if isBinaryExpr() {
            b.OperandR = prsBinary(b.OperandR, precedenceL+1)
        }

        // left to right as correct order of operations
        if precedenceR > precedenceL {
            swap(&b)
        }

        expr = &b
    }

    return expr
}

func swap(expr *ast.BinaryExpr) {
    if expr.Operator.Type == token.Minus {
        expr.Operator.Type = token.Plus
        expr.Operator.Str = "+"

        t := token.Token{ Type: token.Minus, Str: "-" }
        expr.OperandR = &ast.UnaryExpr{ Operator: t, Operand: expr.OperandR }
    }

    tmp := expr.OperandR
    expr.OperandR = expr.OperandL
    expr.OperandL = tmp
}


func prsCallFn() *ast.OpFnCall {
    name := token.Cur()
    token.Next()
    vals := prsPassArgs()

    return &ast.OpFnCall{ FnName: name, Values: vals }
}

func prsPassArgs() []ast.OpExpr {
    if token.Cur().Type != token.ParenL {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"(\" but got \"%s\"(%s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    var values []ast.OpExpr

    if token.Next().Type == token.ParenR {
        return values
    }

    values = append(values, prsExpr())
    for token.Next().Type == token.Comma {
        token.Next()
        values = append(values, prsExpr())
    }

    if token.Cur().Type != token.ParenR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \")\" but got \"%s\"(%s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return values
}
