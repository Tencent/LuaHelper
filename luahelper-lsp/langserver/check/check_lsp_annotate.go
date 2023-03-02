package check

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"strings"
	"time"
)

// 引入注解系统，一些方法

//
func (a *AllProject) checkOneFileType(annotateFile *common.AnnotateFile, fragemnet *common.FragementInfo,
	oneType annotateast.Type) {
	strList, locList := annotateast.GetAllStrAndLocList(oneType)

	var genericMap map[string]struct{} = map[string]struct{}{}
	if fragemnet.GenericInfo != nil {
		for _, oneGeneric := range fragemnet.GenericInfo.GenericInfoList {
			genericMap[oneGeneric.Name] = struct{}{}
		}
	}

	for index, str := range strList {
		if common.GConfig.IsDefaultAnnotateType(str) {
			continue
		}

		if str == "..." {
			continue
		}

		if _, ok := genericMap[str]; ok {
			continue
		}

		if _, ok := a.createTypeMap[str]; ok {
			continue
		}

		errStr := "not define annotate type: " + str
		annotateFile.PushTypeDefineError(errStr, locList[index])
	}
}

// 单个注解文件进行check
func (a *AllProject) checkFileAnnotate(fileStruct *results.FileStruct) {
	annotateFile := fileStruct.AnnotateFile
	if annotateFile == nil {
		return
	}

	// 清除部分错误，常规的注解语法错误保留
	annotateFile.ClearCheckError()

	// 遍历所有的注解代码块
	for _, oneFragment := range annotateFile.FragementMap {
		if oneFragment.AliasInfo != nil {
			for _, oneAlias := range oneFragment.AliasInfo.AliasList {
				aliasType := oneAlias.AliasState.AliasType
				a.checkOneFileType(annotateFile, oneFragment, aliasType)
			}
		}

		if oneFragment.TypeInfo != nil {
			for _, oneType := range oneFragment.TypeInfo.TypeList {
				a.checkOneFileType(annotateFile, oneFragment, oneType)
			}
		}

		if oneFragment.ClassInfo != nil {
			for _, oneClass := range oneFragment.ClassInfo.ClassList {
				for _, oneField := range oneClass.FieldMap {
					a.checkOneFileType(annotateFile, oneFragment, oneField.FiledType)
				}
			}
		}

		if oneFragment.ParamInfo != nil {
			for _, oneParam := range oneFragment.ParamInfo.ParamList {
				a.checkOneFileType(annotateFile, oneFragment, oneParam.ParamType)
			}
		}

		if oneFragment.ReturnInfo != nil {
			for _, oneType := range oneFragment.ReturnInfo.ReturnTypeList {
				a.checkOneFileType(annotateFile, oneFragment, oneType)
			}
		}

		if oneFragment.OverloadInfo != nil {
			for _, oneLoad := range oneFragment.OverloadInfo.OverloadList {
				a.checkOneFileType(annotateFile, oneFragment, oneLoad.OverFunType)
			}
		}

		if oneFragment.VarargInfo != nil {
			oneVararg := oneFragment.VarargInfo.VarargInfo
			if oneVararg != nil {
				a.checkOneFileType(annotateFile, oneFragment, oneVararg.VarargType)
			}
		}
	}

	// 所有的错误告警信息，进行排序，因为在比对注解告警信息的时候，希望是有序的
	annotateFile.RankCheckError()
}

// 判断定义的注解类型是否重复
func (a *AllProject) checkCreateTypeListDuplicate(str string, createList common.CreateTypeList) {
	lenList := len(createList.List)
	if lenList <= 1 {
		return
	}

	for index, oneCreate := range createList.List {
		luaFile, loc := oneCreate.GetFileNameAndLoc()
		annotateFile := a.getNotCacheAnnotateFile(luaFile)
		if annotateFile == nil {
			continue
		}

		errStr := "duplicate annotate type: " + str
		var relateVec []common.RelateCheckInfo
		for i, nextCreate := range createList.List {
			if index == i {
				continue
			}

			relateLuaFile, relateLoc := nextCreate.GetFileNameAndLoc()
			relateCheck := common.RelateCheckInfo{
				LuaFile: relateLuaFile,
				ErrStr:  errStr,
				Loc:     relateLoc,
			}

			relateVec = append(relateVec, relateCheck)
		}

		annotateFile.PushTypeDuplicateError(errStr, loc, relateVec)
	}
}

// 根据文件名称，获取到文件的注解文件，没有缓存出
func (a *AllProject) getNotCacheAnnotateFile(strFile string) (annotateFile *common.AnnotateFile) {
	// 1）先查找该文件是否存在
	fileStruct := a.fileStructMap[strFile]
	if fileStruct == nil {
		log.Error("getAnnotateFile error, not find file=%s", strFile)
		return
	}

	// 2) 获取文件对应的annotateFile
	annotateFile = fileStruct.AnnotateFile
	if annotateFile == nil {
		log.Error("getAnnotateFile annotateFile is nil, file=%s", strFile)
	}

	return annotateFile
}

// 检查所有的注解类型系统，进行告警
// 告警主要分为两方面：1）使用的type类型是否有注解定义。2）使用的注解type是否重复
func (a *AllProject) checkAllAnnotate() {
	if len(a.fileStructMap) == 0 {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorAnnotate) {
		return
	}

	time1 := time.Now()

	// 1) 首先统一校验是type是否有定义
	for _, fileStruct := range a.fileStructMap {
		a.checkFileAnnotate(fileStruct)
	}

	// 2) 再次校验是否出现了重复的type
	for str, createList := range a.createTypeMap {
		a.checkCreateTypeListDuplicate(str, createList)
	}

	ftime := time.Since(time1).Milliseconds()
	log.Debug("checkAllAnnotate time:%d", ftime)
}

// 根据文件名称，获取到文件的注释结构
func (a *AllProject) getAnnotateFile(strFile string) (annotateFile *common.AnnotateFile) {
	// 1）先查找该文件是否存在
	fileStruct, _ := a.GetCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("getAnnotateFile error, not find file=%s", strFile)
		return
	}

	// 2) 获取文件对应的annotateFile
	annotateFile = fileStruct.AnnotateFile
	if annotateFile == nil {
		log.Error("getAnnotateFile annotateFile is nil, file=%s", strFile)
	}

	return annotateFile
}

// 判断类是否有该成员 默认返回true
func (a *AllProject) IsFieldOfClass(className string, fieldName string) bool {

	// 以下是从所属类成员找
	if className == "" {
		return true
	}

	createTypeList, flag := a.createTypeMap[className]
	if !flag || len(createTypeList.List) != 1 {
		return true
	}

	ci := createTypeList.List[0].ClassInfo
	if ci == nil {
		return true
	}

	_, ok := ci.FieldMap[fieldName]
	return ok
}

