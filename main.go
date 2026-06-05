package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

// トークンの種類
type TokenKind int

const (
	TK_RESERVED TokenKind = iota // 記号
	TK_NUM                       // 整数トークン
	TK_EOF                       // 入力の終わりを表すトークン
)

// トークン型
type Token struct {
	kind TokenKind
	next *Token
	val  int
	str  string
	pos  int // userInput中の位置
}

// 入力プログラム
var userInput string

// 現在着目しているトークン
var token *Token

// エラーを報告して終了
func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// エラー箇所を報告して終了
func errorAt(loc int, format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s\n", userInput)
	fmt.Fprintf(os.Stderr, "%*s", loc, "")
	fmt.Fprintf(os.Stderr, "^ ")
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// 次のトークンが期待している記号のときには、トークンを1つ読み進めて
// 真を返す。それ以外の場合には偽を返す。
func consume(op byte) bool {
	if token.kind != TK_RESERVED || token.str[0] != op {
		return false
	}
	token = token.next
	return true
}

// 次のトークンが期待している記号のときには、トークンを1つ読み進める。
// それ以外の場合にはエラーを報告する。
func expect(op byte) {
	if token.kind != TK_RESERVED || token.str[0] != op {
		errorAt(token.pos, "'%c'ではありません", op)
	}
	token = token.next
}

// 次のトークンが数値の場合、トークンを1つ読み進めてその数値を返す。
// それ以外の場合にはエラーを報告する。
func expectNumber() int {
	if token.kind != TK_NUM {
		errorAt(token.pos, "数ではありません")
	}
	val := token.val
	token = token.next
	return val
}

func atEOF() bool {
	return token.kind == TK_EOF
}

// 新しいトークンを作成してcurに繋げる
func newToken(kind TokenKind, cur *Token, str string) *Token {
	tok := &Token{kind: kind, str: str}
	cur.next = tok
	return tok
}

// userInputをトークナイズしてそれを返す
func tokenize() *Token {
	head := &Token{}
	cur := head
	i := 0

	for i < len(userInput) {
		// 空白文字をスキップ
		if unicode.IsSpace(rune(userInput[i])) {
			i++
			continue
		}

		if userInput[i] == '+' || userInput[i] == '-' {
			cur = newToken(TK_RESERVED, cur, userInput[i:i+1])
			cur.pos = i
			i++
			continue
		}

		if unicode.IsDigit(rune(userInput[i])) {
			j := i
			for j < len(userInput) && unicode.IsDigit(rune(userInput[j])) {
				j++
			}
			val, _ := strconv.Atoi(userInput[i:j])
			cur = newToken(TK_NUM, cur, userInput[i:j])
			cur.val = val
			cur.pos = i
			i = j
			continue
		}

		errorAt(i, "トークナイズできません")
	}

	newToken(TK_EOF, cur, "")
	return head.next
}

func main() {
	if len(os.Args) != 2 {
		errorf("引数の個数が正しくありません")
	}

	userInput = os.Args[1]
	token = tokenize()

	// アセンブリの前半部分を出力
	fmt.Println(".intel_syntax noprefix")
	fmt.Println(".globl main")
	fmt.Println("main:")

	// 式の最初は数でなければならないので、それをチェックして
	// 最初のmov命令を出力
	fmt.Printf("  mov rax, %d\n", expectNumber())

	// `+ <数>`あるいは`- <数>`というトークンの並びを消費しつつ
	// アセンブリを出力
	for !atEOF() {
		if consume('+') {
			fmt.Printf("  add rax, %d\n", expectNumber())
			continue
		}
		expect('-')
		fmt.Printf("  sub rax, %d\n", expectNumber())
	}

	fmt.Println("  ret")
}
