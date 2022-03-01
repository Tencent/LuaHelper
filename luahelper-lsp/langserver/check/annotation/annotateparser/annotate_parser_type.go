package annotateparser

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/annotation/annotatelexer"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"strings"
)

// 继续单个的类型，不包含|
func parserSingleType(l *annotatelexer.AnnotateLexer) annotateast.Type {
	lookHeardKind := l.LookAheadKind()
	beginLoc := l.GetNowLoc()

	var subType annotateast.Type
	if lookHeardKind == annotatelexer.ATokenVSepLparen {
		// 为左括号，用于优先级的
		// 删除掉对应的左括号
		l.NextTokenOfKind(annotatelexer.ATokenVSepLparen)

		// 需要递归调用
		subType = parserOneType(l)

		// 删除掉匹配的右括号
		l.NextTokenOfKind(annotatelexer.ATokenVSepRparen)
	} else if lookHeardKind == annotatelexer.ATokenKwFun {
		// 如果为特殊的fun定义类型
		subType = parserFunType(l)
	} else if lookHeardKind == annotatelexer.ATokenKwTable {
		subType = parserTableType(l)
	} else if lookHeardKind == annotatelexer.ATokenKwIdentifier {
		// 为其他的标识符
		nameStr := l.NextTypeIdentifier()
		subType = &annotateast.NormalType{
			StrName:   nameStr,
			NameLoc:   l.GetNowLoc(),
			ShowColor: true,
		}

		l.SetLastNormalTypeLoc(l.GetNowLoc())
	} else if lookHeardKind == annotatelexer.ATokenVararg {
		l.NextToken()
		subType = &annotateast.NormalType{
			StrName:   "...",
			NameLoc:   l.GetNowLoc(),
			ShowColor: true,
		}

		l.SetLastNormalTypeLoc(l.GetNowLoc())
	} else if lookHeardKind == annotatelexer.ATokenString {
		tokenStr := l.GetHeardTokenStr()
		splitStr, quetesFlag := splitStrQuotes(tokenStr)
		l.NextToken()
		subType = &annotateast.ConstType{
			Name:       splitStr,
			Loc:        l.GetNowLoc(),
			QuotesFlag: quetesFlag,
			Comment:    "",
		}
	} else {
		// 无效的
		l.ErrorPrint(annotatelexer.AErrorType, annotatelexer.ATokenEOF, "not find annotate type")
	}

	// 有可能是列表类型，例如 string[]
	if l.LookAheadKind() == annotatelexer.ATokenVSepLbrack {
		// 删除掉对应的[
		l.NextTokenOfKind(annotatelexer.ATokenVSepLbrack)
		l.NextTokenOfKind(annotatelexer.ATokenVSepRbrack)

		endLoc := l.GetNowLoc()
		arrayType := &annotateast.ArrayType{
			ItemType: subType,
			Loc:      lexer.GetRangeLoc(&beginLoc, &endLoc),
		}

		return arrayType
	}

	return subType
}

// 解析最复杂的单个类型，包含 |
func parserOneType(l *annotatelexer.AnnotateLexer) annotateast.Type {
	multiType := &annotateast.MultiType{}
	beginLoc := l.GetHeardLoc()
	for {
		// 解析单个的type
		subType := parserSingleType(l)
		multiType.TypeList = append(multiType.TypeList, subType)

		strEnumFlag := false
		if l.LookAheadKind() != annotatelexer.ATokenBor {
			// 解析完了
			break
		}
		l.NextTokenOfKind(annotatelexer.ATokenBor)

		/*for {
			// 发现是 | 表示几种类型都可以，或者的关系
			l.NextTokenOfKind(annotatelexer.ATokenBor)
			normalType, ok := subType.(*annotateast.NormalType)
			if !ok || normalType.StrName != "string" {
				break
			}

			// 这里为字符串枚举
			// ---@type  string | '"r"' | '"w"' | '"a"' | '"r+"' | '"w+"' | '"a+"' | '"rb"' | '"wb"' | '"ab"' | '"rb+"' | '"wb+"' | '"ab+"'
			if l.LookAheadKind() != annotatelexer.ATokenString {
				break
			}

			strEnumFlag = true
			l.NextTokenOfKind(annotatelexer.ATokenString)
			if l.LookAheadKind() == annotatelexer.ATokenBor {
				continue
			}

			break
		}*/

		if strEnumFlag {
			break
		}
	}
	endLoc := l.GetNowLoc()
	multiType.Loc = lexer.GetRangeLoc(&beginLoc, &endLoc)

	return multiType
}

