package annotatelexer

import (
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// AnnotateErrType 错误的类型
type AnnotateErrType int

const (
	// AErrorOk 正常的
	AErrorOk AnnotateErrType = 0

	// AErrorToken 没有切出正确的切词
	AErrorToken AnnotateErrType = 1

	// AErrorKind 语法错误
	AErrorKind AnnotateErrType = 2

	// AErrorType 注解类型的错误
	AErrorType AnnotateErrType = 3

	// AErrorNormalIdentifier 普通标识符
	AErrorNormalIdentifier AnnotateErrType = 4

	// AErrorTypeIdentifier 类型注解标识符出错
	AErrorTypeIdentifier AnnotateErrType = 5
)

// ParseAnnotateErr check的错误信息
type ParseAnnotateErr struct {
	ErrType  AnnotateErrType // 错误的类型
	NeedKind ATokenType      // 当为AErrorKind类型不匹配的时候，保存需要的kind
	ErrStr   string          // 简单的错误信息
	ShowStr  string          // 完整的错误信息
	ErrToken AnnotateToken   // 出错的之前token
	NowToken AnnotateToken   // 当前的token
	ErrLoc   lexer.Location  // 出错的位置信息
}
