package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
)

func (a *Analysis) checkFuncOfClass(className string, funcName string, loc lexer.Location) {
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorClassField) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorClassField]; !ok {
		return
	}

	if className == "" || funcName == "" {
		return
	}

	// 如果是类的成员函数 先找类变量的定义
	find, varDefine := a.findVarDefineGlobal(className)
	if !find {
		return
	}

	// 再找类上方注解
	classTypeStr, ok := a.Projects.GetVarAnnType(varDefine.FileName, varDefine.Loc.StartLine-1)
	if !ok {
		return
	}

	// 根据注解找成员
	if a.Projects.IsFieldOfClass(classTypeStr, funcName) {
		return
	}

	errStr := fmt.Sprintf("Property '%s' not found in '%s'", funcName, className)
	a.curResult.InsertError(common.CheckErrorClassField, errStr, loc)
}

// CheckTableDecl 根据注解判断table成员合法性 在 t={f1=1,f1=2,} 时使用
func (a *Analysis) CheckTableDecl(strTableName string, strFieldNamelist []string, nodeLoc *lexer.Location, node *ast.TableConstructorExp) {
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorClassField) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorClassField]; !ok {
		return
	}

	if strTableName == "" || len(strFieldNamelist) == 0 || nodeLoc == nil || node == nil {
		return
	}

	isMemberMap, className := a.Projects.IsMemberOfAnnotateClassByLoc(a.curResult.Name, strFieldNamelist, nodeLoc.StartLine-1)
	if len(isMemberMap) == 0 || len(className) == 0 || (className) == "any" {
		return
	}

	for strFieldName, isMember := range isMemberMap {
		if !common.JudgeSimpleStr(strFieldName) {
			continue
		}

		if isMember {
			log.Debug("CheckTableDec currect, tableName=%s, keyName=%s", strTableName, strFieldName)
		} else {
			ok, keyLoc := common.GetTableConstructorExpKeyStrLoc(*node, strFieldName)

			if ok {
				errStr := fmt.Sprintf("Property '%s' not found in '%s'", strFieldName, className)
				a.curResult.InsertError(common.CheckErrorClassField, errStr, keyLoc)
			}
		}
	}
}

// 根据注解判断table成员合法性 包括 t.a t可以是符号 或者函数参数
// 最多只判断三段，如a.b.c.. 当a是import值或者_G时，会判断c是否b的成员，否则只判断b是否a成员
func (a *Analysis) checkTableAccess(node *ast.TableAccessExp) {
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorClassField) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorClassField]; !ok {
		return
	}

	preName := ""
	varName := ""
	keyName := ""
	preLoc := lexer.Location{}
	varLoc := lexer.Location{}

	preName, varName, keyName, preLoc, varLoc, _ = common.GetTableNameInfo(node)
	if preName == "" || varName == "" {
		return
	}

	ok, varInfo, isPreImport := a.findVarDefineWithPre(preName, varName, preLoc, varLoc, false)
	if !ok {
		return
	}

	isMember := true
	className := ""
	useKeyName := keyName
	if isPreImport {
		if keyName == "" {
			//只有a.b a又是import值 这时候不检查成员
			return
		}
		isMember, className = a.Projects.IsMemberOfAnnotateClassByVar(keyName, varName, varInfo)
	} else {
		isMember, className = a.Projects.IsMemberOfAnnotateClassByVar(varName, preName, varInfo)
		useKeyName = varName
	}

	if isMember || len(className) == 0 || (className) == "any" {
		return
	}

	errStr := fmt.Sprintf("Property '%s' not found in '%s'", useKeyName, className)
	a.curResult.InsertError(common.CheckErrorClassField, errStr, node.Loc)
}