// 解析定义的fun 函数
func parserFunType(l *annotatelexer.AnnotateLexer) annotateast.Type {
	// fun(param1:PARAM_TYPE1 [,param2:PARAM_TYPE2]): RETURN_TYPE1[, RETURN_TYPE2]

	// 1) 首先移除掉fun关键字
	l.NextTokenOfKind(annotatelexer.ATokenKwFun)
	beginLoc := l.GetNowLoc()

	// 2) 判断是否为fun<V, K> 包含泛型的
	if l.LookAheadKind() == annotatelexer.ATokenLt {
		l.NextTokenOfKind(annotatelexer.ATokenLt)
		for {
			// 2.1) 解析引用的一个泛型名称
			l.NextFieldName()
			if l.LookAheadKind() == annotatelexer.ATokenSepComma {
				l.NextTokenOfKind(annotatelexer.ATokenSepComma)
				continue
			}

			break
		}
		l.NextTokenOfKind(annotatelexer.ATokenGt)
	}

	// 2) 移除掉函数的左括号
	l.NextTokenOfKind(annotatelexer.ATokenVSepLparen)

	// 定义需要返回的funType
	funType := &annotateast.FuncType{}
	funType.FunLoc = beginLoc

	// 处理函数的参数
	lookHeardKind := l.LookAheadKind()
	if lookHeardKind == annotatelexer.ATokenVSepRparen {
		// 直接匹配到了右侧的), 函数没有参数
	} else {
		// 3) 继续函数的所有参数
		for {
			// 3.1) 解析一个函数参数名称
			paramName := l.NextParamName()
			funType.ParamNameList = append(funType.ParamNameList, paramName)
			funType.ParamNameLocList = append(funType.ParamNameLocList, l.GetNowLoc())

			// 判断前面的是否为 ：,如果是则需要解析类型
			lookHeardKind1 := l.LookAheadKind()
			if lookHeardKind1 == annotatelexer.ATokenOption {
				funType.ParamOptionList = append(funType.ParamOptionList, true)
				l.NextToken()
				lookHeardKind1 = l.LookAheadKind()
			} else {
				funType.ParamOptionList = append(funType.ParamOptionList, false)
			}

			if lookHeardKind1 == annotatelexer.ATokenSepColon {
				// 3.2) 解析 ：
				l.NextTokenOfKind(annotatelexer.ATokenSepColon)

				// 3.3) 解析参数的类型Type
				paramType := parserOneType(l)
				funType.ParamTypeList = append(funType.ParamTypeList, paramType)
			} else {
				// 3.4) 不包含:, 这个参数的类型默认为any
				var paramType annotateast.Type = &annotateast.NormalType{
					StrName:   "any",
					NameLoc:   l.GetNowLoc(),
					ShowColor: false,
				}
				funType.ParamTypeList = append(funType.ParamTypeList, paramType)
			}

			if l.LookAheadKind() == annotatelexer.ATokenSepComma {
				// 如果匹配到 , 表示有多个参数，继续解析其他的参数
				l.NextToken()
			} else {
				break
			}
		}
	}

	// 4) 移除掉函数的右括号
	l.NextTokenOfKind(annotatelexer.ATokenVSepRparen)

	// 5) 解析函数的返回值
	if l.LookAheadKind() == annotatelexer.ATokenSepColon {
		// 有函数的返回值, 继续所有的返回值
		l.NextToken()

		for {
			// 5.1) 解析一个函数返回值
			returnType := parserOneType(l)
			funType.ReturnTypeList = append(funType.ReturnTypeList, returnType)

			if l.LookAheadKind() == annotatelexer.ATokenSepComma {
				// 如果匹配到 , 表示有多个返回值，继续解析其他的返回值
				l.NextToken()
			} else {
				break
			}
		}
	}

	endLoc := l.GetNowLoc()
	funType.Loc = lexer.GetRangeLoc(&beginLoc, &endLoc)
	return funType
}

// 解析table 类型
//---@type table<KEY_TYPE[|KEY_OTHER_TYPE], VALUE_TYPE[|VAULE_OTHER_TYPE]>
//---@type table (table有可能没有key或value)
func parserTableType(l *annotatelexer.AnnotateLexer) annotateast.Type {
	// 1) 首先移除掉table关键字
	l.NextTokenOfKind(annotatelexer.ATokenKwTable)
	beginLoc := l.GetNowLoc()
	tableType := &annotateast.TableType{
		TableStrLoc: beginLoc,
		EmptyFlag:   false,
	}

	// 判断是否是空的table，里面不包含key、value
	if l.LookAheadKind() != annotatelexer.ATokenLt {
		// 空的table
		tableType.EmptyFlag = true
		endLoc := l.GetNowLoc()
		tableType.Loc = lexer.GetRangeLoc(&beginLoc, &endLoc)
		return tableType
	}

	// 2) 移除 < 符号
	l.NextTokenOfKind(annotatelexer.ATokenLt)

	// 3) 解析key值的类型
	keyType := parserOneType(l)
	tableType.KeyType = keyType

	// 4) 解析间隔的逗号
	l.NextTokenOfKind(annotatelexer.ATokenSepComma)

	// 5) 解析value的类型
	valueType := parserOneType(l)
	tableType.ValueType = valueType

	// 6) 移除 > 符号
	l.NextTokenOfKind(annotatelexer.ATokenGt)
	endLoc := l.GetNowLoc()
	tableType.Loc = lexer.GetRangeLoc(&beginLoc, &endLoc)

	return tableType
}

// ParserAliasLine 额外解析alias换行的，例如这样的：---| '"w"' # Write data to this program by `file`.
func parserExtraAliasLine(l *annotatelexer.AnnotateLexer) (constType *annotateast.ConstType) {
	aheadKind := l.LookAheadKind()
	if aheadKind != annotatelexer.ATokenString {
		return
	}

	tokenStr := l.GetHeardTokenStr()
	splitStr, quetesFlag := splitStrQuotes(tokenStr)
	l.NextToken()

	loc := l.GetNowLoc()
	strComment, _ := l.GetRemainComment()
	strComment = strings.TrimLeft(strComment, " ")
	strComment = strings.TrimPrefix(strComment, "#")
	strComment = strings.TrimPrefix(strComment, " ")

	constType = &annotateast.ConstType{
		Name:       splitStr,
		Loc:        loc,
		QuotesFlag: quetesFlag,
		Comment:    strComment,
	}
	return constType
}
