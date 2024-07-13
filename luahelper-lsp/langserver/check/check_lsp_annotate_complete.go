package check

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/annotation/annotatelexer"
	"luahelper-lsp/langserver/check/annotation/annotateparser"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"os/user"
	"strings"
	"time"
)

// 这个文件为因为引入了注解系统，带来的代码补全

// CompleteAnnotateArea ---@ 注解的提示后面的域
func (a *AllProject) CompleteAnnotateArea() {
	// 1) class
	detail := "class struct"
	document := "---@class TYPE[: PARENT_TYPE {, PARENT_TYPE}]  [@comment]"
	document += "\n\n" + "sample:\n---@class People @People class"
	a.completeCache.InsertCompleteNormal("class", detail, document, common.IKAnnotateClass)

	// 2) field
	detail = "class field"
	document = "---@field [public|protected|private] field_name FIELD_TYPE {|OTHER_TYPE} [@comment]"
	document += "\n\n" + "sample:\n---@field age number @age attr is number"
	a.completeCache.InsertCompleteNormal("field", detail, document, common.IKAnnotateClass)

	// 3) type
	detail = "type"
	document = "---@type TYPE{|OTHER_TYPE} [@comment]"
	document += "\n\n" + "sample:\n---@type string @type is string"
	a.completeCache.InsertCompleteNormal("type", detail, document, common.IKAnnotateClass)

	// 4) param
	detail = "param"
	document = "---@param param_name TYPE{|OTHER_TYPE} [@comment]"
	document += "\n\n" + "sample:\n---@param param1 string @param1 is string"
	a.completeCache.InsertCompleteNormal("param", detail, document, common.IKAnnotateClass)

	// 5) return
	detail = "return"
	document = "---@return TYPE{|OTHER_TYPE} [@comment]"
	document += "\n\n" + "sample:\n---@return string @return is string"
	a.completeCache.InsertCompleteNormal("return", detail, document, common.IKAnnotateClass)

	// 6) alias
	detail = "alias"
	document = "---@alias new_type TYPE{|OTHER_TYPE}"
	document += "\n\n" + "sample:\n---@alias Man People @Man is People"
	a.completeCache.InsertCompleteNormal("alias", detail, document, common.IKAnnotateClass)

	// 7) generic
	detail = "generic"
	document = "---@generic T1 [: PARENT_TYPE]"
	document += "\n\n" + "sample:\n---@generic T1 @T1 is generic"
	a.completeCache.InsertCompleteNormal("generic", detail, document, common.IKAnnotateClass)

	// 8) overload
	detail = "overload"
	document = "---@overload fun(param_name : PARAM_TYPE) : RETURN_TYPE"
	document += "\n\n" + "sample:\n---@overload fun(param1 : string) : number"
	a.completeCache.InsertCompleteNormal("overload", detail, document, common.IKAnnotateClass)

	detail = "vararg"
	document = "---@vararg TYPE{|OTHER_TYPE} [@comment]"
	document += "\n\n" + "sample:\n---@vararg number"
	a.completeCache.InsertCompleteNormal("vararg", detail, document, common.IKAnnotateClass)

	// 9) author
	// 插入用户与时间
	userName := ""
	u, err := user.Current()
	if err == nil {
		userName = u.Name
	}
	timeNow := time.Now()
	timeString := timeNow.Format("2006-01-02 15:04:05")
	detail = "author: " + userName + " " + timeString
	a.completeCache.InsertCompleteNormal(detail, detail, document, common.IKAnnotateClass)

	// 10) enum
	detail = "enum start @"
	document = "---@enum start"
	document += "\n\n" + "sample:\n---@enum start @enum start"
	a.completeCache.InsertCompleteNormal(detail, detail, document, common.IKAnnotateClass)

	detail = "enum end"
	document = "---@enum end"
	document += "\n\n" + "sample:\n---@enum end"
	a.completeCache.InsertCompleteNormal(detail, detail, document, common.IKAnnotateClass)
}

// 获取注解输入param时候，提示所有的函数参数名
func (a *AllProject) getFuncParamTypeComplete(strFile string, posLine int) {
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		return
	}

	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		return
	}

	beginLine := posLine + 1
	findFlag := false
	for {
		if fileStruct.FileResult.JugetLineInFragement(beginLine) {
			beginLine = beginLine + 1
			findFlag = true
		} else {
			break
		}
	}

	if !findFlag {
		return
	}

	// 1) 判断是否为for语句引入的参数
	strList := fileStruct.FileResult.GetForLineVarString(beginLine)
	forFlag := true

	// 2) 判断是否为函数的参数
	if len(strList) == 0 {
		funcInfo := fileStruct.FileResult.GetLineFuncInfo(beginLine)
		if funcInfo == nil {
			return
		}

		for _, oneParam := range funcInfo.ParamList {
			if funcInfo.IsColon && oneParam == "this" {
				continue
			}
			strList = append(strList, oneParam)
		}
		forFlag = false
	}

	// 判断这些函数参数是否已经出现过
	fragment := annotateFile.GetBestFragementInfo(beginLine - 2)
	if fragment == nil {
		fragment = annotateFile.GetBestFragementInfo(beginLine)
	}
	var strMap map[string]struct{} = map[string]struct{}{}
	if fragment != nil && fragment.ParamInfo != nil {
		for _, oneParam := range fragment.ParamInfo.ParamList {
			strMap[oneParam.Name] = struct{}{}
		}
	}

	for _, strName := range strList {
		if _, flag := strMap[strName]; flag {
			continue
		}

		if forFlag {
			a.completeCache.InsertCompleteNormal(strName, "for param", "", common.IKAnnotateClass)
		} else {
			a.completeCache.InsertCompleteNormal(strName, "function param", "", common.IKAnnotateClass)
		}
	}
}

