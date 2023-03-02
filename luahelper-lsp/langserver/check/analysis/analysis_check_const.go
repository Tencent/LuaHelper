package analysis

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// 是否给常量赋值 当有a.b.c时，只判断a是否常量 若a是import值，则只判断b是否常量
func (a *Analysis) checkConstAssgin(node ast.Exp) {
	if !a.isNeedCheck() || a.realTimeFlag {
		return
	}

	if common.GConfig.IsGlobalIgnoreErrType(common.CheckErrorConstAssign) {
		return
	}

	if _, ok := common.GConfig.OpenErrorTypeMap[common.CheckErrorConstAssign]; !ok {
		return
	}

	preName := ""
	varName := ""
	preLoc := lexer.Location{}
	varLoc := lexer.Location{}
	switch exp := node.(type) {
	case *ast.NameExp:
		varName = exp.Name
		varLoc = exp.Loc
	case *ast.TableAccessExp:
		preName, varName, _, preLoc, varLoc, _ = common.GetTableNameInfo(exp)
	}

	ok, varInfo, isPreImport, _ := a.findVarDefineWithPre(preName, varName, preLoc, varLoc, false)
	if !ok {
		return
	}
	if varInfo.Loc == varLoc {
		//定义处不检查
		return
	}
	if varInfo.ReferInfo != nil {
		//引用不检查
		return
	}

	if a.Projects.IsAnnotateTypeConst(varName, varInfo) {
		tabName := preName
		if isPreImport || tabName == "" {
			tabName = varName
		}

		//标记了常量，却赋值
		errStr := fmt.Sprintf("'%s' is constant and not assignable", varName)
		a.curResult.InsertError(common.CheckErrorConstAssign, errStr, varLoc)
	}
}
