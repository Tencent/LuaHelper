package annotateparser

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/annotation/annotatelexer"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
	//"strings"
)

// ParseCommentFragment 解析代码注释片段
// frament 是多行的字符串
func ParseCommentFragment(commentInfo *lexer.CommentInfo) (fragment annotateast.AnnotateFragment,
	parseErrVec []annotatelexer.ParseAnnotateErr) {
	// 多行内容整体是一个字符串，拆分成多行数据
	for _, commentLine := range commentInfo.LineVec {
		l := annotatelexer.CreateAnnotateLexer(&commentLine.Str, commentLine.Line, commentLine.Col)

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

		_, flag := annotateState.(*annotateast.AnnotateNotValidState)
		if flag {
			continue
		}

		fragment.Lines = append(fragment.Lines, commentLine.Line)
		fragment.Stats = append(fragment.Stats, annotateState)
	}

	return fragment, parseErrVec
}

// TestComment 注解功能的测试代码
func TestComment() {
	/*strTest1 := "-@type string"
	comment1 := ParseCommentFragment(&strTest1)
	print(len(comment1.Stats))

	strTest2 := "-@type string, number"
	comment2 := ParseCommentFragment(&strTest2)
	print(len(comment2.Stats))

	strTest3 := "-@type string, number, number"
	comment3 := ParseCommentFragment(&strTest3)
	print(len(comment3.Stats))

	strTest4 := "-@type string[]"
	comment4 := ParseCommentFragment(&strTest4)
	print(len(comment4.Stats))

	strTest5 := "-@type string[], number"
	comment5 := ParseCommentFragment(&strTest5)
	print(len(comment5.Stats))

	strTest6 := "-@type table < number , number>"
	comment6 := ParseCommentFragment(&strTest6)
	print(len(comment6.Stats))

	strTest7 := "-@type    fun ( one:string, two: number):one, two"
	comment7 := ParseCommentFragment(&strTest7)
	print(len(comment7.Stats))

	strTest8 := "-@type string | number[]"
	comment8 := ParseCommentFragment(&strTest8)
	print(len(comment8.Stats))

	strTest9 := "-@type (string | number)[dd]"
	comment9 := ParseCommentFragment(&strTest9)
	print(len(comment9.Stats))

	strTest10 := "-@alias Handler fun(type: string | number, data: any):void"
	comment10 := ParseCommentFragment(&strTest10)
	print(len(comment10.Stats))

	strTest11 := "-@class A:B, C @fjsofjsofjo"
	comment11 := ParseCommentFragment(&strTest11)
	print(len(comment11.Stats))

	strTest12 := "-@overload fun(list:table, sep:string, i:number):string|number, bb[]"
	comment12 := ParseCommentFragment(&strTest12)
	print(len(comment12.Stats))

	strTest13 := "-@field public ssss fun(list:table, sep:string, i:number):string|number, bb[] sfsdfsdf"
	comment13 := ParseCommentFragment(&strTest13)
	print(len(comment13.Stats))
	// ---@field [public|protected|private] field_name FIELDLTYPE[|OTHER_TYPE] [@comment]

	strTest14 := "-@param string | fun(list:table, sep:string, i:number):string|number, bb[] sfsdfsdf"
	comment14 := ParseCommentFragment(&strTest14)
	print(len(comment14.Stats))

	strTest15 := "-@return (string | fun(list:table, sep:string, i:number):string|number, bb[]), string sfsdfsdf"
	comment15 := ParseCommentFragment(&strTest15)
	print(len(comment15.Stats))

	strTest16 := "-@generic T : Transport, K sfsdfsdf asfdjosf"
	comment16 := ParseCommentFragment(&strTest16)
	print(len(comment16.Stats))

	strTest17 := "-@vararg string"
	comment17 := ParseCommentFragment(&strTest17)
	print(len(comment17.Stats))
	*/
}

// ParserLine 解析一行注释
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
	}

	return &annotateast.AnnotateNotValidState{}
}