// AnnotateTypeComplete 注解类型代码补全
func (a *AllProject) AnnotateTypeComplete(strFile string, strLine string, strWord string, posLine int) {
	l := annotatelexer.CreateAnnotateLexer(&strLine, 0, 0)

	// 判断这行内容是否以-@开头，是否合法
	if !l.CheckHeardValid() {
		return
	}

	// 后面的内容进行词法解析
	annotateState, parseErr := annotateparser.ParserLine(l)
	_, flag := annotateState.(*annotateast.AnnotateNotValidState)
	// 1) 判断是否为提示@后面的内容
	if flag && parseErr.ErrType == annotatelexer.AErrorOk {
		a.CompleteAnnotateArea()
		return
	}

	// 2) 判断是否要提示类型系统
	lastTypeLoc := l.GetLastNormalTypeLoc()
	log.Debug("last col=%d", lastTypeLoc.EndColumn)
	if lastTypeLoc.EndColumn == len(strLine) {
		// 提示所有的类型，先只获取这个文件的所有类型
		a.getTypeCompleteVecs(strFile, strWord, posLine)
		return
	}

	// 3) 根据错误类型，特殊提示 fun
	if parseErr.ErrType == annotatelexer.AErrorKind && parseErr.NeedKind == annotatelexer.ATokenKwFun {
		a.completeCache.InsertCompleteNormal("fun", "function", "", common.IKAnnotateClass)
		return
	}

	// 4) 判断是否要补充param 后面的名称
	if parseErr.ErrToken.GetTokenKind() == annotatelexer.ATokenKwParam {
		// 补全param后面的参数名
		a.getFuncParamTypeComplete(strFile, posLine)
		return
	}
}

// 获取补全注解系统的所有类型
func (a *AllProject) getTypeCompleteVecs(strFile, strWord string, posLine int) {
	// 1) 传入的strWord 通过.进行切词
	lastIndex := strings.LastIndex(strWord, ".")
	preStr := ""
	if lastIndex >= 0 {
		preStr = strWord[0 : lastIndex+1]
	}

	// 2) 这个文件对应代码段产生的generic 泛型也放入进来
	a.getGenericCompleteVecs(strFile, preStr, posLine+1)

	// 3) 当前文件的所有的class 或alias类型都放入进来
	annotateFile := a.getAnnotateFile(strFile)
	var showMap map[string]bool = map[string]bool{}
	if annotateFile != nil {
		for strName, typeList := range annotateFile.CreateTypeMap {
			showMap[strName] = true
			if preStr != "" {
				if strings.HasPrefix(strName, preStr) {
					strName = strings.TrimPrefix(strName, preStr)
				} else {
					continue
				}
			}

			for _, typeOne := range typeList.List {
				a.completeCache.InsertCompleteInnotateType(strName, typeOne)
			}
		}
	}

	// 4) 获取所有的class或alias类型
	for strName, typeList := range a.createTypeMap {
		if _, flag := showMap[strName]; flag {
			continue
		}

		if preStr != "" {
			if strings.HasPrefix(strName, preStr) {
				strName = strings.TrimPrefix(strName, preStr)
			} else {
				continue
			}
		}

		for _, typeOne := range typeList.List {
			a.completeCache.InsertCompleteInnotateType(strName, typeOne)
		}
	}

	// 5）最后插入fun
	a.completeCache.InsertCompleteNormal("fun", "function", "", common.IKAnnotateClass)
}

// 获取块注释段引入的generic泛型类型系统
func (a *AllProject) getGenericCompleteVecs(strFile, preStr string, posLine int) {
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		return
	}

	fragment := annotateFile.GetBestFragementInfo(posLine)
	if fragment == nil {
		return
	}

	genericInfo := fragment.GenericInfo
	if genericInfo == nil {
		return
	}

	for _, oneGeneric := range genericInfo.GenericInfoList {
		strName := oneGeneric.Name
		if preStr != "" {
			if strings.HasPrefix(strName, preStr) {
				strName = strings.TrimPrefix(strName, preStr)
			} else {
				continue
			}
		}

		a.completeCache.InsertCompleteNormal(strName, "generic", strFile, common.IKAnnotateClass)
	}
}

