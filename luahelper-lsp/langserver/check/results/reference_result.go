package results

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
	"strings"
)

// SelfConvertInterface 接口， self转换的
type SelfConvertInterface interface {
	ChangeSelfToReferVar(strTable string, prefixStr string) (str string)
}

// ReferenceFileResult 第四阶段分析的单个lua文件，查找对变量的引用
type ReferenceFileResult struct {
	StrFile          string                 // 文件的名称
	fileResult       *FileResult            // 单个文件分析的指针
	fileName         string                 // 引用所在的lua文件
	referSuffVec     []string               // 引用的后缀存放字符串数组，例如引用a.b.c, referName存放的值为a，b和c的值存放在数组中
	findSymbol       *common.VarInfo        // 引用如果为局部信息，有值
	secondProjectVec []*SingleProjectResult // 引用的文件所需要处理的第二阶段工程名
	thirdStruct      *AnalysisThird         // 引用所包含的第三阶段散落的文件
	FindLocVec       []lexer.Location       // 找到的引用关系位置
	ignoreDefineLoc  lexer.Location         // 需要忽略的定义位置
}

// CreateReferenceFileResult 创建单个引用的指针
func CreateReferenceFileResult(strFile string) *ReferenceFileResult {
	return &ReferenceFileResult{
		StrFile:         strFile,
		fileResult:      nil,
		findSymbol:      nil,
		referSuffVec:    []string{},
		ignoreDefineLoc: lexer.Location{},
	}
}

// MatchVarInfo 对比遍历的局部变量引用是否是自己想要的
// excludeRequire 表示是否要剔除掉require的影响，require的时候会少点原生table的名称，剔除时候，four.referSuffVec也需要剔除第一个数据
func (r *ReferenceFileResult) MatchVarInfo(selfConvert SelfConvertInterface, strName string, fileName string, varInfo *common.VarInfo,
	fi *common.FuncInfo, strPreExp string, nameExp ast.Exp, excludeRequire bool) bool {
	referSuffVec := r.referSuffVec
	if excludeRequire && len(referSuffVec) > 1 {
		referSuffVec = referSuffVec[1:]
	}

	if r.findSymbol == nil || varInfo == nil {
		return false
	}

	if r.fileName != fileName {
		return false
	}

	if referSuffVec[0] != strName {
		return false
	}

	if !lexer.CompareTwoLoc(&r.findSymbol.Loc, &varInfo.Loc) {
		return false
	}

	lenSuffVec := len(referSuffVec)
	if lenSuffVec > 1 && strPreExp == "" {
		// 多余一层的查找，传入的为table，需要比对前面的table
		return false
	} else if lenSuffVec == 1 && strPreExp != "" {
		return false
	}

	findLoc := varInfo.Loc
	if lenSuffVec > 1 {
		// 第二轮或第三轮判断table取值是否有定义
		strTable := strPreExp
		if strings.Contains(strTable, "#") {
			return false
		}

		strKey := common.GetExpName(nameExp)
		findLoc = common.GetExpLoc(nameExp)
		// 如果不是简单字符，退出
		if !common.JudgeSimpleStr(strKey) {
			return false
		}

		strTableArry := strings.Split(strTable, ".")
		strTableArry = append(strTableArry, strKey)
		if len(strTableArry) != lenSuffVec {
			return false
		}

		// self进行替换
		if strTableArry[0] == "!self" {
			strTableArry[0] = selfConvert.ChangeSelfToReferVar(strTableArry[0], "!")
		}

		_, strTableArry[0] = common.StrRemoveSigh(strTableArry[0])
		if strTableArry[0] == "" {
			return false
		}

		for i := 1; i < lenSuffVec; i++ {
			if referSuffVec[i] != strTableArry[i] {
				return false
			}
		}
	} else {
		findLoc = common.GetExpLoc(nameExp)
	}

	r.FindLocVec = append(r.FindLocVec, findLoc)
	return true
}

func (r *ReferenceFileResult) SetFileResult(fileResult *FileResult) {
	r.fileResult = fileResult
}

// 查找所有引用的全局变量，在所有的第二阶段工程，和三阶段散落的文件集合中查找
func (r *ReferenceFileResult) FindProjectGlobal(selfConvert SelfConvertInterface, strName string, strProPre string,
	fi *common.FuncInfo, strPreExp string, nameExp ast.Exp) bool {
	if !r.findSymbol.IsGlobal() {
		return false
	}

	for _, secondProject := range r.secondProjectVec {
		ok, oneVar := secondProject.FindGlobalGInfo(strName, CheckTermFirst, strProPre)
		if !ok {
			continue
		}

		ok = r.MatchVarInfo(selfConvert, strName, oneVar.FileName, oneVar, fi, strPreExp, nameExp, false)
		if ok {
			return true
		}
	}

	// 向工程的第一阶段全局_G符号表中查找
	if r.thirdStruct != nil {
		ok, oneVar := r.thirdStruct.FindThirdGlobalGInfo(false, strName, strProPre)
		if !ok {
			return false
		}

		ok = r.MatchVarInfo(selfConvert, strName, oneVar.FileName, oneVar, fi, strPreExp, nameExp, false)
		if ok {
			log.Debug("find global Info, thirdStruct, strName=%s", strName)
			return true
		}
	}

	return false
}

// SetFindReferenceInfo 设置需要查找引用的信息
func (r *ReferenceFileResult) SetFindReferenceInfo(strName string, varInfo *common.VarInfo, secondVec []*SingleProjectResult,
	referSuffVec []string, ignoreDefineLoc lexer.Location, thirdStruct *AnalysisThird) {
	r.fileName = strName
	r.findSymbol = varInfo
	r.secondProjectVec = secondVec
	r.referSuffVec = referSuffVec
	r.ignoreDefineLoc = ignoreDefineLoc
	r.thirdStruct = thirdStruct
}
