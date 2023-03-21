package results

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	"strings"
)

// CheckTerm 检查的论数
type CheckTerm int

const (
	// CheckTermFirst 第一轮，单独对文件进行AST构造
	CheckTermFirst CheckTerm = 1

	// CheckTermSecond 第二轮, 分析工程入口文件
	CheckTermSecond CheckTerm = 2

	// CheckTermThird 第三轮, 分析散落的文件
	CheckTermThird CheckTerm = 3

	// CheckTermFour 第四轮, 查找变量的引用
	CheckTermFour CheckTerm = 4

	// CheckTermFive 第五轮，获取全局变量或函数的位置，告诉客户端渲染特定的颜色
	CheckTermFive CheckTerm = 6
)

// FileResult 单个lua文件分析的结果
type FileResult struct {
	Name         string                     // lua文件名，唯一标识
	checkTerm    CheckTerm                  // 分析的轮数，默认从第1轮开始
	entryFile    string                     // 输出错误的时候，是否打印入口文件
	Block        *ast.Block                 // 整个ast结构
	MainFunc     *common.FuncInfo           // ast生成的主function
	ReferVec     []*common.ReferInfo        // 整体引用的vector，按先后顺序插入到其中
	CheckErrVec  []common.CheckError        // 所有检测错误的信息
	GlobalMaps   map[string]*common.VarInfo // 所有的全局信息, 包含没有_G的与含有_G前缀的变量
	ProtocolMaps map[string]*common.VarInfo // 所有的协议前缀符号表
	NodefineMaps map[string]*common.VarInfo // 第一阶段当中，没有定义但是出现的全局变量
	FuncIDVec    []*common.FuncInfo         // 保存的所有funcInfo信息，可以通过id来查找
	funcID       int                        // 自增的funcID，默认值为0，每产生一个新的funcID自增1
	CommentMap   map[int]*lexer.CommentInfo // 第一轮分析时候，保存所有的注释信息, key值为行号
}

// CreateFileResult 创建一个新的文件分析结果
func CreateFileResult(strName string, block *ast.Block, checkTerm CheckTerm, entryFile string) *FileResult {
	loc := lexer.Location{}
	if block != nil {
		loc = block.Loc
	}
	funcInfo := common.CreateFuncInfo(nil, 0, loc, true, nil, strName)

	return &FileResult{
		Name:         strName,
		Block:        block,
		MainFunc:     funcInfo,
		GlobalMaps:   map[string]*common.VarInfo{},
		ProtocolMaps: map[string]*common.VarInfo{},
		NodefineMaps: map[string]*common.VarInfo{},
		checkTerm:    checkTerm,
		funcID:       -1,
		CommentMap:   map[int]*lexer.CommentInfo{},
		entryFile:    entryFile,
	}
}

// InertNewFunc 顺序插入一个新的func
func (f *FileResult) InertNewFunc(newFunc *common.FuncInfo) {
	f.FuncIDVec = append(f.FuncIDVec, newFunc)
	f.funcID++
	newFunc.FuncID = f.funcID
}

func (f *FileResult) GetFileTerm() CheckTerm {
	return f.checkTerm
}

// 在globalMaps中插入一个新的全局定义信息
func (f *FileResult) InsertGlobalVar(name string, varInfo *common.VarInfo) {
	if varInfo.ExtraGlobal.StrProPre != "" {
		beforeVar := f.ProtocolMaps[name]
		varInfo.ExtraGlobal.Prev = beforeVar
		f.ProtocolMaps[name] = varInfo
	} else {
		beforeVar := f.GlobalMaps[name]
		varInfo.ExtraGlobal.Prev = beforeVar
		f.GlobalMaps[name] = varInfo
	}
}

// 在globalMaps中查找一个新的全局定义信息, 带了限定条件
// 找到了就返回true，否则返回false
// gFlag 是否只查找_G的
func (f *FileResult) FindGlobalLimitVar(name string, funcLv int, scopeLv int, loc lexer.Location,
	strProPre string, gFlag bool) (*common.VarInfo, bool) {
	findMaps := f.GlobalMaps
	if strProPre != "" {
		findMaps = f.ProtocolMaps
	}

	oneVar := findMaps[name]
	for {
		if oneVar == nil {
			return nil, false
		}

		extraGlobal := oneVar.ExtraGlobal
		// 判断是否要匹配_G的
		if gFlag && !extraGlobal.GFlag {
			oneVar = extraGlobal.Prev
			continue
		}

		if extraGlobal.StrProPre != strProPre {
			oneVar = extraGlobal.Prev
			continue
		}

		if extraGlobal.FuncLv > funcLv {
			return nil, false
		} else if extraGlobal.FuncLv < funcLv {
			return oneVar, true
		}

		if extraGlobal.ScopeLv > scopeLv {
			return nil, false
		} else if extraGlobal.ScopeLv < scopeLv {
			return oneVar, true
		}

		if oneVar.Loc.StartLine > loc.StartLine {
			return nil, false
		} else if oneVar.Loc.StartLine < loc.StartLine {
			return oneVar, true
		}

		if oneVar.Loc.StartColumn > loc.StartColumn {
			return nil, false
		}

		return oneVar, true
	}
}

