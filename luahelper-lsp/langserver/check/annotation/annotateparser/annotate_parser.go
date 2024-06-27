package annotateparser

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/annotation/annotatelexer"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
)

// ParseCommentFragment 解析代码注释片段
// frament 是多行的字符串
func ParseCommentFragment(commentInfo *lexer.CommentInfo) (fragment annotateast.AnnotateFragment,
	parseErrVec []annotatelexer.ParseAnnotateErr) {
	// 多行内容整体是一个字符串，拆分成多行数据
	for _, commentLine := range commentInfo.LineVec {
		l := annotatelexer.CreateAnnotateLexer(&commentLine.Str, commentLine.Line, commentLine.Col)

		// 判断是否为特殊的alias换行的内容，例如为：
		// ---@alias one
		// ---| '"r"' # Read data from this program by `file`.
		// ---| '"w"' # Write data to this program by `file`.
		if l.CheckAliasHeadValid() {
			constType := parserExtraAliasLine(l)
			if constType == nil {
				continue
			}

			// 如果存在有效的换行alias内容，尝试补充到上一个alias state中
			if len(fragment.Stats) == 0 {
				continue
			}
			lastStat := fragment.Stats[len(fragment.Stats)-1]
			aliasState, ok := lastStat.(*annotateast.AnnotateAliasState)
			if !ok {
				continue
			}
			appendAliasState(aliasState, constType)
			continue
		}

		// 判断这行内容是否以-@开头，是否合法
		if !l.CheckHeardValid() {
			continue
		}

		// 后面的内容进行词法解析
		annotateState, parseErr := ParserLine(l)
		if parseErr.ErrType != annotatelexer.AErrorOk {
			parseErrVec = append(parseErrVec, parseErr)
			continue
		}

		if _, flag := annotateState.(*annotateast.AnnotateNotValidState); flag {
			continue
		}

		fragment.Lines = append(fragment.Lines, commentLine.Line)
		fragment.Stats = append(fragment.Stats, annotateState)
	}

	// 判断是否有空的AliasState，如果有清除掉
	clearEmpytAlias(&fragment)

	return fragment, parseErrVec
}

// constType常量补偿到aliasState中
func appendAliasState(aliasState *annotateast.AnnotateAliasState, constType *annotateast.ConstType) {
	if aliasState.AliasType == nil {
		// 如果之前没有alias其他的类型
		multiType := &annotateast.MultiType{}
		multiType.Loc = constType.Loc
		multiType.TypeList = append(multiType.TypeList, constType)
		aliasState.AliasType = multiType
		return
	}

	multiType, ok := aliasState.AliasType.(*annotateast.MultiType)
	if !ok {
		return
	}

	multiType.TypeList = append(multiType.TypeList, constType)
}

// 判断是否有空的AliasState，如果有清除掉
func clearEmpytAlias(fragment *annotateast.AnnotateFragment) {
	for i := 0; i < len(fragment.Stats); i++ {
		aliasState, ok := fragment.Stats[i].(*annotateast.AnnotateAliasState)
		if !ok {
			continue
		}

		if aliasState.AliasType == nil {
			fragment.Stats = append(fragment.Stats[:i], fragment.Stats[i+1:]...)
			i--
		}
	}
}

// ParserLine 正常解析一行注释
func ParserLine(l *annotatelexer.AnnotateLexer) (oneState annotateast.AnnotateState,
	parseErr annotatelexer.ParseAnnotateErr) {
	parseErr.ErrType = annotatelexer.AErrorOk
	defer func() {
		if err2 := recover(); err2 != nil {
			parseErr = err2.(annotatelexer.ParseAnnotateErr)
			log.Error("ParseAnnotateErr errStr=%s, errShow=%s", parseErr.ErrStr, parseErr.ShowStr)
			oneState = &annotateast.AnnotateNotValidState{}
		}
	}()

	oneState = parserOneState(l)
	return oneState, parseErr
}

func parserOneState(l *annotatelexer.AnnotateLexer) annotateast.AnnotateState {
	switch l.LookAheadKind() {
	case annotatelexer.ATokenKwType:
		return parserTypeState(l)
	case annotatelexer.ATokenKwAlias:
		return parserAliasState(l)
	case annotatelexer.ATokenKwClass:
		return parserClassState(l)
	case annotatelexer.ATokenKwOverload:
		return parserOverloadState(l)
	case annotatelexer.ATokenKwField:
		return parserFieldState(l)
	case annotatelexer.ATokenKwParam:
		return parserParamState(l)
	case annotatelexer.ATokenKwReturn:
		return parserReturnState(l)
	case annotatelexer.ATokenKwGeneric:
		return parserGenericState(l)
	case annotatelexer.ATokenKwVararg:
		return parserVarargState(l)
	case annotatelexer.ATokenKwEnum:
		return parserEnumState(l)
	}

	return &annotateast.AnnotateNotValidState{}
}
