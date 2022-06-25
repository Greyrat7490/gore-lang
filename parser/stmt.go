package prs

import (
    "os"
    "fmt"
    "gorec/ast"
    "gorec/token"
    "gorec/types"
)

func prsStmt() ast.OpStmt {
    switch t := token.Next(); t.Type {
    case token.Dec_var:
        d := prsDecVar()
        return &ast.OpDeclStmt{ Decl: &d }

    case token.Def_var:
        d := prsDefVar()
        return &ast.OpDeclStmt{ Decl: &d }

    case token.BraceL:
        b := prsBlock()
        return &b

    case token.If:
        ifStmt := prsIfStmt()

        if token.Cur().Str == "{" {
            switchStmt := prsCondSwitch(ifStmt)
            return &switchStmt
        }

        return &ifStmt

    case token.While:
        w := prsWhileStmt()
        return &w

    case token.For:
        f := prsForStmt()
        return &f

    case token.Break:
        b := prsBreak()
        return &b

    case token.Continue:
        c := prsContinue()
        return &c

    case token.Number, token.Str, token.Boolean:
        return &ast.OpExprStmt{ Expr: prsLitExpr() }

    case token.UndScr:
        return &ast.OpExprStmt{ Expr: prsIdentExpr() }

    case token.ParenL:
        return &ast.OpExprStmt{ Expr: prsParenExpr() }

    case token.Plus, token.Minus, token.Mul, token.Amp:
        expr := prsUnaryExpr()

        if token.Peek().Type == token.Assign {
            a := prsAssignVar(expr)
            return &a
        } else {
            return &ast.OpExprStmt{ Expr: expr }
        }

    case token.Name:
        if token.Peek().Type == token.Assign {
            name := prsIdentExpr()
            a := prsAssignVar(name)
            return &a
        } else if token.Peek().Type == token.ParenL {
            return &ast.OpExprStmt{ Expr: prsCallFn() }
        } else {
            return &ast.OpExprStmt{ Expr: prsIdentExpr() }
        }

    case token.Elif:
        fmt.Fprintf(os.Stderr, "[ERROR] missing if (elif without an if before)\n")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Else:
        fmt.Fprintf(os.Stderr, "[ERROR] missing if (else without an if before)\n")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Assign:
        fmt.Fprintf(os.Stderr, "[ERROR] no destination for assignment\n")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    case token.Def_fn:
        fmt.Fprintln(os.Stderr, "[ERROR] you are not allowed to define functions inside a function")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}

    default:
        fmt.Fprintf(os.Stderr, "[ERROR] unexpected token \"%s\" (of type %s)\n", token.Cur().Str, token.Cur().Type.Readable())
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
        return &ast.BadStmt{}
    }
}

func prsBlock() ast.OpBlock {
    block := ast.OpBlock{ BraceLPos: token.Cur().Pos }

    for token.Peek().Type != token.BraceR {
        block.Stmts = append(block.Stmts, prsStmt())
    }

    block.BraceRPos = token.Next().Pos

    return block
}

func prsAssignVar(dest ast.OpExpr) ast.OpAssignVar {
    pos := token.Next().Pos
    token.Next()
    val := prsExpr()

    return ast.OpAssignVar{ Pos: pos, Dest: dest, Value: val }
}

func prsIfStmt() ast.IfStmt {
    pos := token.Cur().Pos
    token.Next()
    cond := prsExpr()

    // cond-switch
    if token.Cur().Str == "{" {
        return ast.IfStmt{ Pos: pos, Cond: cond }

    // normal if
    } else {
        token.Next()
        block := prsBlock()

        ifStmt := ast.IfStmt{ Pos: pos, Cond: cond, Block: block }

        if token.Peek().Type == token.Else {
            token.Next()
            elseStmt := prsElse()
            ifStmt.Else = &elseStmt
        } else if token.Peek().Type == token.Elif {
            token.Next()
            elifStmt := prsElif()
            ifStmt.Elif = &elifStmt
        }

        return ifStmt
    }
}