// 查找成员函数的返回值类型
func (a *AllProject) GetFuncReturnTypeByClass(className string, funcName string) (retVec [][]string) {

	// 以下是从所属类成员找
	if className == "" || funcName == "" {
		return
	}

	createTypeList, flag := a.createTypeMap[className]
	if !flag || len(createTypeList.List) != 1 {
		return
	}

	ci := createTypeList.List[0].ClassInfo
	if ci == nil {
		return
	}

	ann, ok := ci.FieldMap[funcName]
	if !ok {
		return
	}

	//一路解析下去
	multiType, ok := ann.FiledType.(*annotateast.MultiType)
	if !ok || len(multiType.TypeList) != 1 {
		return
	}

	funcType, ok := multiType.TypeList[0].(*annotateast.FuncType)
	if len(funcType.ParamNameList) != len(funcType.ParamTypeList) {
		return
	}

	for _, v := range funcType.ReturnTypeList {
		oneRetType := []string{}

		mt, ok := v.(*annotateast.MultiType)
		if !ok {
			// 这里要补空的
			retVec = append(retVec, oneRetType)
			continue
		}

		for _, typeone := range mt.TypeList {
			nt, ok := typeone.(*annotateast.NormalType)
			if !ok {
				continue
			}
			oneRetType = append(oneRetType, nt.StrName)
		}
		retVec = append(retVec, oneRetType)
	}

	return
}

// 查找成员函数的参数类型
func (a *AllProject) GetFuncParamTypeByClass(className string, funcName string) (retMap map[string][]string) {
	retMap = make(map[string][]string)

	// 以下是从所属类成员找
	if className == "" || funcName == "" {
		return
	}

	createTypeList, flag := a.createTypeMap[className]
	if !flag || len(createTypeList.List) != 1 {
		return
	}

	ci := createTypeList.List[0].ClassInfo
	if ci == nil {
		return
	}

	ann, ok := ci.FieldMap[funcName]
	if !ok {
		return
	}

	//一路解析下去
	multiType, ok := ann.FiledType.(*annotateast.MultiType)
	if !ok || len(multiType.TypeList) != 1 {
		return
	}

	funcType, ok := multiType.TypeList[0].(*annotateast.FuncType)
	if len(funcType.ParamNameList) != len(funcType.ParamTypeList) {
		return
	}

	for i, v := range funcType.ParamNameList {
		mt, ok := funcType.ParamTypeList[i].(*annotateast.MultiType)
		if !ok {
			continue
		}

		retMap[v] = []string{}
		for _, typeone := range mt.TypeList {
			nt, ok := typeone.(*annotateast.NormalType)
			if !ok {
				continue
			}
			retMap[v] = append(retMap[v], nt.StrName)
		}
	}

	return
}

// GetFuncParamInfo 获取前面注释行的所有参数返回
func (a *AllProject) GetFuncParamInfo(fileName string, lastLine int) (paramInfo *common.FragementParamInfo) {
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		return
	}

	// 2) 获取注解文件指定行号的注释块信息
	fragmentInfo := annotateFile.GetLineFragementInfo(lastLine)
	if fragmentInfo == nil {
		return
	}

	// 3) 判断是否有函数返回信息
	return fragmentInfo.ParamInfo
}

//  获取前面注释行的所有返回值
func (a *AllProject) GetFuncReturnInfo(fileName string, lastLine int) (paramInfo *common.FragementReturnInfo) {
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		return
	}

	// 2) 获取注解文件指定行号的注释块信息
	fragmentInfo := annotateFile.GetLineFragementInfo(lastLine)
	if fragmentInfo == nil {
		return
	}

	// 3) 判断是否有函数返回信息
	return fragmentInfo.ReturnInfo
}

// GetAstTypeFuncType 获取注解astType具体的指向的注解函数
func (a *AllProject) GetAstTypeFuncType(astType annotateast.Type, fileName string,
	lastLine int) (funcType *annotateast.FuncType) {
	if astType == nil {
		return
	}

	// 1) 首先判断是否为直接的注解函数类型
	// 例如下面的
	//---@type fun(a:number):D
	oneFuncType := annotateast.GetAllFuncType(astType)
	if oneFuncType != nil {
		subFuncType, _ := oneFuncType.(*annotateast.FuncType)
		return subFuncType
	}

	// 2)
	repeatTypeList := &common.CreateTypeList{
		List: []*common.CreateTypeInfo{},
	}

	// 2) 因此判断某一个简单的字符串进行处理
	strMap := map[string]bool{}
	funcType = a.getInlineAllNormalFuncType(astType, fileName, lastLine, repeatTypeList, strMap)
	return funcType
}

func (a *AllProject) getInlineAllNormalFuncType(astType annotateast.Type, fileName string,
	lastLine int, repeatTypeList *common.CreateTypeList, strMap map[string]bool) (funcType *annotateast.FuncType) {

	strSimpleList := annotateast.GetAllNormalStrList(astType)
	if len(strSimpleList) == 0 {
		return
	}

	// 2) 因此判断某一个简单的字符串进行处理
	for _, strSimple := range strSimpleList {
		// 2.1) 判断这个是否处理重复了
		if _, ok := strMap[strSimple]; ok {
			continue
		}

		strMap[strSimple] = true

		// 2.2) 这个字符串去查找到对应的OneClassInfo list
		funcType = a.getClassTypeInfoFuncType(strSimple, fileName, lastLine, repeatTypeList, strMap)
		if funcType != nil {
			break
		}
	}

	return funcType
}

// 通过字符名查找到ClassTypeInfo 列表
// strName 为字符串的名称
// fileName 为strName所在的lua文件名
// lastLine 为strName所在的lua文件行号
// repeatTypeList 主要是为了防止重复，记录所有已经查找的
func (a *AllProject) getClassTypeInfoFuncType(strName string, fileName string, lastLine int,
	repeatTypeList *common.CreateTypeList, strMap map[string]bool) (funcType *annotateast.FuncType) {
	// 1) 优先在当前文件查找，如果没有找到，再找全局的文件注释
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", fileName)
		return
	}

	// 1.1) 这个文件查找这个符号
	createBestType := annotateFile.GetBestCreateTypeInfo(strName, lastLine)
	if createBestType != nil {
		// 1.2) 判重
		if repeatTypeList.IsRepeateTypeInfo(createBestType) {
			// 如果有重复的
			return
		}

		repeatTypeList.List = append(repeatTypeList.List, createBestType)

		// 1.4) 判断是否为OneAliasInfo
		if createBestType.AliasInfo != nil {
			aliasType := createBestType.AliasInfo.AliasState.AliasType
			oneFuncType := annotateast.GetAllFuncType(aliasType)
			if oneFuncType != nil {
				subFuncType, _ := oneFuncType.(*annotateast.FuncType)
				return subFuncType
			}

			funcType = a.getInlineAllNormalFuncType(aliasType, fileName, createBestType.LastLine,
				repeatTypeList, strMap)
		}
		return
	}

	// 2.1) 获取所有的全局信息
	createTypeList, flag := a.createTypeMap[strName]
	if !flag || len(createTypeList.List) == 0 {
		return
	}

	for _, oneCreate := range createTypeList.List {
		// 2.2) 判重
		if repeatTypeList.IsRepeateTypeInfo(oneCreate) {
			// 如果有重复的
			continue
		}

		repeatTypeList.List = append(repeatTypeList.List, oneCreate)

		if oneCreate.AliasInfo != nil {
			aliasType := oneCreate.AliasInfo.AliasState.AliasType
			oneFuncType := annotateast.GetAllFuncType(aliasType)
			if oneFuncType != nil {
				subFuncType, _ := oneFuncType.(*annotateast.FuncType)
				return subFuncType
			}

			funcType = a.getInlineAllNormalFuncType(aliasType, fileName, oneCreate.LastLine, repeatTypeList, strMap)
			if funcType != nil {
				break
			}
		}
	}

	return funcType
}

