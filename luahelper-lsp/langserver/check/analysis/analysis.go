package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/check/projects"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
)

// IgnoreAssignOR 忽略a = a or 0 其中a没有定义的告警
type IgnoreAssignOR struct {
	strName string // 变量定义的名字
	line    int    // 行号
}

// IgnoreInfo 忽略的信息
type IgnoreInfo struct {
	strName    string // 变量定义的名字
	line       int    // 行号
	inif       bool   // 判断是否在if语句块中
	assignName string // 赋值语句的部分
}

// Analysis 分析的结构, 遍历AST抽象语法树
type Analysis struct {
	// 分析的轮数, 目前有第一轮、第二轮、第三轮、第四轮、第五轮， 如果为第一轮分析，存放的结构为第一轮的
	checkTerm results.CheckTerm

	// 如果为第二轮，表示是否为入口工程文件分析
	entryFile string

	// 当checkTerm表示第一轮的时候，是否为实时分析，是否分析为实时敲入代码，此时很多检查不开展，节约时间
	// 当保存文件后，再进行第一轮分析，此时realTimeFlag为false
	realTimeFlag bool

	// 第二阶段中，忽略告警的错误
	ignoreInfo IgnoreInfo

	// 当前的FuncInfo指针
	curFunc *common.FuncInfo

	// 当前的ScopeInfo指针
	curScope *common.ScopeInfo

	// 当前的FileResult指针
	curResult *results.FileResult

	// 如果为第二阶段分析工程，指向的指针
	SingleProjectResult *results.SingleProjectResult

	// 如果为第三阶段分析散落的文件，指向散落文件的指针
	AnalysisThird *results.AnalysisThirdFile

	// 如果为第四阶段，查找变量的引用，指向四阶段文件的指针
	ReferenceResult *results.ReferenceFileResult

	// 如果为第五阶段，构造变量的所有层级调用，例如ss.data,方便做代码提示
	CompleteResult *results.CompleteFileResult

	// 如果为第六阶段，查找当前文件所有的全局变量或函数位置信息
	ColorResult *results.ColorFileResult

	// 第二阶段分析工程，需要找全局第一轮已经分析完的文件，指针指向；
	// 第一阶段也需要，当前lua文件引入了其他文件的符号，例如require("one") 或是dofile("one.lua")
	Projects projects.Projects
}

// CreateAnalysis 创建一个分析的结构
func CreateAnalysis(checkTerm results.CheckTerm, entryFile string) *Analysis {
	return &Analysis{
		checkTerm: checkTerm,
		entryFile: entryFile,
		ignoreInfo: IgnoreInfo{
			inif:       false,
			assignName: "",
		},
		SingleProjectResult: nil,
		Projects:            nil,
		AnalysisThird:       nil,
		ReferenceResult:     nil,
		CompleteResult:      nil,
		realTimeFlag:        false,
		curFunc:             nil,
	}
}

// 判断是否为第一轮分析，AST构造
func (a *Analysis) isFirstTerm() bool {
	return a.checkTerm == results.CheckTermFirst
}

// 判断是否为第二轮分析，工程入口人家
func (a *Analysis) isSecondTerm() bool {
	return a.checkTerm == results.CheckTermSecond
}

// 判断是否为第三轮，散落的文件分析
func (a *Analysis) isThirdTerm() bool {
	return a.checkTerm == results.CheckTermThird
}

// 判断是否为第四轮，单个文件的引用分析
func (a *Analysis) isFourTerm() bool {
	return a.checkTerm == results.CheckTermFour
}

// 判断是否为第五轮，单个文件的成员变量分析，获取所有的.调用
func (a *Analysis) isFiveTerm() bool {
	return a.checkTerm == results.CheckTermFive
}

// 判断是否为第六轮，返回所有全局的颜色
func (a *Analysis) isSixTerm() bool {
	return a.checkTerm == results.CheckTermSix
}

// 判断是否需要检查告警项，只在第二轮或第三轮才检查
func (a *Analysis) isNeedCheck() bool {
	if a.isSecondTerm() || a.isThirdTerm() {
		return true
	}

	return false
}

// 进入一个scope
func (a *Analysis) enterScope() {
	a.curFunc.EnterScope()
}