// InsertRelateError 插入完整的关联的错误
func (f *FileResult) InsertRelateError(errType common.CheckErrorType, errStr string,
	loc lexer.Location, relateVec []common.RelateCheckInfo) {
	// 错误信息，只有在第一轮、第二轮、第三轮才会产生
	if f.checkTerm != CheckTermFirst && f.checkTerm != CheckTermSecond && f.checkTerm != CheckTermThird {
		return
	}

	oneCheckError := common.CheckError{
		ErrType:   errType,
		ErrStr:    errStr,
		Loc:       loc,
		EntryFile: f.entryFile,
	}

	// 如果不是语法错误，判断是否要忽略该文件错误
	if common.GConfig.IsIgnoreErrorFile(f.Name, errType) {
		log.Debug("ErrType:%d, luaFile:%s, errorInfo:%s, loc[(%d, %d), (%d, %d)] is ignore", errType,
			f.Name, errStr, loc.StartLine, loc.StartColumn, loc.EndLine, loc.EndColumn)
		return
	}

	log.Error("ErrType:%d, luaFile:%s, errorInfo:%s, loc[(%d, %d), (%d, %d)]", errType,
		f.Name, errStr, loc.StartLine, loc.StartColumn, loc.EndLine, loc.EndColumn)

	if len(relateVec) > 0 {
		oneCheckError.RelateVec = relateVec
	}

	f.CheckErrVec = append(f.CheckErrVec, oneCheckError)
}

// InsertError 文件的分析结构中，插入一个错误信息
func (f *FileResult) InsertError(errType common.CheckErrorType, errStr string, loc lexer.Location) {
	f.InsertRelateError(errType, errStr, loc, nil)
}

// 在全局变量中，查找变量是否存在、指向的refer、指向的函数定义
func (f *FileResult) FindGlobalVarInfo(strName string, gFlag bool, strProPre string) (bool, *common.VarInfo) {
	var globalVar *common.VarInfo
	ok := false
	if strProPre == "" {
		globalVar, ok = f.GlobalMaps[strName]
	} else {
		globalVar, ok = f.ProtocolMaps[strName]
	}
	if !ok {
		return false, nil
	}

	for {
		// 判断是否只找_G的
		if gFlag && !globalVar.ExtraGlobal.GFlag {
			if globalVar.ExtraGlobal.Prev == nil {
				return false, nil
			}

			globalVar = globalVar.ExtraGlobal.Prev
		} else {
			if globalVar.ExtraGlobal.StrProPre == strProPre {
				return true, globalVar
			}

			if globalVar.ExtraGlobal.Prev == nil {
				return false, nil
			}

			globalVar = globalVar.ExtraGlobal.Prev
		}
	}
}