// 对外的接口
// 通过传人的类型，获取所有的ClassInfo 列表，返回的是一个列表, 包括所有的父类型也统一放在列表中
// astType 传人的为类型
// fileName 为注解的文件名
// lastLine 为注解出现的行号
func (a *AllProject) getAllNormalAnnotateClass(astType annotateast.Type, fileName string,
	lastLine int) (classList []*common.OneClassInfo) {
	// 1) 对应的类型转换为简单的字符串，可能多有个字符串，类型是或者的关系
	if astType == nil {
		return
	}

	strSimpleList := annotateast.GetAllNormalStrList(astType)
	if len(strSimpleList) == 0 {
		return
	}

	repeatTypeList := &common.CreateTypeList{
		List: []*common.CreateTypeInfo{},
	}

	// 2) 因此判断某一个简单的字符串进行处理
	strMap := map[string]bool{}
	strMap["any"] = true
	classList = a.getInLineAllNormalAnnotateClass(astType, fileName, lastLine, repeatTypeList, strMap)
	return classList
}

// 对内的接口
func (a *AllProject) getInLineAllNormalAnnotateClass(astType annotateast.Type, fileName string,
	lastLine int, repeatTypeList *common.CreateTypeList, strMap map[string]bool) (classList []*common.OneClassInfo) {
	// 1) 对应的类型转换为简单的字符串，可能多有个字符串，类型是或者的关系
	if astType == nil {
		return
	}

	strSimpleList := annotateast.GetAllNormalStrList(astType)
	if len(strSimpleList) == 0 {
		return
	}

	// 2) 因此判断某一个简单的字符串进行处理
	for _, strSimple := range strSimpleList {
		// 2.1) 判断这个是否处理重复了
		if _, ok := strMap[strSimple]; ok {
			continue
		}

		strMap[strSimple] = true

		// 2.2) 这个字符串去查找到对应的OneClassInfo list
		classListTmp := a.getClassTypeInfoList(strSimple, fileName, lastLine, repeatTypeList, strMap)
		classList = append(classList, classListTmp...)
	}

	return classList
}

// 通过字符名查找到ClassTypeInfo 列表
// strName 为字符串的名称
// fileName 为strName所在的lua文件名
// lastLine 为strName所在的lua文件行号
// repeatTypeList 主要是为了防止重复，记录所有已经查找的
func (a *AllProject) getClassTypeInfoList(strName string, fileName string, lastLine int,
	repeatTypeList *common.CreateTypeList, strMap map[string]bool) (classList []*common.OneClassInfo) {
	// 1) 优先在当前文件查找，如果没有找到，再找全局的文件注释
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", fileName)
		return
	}

	// 1.1) 这个文件查找这个符号
	createBestType := annotateFile.GetBestCreateTypeInfo(strName, lastLine)
	if createBestType != nil {
		// 1.2) 判重
		if repeatTypeList.IsRepeateTypeInfo(createBestType) {
			// 如果有重复的
			return classList
		}

		repeatTypeList.List = append(repeatTypeList.List, createBestType)

		// 1.3) 查找到了，判断是否有ClassInfo
		if createBestType.ClassInfo != nil {
			classList = append(classList, createBestType.ClassInfo)

			// 所有父类型也处理下，再次递归获取
			for _, strParent := range createBestType.ClassInfo.ClassState.ParentNameList {
				if _, ok := strMap[strParent]; ok {
					continue
				}

				if strName == strParent {
					continue
				}

				classListTmp := a.getClassTypeInfoList(strParent, fileName, createBestType.LastLine,
					repeatTypeList, strMap)
				classList = append(classList, classListTmp...)
			}
		}

		// 1.4) 判断是否为OneAliasInfo
		if createBestType.AliasInfo != nil {
			aliasType := createBestType.AliasInfo.AliasState.AliasType
			classListTmp := a.getInLineAllNormalAnnotateClass(aliasType, fileName, createBestType.LastLine,
				repeatTypeList, strMap)
			classList = append(classList, classListTmp...)
		}

		return
	}

	// 2.1) 获取所有的全局信息
	createTypeList, flag := a.createTypeMap[strName]
	if !flag || len(createTypeList.List) == 0 {
		return
	}

	for _, oneCreate := range createTypeList.List {
		// 2.2) 判重
		if repeatTypeList.IsRepeateTypeInfo(oneCreate) {
			// 如果有重复的
			continue
		}

		repeatTypeList.List = append(repeatTypeList.List, oneCreate)

		if oneCreate.ClassInfo != nil {
			classInfo := oneCreate.ClassInfo
			classList = append(classList, classInfo)

			// 所有父类型也处理下，再次递归获取
			for _, strParent := range classInfo.ClassState.ParentNameList {
				if _, ok := strMap[strParent]; ok {
					continue
				}

				if strName == strParent {
					continue
				}

				classListTmp := a.getClassTypeInfoList(strParent, classInfo.LuaFile,
					oneCreate.LastLine, repeatTypeList, strMap)
				classList = append(classList, classListTmp...)
			}
		}

		if oneCreate.AliasInfo != nil {
			aliasInfo := oneCreate.AliasInfo
			aliasType := aliasInfo.AliasState.AliasType
			classListTmp := a.getInLineAllNormalAnnotateClass(aliasType, aliasInfo.LuaFile,
				oneCreate.LastLine, repeatTypeList, strMap)
			classList = append(classList, classListTmp...)
		}
	}

	return classList
}

// 追踪出所以的函数类型
// 通过传人的类型，获取所有的ClassInfo 列表，返回的是一个列表, 包括所有的父类型也统一放在列表中
// astType 传人的为类型
// fileName 为注解的文件名
// lastLine 为注解出现的行号
func (a *AllProject) getAllFuncAnnotateType(astType annotateast.Type, fileName string,
	lastLine int) (typeList []annotateast.Type) {
	// 1) 对应的类型转换为简单的字符串，可能多有个字符串，类型是或者的关系
	if astType == nil {
		return
	}

	repeatTypeList := &common.CreateTypeList{
		List: []*common.CreateTypeInfo{},
	}

	// 2) 调用内部的函数，里面又包含了多层次的递归处理
	strMap := map[string]bool{}
	typeList = a.getInlineAllFuncAnnotateType(astType, fileName, lastLine, repeatTypeList, strMap)
	return
}

// 获取函数泛型的返回，如果有泛型的返回，需要推导其关联的值
func (a *AllProject) getFuncGenericVarInfo(oldSymbol *common.Symbol, fragment *common.FragementInfo,
	funcAnnotateType annotateast.Type, node *ast.FuncCallExp, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile) (findSymbol *common.Symbol) {
	if oldSymbol.VarInfo == nil {
		return
	}

	funcInfo := oldSymbol.VarInfo.ReferFunc
	if funcInfo == nil {
		return
	}

	if node == nil || fragment == nil || fragment.GenericInfo == nil || fragment.ParamInfo == nil {
		return
	}

	normalStrList := annotateast.GetAllNormalStrList(funcAnnotateType)
	for _, oneStr := range normalStrList {
		// 判断这个字符串是否有关联到泛型
		flag := false
		for _, genericInfo := range fragment.GenericInfo.GenericInfoList {
			if oneStr == genericInfo.Name {
				flag = true
				continue
			}
		}

		// 没有匹配到这个泛型
		if !flag {
			continue
		}

		// 看哪个参数有匹配到这个泛型，指考虑下面这种简单的类型
		// ---@generic AA
		// ---@param one AA
		// ---@return AA

		// 复杂的这种，暂时不考虑
		// ---@generic AA
		// ---@param one AA[]
		// ---@return AA

		// 一个个参数的匹配
		// node.Args
		var findIndex int = -1
		for _, oneParam := range fragment.ParamInfo.ParamList {
			strParamNormalType := annotateast.TraverseOneType(oneParam.ParamType)
			if strParamNormalType != oneStr {
				continue
			}

			for index, strParam := range funcInfo.ParamList {
				if strParam == oneParam.Name {
					findIndex = index
					break
				}
			}

			if findIndex >= 0 {
				break
			}
		}

		if findIndex < 0 {
			continue
		}

		// 找到这个findIndex对应的参数表达式
		if findIndex >= len(node.Args) {
			log.Error("findIndex =%d error, len=%d", findIndex, len(node.Args))
			break
		}

		paramExp := node.Args[findIndex]
		paramVarFile := a.FindVarReferSymbol(oldSymbol.FileName, paramExp, comParam, findExpList, 1)
		if paramVarFile != nil {
			return paramVarFile
		}
	}

	return findSymbol
}