// 出一个scope
func (a *Analysis) exitScope() {
	a.curFunc.ExitScope()

	// 如果是一轮校验，判断是否要校验局部变量是否定义未使用
	if !a.isFirstTerm() || a.realTimeFlag {
		return
	}

	// 判断是否开启了局部变量定义了是否未使用的告警
	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorLocalNoUse) {
		return
	}

	scope := a.curScope
	if scope == nil {
		return
	}

	fileResult := a.curResult

	// 扫描当前scope，判断哪些局部变量定义了未使用
	for varName, varInfoList := range scope.LocVarMap {
		// _ 局部变量忽略, _G也忽略
		if varName == "_" || varName == "_G" {
			continue
		}

		if common.GConfig.IsIgnoreLocNotUseVar(varName) {
			// 如果为系统忽略的局部变量定义了，未使用的，忽略掉
			continue
		}

		for _, oneVar := range varInfoList.VarVec {
			if oneVar.IsUse || oneVar.IsClose {
				continue
			}

			// 定义的局部函数忽略
			if oneVar.ReferFunc != nil {
				continue
			}

			// 判断指向的关联变量，是否为系统的函数或模块
			// 例如 local math = math 这样的忽略掉
			expName := common.GetExpName(oneVar.ReferExp)

			// 1) 判断是否直接关联到的系统模块或函数
			oneStr := common.GetExpSubKey(expName)
			if oneStr != "" {
				if common.GConfig.IsInSysNotUseMap(oneStr) {
					// 为系统的模块或函数名，忽略掉
					continue
				}
			}

			// 2) 判断是否关联到系统模块的成员， 例如：local concat = table.concat
			flagG, strRet := common.StrRemovePreG(expName)
			if flagG {
				expName = "!" + strRet
			}
			moduleName, keyName := common.GetTableStrTwoStr(expName)
			if moduleName != "" && keyName != "" {
				if common.GConfig.IsInSysNotUseMap(moduleName) {
					// 为系统的模块或函数名，忽略掉
					continue
				}
			}

			errorStr := fmt.Sprintf("%s declared and not used", varName)
			fileResult.InsertError(common.CheckErrorLocalNoUse, errorStr, oneVar.Loc)

			// 遍历所有的定义了未使用，只是简单的赋值
			for _, subVar := range oneVar.NoUseAssignLocs {
				errorStr := fmt.Sprintf("%s declared and not used, this just assign", varName)
				fileResult.InsertError(common.CheckErrorNoUseAssign, errorStr, subVar)
			}

			// 清除掉
			oneVar.NoUseAssignLocs = nil
		}
	}
}

// 向分析的单个工程中，添加全局定义的变量
func (a *Analysis) insertAnalysisGlobalVar(strName string, varInfo *common.VarInfo) {
	fileResult := a.curResult

	// 首先插入到单个文件结构中的全局表中
	fileResult.InsertGlobalVar(strName, varInfo)

	// 第二轮为工程check，才放入到工程结构的全局符号表中
	if !a.isSecondTerm() {
		return
	}

	// 判断是否为_G的符号或是协议前缀
	if !varInfo.ExtraGlobal.GFlag && varInfo.ExtraGlobal.StrProPre == "" {
		return
	}

	// 第二阶段工程check中，每解析到一个_G的全局变量，放入到第二阶段全局map中
	a.SingleProjectResult.InsertGlobalGMaps(strName, varInfo, results.CheckTermSecond)
}

// 向二阶段工程分析中，插入第二阶段的AnalysisFileResult
func (a *Analysis) insertSecondProjectAnalysis(strFile string, secondAnalysisFile *results.FileResult) {
	a.SingleProjectResult.AnalysisFileMap[strFile] = secondAnalysisFile
}

// 判断二阶段工程分析中，AnalysisFileResult是否已经存在了
func (a *Analysis) isExistSecondProjectAnalysis(strFile string) bool {
	fileResult := a.SingleProjectResult.AnalysisFileMap[strFile]
	if fileResult == nil {
		return false
	}

	return true
}

// 第一轮遍历AST的处理
func (a *Analysis) HandleFirstTraverseAST(fileResult *results.FileResult) {
	a.curResult = fileResult
	a.curFunc = fileResult.MainFunc
	a.curScope = fileResult.MainFunc.MainScope
	a.cgBlock(fileResult.Block)
	a.exitScope()
}

