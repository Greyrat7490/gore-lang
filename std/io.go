package std

import (
    "os"
    "fmt"
)

const STDOUT = 1

func definePrintStr(asm *os.File) {
    asm.WriteString("printStr:\n")

    asm.WriteString("mov rdx, rsi\n")
    asm.WriteString("mov esi, edi\n")
    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func definePrintInt(asm *os.File) {
    asm.WriteString("printInt:\n")
    asm.WriteString("movsxd rax, edi\n")   // mov edi into eax and sign extends upper half of rax
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func definePrintPtr(asm *os.File) {
    asm.WriteString("printPtr:\n")
    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _int_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func definePrintBool(asm *os.File) {
    asm.WriteString("printBool:\n")

    asm.WriteString("mov rax, rdi\n")
    asm.WriteString("call _bool_to_str\n")

    asm.WriteString(fmt.Sprintf("mov rdi, %d\n", STDOUT))
    asm.WriteString("mov rdx, rax\n")
    asm.WriteString("mov rsi, rbx\n")
    syscall(asm, SYS_WRITE)

    asm.WriteString("ret\n\n")
}

func defineBtoS(asm *os.File) {
    asm.WriteString(`; rax = input int
; rbx = output string pointer
; rax = output string length
_bool_to_str:
    cmp rax, 0
    jne .c1
    mov rbx, _false
    mov rax, 5
    ret
    .c1:
        mov rbx, _true
        mov rax, 4
        ret

`)
}

func defineItoS(asm *os.File) {
    asm.WriteString(`; rax = input int
; rbx = output string pointer
; rax = output string length
_uint_to_str:
    push rcx
    push rdx

    mov ecx, 10

    mov rbx, _intBuf + 20
    .l1:
        xor edx, edx
        div ecx
        add dl, 48
        dec rbx
        mov byte [rbx], dl
        cmp eax, 0
        jne .l1

    mov rax, rbx
    sub rax, _intBuf
    inc rax
    pop rdx
    pop rcx
    ret

_int_to_str:
    push rcx
    push rdx
    push rax

    mov ecx, 10
    mov rbx, _intBuf + 20

    cmp rax, 0
    jge .l1

    neg rax

    .l1:
        xor edx, edx
        div ecx
        add dl, 48
        dec rbx
        mov byte [rbx], dl
        cmp eax, 0
        jne .l1

    pop rax
    cmp rax, 0
    jge .end

    dec rbx
    mov byte [rbx], 0x2d

    .end:
        mov rax, rbx
        sub rax, _intBuf
        inc rax
        pop rdx
        pop rcx
        ret

`)
}