// 获取函数注释块是否有return语句, 如果有return语句，获取对应的type
func (a *AllProject) getFuncReturnOneType(oldSymbol *common.Symbol, varIndex uint8, node *ast.FuncCallExp,
	comParam *CommonFuncParam, findExpList *[]common.FindExpFile) (flag bool, symbol *common.Symbol) {
	// 判断注解类型是否存在
	// 首先获取变量是否直接注解为函数的返回
	flag, fragment, typeList, _ := a.getFuncReturnAnnotateTypeList(oldSymbol)
	if !flag {
		return
	}

	if len(typeList) < (int)(varIndex) {
		return
	}

	oneFunType := typeList[varIndex-1]
	// 尝试获取泛型的推导
	genericVarFile := a.getFuncGenericVarInfo(oldSymbol, fragment, oneFunType, node, comParam, findExpList)
	if genericVarFile != nil {
		return flag, genericVarFile
	}

	symbol = &common.Symbol{
		FileName:     oldSymbol.FileName, // todo这里的文件名不太准确
		VarInfo:      nil,
		AnnotateType: oneFunType,
		VarFlag:      common.FirstAnnotateFlag,            // 默认还是先获取的变量
		AnnotateLine: oldSymbol.VarInfo.Loc.StartLine - 1, // todo这里的行号不太准确
	}

	return flag, symbol
}

// 获取一个函数返回的类型
// varIndex 表示获取第一个函数返回值
func (a *AllProject) getFuncIndexReturnSymbol(oldSymbol *common.Symbol, varIndex uint8,
	node *ast.FuncCallExp, comParam *CommonFuncParam, findExpList *[]common.FindExpFile) (flag bool,
	symbol *common.Symbol) {
	// 1) 首先判断注解地方，是否有return返回值
	flag, symbol = a.getFuncReturnOneType(oldSymbol, varIndex, node, comParam, findExpList)
	if flag {
		return
	}

	// 2) 根据是否有@type 注解类型指向函数的
	astType := oldSymbol.AnnotateType
	fileName := oldSymbol.FileName
	lastLine := oldSymbol.GetLine()
	returnTypeList := a.getAllFuncAnnotateType(astType, fileName, lastLine)
	for _, oneType := range returnTypeList {
		// 获取对应的函数返回类型
		funType, flag1 := oneType.(*annotateast.FuncType)
		if !flag1 {
			continue
		}

		flag = true
		if len(funType.ReturnTypeList) < (int)(varIndex) {
			continue
		}

		oneReturnType := funType.ReturnTypeList[varIndex-1]
		symbol = &common.Symbol{
			FileName:     fileName, // todo这里的文件名不太准确
			VarInfo:      oldSymbol.VarInfo,
			AnnotateType: oneReturnType,
			VarFlag:      common.FirstAnnotateFlag,
			AnnotateLine: lastLine, // todo这里的行号不太准确
		}

		return
	}

	return
}

// 对内的函数的
// 追踪出所以的函数类型
// 通过传人的类型，获取所有的ClassInfo 列表，返回的是一个列表, 包括所有的父类型也统一放在列表中
// astType 传人的为类型
// fileName 为注解的文件名
// lastLine 为注解出现的行号
func (a *AllProject) getInlineAllFuncAnnotateType(astType annotateast.Type, fileName string,
	lastLine int, repeatTypeList *common.CreateTypeList, strMap map[string]bool) (typeList []annotateast.Type) {
	// 1) 对应的类型转换为简单的字符串，可能多有个字符串，类型是或者的关系
	if astType == nil {
		return
	}

	// 1) 首先判断是否关联的一个简单的函数类型
	funcType := annotateast.GetAllFuncType(astType)
	if funcType != nil {
		// 函数的类型不会nil，获取所有的函数类型
		typeList = append(typeList, funcType)
		return
	}

	// 2) 判断是否关联到的alias，如果是alias那是一个简单的类型
	strSimpleList := annotateast.GetAllNormalStrList(astType)
	if len(strSimpleList) == 0 {
		return
	}

	// 2) 因此判断某一个简单的字符串进行处理
	for _, strSimple := range strSimpleList {
		// 2.1) 判断这个是否处理重复了
		if _, ok := strMap[strSimple]; ok {
			continue
		}

		strMap[strSimple] = true

		// 2.2) 这个字符串去查找到对应的OneClassInfo list
		typeListTmp := a.getFuncTypeInfoList(strSimple, fileName, lastLine, repeatTypeList, strMap)
		typeList = append(typeList, typeListTmp...)
	}

	return
}

// 通过字符名查找到类型，直到是一个指向函数的变量
// strName 为字符串的名称
// fileName 为strName所在的lua文件名
// lastLine 为strName所在的lua文件行号
// repeatTypeList 主要是为了防止重复，记录所有已经查找的
func (a *AllProject) getFuncTypeInfoList(strName string, fileName string, lastLine int,
	repeatTypeList *common.CreateTypeList, strMap map[string]bool) (typeList []annotateast.Type) {
	// 1) 优先在当前文件查找，如果没有找到，再找全局的文件注释
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", fileName)
		return
	}

	// 1.1) 这个文件查找这个符号
	createBestType := annotateFile.GetBestCreateTypeInfo(strName, lastLine)
	if createBestType != nil {
		// 1.2) 判重
		if repeatTypeList.IsRepeateTypeInfo(createBestType) {
			// 如果有重复的
			return typeList
		}

		repeatTypeList.List = append(repeatTypeList.List, createBestType)

		// 1.3) 判断是否为OneAliasInfo
		if createBestType.AliasInfo != nil {
			aliasInfo := createBestType.AliasInfo
			aliasType := aliasInfo.AliasState.AliasType
			typeListTmp := a.getInlineAllFuncAnnotateType(aliasType, fileName, createBestType.LastLine, repeatTypeList,
				strMap)
			typeList = append(typeList, typeListTmp...)
		}

		return typeList
	}

	// 2.1) 获取所有的全局信息
	createTypeList, flag := a.createTypeMap[strName]
	if !flag || len(createTypeList.List) == 0 {
		return
	}

	for _, oneCreate := range createTypeList.List {
		// 2.2) 判重
		if repeatTypeList.IsRepeateTypeInfo(oneCreate) {
			// 如果有重复的
			continue
		}

		repeatTypeList.List = append(repeatTypeList.List, oneCreate)

		if oneCreate.AliasInfo != nil {
			aliasInfo := createBestType.AliasInfo
			aliasType := aliasInfo.AliasState.AliasType
			typeListTmp := a.getInlineAllFuncAnnotateType(aliasType, aliasInfo.LuaFile, createBestType.LastLine,
				repeatTypeList, strMap)
			typeList = append(typeList, typeListTmp...)
		}
	}

	return typeList
}