// 查找引用一个文件的结果
// allFilesMap 为所有加载文件map
func (f *FileResult) CheckReferFile(referInfo *common.ReferInfo, allFilesMap map[string]string, fileIndexInfo *common.FileIndexInfo) {
	strFile := referInfo.ReferStr
	strFile = pathpre.GetRemovePreStr(strFile)
	curFile := f.Name

	// 如果该文件是需要被忽略，直接返回
	if common.GConfig.IsIngoreNeedReadFile(strFile) {
		referInfo.Valid = false
		return
	}

	dirManager := common.GConfig.GetDirManager()

	// 判断引入方式，是否要包含后缀的方式
	suffixFlag := common.JudgeReferSuffixFlag(referInfo.ReferType, referInfo.ReferTypeStr)
	// 1) 是带后缀的
	if suffixFlag {
		// 1.1) 如果是包含了后缀，查看路径下面是否直接有, 直接存在返回true
		strFileMatch := dirManager.MatchCompleteReferFile(curFile, strFile)
		if strFileMatch != "" {
			// lua文件存在，正常
			referInfo.ReferValidStr = strFileMatch
			return
		}

		// 1.2) 非全路径匹配
		if !common.GConfig.ReferMatchPathFlag {
			// 如果配置为非全路径匹配，尝试模糊匹配路径
			bestFilePath := common.GetBestMatchReferFile(curFile, strFile, allFilesMap, fileIndexInfo)
			if bestFilePath != "" {
				// lua文件存在，正常
				referInfo.ReferValidStr = bestFilePath
				return
			}
		}

		// 3) 没有读到该文件
		if f.checkTerm == CheckTermFirst {
			// 没有读到文件，报错
			errStr := fmt.Sprintf("%s file error, not find file:%s", referInfo.ReferTypeStr, strFile)
			f.InsertError(common.CheckErrorNoFile, errStr, referInfo.Loc)
		}
		referInfo.Valid = false
		return
	}

	// 判断是否忽略require 该模块变量
	if referInfo.ReferType == common.ReferTypeRequire && common.GConfig.IsIgnoreRequireModuleError(strFile) {
		referInfo.Valid = false
		return
	}

	// 下面处理，引入不包含后缀的
	// require可能包含. 替换为/
	strNewFile := strings.Replace(strFile, ".", "/", -1)
	if common.GConfig.ReferMatchPathFlag {
		// a) 优先尝试找so
		// 如果不匹配，但是下面存在.so的，优先匹配到so的
		soFile := strNewFile + ".so"
		matchAllDirFile := dirManager.MatchAllDirReferFile(curFile, soFile)
		if matchAllDirFile != "" {
			// 如果so存在, 返回
			referInfo.Valid = false
			return
		}

		// 如果是引用全路径
		// b) 用它去匹配对应的文件 suffixStrFile，匹配对应的文件
		referLuaFile := strNewFile + ".lua"
		matchAllDirFile = dirManager.MatchAllDirReferFile(curFile, referLuaFile)
		if matchAllDirFile != "" {
			// 匹配到了
			referInfo.ReferValidStr = matchAllDirFile
			return
		}

		// c) 判断拼接的 init.lua是否存在
		initFile := strNewFile + "/init.lua"
		matchAllDirFile = dirManager.MatchAllDirReferFile(curFile, initFile)
		if matchAllDirFile != "" {
			// 匹配到了
			referInfo.ReferValidStr = matchAllDirFile
			return
		}

		if f.checkTerm == CheckTermFirst {
			// 没有读到文件，报错
			errStr := fmt.Sprintf("%s file error, not find file:%s", referInfo.ReferTypeStr, strFile)
			f.InsertError(common.CheckErrorNoFile, errStr, referInfo.Loc)
		}

		// 都没有找到
		referInfo.Valid = false
		return
	}

	// 下面的为模糊匹配
	// a) 路径下面存在.so的，优先匹配到so的
	soFile := strNewFile + ".so"
	matchAllDirFile := dirManager.MatchAllDirReferFile(curFile, soFile)
	if matchAllDirFile != "" {
		// 如果so存在, 返回
		referInfo.Valid = false
		return
	}

	// b) 如果配置为非全路径匹配，尝试模糊匹配路径
	strBestFileTmp := common.GetBestMatchReferFile(curFile, strNewFile, allFilesMap, fileIndexInfo)
	if strBestFileTmp != "" {
		// 匹配到了，判断对应的文件，是否存在
		// lua文件存在，正常
		referInfo.ReferValidStr = strBestFileTmp
		return
	}

	// c) suffixStrFile = strOldFile + "/init.lua"
	initFile := strNewFile + "/init.lua"
	// 如果配置为全路径匹配，尝试模糊匹配路径
	strBestFileTmp = common.GetBestMatchReferFile(curFile, initFile, allFilesMap, fileIndexInfo)
	if strBestFileTmp != "" {
		referInfo.ReferValidStr = strBestFileTmp
		return
	}

	if f.checkTerm == CheckTermFirst {
		// 没有读到文件，报错
		errStr := fmt.Sprintf("%s file error, not find file:%s", referInfo.ReferTypeStr, strFile)
		f.InsertError(common.CheckErrorNoFile, errStr, referInfo.Loc)
	}

	// 都没有找到
	referInfo.Valid = false
}

// isHasErrorNoFile 判断文件是否包含引用错误
func (f *FileResult) isHasErrorNoFile() bool {
	for _, oneErr := range f.CheckErrVec {
		if oneErr.ErrType == common.CheckErrorNoFile {
			return true
		}
	}

	return false
}

