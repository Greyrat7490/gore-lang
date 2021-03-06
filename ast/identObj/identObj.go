package identObj

import (
    "os"
    "fmt"
    "gorec/token"
    "gorec/types"
    "gorec/ast/identObj/func"
    "gorec/ast/identObj/vars"
    "gorec/ast/identObj/consts"
)

type IdentObj interface {
    GetName() string
    GetPos() token.Pos
    Addr(fieldNum int) string
}

func DecVar(name token.Token, t types.Type) vars.Var {
    if InGlobalScope() {
        v := vars.CreateGlobalVar(name, t)
        curScope.identObjs[name.Str] = &v
        return &v
    } else {
        v := vars.CreateLocal(name, t, curScope.frameSize)
        curScope.frameSize += v.GetType().Size()
        curScope.identObjs[name.Str] = &v
        return &v
    }
}

func DecConst(name token.Token, t types.Type) *consts.Const {
    checkName(name)

    c := consts.CreateConst(name, t)
    curScope.identObjs[name.Str] = &c
    return &c
}

func DecFunc(name token.Token) *fn.Func {
    checkName(name)

    f := fn.CreateFunc(name, nil)
    curScope.identObjs[name.Str] = &f
    return &f
}

func AddBuildIn(name string, argname string, argtype types.Type) {
    if !InGlobalScope() {
        fmt.Fprintln(os.Stderr, "[ERROR] AddBuildIn has to be called in the global scope")
        os.Exit(1)
    }

    f := fn.CreateFunc(
        token.Token{ Str: name, Type: token.Name },
        []types.Type{ argtype },
    )

    curScope.identObjs[name] = &f
}
