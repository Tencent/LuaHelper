package check

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/compiler/parser"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"strings"
)

// FileEventType 定义的文件变更类型
type FileEventType int

const (
	// FileEventCreated create
	FileEventCreated = 1
	// FileEventChanged change
	FileEventChanged = 2
	// FileEventDeleted delete
	FileEventDeleted = 3
)

// FileEventStruct 监控的文件变化对象
type FileEventStruct struct {
	StrFile string // 文件名
	Type    int    // 变化类型，1创建，2变化，3删除
}

// DefineStruct 文件定义的返回单个结构
type DefineStruct struct {
	StrFile string         // 文件名
	Loc     lexer.Location // 位置信息， 行从1开始，列从0开始
}

// 拷贝文件内的所有分析错误
// fileStrMap 为文件的唯一的错误信息map
func (a *AllProject) copyFileErr(strFile string, fileError []common.CheckError,
	fileErrorMap map[string][]common.CheckError, fileStrMap map[string]bool) {
	if len(fileError) == 0 {
		return
	}

	checkErrList, ok := fileErrorMap[strFile]
	if !ok {
		fileErrorMap[strFile] = make([]common.CheckError, 0)
		checkErrList = fileErrorMap[strFile]
	}

	for _, oneError := range fileError {
		strOnly := oneError.ToString()
		if _, ok := fileStrMap[strOnly]; ok {
			continue
		}

		fileStrMap[strOnly] = true
		checkErrList = append(checkErrList, oneError)
	}

	fileErrorMap[strFile] = checkErrList
}

// GetAllFileErrorInfo 获取所有检测的错误返回，按类型来排
func (a *AllProject) GetAllFileErrorInfo() map[string][]common.CheckError {
	fileErrorMap := make(map[string][]common.CheckError)

	// 所有分析文件去重的信息map
	onlyStrFileMap := make(map[string](map[string]bool))

	getFileStrMap := func(strFile string) map[string]bool {
		fileStrMap, ok := onlyStrFileMap[strFile]
		if !ok {
			onlyStrFileMap[strFile] = make(map[string]bool)
			fileStrMap = onlyStrFileMap[strFile]
		}

		return fileStrMap
	}

	// 1) 获取第一阶段的错误
	for strFile, fileStruct := range a.fileStructMap {
		if fileStruct.HandleResult == results.FileHandleReadErr {
			log.Error("read file:%s error", strFile)
			continue
		}

		fileResult := fileStruct.FileResult
		if fileResult != nil && len(fileResult.CheckErrVec) > 0 {
			fileStrMap := getFileStrMap(strFile)
			a.copyFileErr(strFile, fileResult.CheckErrVec, fileErrorMap, fileStrMap)
		}

		// 拷贝所有的注解错误
		annotateFile := fileStruct.AnnotateFile
		if annotateFile != nil {
			checkErrVec := annotateFile.GetErrorVec()
			if len(checkErrVec) > 0 {
				fileStrMap := getFileStrMap(strFile)
				a.copyFileErr(strFile, checkErrVec, fileErrorMap, fileStrMap)
			}
		}
	}

	// 若为非常规工程，则不做第二， 第三阶段诊断
	dirManager := common.GConfig.GetDirManager()
	mainDir := dirManager.GetMainDir()
	if mainDir == "" {
		return fileErrorMap
	}

	// 2) 获取第二阶段的错误
	for _, analysisSecond := range a.analysisSecondMap {
		for strFile, fileAnalsisError := range analysisSecond.FileErrorMap {
			if len(fileAnalsisError) == 0 {
				continue
			}

			fileStrMap := getFileStrMap(strFile)
			a.copyFileErr(strFile, fileAnalsisError, fileErrorMap, fileStrMap)
		}
	}

	if a.thirdStruct != nil {
		// 3) 获取第三阶段的错误
		for strFile, fileAnalsisError := range a.thirdStruct.FileErrorMap {
			if len(fileAnalsisError) == 0 {
				continue
			}

			fileStrMap := getFileStrMap(strFile)
			a.copyFileErr(strFile, fileAnalsisError, fileErrorMap, fileStrMap)
		}
	}

	return fileErrorMap
}

// IsNeedHandle 给一个文件名，判断是否要进行处理
func (a *AllProject) IsNeedHandle(strFile string) bool {
	// 判断该文件是否是忽略处理的
	if common.GConfig.IsIgnoreCompleteFile(strFile) {
		log.Debug("strFile=%s ignore", strFile)
		return false
	}

	return true
}