// 判断引用的文件是否包含指定的文件列表
func (f *FileResult) isReferFileContainFiles(needReferFileMap map[string]struct{}) bool {
	for _, oneRefer := range f.ReferVec {
		if _, ok := needReferFileMap[oneRefer.ReferValidStr]; ok {
			log.Debug("strFile=%s has change refer=%s", f.Name, oneRefer.ReferValidStr)
			return true
		}

		strFile := pathpre.GetRemovePreStr(oneRefer.ReferStr)
		for oneStr := range needReferFileMap {
			if strings.HasSuffix(oneStr, strFile) {
				return true
			}

			if !strings.HasSuffix(strFile, ".lua") {
				strTemp := strFile + ".lua"
				if strings.HasSuffix(oneStr, strTemp) {
					return true
				}
			}
		}
	}

	return false
}

// ReanalyseReferInfo 重新分析这个文件的所有引用关系，引用有文件变动（有文件增加或减少）
// allFilesMap map[string]bool 为所有加载的文件列表
func (f *FileResult) ReanalyseReferInfo(needReferFileMap map[string]struct{}, allFilesMap map[string]string, fileIndexInfo *common.FileIndexInfo) {
	// 1) 首先判断是否有包含改动的引用关系
	if !f.isHasErrorNoFile() && !f.isReferFileContainFiles(needReferFileMap) {
		return
	}

	log.Debug("strFile=%s has change refer", f.Name)

	// 2) 如果该文件有变动的引用关系，引用关系重新梳理，清除掉之前的引用关系错误
	var newErrVec []common.CheckError
	for _, oneError := range f.CheckErrVec {
		if oneError.ErrType == common.CheckErrorNoFile {
			continue
		}

		newErrVec = append(newErrVec, oneError)
	}
	f.CheckErrVec = newErrVec

	// 3) 然后重新扫描所有的引用关系
	for _, oneRefer := range f.ReferVec {
		oneRefer.Valid = true
		f.CheckReferFile(oneRefer, allFilesMap, fileIndexInfo)
	}
}

// FindASTNode 给定行与列， 查找AST上的节点
// posLine 从0开始
// posCh 从0开始
func (f *FileResult) FindASTNode(posLine, posCh int) (*common.ScopeInfo, *common.FuncInfo) {
	mainAst := f.Block
	if mainAst == nil {
		log.Error("strFile=%s mainAst is nil", f.Name)
		return nil, nil
	}

	// 1) 先通过行与列，找到最小单位的ScopeInfo
	minScope := f.MainFunc.MainScope.FindMinScope(posLine+1, posCh)
	if minScope == nil {
		log.Debug("strFile=%s find minScope is nil", f.Name)
		minScope = f.MainFunc.MainScope
	}
	log.Debug("strFile=%s find minScope is[(%d, %d), (%d, %d)]", f.Name, minScope.Loc.StartLine,
		minScope.Loc.StartColumn, minScope.Loc.EndLine, minScope.Loc.EndColumn)

	// 2) 通过最新的scopeInfo，定位到最小的funcInfo
	minFunc := minScope.FindMinFunc()
	if minFunc == nil {
		log.Debug("strFile=%s find minMin is nil", f.Name)
		return f.MainFunc.MainScope, f.MainFunc
	}
	log.Debug("strFile=%s find minFunc is[(%d, %d), (%d, %d)]", f.Name, minFunc.Loc.StartLine,
		minFunc.Loc.StartColumn, minFunc.Loc.EndLine, minFunc.Loc.EndColumn)

	return minScope, minFunc
}