func prsElif() ast.ElifStmt {
    return ast.ElifStmt(prsIfStmt())
}

func prsElse() ast.ElseStmt {
    pos := token.Cur().Pos
    token.Next()
    block := prsBlock()

    return ast.ElseStmt{ ElsePos: pos, Block: block }
}

func prsWhileStmt() ast.WhileStmt {
    var op ast.WhileStmt = ast.WhileStmt{ WhilePos: token.Cur().Pos, InitVal: nil }

    if token.Peek().Type == token.Name && token.Peek2().Type == token.Typename {
        op.Dec = prsDecVar()

        if token.Next().Type != token.Comma {
            fmt.Fprintln(os.Stderr, "[ERROR] missing \",\"")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
        }

        token.Next()
        expr := prsExpr()

        if token.Next().Type == token.Comma {
            token.Next()
            op.Cond = prsExpr()
            op.InitVal = expr
            token.Next()
        } else {
            op.InitVal = &ast.LitExpr{ Val: token.Token{ Type: token.Number, Str: "0" }, Type: types.I32Type{} }
            op.Cond = expr
        }
    } else {
        token.Next()
        op.Cond = prsExpr()
        token.Next()
    }

    op.Block = prsBlock()

    return op
}

func prsForStmt() ast.ForStmt {
    var op ast.ForStmt = ast.ForStmt{
        ForPos: token.Cur().Pos,
        Limit: nil,
        Start: &ast.LitExpr{
            Val: token.Token{ Str: "0", Type: token.Number },
            Type: types.I32Type{},
        },
    }

    op.Dec = prsDecVar()

    op.Step = &ast.BinaryExpr{
        Operator: token.Token{ Type: token.Plus },
        OperandL: &ast.IdentExpr{ Ident: op.Dec.Varname },
        OperandR: &ast.LitExpr{
            Val: token.Token{ Str: "1", Type: token.Number },
            Type: types.I32Type{},
        },
    }

    if token.Next().Type == token.Comma {
        token.Next()
        op.Limit = prsExpr()

        if token.Next().Type == token.Comma {
            token.Next()
            op.Start = prsExpr()

            if token.Next().Type == token.Comma {
                token.Next()
                op.Step = prsExpr()
                token.Next()
            }
        }
    }

    op.Block = prsBlock()

    return op
}

func prsBreak() ast.BreakStmt {
    var op ast.BreakStmt = ast.BreakStmt{ Pos: token.Cur().Pos }
    return op
}

func prsContinue() ast.ContinueStmt {
    var op ast.ContinueStmt = ast.ContinueStmt{ Pos: token.Cur().Pos }
    return op
}


func getPlaceholder(cond ast.OpExpr) (expr *ast.OpExpr) {
    for {
        if b, ok := cond.(*ast.BinaryExpr); !ok {
            fmt.Fprintln(os.Stderr, "[ERROR] expected condition to be a BinaryExpr")
            fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
            os.Exit(1)
            return nil
        } else {
            if b.Operator.Type == token.Eql || b.Operator.Type == token.Neq ||
               b.Operator.Type == token.Grt || b.Operator.Type == token.Lss ||
               b.Operator.Type == token.Geq || b.Operator.Type == token.Leq {
                return &b.OperandR
            }

            cond = b.OperandR
        }
    }
}

func completeCond(placeholder *ast.OpExpr, baseCond ast.OpExpr, expr *ast.OpExpr) {
    if b, ok := baseCond.(*ast.BinaryExpr); ok {
        if ident, ok := (*expr).(*ast.IdentExpr); ok {
            if ident.Ident.Type == token.UndScr {
                *expr = nil
                return
            }
        }

        *placeholder = *expr
        newCond := *b
        *expr = &newCond
    } else {
        fmt.Fprintln(os.Stderr, "[ERROR] expected condition to be a BinaryExpr")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }
}