// GetCacheFileStruct 获取第一阶段的文件结构，如果有cache，优先获取cache的内容，cache的会相对新
func (a *AllProject) GetCacheFileStruct(strFile string) (*results.FileStruct, bool) {
	// 优先获取cache中的
	v, flag, _ := a.fileLRUMap.Get(strFile)
	if flag {
		// 有cache
		fileStruct := v.(*results.FileStruct)
		return fileStruct, true
	}

	return a.GetFirstFileStuct(strFile)
}

// getSpecialLineComment 获取某一行文件的注释
// inLuaFile 表示获取注释的文件
// lastLine 获取注释的行
// headFlag 表示是否为头部注释，true表示是，头部注释为注释在前面的一行
func (a *AllProject) getSpecialLineComment(inLuaFile string, lastLine int, headFlag bool) string {
	fileStruct, _ := a.GetCacheFileStruct(inLuaFile)
	if fileStruct != nil && fileStruct.FileResult != nil {
		oneComment := fileStruct.FileResult.GetFileLineComment(lastLine)
		if oneComment == nil {
			return ""
		}

		if oneComment.HeadFlag != headFlag {
			return ""
		}

		strComment := ""
		// 拼接新的注释
		for _, oneLine := range oneComment.LineVec {
			if strComment == "" {
				strComment = oneLine.Str
			} else {
				strComment = strComment + "\n" + oneLine.Str
			}
		}

		return strComment
	}

	return ""
}

// GetLineComment 获取某一行的注释
// luaFile 表示获取注释的文件
// line 获取注释的行
func (a *AllProject) GetLineComment(luaFile string, line int) string {
	// 先尝试获取同行的注释，如果没有；再获取前面一行的注释
	strComment := a.getSpecialLineComment(luaFile, line, false)
	if strComment == "" {
		strComment = a.getSpecialLineComment(luaFile, line-1, true)
	}

	return strComment
}

// insertFirstFileStruct 插入一个第一阶段处理的结果
func (a *AllProject) insertFirstFileStruct(strFile string, fileStruct *results.FileStruct) {
	a.fileStructMutex.Lock()
	defer a.fileStructMutex.Unlock()
	a.fileStructMap[strFile] = fileStruct
}

// GetFirstReferFileResult 给一个引用关系，找第一阶段引用lua文件
func (a *AllProject) GetFirstReferFileResult(referInfo *common.ReferInfo) *results.FileResult {
	if !referInfo.Valid || referInfo.ReferValidStr == "" {
		return nil
	}

	// 完整的引用文件，可能是调用的require("one"), 需要找到的文件为one.lua
	strFile := referInfo.ReferValidStr

	fileStruct, _ := a.GetFirstFileStuct(strFile)
	if fileStruct == nil {
		log.Debug("refer file %s  not find, line=%d", strFile, referInfo.Loc.StartLine)
		return nil
	}

	if fileStruct.HandleResult != results.FileHandleOk {
		log.Debug("refer file %s  first file struct err, line=%d", strFile, referInfo.Loc.StartLine)
		return nil
	}

	fileResult := fileStruct.FileResult
	if fileResult == nil {
		log.Debug("refer file %s first analysis result err, line=%d", strFile, referInfo.Loc.StartLine)
		return nil
	}

	if strFile != fileResult.Name {
		log.Debug("refer strFile error, oneName=%s, OtherName=%s, line=%d", strFile,
			fileResult.Name, referInfo.Loc.StartLine)
		return nil
	}

	return fileResult
}

// setCheckTerm 设置整体的轮数
func (a *AllProject) setCheckTerm(checkTerm results.CheckTerm) {
	a.checkTerm = checkTerm
}

// GetFuncDefaultParamInfo 获取默认参数标记
func (a *AllProject) GetFuncDefaultParamInfo(fileName string, lastLine int, paramNameList []string) (paramDefaultNum int) {
	annotateParamInfo := a.GetFuncParamInfo(fileName, lastLine)
	if annotateParamInfo == nil {
		return -1
	}

	for _, paramName := range paramNameList {
		for _, oneParam := range annotateParamInfo.ParamList {
			if paramName == oneParam.Name && oneParam.IsOptional {
				paramDefaultNum++
				break
			}
		}
	}

	return paramDefaultNum
}

//
func (a *AllProject) GetAnnotateClassAllFieldOfStrict(astType annotateast.Type, fileName string,
	lastLine int) (isStrict bool, retMap map[string]bool, className string) {

	isStrict = false
	retMap = map[string]bool{}
	className = ""

	classInfoList := a.getAllNormalAnnotateClass(astType, fileName, lastLine)

	//暂不处理多个class情况
	if len(classInfoList) == 1 {
		isStrict = classInfoList[0].ClassState.IsStrict
		for k := range classInfoList[0].FieldMap {
			retMap[k] = true
		}
		className = classInfoList[0].ClassState.Name
		return isStrict, retMap, className
	}

	return isStrict, retMap, className

}