// HandleSecondProjectTraverseAST 第二轮深度遍历AST的处理（带工程的方式）或是第三轮遍历单个文件
func (a *Analysis) HandleSecondProjectTraverseAST(initialResult *results.FileResult, parent *results.FileResult) {
	strFile := initialResult.Name
	mainAst := initialResult.Block

	dirManager := common.GConfig.GetDirManager()
	entryFile := dirManager.RemovePathDirPre(a.entryFile)

	fileResult := results.CreateFileResult(strFile, mainAst, results.CheckTermSecond, entryFile)

	// 插入主函数
	fileResult.InertNewFunc(fileResult.MainFunc)
	a.insertSecondProjectAnalysis(strFile, fileResult)

	a.curResult = fileResult
	a.curFunc = fileResult.MainFunc
	a.curScope = fileResult.MainFunc.MainScope
	// 开始遍历
	a.cgBlock(fileResult.Block)
	a.exitScope()
}

// 第二轮深度遍历过程中，其他地方又引入了其他的lua文件，这里统一处理
// 例如，lua文件中 第三行 local a = import("two.lua")
// 因此，执行到底三行的时候，需要跟入进去，执行two.lua文件
func (a *Analysis) deepHanleReferFile(referInfo *common.ReferInfo) {
	strFile := referInfo.ReferValidStr
	fileResult := a.GetReferFileResult(referInfo, results.CheckTermFirst)
	if fileResult == nil {
		return
	}

	if a.isExistSecondProjectAnalysis(strFile) {
		return
	}

	// 下面的为深度遍历，也保存
	backupAnFile := a.curResult
	backupFunc := a.curFunc
	backupScope := a.curScope

	// 深度递归处理加载的文件
	a.HandleSecondProjectTraverseAST(fileResult, backupAnFile)

	// 再还原
	a.curResult = backupAnFile
	a.curFunc = backupFunc
	a.curScope = backupScope
}

// InsertRequireInfoGlobalVars 第二轮，尝试把引入文件的require("two") 或 dofile("two.lua")
// 需要把他们的符号表引入到_G符号表中
func (a *Analysis) InsertRequireInfoGlobalVars(referInfo *common.ReferInfo, checkTerm results.CheckTerm) {
	subReferType := a.Projects.GetReferFrameType(referInfo)
	if subReferType == common.RtypeImport {
		return
	}

	referFile := a.GetReferFileResult(referInfo, results.CheckTermFirst)
	strFile := referInfo.ReferValidStr
	if referFile == nil {
		return
	}

	if subReferType == common.RtypeRequire {
		var requireFilesMap map[string]bool
		if checkTerm == results.CheckTermFirst {
			requireFilesMap = a.SingleProjectResult.FirstRequireFileMap
		} else if checkTerm == results.CheckTermSecond {
			requireFilesMap = a.SingleProjectResult.SecondRequireFileMap
		}

		if requireFilesMap[strFile] {
			return
		}

		requireFilesMap[strFile] = true
	}

	for strName, oneVar := range referFile.GlobalMaps {
		// _G的全局变量，先忽略，已经加载过了
		if oneVar.ExtraGlobal.GFlag {
			continue
		}

		// 文件分析结构中，插入引用到的其他全局变量
		if checkTerm == results.CheckTermFirst {
			// 第二阶段工程check中，每解析到一个_G的全局变量，放入到第二阶段全局map中
			a.SingleProjectResult.InsertGlobalGMaps(strName, oneVar, results.CheckTermFirst)
		} else if checkTerm == results.CheckTermSecond {
			a.SingleProjectResult.InsertGlobalGMaps(strName, oneVar, results.CheckTermSecond)
		}
	}
}

// 查找第一阶段，对应文件的FileResult结构
func (a *Analysis) getFirstFileResult(strName string) *results.FileResult {
	fileStruct, _ := a.Projects.GetFirstFileStuct(strName)

	if fileStruct == nil {
		log.Error("getFirstFileResult file=%s error", strName)
		panic("getFirstFileResult error")
	}

	return fileStruct.FileResult
}

// GetReferFileResult 给一个引用关系，找到需要跟进去分析的引用lua文件
// checkTerm 表示第几轮，值为1或是2
func (a *Analysis) GetReferFileResult(referInfo *common.ReferInfo, checkTerm results.CheckTerm) *results.FileResult {
	if !referInfo.Valid || referInfo.ReferValidStr == "" {
		return nil
	}

	// 完整的引用文件，可能是调用的require("one"), 需要找到的文件为one.lua
	strFile := referInfo.ReferValidStr
	if checkTerm != results.CheckTermSecond {
		return a.Projects.GetFirstReferFileResult(referInfo)
	}
	secondFile := a.SingleProjectResult.AnalysisFileMap[strFile]
	if secondFile == nil {
		log.Debug("refer file %s first analysis result err, line=%d", strFile, referInfo.Loc.StartLine)
		return nil
	}

	if strFile != secondFile.Name {
		log.Debug("refer strFile error, oneName=%s, OtherName=%s, in lua file, line=%d", strFile,
			secondFile.Name, referInfo.Loc.StartLine)
		panic("error...")
	}

	return secondFile
}