// FindAllSymbol 单个文件查找所有的符号
func (f *FileResult) FindAllSymbol() (symbolVec []common.FileSymbolStruct) {
	// 查找当前文件下所有的全局变量scope
	gScopes := f.FindGMapsScopes()

	// 1) 首先找所有的局部变量
	symbolVec = f.MainFunc.MainScope.FindAllLocalVal(gScopes)

	// 2) 查找这个文件的所有全局变量
	for strVar, oneVar := range f.GlobalMaps {
		if common.GConfig.IsStrProtocol(strVar) {
			continue
		}

		containerName := ""
		if oneVar.ExtraGlobal.GFlag {
			containerName = "_G"
		}
		oneSymbol := common.FileSymbolStruct{
			Name:          strVar,
			Kind:          common.IKVariable,
			Loc:           oneVar.Loc,
			ContainerName: containerName,
		}

		// 若是函数
		if oneVar.ReferFunc != nil {
			oneSymbol.Loc = oneVar.ReferFunc.Loc
			oneSymbol.Kind = 2
			paramStr := ""
			// oneSymbol.Children = oneVar.ReferFunc.MainScope.FindAllLocalVal(nil)
			for _, param := range oneVar.ReferFunc.ParamList {
				if paramStr == "" {
					paramStr += param
				} else {
					paramStr += ", " + param
				}
			}
			if paramStr != "" {
				oneSymbol.Name = oneSymbol.Name + "(" + paramStr + ")"
			}
		} else {
			if oneVar.SubMaps == nil {
				symbolVec = append(symbolVec, oneSymbol)
				continue
			}

			var maxLoc lexer.Location
			maxLoc.EndLine = oneSymbol.Loc.EndLine
			maxLoc.EndColumn = oneSymbol.Loc.EndColumn

			for strName, varInfo := range oneVar.SubMaps {
				subOneSymbol := varInfo.FindAllVar(strName, strVar)
				if oneSymbol.Children == nil {
					oneSymbol.Children = make([]common.FileSymbolStruct, 0)
				}
				if subOneSymbol.Name != "" {
					oneSymbol.Children = append(oneSymbol.Children, subOneSymbol)

					if subOneSymbol.Loc.EndLine > maxLoc.EndLine {
						maxLoc.EndLine = subOneSymbol.Loc.EndLine
						maxLoc.EndColumn = subOneSymbol.Loc.EndColumn
					} else if subOneSymbol.Loc.EndLine == maxLoc.EndLine && subOneSymbol.Loc.EndColumn > maxLoc.EndColumn {
						maxLoc.EndColumn = subOneSymbol.Loc.EndColumn
					}
				}
			}

			oneSymbol.Loc.EndLine = maxLoc.EndLine
			oneSymbol.Loc.StartColumn = maxLoc.EndColumn
		}
		symbolVec = append(symbolVec, oneSymbol)
	}

	// todo 包含局部变量的时候，定位不准

	// 3) 查找这个文件的所有协议前缀
	if len(f.ProtocolMaps) != 0 {
		protocolSymbols := make(map[string]common.FileSymbolStruct)
		for strVar, oneVar := range f.ProtocolMaps {
			for {
				oneSymbol := common.FileSymbolStruct{
					Kind: common.IKVariable,
					Loc:  oneVar.Loc,
				}

				strProPre := oneVar.ExtraGlobal.StrProPre
				oneSymbol.Name = strProPre + "." + strVar
				if oneVar.ReferFunc != nil {
					oneSymbol.Kind = common.IKFunction
					oneSymbol.Loc = oneVar.ReferFunc.Loc
					paramStr := ""
					// oneSymbol.Children = oneVar.ReferFunc.MainScope.FindAllLocalVal(nil)
					for _, param := range oneVar.ReferFunc.ParamList {
						if paramStr == "" {
							paramStr += param
						} else {
							paramStr += ", " + param
						}
					}
					if paramStr != "" {
						oneSymbol.Name = oneSymbol.Name + "(" + paramStr + ")"
					}
				}

				if symbols, ok := protocolSymbols[strProPre]; ok {
					symbols.Children = append(symbols.Children, oneSymbol)
					protocolSymbols[strProPre] = symbols
				} else {
					protocolVar := f.GlobalMaps[strProPre]
					preSybmbol := common.FileSymbolStruct{
						Name: strProPre,
						Kind: common.IKVariable,
					}
					if protocolVar != nil {
						preSybmbol.Loc = protocolVar.Loc
					} else {
						preSybmbol.Loc = oneVar.Loc
					}

					preSybmbol.Children = append(preSybmbol.Children, oneSymbol)
					protocolSymbols[strProPre] = preSybmbol
				}

				if oneVar.ExtraGlobal.Prev == nil {
					break
				}

				oneVar = oneVar.ExtraGlobal.Prev
			}
		}

		for _, oneProtocolSymbol := range protocolSymbols {
			symbolVec = append(symbolVec, oneProtocolSymbol)
		}
	}

	return
}

