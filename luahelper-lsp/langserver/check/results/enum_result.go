package results

import (
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
)

// 定义枚举类型关联的值
type enumVar struct {
	varStr  string          // 枚举变量的名称
	varInfo *common.VarInfo // 枚举变量的信息
}

// 定义枚举类型的值列表
type enumVacList struct {
	enumVarVec []*enumVar
}

// AddEnumVar 添加枚举变量
func (e *enumVacList) AddEnumVar(varStr string, varInfo *common.VarInfo) {
	e.enumVarVec = append(e.enumVarVec, &enumVar{
		varStr:  varStr,
		varInfo: varInfo,
	})
}

// CheckEnumVar 检查枚举变量指向的值是否存在
func (e *enumVacList) CheckEnumVar(varStr string, varInfo *common.VarInfo) (oldStr string, oldVar *common.VarInfo,
	flag bool) {
	for _, enumVar := range e.enumVarVec {
		if varInfo.CheckVarIsSample(enumVar.varInfo) {
			return enumVar.varStr, enumVar.varInfo, true
		}
	}

	return "", nil, false
}

// Len 注解错误排序相关的获取长度的函数
func (e *enumVacList) Len() int { return len(e.enumVarVec) }

// Less 注解错误排序比较的函数
func (e *enumVacList) Less(i, j int) bool {
	oneLoc := e.enumVarVec[i].varInfo.Loc
	twoLoc := e.enumVarVec[j].varInfo.Loc
	if oneLoc.StartLine == twoLoc.StartLine {
		return oneLoc.StartColumn < twoLoc.StartColumn
	}

	return oneLoc.StartLine < twoLoc.StartLine
}

// Swap 注解错误排序较好的函数
func (e *enumVacList) Swap(i, j int) {
	e.enumVarVec[i], e.enumVarVec[j] = e.enumVarVec[j], e.enumVarVec[i]
}

type tableEnumItem struct {
	fieldStr string
	keyExp   ast.Exp
	valueExp ast.Exp
}

// 定义枚举类型的值列表
type tableEnumList struct {
	enumItemVec []*tableEnumItem
}

// AddEnumVar 添加枚举变量
func (t *tableEnumList) AddEnumVar(varStr string, keyExp, valueExp ast.Exp) {
	t.enumItemVec = append(t.enumItemVec, &tableEnumItem{
		fieldStr: varStr,
		keyExp:   keyExp,
		valueExp: valueExp,
	})
}

// CheckEnumVar 检查枚举变量指向的值是否存在
func (t *tableEnumList) CheckEnumExp(varStr string, exp ast.Exp) (enumItem *tableEnumItem,
	flag bool) {
	for _, enumItem := range t.enumItemVec {
		if common.CheckExpRefIsSample(exp, enumItem.valueExp) {
			return enumItem, true
		}
	}

	return nil, false
}