func prsCondSwitchBlock() (stmts []ast.OpStmt) {
    if token.Peek().Type == token.Colon {
        fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
        fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
        os.Exit(1)
    }

    for token.Peek().Type != token.BraceR {
        var stmt ast.OpStmt
        if token.Peek().Type == token.Comma {
            token.Next()

            if token.Peek().Type == token.Colon {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
                fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
                os.Exit(1)
            }

            stmt = prsStmt()

            if token.Peek().Type != token.Colon {
                fmt.Fprintln(os.Stderr, "[ERROR] expected \",\" before a case condition to indicate the end of the current")
                fmt.Fprintln(os.Stderr, "\t" + token.Peek().At())
                os.Exit(1)
            }
        } else {
            stmt = prsStmt()
        }

        if token.Peek().Type == token.Colon {
            token.Next()

            if nextColon := token.FindNext(token.Colon); token.Cur().Pos.Line == nextColon.Line {
                nextComma := token.FindNext(token.Comma)

                if nextComma.Line == -1 || (nextComma.Line == nextColon.Line && nextComma.Col > nextColon.Col) {
                    fmt.Fprintln(os.Stderr, "[ERROR] multiple cases in a line should be separated with a \",\"")
                    fmt.Fprintln(os.Stderr, "\t" + nextColon.At())
                    os.Exit(1)
                }
            }

            if token.Last().Pos.Line < token.Cur().Pos.Line {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: nothing before \":\"")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }

            if e, ok := stmt.(*ast.OpExprStmt); ok {
                stmt = &ast.CaseStmt{ Cond: e.Expr, ColonPos: token.Cur().Pos }
            } else {
                fmt.Fprintln(os.Stderr, "[ERROR] invalid case condition: expected an expr before \":\"")
                fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
                os.Exit(1)
            }

        }

        stmts = append(stmts, stmt)
    }

    if token.Next().Type != token.BraceR {
        fmt.Fprintf(os.Stderr, "[ERROR] expected \"}\" at the end of the cond-switch " +
            "but got \"%s\"(%v)\n", token.Cur().Str, token.Cur().Type)
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    return stmts
}

func toCases(baseCond ast.OpExpr, stmts []ast.OpStmt) (caseStmts []ast.CaseStmt) {
    if len(stmts) == 0 {
        fmt.Fprintln(os.Stderr, "[ERROR] empty cond-switch")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    if _, ok := stmts[0].(*ast.CaseStmt); !ok {
        fmt.Fprintln(os.Stderr, "[ERROR] missing case at the beginning of the cond-switch")
        fmt.Fprintln(os.Stderr, "\t" + token.Cur().At())
        os.Exit(1)
    }

    placeholder := getPlaceholder(baseCond)

    var curCase *ast.CaseStmt = nil
    for i := range stmts {
        if c, ok := stmts[i].(*ast.CaseStmt); ok {
            completeCond(placeholder, baseCond, &c.Cond)

            caseStmts = append(caseStmts, *c)
            curCase = &caseStmts[len(caseStmts)-1]
            continue
        }

        curCase.Stmts = append(curCase.Stmts, stmts[i])
    }

    for _,c := range caseStmts {
        if len(c.Stmts) == 0 {
            fmt.Fprintln(os.Stderr, "[ERROR] no stmts provided for this case")
            fmt.Fprintln(os.Stderr, "\t" + c.ColonPos.At())
            os.Exit(1)
        }
    }

    return
}

func prsCondSwitch(ifStmt ast.IfStmt) ast.SwitchStmt {
    switchStmt := ast.SwitchStmt{ Pos: ifStmt.Pos }

    stmts := prsCondSwitchBlock()
    switchStmt.Cases = toCases(ifStmt.Cond, stmts)

    return switchStmt
}