// IsExistGlobalVarTableStrKey 查找文件中的所有的全局变量，判断是否有定义的table，包含下面的字符串key
// 返回前两层的字符串
func (f *FileResult) IsExistGlobalVarTableStrKey(strTableKey string, line int, charactor int) (firstStr string,
	secondStr string) {
	for strName, globaInfo := range f.GlobalMaps {
		for {
			firstStr, secondStr = common.GetSubMapStrKey(globaInfo.SubMaps, strName, strTableKey, line, charactor)
			if firstStr != "" {
				return firstStr, secondStr
			}

			if globaInfo.ExtraGlobal.Prev != nil {
				globaInfo = globaInfo.ExtraGlobal.Prev
			} else {
				break
			}
		}
	}

	return "", ""
}

// GetGlobalVarTableStrKey 代码补全时候，变量是否指向一个table，且它的key是一个strTableKey；或是是key为nil，value为strTableKey
func (f *FileResult) GetGlobalVarTableStrKey(strTableKey string, line int, charactor int) (string, *common.VarInfo) {
	for strName, globaInfo := range f.GlobalMaps {
		for {
			if globaInfo.IsHasReferTableKey(strTableKey, line, charactor) {
				return strName, globaInfo
			}

			if globaInfo.ExtraGlobal.Prev != nil {
				globaInfo = globaInfo.ExtraGlobal.Prev
			} else {
				break
			}
		}
	}

	return "", nil
}

// GetAstCheckError 获取AST的错误
func (f *FileResult) GetAstCheckError() (errList []common.CheckError) {
	fileError := f.CheckErrVec
	for _, oneErr := range fileError {
		if oneErr.ErrType == common.CheckErrorSyntax {
			errList = append(errList, oneErr)
		}
	}

	return errList
}

// GetFileLineComment 获取某一行的注释
func (f *FileResult) GetFileLineComment(line int) *lexer.CommentInfo {
	if line <= 0 {
		return nil
	}

	oneComment := f.CommentMap[line]
	return oneComment
}

// GetLineFuncInfo 获取指定行的函数信息
func (f *FileResult) GetLineFuncInfo(line int) *common.FuncInfo {
	var lastFuncInfo *common.FuncInfo
	for _, funcInfo := range f.FuncIDVec {
		if funcInfo.Loc.StartLine == line {
			lastFuncInfo = funcInfo
		}

		if funcInfo.Loc.StartLine > line {
			break
		}
	}

	return lastFuncInfo
}

// GetFuncInfoReferGlobalName 给定一个函数指针，判断是否是否关联到了对应的全局变量名称
func (f *FileResult) GetFuncInfoReferGlobalName(funcInfo *common.FuncInfo) string {
	if funcInfo == nil {
		return ""
	}

	// 1) 遍历所有的全局变量，判断指针是否一样
	for strName, oneFuncInfo := range f.GlobalMaps {
		if oneFuncInfo.ReferFunc == funcInfo {
			return strName
		}
	}

	// 2) 遍历所有的协议全局全局变量，判断指针是否一样
	for strName, oneFuncInfo := range f.ProtocolMaps {
		if oneFuncInfo.ReferFunc == funcInfo {
			return strName
		}
	}

	return ""
}

// FindGMapsScopes 找到当前全局变量 和 全局协议变量下的所有Scope
func (f *FileResult) FindGMapsScopes() (scopes []*common.ScopeInfo) {
	for strName, varInfo := range f.GlobalMaps {
		if common.GConfig.IsStrProtocol(strName) {
			continue
		}

		// 分析每个全局变量的referfunc
		if varInfo.ReferFunc != nil {
			scopes = append(scopes, varInfo.ReferFunc.MainScope)
		}

		// 分析每个全局变量的subMaps
		for _, subMapVarInfo := range varInfo.SubMaps {
			if subMapVarInfo.ReferFunc != nil {
				scopes = append(scopes, subMapVarInfo.ReferFunc.MainScope)
			}
		}
	}

	for _, varInfo := range f.ProtocolMaps {
		// 分析每个协议全局变量的referfunc
		if varInfo.ReferFunc != nil {
			scopes = append(scopes, varInfo.ReferFunc.MainScope)
		}
	}
	return
}

// JugetLineInFragement 判断给定的输入行是否在注解里面
func (f *FileResult) JugetLineInFragement(line int) bool {
	for _, comment := range f.CommentMap {
		if !comment.HeadFlag {
			continue
		}

		for _, oneLine := range comment.LineVec {
			if oneLine.Line == line {
				return true
			}
		}
	}

	return false
}

// GetForLineVarString 获取指定行的for语句引起的局部变量名称
func (f *FileResult) GetForLineVarString(line int) (strList []string) {
	mainScope := f.MainFunc.MainScope
	strList = mainScope.ForLineVarString(line)
	return
}
