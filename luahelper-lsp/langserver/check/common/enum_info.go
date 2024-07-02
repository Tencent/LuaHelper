package common

import (
	"luahelper-lsp/langserver/check/compiler/ast"
)

// 定义枚举类型关联的值
type enumVar struct {
	VarStr  string   // 枚举变量的名称
	VarInfo *VarInfo // 枚举变量的信息
}

// 定义枚举类型的值列表
type EnumVacList struct {
	EnumVarVec []*enumVar
}

// AddEnumVar 添加枚举变量
func (e *EnumVacList) AddEnumVar(varStr string, varInfo *VarInfo) {
	e.EnumVarVec = append(e.EnumVarVec, &enumVar{
		VarStr:  varStr,
		VarInfo: varInfo,
	})
}

// CheckEnumVar 检查枚举变量指向的值是否存在
func (e *EnumVacList) CheckEnumVar(varStr string, varInfo *VarInfo) (oldStr string, oldVar *VarInfo,
	flag bool) {
	for _, enumVar := range e.EnumVarVec {
		if varInfo.CheckVarIsSample(enumVar.VarInfo) {
			return enumVar.VarStr, enumVar.VarInfo, true
		}
	}

	return "", nil, false
}

// Len 注解错误排序相关的获取长度的函数
func (e *EnumVacList) Len() int { return len(e.EnumVarVec) }

// Less 注解错误排序比较的函数
func (e *EnumVacList) Less(i, j int) bool {
	oneLoc := e.EnumVarVec[i].VarInfo.Loc
	twoLoc := e.EnumVarVec[j].VarInfo.Loc
	if oneLoc.StartLine == twoLoc.StartLine {
		return oneLoc.StartColumn < twoLoc.StartColumn
	}

	return oneLoc.StartLine < twoLoc.StartLine
}

// Swap 注解错误排序较好的函数
func (e *EnumVacList) Swap(i, j int) {
	e.EnumVarVec[i], e.EnumVarVec[j] = e.EnumVarVec[j], e.EnumVarVec[i]
}

type tableEnumItem struct {
	FieldStr string
	KeyExp   ast.Exp
	ValueExp ast.Exp
}

// 定义枚举类型的值列表
type TableEnumList struct {
	enumItemVec []*tableEnumItem
}

// AddEnumVar 添加枚举变量
func (t *TableEnumList) AddEnumVar(varStr string, keyExp, valueExp ast.Exp) {
	t.enumItemVec = append(t.enumItemVec, &tableEnumItem{
		FieldStr: varStr,
		KeyExp:   keyExp,
		ValueExp: valueExp,
	})
}

// CheckEnumVar 检查枚举变量指向的值是否存在
func (t *TableEnumList) CheckEnumExp(varStr string, exp ast.Exp) (enumItem *tableEnumItem,
	flag bool) {
	for _, enumItem := range t.enumItemVec {
		if CheckExpRefIsSample(exp, enumItem.ValueExp) {
			return enumItem, true
		}
	}

	return nil, false
}