// 获取注释是否为函数的返回字段列表
// flag 表示是否获取的为一个函数返回
// astTypeList 为所有的函数返回字段列表，函数可能有多个返回参数
func (a *AllProject) getFuncReturnAnnotateTypeList(symbol *common.Symbol) (flag bool,
	fragment *common.FragementInfo, astTypeList []annotateast.Type, commentList []string) {
	if symbol == nil || symbol.VarInfo == nil {
		return
	}

	strFile := symbol.FileName

	// 1) 获取文件对应的annotateFile
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		log.Error("getFuncReturnAnnotateTypeList annotateFile is nil, file=%s", strFile)
		return
	}

	// 2) 获取注解文件指定行号的注释块信息
	fragmentInfo := annotateFile.GetLineFragementInfo(symbol.VarInfo.Loc.StartLine - 1)
	if fragmentInfo == nil {
		return
	}

	// 3) 判断是否有函数返回信息
	if fragmentInfo.ReturnInfo != nil {
		flag = true
		return flag, fragmentInfo, fragmentInfo.ReturnInfo.ReturnTypeList, fragmentInfo.ReturnInfo.CommentList
	}

	return
}

// ast表达式关联到注解类型
func (a *AllProject) astConvertToAnnotateType(luaFile string, varInfo *common.VarInfo, exp ast.Exp) (astType annotateast.Type) {
	comParam := a.getCommFunc(luaFile, varInfo.Loc.StartLine, varInfo.Loc.StartColumn)
	if comParam == nil {
		return
	}

	findExpList := []common.FindExpFile{}
	varIndex := varInfo.VarIndex

	// 下面的例子为这样的, 对getCatsMap()函数返回值进行追踪时，默认追踪第一个参数
	// for catId, catData in pairs(getCatsMap()) do
	//
	// end
	if _, ok := exp.(*ast.FuncCallExp); ok {
		varIndex = 1
	}
	// 这里也需要做判断，函数返回的变量逐层跟踪，目前只跟踪了一层
	symList := a.FindDeepSymbolList(luaFile, exp, comParam, &findExpList, false, varIndex)
	for _, oneSymbol := range symList {
		// todo 这里的strName默认为""
		getAnnotateType, _, _ := a.getInfoFileAnnotateType("", oneSymbol)
		if getAnnotateType != nil {
			return getAnnotateType
		}
	}

	return
}

// GetAllTableKeyType 获取Type对应的TableType的key类型
// todo 这里后面还需要重新组装
func (a *AllProject) GetAllTableKeyType(strFile string, astType annotateast.Type) (arrayType annotateast.Type) {
	switch subAst := astType.(type) {
	case *annotateast.MultiType:
		if len(subAst.TypeList) == 0 {
			return
		}

		// 有多种类型，是或者的关系，因此获取多个
		for _, oneType := range subAst.TypeList {
			getType := a.GetAllTableKeyType(strFile, oneType)
			if getType != nil {
				return getType
			}
		}

		return
	case *annotateast.TableType:
		if !subAst.EmptyFlag {
			// 非空的情况，才能取到value的类型
			return subAst.KeyType
		}

		return

	case *annotateast.NormalType:
		// 如果是普通类型，判断是否为alias
		// 1) 获取文件对应的annotateFile
		annotateFile := a.getAnnotateFile(strFile)
		if annotateFile == nil {
			log.Error("annotateFile is nil, file=%s", strFile)
			return
		}

		//  判断对应的是否包含alias信息
		createTypeList1, ok1 := annotateFile.CreateTypeMap[subAst.StrName]

		var typeList []*common.CreateTypeInfo
		if ok1 {
			typeList = createTypeList1.List
		} else {
			createTypeList2, ok2 := a.createTypeMap[subAst.StrName]
			if ok2 {
				typeList = createTypeList2.List
			}
		}

		for _, createTypeInfo := range typeList {
			if createTypeInfo.AliasInfo != nil {
				aliasType := createTypeInfo.AliasInfo.AliasState.AliasType
				return a.GetAllTableKeyType(createTypeInfo.AliasInfo.LuaFile, aliasType)
			}
		}

		return
	}

	return
}

// GetAllTableType 获取Type对应的TableType的value类型
// todo 这里后面还需要重新组装
func (a *AllProject) GetAllTableType(strFile string, astType annotateast.Type) (arrayType annotateast.Type) {
	switch subAst := astType.(type) {
	case *annotateast.MultiType:
		if len(subAst.TypeList) == 0 {
			return
		}

		// 有多种类型，是或者的关系，因此获取多个
		for _, oneType := range subAst.TypeList {
			getType := a.GetAllTableType(strFile, oneType)
			if getType != nil {
				return getType
			}
		}

		return
	case *annotateast.TableType:
		if !subAst.EmptyFlag {
			// 非空的情况，才能取到value的类型
			return subAst.ValueType
		}

		return

	case *annotateast.NormalType:
		// 如果是普通类型，判断是否为alias
		// 1) 获取文件对应的annotateFile
		annotateFile := a.getAnnotateFile(strFile)
		if annotateFile == nil {
			log.Error("annotateFile is nil, file=%s", strFile)
			return
		}

		//  判断对应的是否包含alias信息
		createTypeList1, ok1 := annotateFile.CreateTypeMap[subAst.StrName]

		var typeList []*common.CreateTypeInfo
		if ok1 {
			typeList = createTypeList1.List
		} else {
			createTypeList2, ok2 := a.createTypeMap[subAst.StrName]
			if ok2 {
				typeList = createTypeList2.List
			}
		}

		for _, createTypeInfo := range typeList {
			if createTypeInfo.AliasInfo != nil {
				aliasType := createTypeInfo.AliasInfo.AliasState.AliasType
				return a.GetAllTableType(createTypeInfo.AliasInfo.LuaFile, aliasType)
			}
		}

		return
	}

	return
}

// GetAllArrayType 获取Type对应的ArrayType 类型
// todo 这里后面还需要重新组装
func (a *AllProject) GetAllArrayType(strFile string, astType annotateast.Type) (arrayType annotateast.Type) {
	switch subAst := astType.(type) {
	case *annotateast.MultiType:
		if len(subAst.TypeList) == 0 {
			return
		}

		// 有多种类型，是或者的关系，因此获取多个
		for _, oneType := range subAst.TypeList {
			getType := a.GetAllArrayType(strFile, oneType)
			if getType != nil {
				return getType
			}
		}

		return
	case *annotateast.ArrayType:
		return subAst.ItemType

	case *annotateast.NormalType:
		// 如果是普通类型，判断是否为alias
		// 1) 获取文件对应的annotateFile
		annotateFile := a.getAnnotateFile(strFile)
		if annotateFile == nil {
			log.Error("annotateFile is nil, file=%s", strFile)
			return
		}

		//  判断对应的是否包含alias信息
		createTypeList1, ok1 := annotateFile.CreateTypeMap[subAst.StrName]

		var typeList []*common.CreateTypeInfo
		if ok1 {
			typeList = createTypeList1.List
		} else {
			createTypeList2, ok2 := a.createTypeMap[subAst.StrName]
			if ok2 {
				typeList = createTypeList2.List
			}
		}

		for _, createTypeInfo := range typeList {
			if createTypeInfo.AliasInfo != nil {
				aliasType := createTypeInfo.AliasInfo.AliasState.AliasType
				return a.GetAllArrayType(createTypeInfo.AliasInfo.LuaFile, aliasType)
			}
		}

		return
	}

	return
}