// HandleTermTraverseAST 第三轮、第四轮、第五轮遍历,不会跟进去引用的文件
func (a *Analysis) HandleTermTraverseAST(checkTerm results.CheckTerm, firstFile *results.FileResult, parent *results.FileResult) {
	strFile := firstFile.Name
	mainAst := firstFile.Block
	fileResult := results.CreateFileResult(strFile, mainAst, checkTerm, "")

	if checkTerm == results.CheckTermThird {
		a.AnalysisThird.FileResult = fileResult
	} else if checkTerm == results.CheckTermFour {
		a.ReferenceResult.SetFileResult(fileResult)
	} else if checkTerm == results.CheckTermFive {
		//analysis.analysisFive.fileResult = fileResult
	} else if checkTerm == results.CheckTermSix {
		a.ColorResult.FileResult = fileResult
	} else {
		log.Error("checkTerm error")
		return
	}

	// 插入主函数
	fileResult.InertNewFunc(fileResult.MainFunc)
	a.curResult = fileResult
	a.curFunc = fileResult.MainFunc
	a.curScope = fileResult.MainFunc.MainScope
	// 开始遍历
	a.cgBlock(fileResult.Block)
	a.exitScope()
}

func (a *Analysis) SetRealTimeFlag(flag bool) {
	a.realTimeFlag = flag
}

// 根据注解判断table成员合法性 在local t={f1=1,f1=2,} 时使用，全局符号todo
func (a *Analysis) CheckClassField(strTableName string, strFieldName string, nodeLoc lexer.Location) {
	// 下面的判断只在第3轮，且是非实时检查时才触发
	if !a.isThirdTerm() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorClassField) {
		return
	}

	if strTableName == "" || strFieldName == "" {
		return
	}

	if !common.JudgeSimpleStr(strFieldName) {
		return
	}

	isStrict, fieldMap := a.Projects.GetAnnotateClass(a.curResult.Name, strTableName, &nodeLoc, nil, a.curScope)
	if !isStrict || len(fieldMap) == 0 {
		return
	}

	if fieldMap[strFieldName] {
		log.Debug("CheckClassField currect, tableName=%s, keyName=%s", strTableName, strFieldName)
	} else {
		errStr := fmt.Sprintf("the field (%s), is not a member of (%s)", strFieldName, strTableName)
		//a.curResult.InsertError(common.CheckErrorSelfAssign, errStr, nodeLoc)
		a.curResult.InsertError(common.CheckErrorClassField, errStr, nodeLoc)
	}

	return
}

// 根据注解判断table成员合法性 包括 t.a t可以是local符号 或者函数参数 ，全局符号todo
func (a *Analysis) checkTableAccess(node *ast.TableAccessExp) {
	// 下面的判断只在第3轮，且是非实时检查时才触发
	if !a.isThirdTerm() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorClassField) {
		return
	}

	strTable := common.GetExpName(node.PrefixExp)
	strTableName := common.GetSimpleValue(strTable)
	if strTableName == "" {
		return
	}

	strKey := common.GetExpName(node.KeyExp)
	// 如果不是简单字符，退出
	if !common.JudgeSimpleStr(strKey) || strKey == "" {
		return
	}

	isStrict, fieldMap := a.Projects.GetAnnotateClass(a.curResult.Name, strTableName, nil, &node.Loc, a.curScope)
	if !isStrict || len(fieldMap) == 0 {
		return
	}

	if fieldMap[strKey] {
		log.Debug("checkTableAccess currect, tableName=%s, keyName=%s", strTableName, strKey)
	} else {
		errStr := fmt.Sprintf("the field (%s), is not a member of (%s)", strKey, strTableName)
		//a.curResult.InsertError(common.CheckErrorSelfAssign, errStr, node.Loc)
		a.curResult.InsertError(common.CheckErrorClassField, errStr, node.Loc)
	}

	return

}
