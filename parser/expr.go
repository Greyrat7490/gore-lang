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
    COMPARE_PRECEDENCE  precedence = 0
    ADD_SUB_PRECEDENCE  precedence = 1
    MUL_DIV_PRECEDENCE  precedence = 2
    EXP_ROOT_PRECEDENCE precedence = 3
    PAREN_PRECEDENCE    precedence = 4
)

func prsExpr() ast.OpExpr {
    var expr ast.OpExpr
    switch token.Cur().Type {
    case token.Number, token.Str, token.Boolean:
        expr = prsLitExpr()

    case token.Name:
        expr = prsIdentExpr()

    case token.ParenL:
        expr = prsParenExpr()

    case token.Plus, token.Minus:
        expr = prsUnaryExpr()

    // TODO: OpFnCall

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
    return  token.Cur().Type == token.Plus || token.Cur().Type == token.Minus
}

func isParenExpr() bool {
    return  token.Cur().Type == token.ParenL
}

func isBinaryExpr() bool {
    return  token.Peek().Type == token.Plus || token.Peek().Type == token.Minus ||
            token.Peek().Type == token.Mul  || token.Peek().Type == token.Div ||
            isComparison()
}

func isComparison() bool {
    return  token.Peek().Type == token.Eql || token.Peek().Type == token.Neq ||
            token.Peek().Type == token.Grt || token.Peek().Type == token.Lss ||
            token.Peek().Type == token.Geq || token.Peek().Type == token.Leq
}

func getPrecedence() precedence {
    switch {
    case isComparison():
        return COMPARE_PRECEDENCE
    case token.Peek().Type == token.Plus || token.Peek().Type == token.Minus:
        return ADD_SUB_PRECEDENCE
    case token.Peek().Type == token.Mul || token.Peek().Type == token.Div:
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

    token.Next()
    expr.Operand = prsValue()

    return &expr
}

func prsBinary(expr ast.OpExpr, min_precedence precedence) ast.OpExpr {
    for isBinaryExpr() && getPrecedence() >= min_precedence {
        var b ast.BinaryExpr
        b.OperandL = expr

        precedenceL := getPrecedence()
        b.Operator = token.Next()

        token.Next()
        precedenceR := getPrecedence()

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


func prsCallFn() ast.OpFnCall {
    name := token.Cur()
    token.Next()
    vals := prsPassArgs()

    return ast.OpFnCall{ FnName: name, Values: vals }
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