// oneClassInfo 转换为对应的completeVecs
func (a *AllProject) convertClassInfoToCompleteVecs(oneClass *common.OneClassInfo, colonFlag bool) {
	className := oneClass.ClassState.Name

	// 1) oneClass所有的成员
	for strName, fieldState := range oneClass.FieldMap {
		// 注解类型的优先级较高，如果存在重复的进行替换
		// if a.completeCache.ExistStr(strName) {
		// 	continue
		// }

		if colonFlag {
			// 检查下面的逻辑是否类似下面的，下面的含义表示ClassA有一个：的FunctionC 函数，隐藏的用法部分时候要去掉第一个self参数
			// ---@class ClassA
			// ---@field FunctionC fun(self:ClassA):void
			oneFuncType := annotateast.GetAllFuncType(fieldState.FiledType)
			if oneFuncType != nil {
				subFuncType, _ := oneFuncType.(*annotateast.FuncType)

				if subFuncType != nil && len(subFuncType.ParamNameList) > 0 && len(subFuncType.ParamTypeList) > 0 &&
					subFuncType.ParamNameList[0] == "self" && annotateast.TypeConvertStr(subFuncType.ParamTypeList[0]) == className {
					a.completeCache.InsertCompleteClassField(oneClass.LuaFile, strName, fieldState, annotateast.FieldColonHide)
					continue
				}
			}
		}

		// 注解: 不能调用.的成员
		if colonFlag && fieldState.FieldColonType != annotateast.FieldColonYes {
			continue
		}

		// 检查下面的逻辑是否类似下面的，下面的含义表示ClassA有一个：的FunctionC 函数, 这种写法暂时不支持.调用:的函数
		// ---@class ClassA
		// ---@field FunctionC : fun():void

		// 注解.也不能调用显示的:的成员
		if !colonFlag && fieldState.FieldColonType != annotateast.FieldColonNo {
			continue
		}

		a.completeCache.InsertCompleteClassField(oneClass.LuaFile, strName, fieldState, fieldState.FieldColonType)
	}

	// 2) oneClass关联的变量
	if oneClass.RelateVar != nil {
		a.getVarCompleteExt(oneClass.LuaFile, oneClass.RelateVar, colonFlag)
	}
}

// getVarInfoCompleteExt 获取变量关联的所有子成员信息，用于代码补全
// excludeMap 表示这次因为冒号语法剔除掉的字符串
// colonFlag 表示是否获取冒号成员
func (a *AllProject) getVarInfoCompleteExt(symbol *common.Symbol, colonFlag bool) {
	if symbol == nil {
		return
	}

	// 1) 判断注解开关是否有打开, 如果注解打开获取注解的里面的信息
	if symbol.AnnotateType != nil {
		classList := a.getAllNormalAnnotateClass(symbol.AnnotateType, symbol.FileName, symbol.GetLine())
		for _, oneClass := range classList {
			a.convertClassInfoToCompleteVecs(oneClass, colonFlag)
		}

		// 注释掉，同时补全注解类型与变量的类型
		// return
	}

	// 2) 没有注解的模式
	a.getVarCompleteExt(symbol.FileName, symbol.VarInfo, colonFlag)
}

// 获取函数的返回代码补全
func (a *AllProject) getFuncReturnCompleteExt(symbol *common.Symbol, colonFlag bool, comParam *CommonFuncParam) {
	if symbol == nil {
		return
	}

	// 判断注解类型是否存在
	// 首先获取变量是否直接注解为函数的返回
	if ok, _, typeList, _ := a.getFuncReturnAnnotateTypeList(symbol); ok {
		if len(typeList) > 0 {
			returnType := typeList[0]
			line := symbol.VarInfo.Loc.StartLine - 1

			classList := a.getAllNormalAnnotateClass(returnType, symbol.FileName, line)
			for _, oneClass := range classList {
				a.convertClassInfoToCompleteVecs(oneClass, colonFlag)
			}
		}
	} else {
		if symbol.AnnotateType != nil {
			// 包含注解类型，这个注解类型跟踪出函数的类型
			returnTypeList := a.getAllFuncAnnotateType(symbol.AnnotateType, symbol.FileName,
				symbol.AnnotateLine)
			for _, oneType := range returnTypeList {
				// 获取对应的函数返回类型
				funType, flag := oneType.(*annotateast.FuncType)
				if !flag {
					continue
				}

				if len(funType.ReturnTypeList) == 0 {
					continue
				}

				oneFunType := funType.ReturnTypeList[0]
				classList := a.getAllNormalAnnotateClass(oneFunType, symbol.FileName, symbol.AnnotateLine)
				for _, oneClass := range classList {
					a.convertClassInfoToCompleteVecs(oneClass, colonFlag)
				}
			}
		}
	}

	// 处理func(*).的提示，判断是否有指向普通的变量
	if symbol.VarInfo == nil || symbol.VarInfo.ReferFunc == nil {
		return
	}

	find, returnExp := symbol.VarInfo.ReferFunc.GetOneReturnExp()
	if !find {
		return
	}

	findExpList := []common.FindExpFile{}
	// 这里也需要做判断，函数返回的变量逐层跟踪，目前只跟踪了一层
	symList := a.FindDeepSymbolList(symbol.FileName, returnExp, comParam, &findExpList, true, 1)
	for _, subSym := range symList {
		a.getVarInfoCompleteExt(subSym, colonFlag)
	}
}