func (a *AllProject) getForCycleAnnotateType(luaFile string, varInfo *common.VarInfo) (astType annotateast.Type,
	strComment string) {
	forCycle := varInfo.ForCycle
	if forCycle == nil {
		return
	}

	forExpType := a.astConvertToAnnotateType(luaFile, varInfo, forCycle.Exp)

	// for ipairs 的key值，为整型
	if forCycle.IpairsFlag {
		if varInfo.VarIndex == 1 {
			normalAst := &annotateast.NormalType{
				StrName: "number",
			}

			return normalAst, ""
		}

		// 下面的都是为value
		valueAst := a.GetAllArrayType(luaFile, forExpType)
		if valueAst != nil {
			return valueAst, ""
		}

		valueTableAst := a.GetAllTableType(luaFile, forExpType)
		if valueTableAst != nil {
			return valueTableAst, ""
		}

		return
	}

	// 下面的为for pairs
	if varInfo.VarIndex == 1 {
		// 为key值， 先获取key
		valueAst := a.GetAllArrayType(luaFile, forExpType)
		if valueAst != nil {
			normalAst := &annotateast.NormalType{
				StrName: "number",
			}

			return normalAst, ""
		}

		// 获取table的key
		valueKeyAst := a.GetAllTableKeyType(luaFile, forExpType)
		if valueKeyAst != nil {
			return valueKeyAst, ""
		}
	} else {
		// 为value值，先获取array的
		valueAst := a.GetAllArrayType(luaFile, forExpType)
		if valueAst != nil {
			return valueAst, ""
		}

		// 再获取table的value
		valueTableAst := a.GetAllTableType(luaFile, forExpType)
		if valueTableAst != nil {
			return valueTableAst, ""
		}
	}

	return
}

// 查找变量直接关联的类型以及注解带来的注释
// strName 为变量的名称
// strPreComment 为注解引入的前置说明
func (a *AllProject) getInfoFileAnnotateType(strName string, symbol *common.Symbol) (astType annotateast.Type,
	strComment string, strPreComment string) {
	if symbol == nil {
		return
	}

	// 0) 首先判断是否包含注解类型
	if symbol.AnnotateType != nil {
		return symbol.AnnotateType, symbol.AnnotateComment, symbol.StrPreComment
	}

	strFile := symbol.FileName

	// 1) 获取文件对应的annotateFile
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", strFile)
		return
	}

	// 2) 获取注解文件指定行号的注释块信息
	fragmentInfo := annotateFile.GetLineFragementInfo(symbol.VarInfo.Loc.StartLine - 1)

	if fragmentInfo == nil {
		// lua变量初始赋值的时候为nil，后面可能会再赋值，再赋值有注解
		// 2.1) 获取执行出exp
		referExp := symbol.VarInfo.ReferExp
		referLoc := common.GetExpLoc(referExp)
		fragmentInfo = annotateFile.GetLineFragementInfo(referLoc.StartLine - 1)
		if fragmentInfo != nil {
			log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", strFile)
		}
	}

	if fragmentInfo == nil {
		// 判断是否关联了for pairs或ipairs里的变量
		if symbol.VarInfo.IsForParam {
			astType, strComment = a.getForCycleAnnotateType(symbol.FileName, symbol.VarInfo)
			return astType, strComment, "param"
		}

		return
	}

	// 3) 判断是否为函数参数的注解
	// 函数参数的注解，会额外的用下面的注解类型
	// ---@param one string
	// todo 这里如果为函数参数的，函数的注释一定要在前面一行，如果函数的参数占用两行，整个会有问题
	if symbol.VarInfo.IsParam || symbol.VarInfo.IsForParam {
		astType, strComment = fragmentInfo.GetParamTypeInfo(strName)
		if astType != nil {
			return astType, strComment, "param"
		}
	}

	// 再次判断是否为for 函数的参数
	if fragmentInfo.TypeInfo == nil && symbol.VarInfo.IsForParam {
		astType, strComment = a.getForCycleAnnotateType(symbol.FileName, symbol.VarInfo)
		return astType, strComment, "param"
	}

	// 4.1) 查找对应的注解类型@type
	if fragmentInfo.TypeInfo != nil {
		varIndex := (int)(symbol.VarInfo.VarIndex)
		if varIndex == 0 {
			varIndex = 1
		}

		for index, oneType := range fragmentInfo.TypeInfo.TypeList {
			if (index + 1) >= varIndex {
				return oneType, fragmentInfo.TypeInfo.CommentList[index], "type"
			}
		}

		return
	}

	// 4.2) 再判断变量是否有class域
	if fragmentInfo.ClassInfo != nil {
		oneClass := fragmentInfo.GetFirstOneClassInfo()
		if oneClass == nil {
			return
		}

		normalType := &annotateast.NormalType{
			StrName: oneClass.ClassState.Name,
			NameLoc: oneClass.ClassState.NameLoc,
		}

		astType = normalType
		return astType, oneClass.ClassState.Comment, "class"
	}

	return
}

// 设置变量关联的注解类型
func (a *AllProject) setInfoFileAnnotateTypes(strName string, symbol *common.Symbol) {
	astType, strComment, strPreComment := a.getInfoFileAnnotateType(strName, symbol)
	if astType == nil {
		return
	}

	symbol.AnnotateComment = strComment
	symbol.StrPreComment = strPreComment
	symbol.AnnotateLine = symbol.VarInfo.Loc.StartLine - 1
	symbol.AnnotateType = astType
	symbol.AnnotateLoc = annotateast.GetAstTypeLoc(astType)
}

// 判断函数的一个注解返回值，是否包含子key
func (a *AllProject) getAnnotateFunStrKey(symbol *common.Symbol, strKey string, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile) (flag bool, subSymbol *common.Symbol) {
	// 如果找到了，判断是否有注解类型
	// 获取对应的函数注释返回值
	flag1, varFuncFile := a.getFuncIndexReturnSymbol(symbol, 1, nil, comParam, findExpList)
	flag = flag1
	if varFuncFile == nil {
		return
	}

	subSymbol = a.symbolHasSubKey(varFuncFile, strKey, comParam, findExpList)
	return
}

