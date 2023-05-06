package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
	"strings"
)

// CheckTableDecl 根据注解判断table成员合法性 在 t={f1=1,f1=2,} 时使用
func (a *Analysis) CheckTableDecl(strTableName string, strFieldNamelist []string, nodeLoc *lexer.Location, node *ast.TableConstructorExp) {

	a.CheckTableClassField(strTableName, strFieldNamelist, nodeLoc, node)
	a.CheckTableDeclAssign(node, nodeLoc)
}

// CheckTableDecl 根据注解判断table成员合法性 在 t={f1=1,f1=2,} 时使用
func (a *Analysis) CheckTableDeclAssign(tcExp *ast.TableConstructorExp, loc *lexer.Location) {
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorAssignType) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorAssignType]; !ok {
		return
	}

	retFieldTypeMap := map[string][]string{}
	a.Projects.GetFieldAnnotateType(a.curResult.Name, loc.StartLine-1, retFieldTypeMap)

	for i, key := range tcExp.KeyExps {
		if i >= len(tcExp.ValExps) {
			break
		}

		strKey := common.GetExpName(key)
		keyTypes, ok := retFieldTypeMap[strKey]
		if !ok {
			continue
		}

		// var keyTypes []string
		// for _, keyAnnType := range keyAnnTypes {
		// 	keyTypes = append(keyTypes, annotateast.GetAstTypeName(keyAnnType))
		// }

		valueTypes := a.GetAnnTypeByExp(tcExp.ValExps[i], -1)

		if len(keyTypes) == 0 || len(valueTypes) == 0 {
			//字段或者值取不到类型时 不检查
			continue
		}

		hasMatch := false
		for _, valueType := range valueTypes {
			for _, keyType := range keyTypes {
				if a.CompAnnTypeForAssign(keyType, valueType) {
					hasMatch = true
					break
				}
			}
		}

		if !hasMatch {
			loc := common.GetExpLoc(key)
			keyTypesStr := strings.Join(keyTypes, "|")
			valueTypesStr := strings.Join(valueTypes, "|")

			errStr := fmt.Sprintf("Type '%s' can not be assigned '%s'", keyTypesStr, valueTypesStr)
			a.curResult.InsertError(common.CheckErrorAssignType, errStr, loc)
		}
	}
}

// CheckTableDecl 根据注解判断table成员合法性 在 t={f1=1,f1=2,} 时使用
func (a *Analysis) CheckTableClassField(strTableName string, strFieldNamelist []string, nodeLoc *lexer.Location, node *ast.TableConstructorExp) {
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
	if len(isMemberMap) == 0 || len(className) == 0 || (className) == "any" || (className) == "table" {
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

	ok, varInfo, isPreImport, varType := a.findVarDefineWithPre(preName, varName, preLoc, varLoc, false)
	if !ok {
		return
	}

	if varInfo.IsParam {
		for _, annType := range varType {
			print(annType) //todo
		}
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

	if isMember || len(className) == 0 || (className) == "any" || (className) == "table" {
		return
	}

	errStr := fmt.Sprintf("Property '%s' not found in '%s'", useKeyName, className)
	a.curResult.InsertError(common.CheckErrorClassField, errStr, node.Loc)
}