// 获取注解class
func (a *AllProject) GetAnnotateClassByVar(strName string, varInfo *common.VarInfo) (isStrict bool, retMap map[string]bool, className string) {
	isStrict = false
	retMap = map[string]bool{}
	className = ""

	if varInfo == nil {
		return isStrict, retMap, className
	}

	strFile := varInfo.FileName

	// 1) 获取文件对应的annotateFile
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		log.Error("GetAnnotateClassByVar annotateFile is nil, file=%s", strFile)
		return isStrict, retMap, className
	}

	decLine := varInfo.Loc.StartLine - 1

	// 这里判断是否为函数参数的注解
	// 函数参数的注解，会额外的用下面的注解类型
	// ---@param one class
	// todo 这里如果为函数参数的，函数的注释一定要在前面一行，如果函数的参数占用两行，整个会有问题
	if varInfo.IsParam || varInfo.IsForParam {

		fragmentInfo := annotateFile.GetLineFragementInfo(decLine)
		if fragmentInfo != nil &&
			fragmentInfo.ParamInfo != nil &&
			len(fragmentInfo.ParamInfo.ParamList) > 0 {

			for i := 0; i < len(fragmentInfo.ParamInfo.ParamList); i++ {
				paramLine := fragmentInfo.ParamInfo.ParamList[i]

				//找到对应的那行---@param one class 再获取class
				if paramLine.Name == strName {
					return a.GetAnnotateClassAllFieldOfStrict(paramLine.ParamType, strFile, decLine)
				}
			}
		}
	} else {
		fragmentInfo := annotateFile.GetLineFragementInfo(varInfo.Loc.StartLine - 1)

		if fragmentInfo != nil &&
			fragmentInfo.TypeInfo != nil &&
			len(fragmentInfo.TypeInfo.TypeList) > 0 &&
			fragmentInfo.TypeInfo.TypeList[0] != nil {
			return a.GetAnnotateClassAllFieldOfStrict(fragmentInfo.TypeInfo.TypeList[0], strFile, varInfo.Loc.StartLine-1)
		}
	}

	//没找到返回空的
	return isStrict, retMap, className
}

// 获取注解class
func (a *AllProject) GetAnnotateClassByLoc(strFile string, lineForGetAnnotate int) (isStrict bool, retMap map[string]bool, className string) {
	isStrict = false
	retMap = map[string]bool{}
	className = ""

	// 1) 获取文件对应的annotateFile
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		log.Error("GetAnnotateClassByLoc annotateFile is nil, file=%s", strFile)
		return isStrict, retMap, className
	}

	fragmentInfo := annotateFile.GetLineFragementInfo(lineForGetAnnotate)
	if fragmentInfo != nil &&
		fragmentInfo.TypeInfo != nil &&
		len(fragmentInfo.TypeInfo.TypeList) > 0 &&
		fragmentInfo.TypeInfo.TypeList[0] != nil {
		return a.GetAnnotateClassAllFieldOfStrict(fragmentInfo.TypeInfo.TypeList[0], strFile, lineForGetAnnotate)
	}

	return isStrict, retMap, className
}

// 获取注解 ---type
func (a *AllProject) IsAnnotateTypeConst(varInfo *common.VarInfo) (isConst bool) {
	isConst = false

	// 1) 获取文件对应的annotateFile
	annotateFile := a.getAnnotateFile(varInfo.FileName)
	if annotateFile == nil {
		log.Error("GetAnnotateType annotateFile is nil, file=%s", varInfo.FileName)
		return isConst
	}

	fragmentInfo := annotateFile.GetLineFragementInfo(varInfo.Loc.StartLine - 1)

	if fragmentInfo != nil &&
		fragmentInfo.TypeInfo != nil &&
		len(fragmentInfo.TypeInfo.ConstList) > 0 {
		return fragmentInfo.TypeInfo.ConstList[0]
	}

	//没找到返回空的
	return isConst
}

// GetFirstFileStuct 获取第一阶段文件处理的结果
func (a *AllProject) GetFirstFileStuct(strFile string) (*results.FileStruct, bool) {
	if a.checkTerm == results.CheckTermFirst {
		a.fileStructMutex.Lock()
		defer a.fileStructMutex.Unlock()
	}

	fileStruct, ok := a.fileStructMap[strFile]
	return fileStruct, ok
}