// findFlag 返回值0表示没有包含的，1表示有包含，2表示需要递归判断 lastExp
func (a *AllProject) getVarInfoFuncHasKey(lastSymbol *common.Symbol, strKey string, comParam *CommonFuncParam,
	findExpList *[]common.FindExpFile) (findFlag int, subSymbol *common.Symbol, lastExp ast.Exp,
	luaFile string) {
	hasFlag, findSymbol := a.getAnnotateFunStrKey(lastSymbol, strKey, comParam, findExpList)
	if hasFlag {
		findFlag = 0
		// 通过注解找到了信息
		if findSymbol != nil {
			findFlag = 1
		}
		return findFlag, findSymbol, nil, ""
	}

	// 没有原始的变量，直接返回
	if lastSymbol.VarInfo == nil {
		return 0, nil, nil, ""
	}

	// 如果是函数的返回，判断是否指向了函数返回
	if lastSymbol.VarInfo.ReferFunc != nil {
		// 首先直接判断是否为指向了函数
		flag, expReturn := lastSymbol.VarInfo.ReferFunc.GetOneReturnExp()
		if !flag {
			return 0, nil, nil, ""
		}

		findFlag = 2
		lastExp = expReturn
		luaFile = lastSymbol.FileName

		return findFlag, nil, expReturn, luaFile
	}

	// lastSymbol.VarInfo 没有直接指向函数，判断它关联的变量是否指向了函数
	tmpList := a.FindDeepSymbolList(lastSymbol.FileName, lastSymbol.VarInfo.ReferExp, comParam,
		findExpList, false, lastSymbol.VarInfo.VarIndex)
	for _, oneSymbol := range tmpList {
		// 如果找到了，判断是否有注解类型
		if ok, findSymbol := a.getAnnotateFunStrKey(oneSymbol, strKey, comParam, findExpList); ok {
			findFlag = 0
			// 通过注解找到了信息
			if findSymbol != nil {
				findFlag = 1
			}
			return findFlag, findSymbol, nil, ""
		}

		if oneSymbol.VarInfo == nil || oneSymbol.VarInfo.ReferFunc == nil {
			continue
		}

		flag, expReturn := oneSymbol.VarInfo.ReferFunc.GetOneReturnExp()
		if !flag {
			continue
		}

		// 关联到了
		findFlag = 2
		lastExp = expReturn
		luaFile = oneSymbol.FileName

		return findFlag, nil, expReturn, luaFile
	}

	return 0, nil, nil, ""
}

// 查找变量关联的深层次引用关系，所有引用的变量都跟踪出来
func (a *AllProject) getDeepVarList(oldSymbol *common.Symbol, varStruct *common.DefineVarStruct,
	comParam *CommonFuncParam) (symList []*common.Symbol) {
	// 1) 只有一层，不用查找子成员，开始追踪引用变量
	findExpList := []common.FindExpFile{}
	if len(varStruct.StrVec) == 1 {
		symList = append(symList, oldSymbol)
		tmpList := a.FindDeepSymbolList(oldSymbol.FileName, oldSymbol.VarInfo.ReferExp, comParam, &findExpList,
			false, oldSymbol.VarInfo.VarIndex)
		symList = append(symList, tmpList...)
		return symList
	}

	// 2.1) 有多层，逐层查找传人的子成员
	lastSymbol := oldSymbol
	for i := 0; i < len(varStruct.StrVec)-1; i++ {
		// 判断这个的值是否为函数的返回。例如字符串为：func().a， 那么第一个值是函数返回，第二个值为否
		funcFlag := false
		if len(varStruct.IsFuncVec) > i {
			funcFlag = varStruct.IsFuncVec[i]
		}

		var lastExp ast.Exp    // 最终查找关联到的表达式
		var lastLuaFile = ""   // 最终查找关联到的表达式所在的文件
		var varIndex uint8 = 1 // 多次申明变量时候，变量的index

		strKey := varStruct.StrVec[i+1]
		if funcFlag {
			// 是函数的查找
			findFlag, subSymbol, exp, luaFile := a.getVarInfoFuncHasKey(lastSymbol, strKey, comParam, &findExpList)
			if findFlag == 0 {
				// 没有找到
				return symList
			}

			if findFlag == 1 {
				// 找到了
				lastSymbol = subSymbol
				continue
			}

			// 需要继续关联表达式查找
			lastExp = exp
			lastLuaFile = luaFile
			varIndex = 1
		} else {
			subSymbol := a.symbolHasSubKey(lastSymbol, strKey, comParam, &findExpList)
			if subSymbol != nil {
				// 如果找到了变量，优先推导注解，然后推导普通的table内的成员
				if subSymbol.VarFlag == common.FirstAnnotateFlag {
					// 如果已经是注解了，直接返回
					lastSymbol = subSymbol
					continue
				}

				// 找到了变量的普通推导类型，尝试向上找注解，优先匹配注解的
				if lastSymbol.VarInfo == nil {
					lastSymbol = subSymbol
					continue
				}

				lastExp = lastSymbol.VarInfo.ReferExp
				lastLuaFile = lastSymbol.FileName
				varIndex = lastSymbol.VarInfo.VarIndex

				lastSymbol = subSymbol

				// 没有找到，那么这个变量，它关联的上一层变量呢
				// 递归进行查找
				var secondSymbol *common.Symbol = nil
				tmpList := a.FindDeepSymbolList(lastLuaFile, lastExp, comParam, &findExpList, false, varIndex)
				for _, oneSymbol := range tmpList {
					if subSymbol := a.symbolHasSubKey(oneSymbol, strKey, comParam, &findExpList); subSymbol != nil {
						secondSymbol = subSymbol
						break
					}
				}

				// 先判断第二次是否有找到，且是否为注解，如果是注解优先返回注解
				if secondSymbol != nil && secondSymbol.VarFlag == common.FirstAnnotateFlag {
					lastSymbol = secondSymbol
				}
				continue
			}

			if lastSymbol.VarInfo == nil {
				return symList
			}

			lastExp = lastSymbol.VarInfo.ReferExp
			lastLuaFile = lastSymbol.FileName
			varIndex = lastSymbol.VarInfo.VarIndex
		}

		// 没有找到，那么这个变量，它关联的上一层变量呢
		// 递归进行查找
		subFindFlag := false
		tmpList := a.FindDeepSymbolList(lastLuaFile, lastExp, comParam, &findExpList, true, varIndex)
		for _, oneSymbol := range tmpList {
			if subSymbol := a.symbolHasSubKey(oneSymbol, strKey, comParam, &findExpList); subSymbol != nil {
				lastSymbol = subSymbol
				subFindFlag = true
				break
			}
		}

		// 最终也是没有找到
		if !subFindFlag {
			return symList
		}
	}

	// 2.2) 有多层，追踪最后一层的引用变量
	symList = append(symList, lastSymbol)
	if lastSymbol.VarInfo == nil {
		return symList
	}

	tmpList := a.FindDeepSymbolList(lastSymbol.FileName, lastSymbol.VarInfo.ReferExp, comParam,
		&findExpList, true, lastSymbol.VarInfo.VarIndex)
	symList = append(symList, tmpList...)
	return symList
}

// GetAnnotateFileSymbolStruct 获取文件的注解类型信息, 用于显示当前文档的符号
func (a *AllProject) GetAnnotateFileSymbolStruct(fileName string) (symbolVec []common.FileSymbolStruct) {
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		return
	}

	for strName, typeList := range annotateFile.CreateTypeMap {
		for _, one := range typeList.List {
			if one.AliasInfo != nil {
				oneSymbol := common.FileSymbolStruct{
					Name: strName,
					Kind: common.IKAnnotateAlias,
					Loc:  one.AliasInfo.AliasState.NameLoc,
				}

				symbolVec = append(symbolVec, oneSymbol)
			}

			if one.ClassInfo != nil {
				oneSymbol := common.FileSymbolStruct{
					Name: strName,
					Kind: common.IKAnnotateClass,
					Loc:  one.ClassInfo.ClassState.NameLoc,
				}

				symbolVec = append(symbolVec, oneSymbol)
			}
		}
	}

	return
}

