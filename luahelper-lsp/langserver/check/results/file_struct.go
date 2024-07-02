package results

import (
	"fmt"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
)

// FileHandleResult 表示文件处理的结果
type FileHandleResult int

const (
	// results.FileHandleOk 表示成功, 构造AST成功
	FileHandleOk FileHandleResult = 1
	// results.FileHandleReadErr 读文件失败
	FileHandleReadErr FileHandleResult = 2
	// results.FileHandleSyntaxErr 构造AST失败（lua语法分析错误）
	FileHandleSyntaxErr FileHandleResult = 3
)

// FileStruct 以文件为单位，存放第一阶段单个文件分析的所有结构
type FileStruct struct {
	StrFile      string               // 文件的名称
	HandleResult FileHandleResult     // 第一阶段处理的结果，是否有错误
	FileResult   *FileResult          // 第一遍单个lua文件扫描AST的结果
	AnnotateFile *common.AnnotateFile // 这个文件对应注解信息
	IsCommonFile bool                 // 是否是常规工程下的文件
	Contents     []byte               // 文件内容的切片，用于比较文件是否有变动
}

// CreateFileStruct 创建一个新的FileStruct
func CreateFileStruct(strFile string) *FileStruct {
	return &FileStruct{
		StrFile:      strFile,
		HandleResult: FileHandleOk,
		FileResult:   nil,
		AnnotateFile: common.CreateAnnotateFile(strFile),
		IsCommonFile: true,
	}
}

// GetFileHandleErr 获取是否有错误
func (f *FileStruct) GetFileHandleErr() FileHandleResult {
	return f.HandleResult
}

// checkMapEnum 检查文件map的enum枚举段定义是否有重复的值
func (f *FileStruct) CheckMapEnum() {
	annotateFile := f.AnnotateFile
	// 判断是否有枚举的定义
	if !annotateFile.IsHasEnumType() {
		return
	}

	// 1) 检查全局变量table的枚举定义
	for _, globalVar := range f.FileResult.GlobalMaps {
		if globalVar.VarType != common.LuaTypeTable {
			continue
		}

		f.checkVarAnnotateEnum(globalVar)
	}

	// 2) 检查局部变量table的枚举定义
	f.checkScopeEnumVar(f.FileResult.MainFunc.MainScope)
}

// checkScopeEnumVar 检查局部变量table的枚举定义
func (f *FileStruct) checkScopeEnumVar(scopeInfo *common.ScopeInfo) {
	for _, varInfoList := range scopeInfo.LocVarMap {
		for _, varInfo := range varInfoList.VarVec {
			if varInfo.VarType != common.LuaTypeTable {
				continue
			}

			f.checkVarAnnotateEnum(varInfo)
		}
	}

	for _, subScope := range scopeInfo.SubScopes {
		f.checkScopeEnumVar(subScope)
	}
}

func (f *FileStruct) checkVarAnnotateEnum(varInfo *common.VarInfo) {
	var tableEnumList common.TableEnumList
	expTable, ok := varInfo.ReferExp.(*ast.TableConstructorExp)
	if !ok {
		return
	}

	// 判断是否有枚举的定义
	if !f.isVarAnnotateEnum(varInfo) {
		return
	}

	for i, keyExp := range expTable.KeyExps {
		valExp := expTable.ValExps[i]
		if valExp == nil {
			continue
		}

		strKeySimple := ""
		strExp, ok := keyExp.(*ast.StringExp)
		if !ok {
			continue
		}
		strKeySimple = strExp.Str

		itemNum, flag := tableEnumList.CheckEnumExp(strKeySimple, valExp)
		if !flag {
			tableEnumList.AddEnumVar(strKeySimple, keyExp, valExp)
			continue
		}

		errStr := fmt.Sprintf("table: enum %s and %s contains duplicate value", itemNum.FieldStr, strKeySimple)

		var relateVec []common.RelateCheckInfo
		oldKeyLoc := common.GetExpLoc(itemNum.KeyExp)
		oldValueLoc := common.GetExpLoc(itemNum.ValueExp)
		oldRangeLoc := lexer.GetRangeLoc(&oldKeyLoc, &oldValueLoc)

		relateVec = append(relateVec, common.RelateCheckInfo{
			LuaFile: f.FileResult.Name,
			ErrStr:  errStr,
			Loc:     oldRangeLoc,
		})

		expKeyLoc := common.GetExpLoc(keyExp)
		expValueLoc := common.GetExpLoc(valExp)
		rangeLoc := lexer.GetRangeLoc(&expKeyLoc, &expValueLoc)
		f.FileResult.InsertRelateError(common.CheckErrorEnumValue, errStr, rangeLoc, relateVec)
	}
}

// 判断变量关联的注解是否有标明为枚举的类型
func (f *FileStruct) isVarAnnotateEnum(varInfo *common.VarInfo) bool {
	fragmentInfo := f.AnnotateFile.GetLineFragementInfo(varInfo.Loc.StartLine - 1)
	if fragmentInfo == nil || fragmentInfo.TypeInfo == nil {
		return false
	}
	if len(fragmentInfo.TypeInfo.EnumList) == 0 {
		return false
	}

	return fragmentInfo.TypeInfo.EnumList[0]
}