// GetReferFrameType 如果为自定义的引入其他lua文件的方式，获取真实的引入的子类型
func (a *AllProject) GetReferFrameType(referInfo *common.ReferInfo) (subReferType common.ReferFrameType) {
	subReferType = common.RtypeNotValid
	if referInfo == nil {
		return subReferType
	}

	if referInfo.ReferType == common.ReferTypeRequire {
		subReferType = common.RtypeRequire
	}

	if referInfo.ReferType != common.ReferTypeFrame {
		return subReferType
	}

	// 如果是框架子定义的引入方式
	subReferType = common.GConfig.GetReferFrameSubType(referInfo.ReferTypeStr)
	if subReferType != common.RtypeAuto {
		return subReferType
	}

	// 如果为自动的，判断引入的文件是否返回值为table。如果为table，引入的方式即为require，否则为import
	referFile := a.GetFirstReferFileResult(referInfo)
	if referFile == nil {
		// 如果引入的文件不存在, 表示引入的无效
		subReferType = common.RtypeNotValid
		return
	}

	find, _ := referFile.MainFunc.GetOneReturnExp()
	if !find {
		// 如果函数没有返回值，为import
		subReferType = common.RtypeImport
		return
	}

	// 否则为require
	subReferType = common.RtypeRequire
	return subReferType
}

// VarInfoFlie 转换为DefineStruct定义的信息
func convertVarInfoFlieToDefine(symbol *common.Symbol) (flag bool, defineStruct DefineStruct) {
	// 1) 如果两个都有，优先看注解
	if symbol.VarInfo != nil && symbol.AnnotateType != nil {
		if symbol.VarFlag == common.FirstVarFlag {
			defineStruct = DefineStruct{
				StrFile: symbol.FileName,
				Loc:     symbol.VarInfo.Loc,
			}
		} else {
			defineStruct = DefineStruct{
				StrFile: symbol.FileName,
				Loc:     symbol.AnnotateLoc,
			}
		}

		return true, defineStruct
	}

	// 2) 其次优化看直接变量信息
	if symbol.VarInfo != nil {
		defineStruct = DefineStruct{
			StrFile: symbol.FileName,
			Loc:     symbol.VarInfo.Loc,
		}

		return true, defineStruct
	}

	// 3) 最后看注解
	if symbol.AnnotateType != nil {
		defineStruct = DefineStruct{
			StrFile: symbol.FileName,
			Loc:     symbol.AnnotateLoc,
		}

		return true, defineStruct
	}

	return false, defineStruct
}

// 获取有效的
func (a *AllProject) getVailidCacheFileStruct(strFile string) *results.FileStruct {
	// 1）先查找该文件是否存在
	fileStruct, _ := a.GetCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("not find file=%s", strFile)
		return nil
	}

	// 2) 文件结构是否正常
	if fileStruct.GetFileHandleErr() != results.FileHandleOk {
		log.Error("fileStruct not ok file=%s", strFile)
		return nil
	}
	if fileStruct.FileResult == nil {
		log.Error(" fileResult not ok file=%s", strFile)
		return nil
	}

	return fileStruct
}

// 删除注释一些空格和其他的多余的格式
// beforeEmptyLine 表示前面是否要插入空行
func getFinalStrComment(strComment string, beforeEmptyLine bool) string {
	if strComment == "" {
		return strComment
	}

	// 删除注释一些空格和其他的多余的格式
	splitStrArr := strings.Split(strComment, "\n")
	for index, oneStr := range splitStrArr {
		oneStr = strings.TrimPrefix(oneStr, "-*")
		oneStr = strings.TrimPrefix(oneStr, "*")
		oneStr = strings.TrimPrefix(oneStr, "-")
		oneStr = strings.TrimLeft(oneStr, " ")
		splitStrArr[index] = oneStr

		// 最后一个是空的字符串，删除掉
		if (index == len(splitStrArr)-1) && oneStr == "" {
			splitStrArr = splitStrArr[0 : len(splitStrArr)-1]
		}
	}
	if beforeEmptyLine {
		strComment = "\n" + strings.Join(splitStrArr, "\n")
	} else {
		strComment = strings.Join(splitStrArr, "\n")
	}

	return strComment
}