// typeStrFile 类型对应的文件信息，根据一个类型名查找定义时候用到
type typeStrFile struct {
	FileName  string         // lua所在的文件名
	Loc       lexer.Location //位置信息
	StrDetail string         // 额外的详细描述
}

// 判断是否为泛型的
func (a *AllProject) getStrNameGenericLocVec(strName string, fileName string, lastLine int) (list []typeStrFile) {
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", fileName)
		return
	}

	// 2) 判断是否有对应的generic符号
	fragment := annotateFile.GetBestFragementInfo(lastLine)
	if fragment == nil {
		return
	}

	genericInfo := fragment.GenericInfo
	if genericInfo == nil {
		return
	}

	for _, oneGeneric := range genericInfo.GenericInfoList {
		if oneGeneric.Name != strName {
			continue
		}

		oneFile := typeStrFile{
			FileName:  fileName,
			Loc:       oneGeneric.NameLoc,
			StrDetail: "generic info",
		}

		list = append(list, oneFile)
	}

	return list
}

// 通过字符名查找到所有的定义关系
func (a *AllProject) getStrNameDefineLocVec(strName string, fileName string, lastLine int) (list []typeStrFile) {
	// 1) 优先在当前文件查找，如果没有找到，再找全局的文件注释
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", fileName)
		return
	}

	// 2) 首先判断是否对应的generic泛型的名称
	list = a.getStrNameGenericLocVec(strName, fileName, lastLine)
	if len(list) > 0 {
		return list
	}

	// 1.1) 这个文件查找这个符号
	createBestType := annotateFile.GetBestCreateTypeInfo(strName, lastLine)
	if createBestType != nil {
		// 1.4) 判断是否为OneAliasInfo
		if createBestType.AliasInfo != nil {
			oneFile := typeStrFile{
				FileName:  fileName,
				Loc:       createBestType.AliasInfo.AliasState.NameLoc,
				StrDetail: "alias info",
			}

			list = append(list, oneFile)
		}

		if createBestType.ClassInfo != nil {
			oneFile := typeStrFile{
				FileName:  fileName,
				Loc:       createBestType.ClassInfo.ClassState.NameLoc,
				StrDetail: "class info",
			}

			list = append(list, oneFile)
		}
		return
	}

	// 2.1) 获取所有的全局信息
	createTypeList, flag := a.createTypeMap[strName]
	if !flag || len(createTypeList.List) == 0 {
		return
	}

	for _, oneCreate := range createTypeList.List {
		// 1.4) 判断是否为OneAliasInfo
		if oneCreate.AliasInfo != nil {
			oneFile := typeStrFile{
				FileName:  oneCreate.AliasInfo.LuaFile,
				Loc:       oneCreate.AliasInfo.AliasState.NameLoc,
				StrDetail: "alias info",
			}

			list = append(list, oneFile)
		}

		if oneCreate.ClassInfo != nil {
			oneFile := typeStrFile{
				FileName:  oneCreate.ClassInfo.LuaFile,
				Loc:       oneCreate.ClassInfo.ClassState.NameLoc,
				StrDetail: "class info",
			}

			list = append(list, oneFile)
		}
	}

	return list
}

// 获取注解字符串对应的信息
func (a *AllProject) getAnnotateStrTypeInfo(strName string, fileName string, lastLine int) (createType *common.CreateTypeInfo) {
	// 1) 优先在当前文件查找，如果没有找到，再找全局的文件注释
	annotateFile := a.getAnnotateFile(fileName)
	if annotateFile == nil {
		log.Error("getInfoFileAnnotateType annotateFile is nil, file=%s", fileName)
		return
	}

	// 1.1) 这个文件查找这个符号
	createBestType := annotateFile.GetBestCreateTypeInfo(strName, lastLine)
	if createBestType != nil {
		createType = createBestType
		return
	}

	// 2.1) 获取所有的全局信息
	createTypeList, flag := a.createTypeMap[strName]
	if !flag || len(createTypeList.List) == 0 {
		return
	}

	createType = createTypeList.List[0]
	return
}

// 判断注解类型的字符串是否存在
func (a *AllProject) judgeExistAnnoteTypeStr(strName string) bool {
	_, ok := a.createTypeMap[strName]
	return ok
}

// 判断是否alias多个候选词
// ---@alias exitcode2 '"exit"' | '"signal"'
func (a *AllProject) getAliasMultiCandidate(className string, fileName string, line int) (str string) {
	creatType := a.getAnnotateStrTypeInfo(className, fileName, line)
	if creatType == nil || creatType.AliasInfo == nil {
		return
	}

	aliasState := creatType.AliasInfo.AliasState
	if aliasState == nil {
		return
	}

	multiType, ok := aliasState.AliasType.(*annotateast.MultiType)
	if !ok {
		return
	}

	var multiArr []string
	for _, one := range multiType.TypeList {
		constType, ok := one.(*annotateast.ConstType)
		if !ok {
			continue
		}

		oneStr := "    | "
		if constType.QuotesFlag {
			oneStr = oneStr + "\"" + constType.Name + "\""
		} else {
			oneStr = oneStr + constType.Name
		}

		if constType.Comment != "" {
			oneStr = oneStr + " -- " + constType.Comment
		}

		multiArr = append(multiArr, oneStr)
	}

	if len(multiArr) == 0 {
		return
	}

	str = "\n" + strings.Join(multiArr, "\n")
	return
}

// 判断是否alias多个候选词
// ---@alias exitcode2 '"exit"' | '"signal"'
func (a *AllProject) getAliasMultiCandidateMap(className string, fileName string, line int, strMap map[string]string) {
	creatType := a.getAnnotateStrTypeInfo(className, fileName, line)
	if creatType == nil || creatType.AliasInfo == nil {
		return
	}

	aliasState := creatType.AliasInfo.AliasState
	if aliasState == nil {
		return
	}

	multiType, ok := aliasState.AliasType.(*annotateast.MultiType)
	if !ok {
		return
	}
	for _, one := range multiType.TypeList {
		constType, ok := one.(*annotateast.ConstType)
		if !ok {
			continue
		}

		oneStr := ""
		if constType.QuotesFlag {
			oneStr = oneStr + "\"" + constType.Name + "\""
		} else {
			oneStr = oneStr + constType.Name
		}

		strMap[oneStr] = constType.Comment
	}
}

func (a *AllProject) getSymbolAliasMultiCandidate(annotateType annotateast.Type, fileName string, line int) (str string) {
	if annotateType == nil {
		return
	}

	multiType, ok := annotateType.(*annotateast.MultiType)
	if !ok {
		return
	}

	for _, oneType := range multiType.TypeList {
		// 判断是否在几个候选词中
		if constType, ok := oneType.(*annotateast.ConstType); ok {
			oneStr := ""
			if constType.QuotesFlag {
				oneStr = "\"" + constType.Name + "\""
			} else {
				oneStr = constType.Name
			}

			if str != "" {
				str = str + " | " + oneStr
			} else {
				str = oneStr
			}
			continue
		}

		simpleType, ok := oneType.(*annotateast.NormalType)
		if !ok {
			continue
		}

		if isDefaultType(simpleType.StrName) {
			continue
		}

		str = str + a.getAliasMultiCandidate(simpleType.StrName, fileName, line)
		if str != "" {
			return str
		}
	}

	return str
}