func recurseExpToDefine(exp ast.Exp, defineVar *common.DefineVarStruct) {
	switch expV := exp.(type) {
	case *ast.NameExp:
		defineVar.StrVec = append(defineVar.StrVec, expV.Name)
		defineVar.IsFuncVec = append(defineVar.IsFuncVec, false)
	case *ast.FuncCallExp:
		if subExp, flag := expV.PrefixExp.(*ast.NameExp); flag {
			if subExp.Name == "require" {
				defineVar.Exp = exp
			}
		}

		recurseExpToDefine(expV.PrefixExp, defineVar)
		if expV.NameExp == nil {
			if len(defineVar.IsFuncVec) > 0 {
				// 修改为true
				defineVar.IsFuncVec[len(defineVar.IsFuncVec)-1] = true
			}

		} else {
			defineVar.StrVec = append(defineVar.StrVec, expV.NameExp.Str)
			defineVar.IsFuncVec = append(defineVar.IsFuncVec, true)
		}
	case *ast.TableAccessExp:
		recurseExpToDefine(expV.PrefixExp, defineVar)

		defineVar.StrVec = append(defineVar.StrVec, common.GetExpName(expV.KeyExp))
		defineVar.IsFuncVec = append(defineVar.IsFuncVec, false)
	default:
		defineVar.StrVec = append(defineVar.StrVec, "$1")
		defineVar.IsFuncVec = append(defineVar.IsFuncVec, false)
	}
}

// ExpToDefineVarStruct 转换成VarStruct
func ExpToDefineVarStruct(exp ast.Exp) (defineVar common.DefineVarStruct) {
	switch expV := exp.(type) {
	case *ast.NameExp:
		defineVar.StrVec = append(defineVar.StrVec, expV.Name)
		defineVar.IsFuncVec = append(defineVar.IsFuncVec, false)
		defineVar.ValidFlag = true
	case *ast.FuncCallExp:
		defineVar.ValidFlag = true
		if subExp, flag := expV.PrefixExp.(*ast.NameExp); flag {
			if subExp.Name == "require" {
				defineVar.Exp = exp
			}
		}
		recurseExpToDefine(expV.PrefixExp, &defineVar)
		if expV.NameExp == nil {
			if len(defineVar.IsFuncVec) > 0 {
				// 修改为true
				defineVar.IsFuncVec[len(defineVar.IsFuncVec)-1] = true
			}
		} else {
			defineVar.StrVec = append(defineVar.StrVec, expV.NameExp.Str)
			defineVar.IsFuncVec = append(defineVar.IsFuncVec, true)
			defineVar.ColonFlag = true
		}
	case *ast.TableAccessExp:
		defineVar.ValidFlag = true
		recurseExpToDefine(expV.PrefixExp, &defineVar)
		defineVar.StrVec = append(defineVar.StrVec, common.GetExpName(expV.KeyExp))
		defineVar.IsFuncVec = append(defineVar.IsFuncVec, false)

	case *ast.ParensExp:
		defineVar.ValidFlag = true
		recurseExpToDefine(expV.Exp, &defineVar)
	default:
		break
	}

	return defineVar
}

// StrToDefineVarStruct change str to defineVarStruct
func StrToDefineVarStruct(str string) (defineVar common.DefineVarStruct) {
	defineVar.ValidFlag = false

	_, ok1 := common.GConfig.CompKeyMap[str]
	_, ok2 := common.GConfig.CompSnippetMap[str]
	if ok1 || ok2 {
		defineVar.ValidFlag = true
		defineVar.ColonFlag = false
		defineVar.StrVec = append(defineVar.StrVec, str)
		defineVar.IsFuncVec = append(defineVar.IsFuncVec, false)
		return defineVar
	}

	newParser := parser.CreateParser([]byte(str), "")
	exp := newParser.BeginAnalyzeExp()
	errList := newParser.GetErrList()
	if exp == nil || len(errList) > 0 {
		return defineVar
	}

	defineVar = ExpToDefineVarStruct(exp)
	return defineVar
}

// 判断是否为默认的exp，转换成对应的注解的类型
func extChangeVarInfo(exp ast.Exp, fileName string) (symbol *common.Symbol) {
	expType := common.GetSimpleExpType(exp)
	if expType == common.LuaTypeAll || expType == common.LuaTypeRefer {
		return nil
	}

	strName := "string"

	if expType == common.LuaTypeBool {
		strName = "boolean"
	} else if expType == common.LuaTypeNumber || expType == common.LuaTypeInter || expType == common.LuaTypeFloat {
		strName = "number"
	}
	normalAst := &annotateast.NormalType{
		StrName: strName,
	}

	symbol = &common.Symbol{
		FileName:     fileName,
		VarInfo:      nil,
		AnnotateType: normalAst,
		VarFlag:      common.FirstAnnotateFlag,
		AnnotateLine: 0, // todo这里的行号不太准确
	}

	return symbol
}
