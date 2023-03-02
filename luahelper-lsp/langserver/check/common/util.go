package common

import (
	"fmt"
	"luahelper-lsp/langserver/check/compiler/ast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"math"
	"strconv"
	"strings"
)

// GetTableAccessName 拼接table的名称
func GetTableAccessName(tabExp *ast.TableAccessExp) string {
	strPre := GetExpName(tabExp.PrefixExp)
	strKey := GetExpName(tabExp.KeyExp)
	return strPre + "." + strKey
}

// GetTableNameInfo 获取表达式的前三个值
// 例如 moduleA.var.key.a.b.c... 得到三个值 moduleA,var,key
// 当第一个词是_G,取前两个值，例如_G.var.key... 得到两个值 var,key
// 当只有两个值的时候 var.key 就取这两个
func GetTableNameInfo(tabExp *ast.TableAccessExp) (string, string, string, lexer.Location, lexer.Location, bool) {
	preExp := tabExp.PrefixExp
	varExp := tabExp.PrefixExp
	keyExp := tabExp.KeyExp
	srcKeyExp := keyExp
	isWhole := false
	for {
		subExp, ok := tabExp.PrefixExp.(*ast.TableAccessExp)
		if !ok {
			break
		}

		preExp = subExp.PrefixExp
		varExp = subExp.KeyExp
		keyExp = tabExp.KeyExp
		tabExp = subExp
	}

	preName := GetExpName(preExp)
	varName := GetExpName(varExp)
	keyName := GetExpName(keyExp)
	srcKeyName := GetExpName(srcKeyExp)

	if srcKeyName == keyName {
		isWhole = true
	}

	// 当是a[b]这种形式，b的值类型更多是运行时决定，这里截断
	if _, ok := keyExp.(*ast.StringExp); !ok {
		isWhole = false
		keyName = ""
	}

	if preName == "_G" || preName == varName {
		return GetSimpleValue(varName), keyName, "", GetExpLoc(varExp), GetExpLoc(keyExp), isWhole
	}

	// 当是a[b]这种形式，b的值类型更多是运行时决定，这里截断
	if _, ok := varExp.(*ast.StringExp); !ok {
		isWhole = false
		varName = ""
	}

	return GetSimpleValue(preName), varName, keyName, GetExpLoc(preExp), GetExpLoc(varExp), isWhole
}

// GetExpType 获取exp的类型
func GetExpType(node ast.Exp) LuaType {
	switch exp := node.(type) {
	case *ast.NilExp:
		return LuaTypeNil
	case *ast.FalseExp:
		return LuaTypeBool
	case *ast.TrueExp:
		return LuaTypeBool
	case *ast.IntegerExp:
		return LuaTypeInter
	case *ast.FloatExp:
		return LuaTypeFloat
	case *ast.StringExp:
		return LuaTypeString
	case *ast.FuncDefExp:
		return LuaTypeFunc
	case *ast.TableConstructorExp:
		return LuaTypeTable

	case *ast.BinopExp:
		if exp.Op == lexer.TkOpEq || exp.Op == lexer.TkOpNe || exp.Op == lexer.TkOpLt || exp.Op == lexer.TkOpLe ||
			exp.Op == lexer.TkOpGt || exp.Op == lexer.TkOpGe {
			return LuaTypeBool
		}

		if exp.Op == lexer.TkOpOr {

			oneType := GetExpType(exp.Exp1)
			if oneType != LuaTypeAll && oneType != LuaTypeRefer {
				return oneType
			} else {
				return GetExpType(exp.Exp2)
			}
		} else if exp.Op == lexer.TkOpAnd {
			return GetExpType(exp.Exp2)
		}

		oneType := GetExpType(exp.Exp1)
		if oneType != LuaTypeAll && oneType != LuaTypeRefer {
			return oneType
		}

		twoType := GetExpType(exp.Exp2)
		if twoType != LuaTypeAll && twoType != LuaTypeRefer {
			return twoType
		}

		opValue := exp.Op
		if opValue == lexer.TkOpAdd || opValue == lexer.TkOpSub || opValue == lexer.TkOpMinus ||
			opValue == lexer.TkOpMul || opValue == lexer.TkOpDiv || opValue == lexer.TkOpMod ||
			opValue == lexer.TkOpShr || opValue == lexer.TkOpShl || opValue == lexer.TkOpIdiv ||
			opValue == lexer.TkOpPow || opValue == lexer.TkOpWave {
			return LuaTypeFloat
		}

		if opValue == lexer.TkOpConcat {
			return LuaTypeString
		}

		return LuaTypeRefer

	case *ast.UnopExp:
		if exp.Op == lexer.TkOpNot {
			return LuaTypeBool
		}

		oneType := GetExpType(exp.Exp)
		if oneType != LuaTypeAll {
			return oneType
		}

		if exp.Op == lexer.TkOpNen || exp.Op == lexer.TkOpUnm {
			return LuaTypeInter
		}

	case *ast.NameExp:
		return LuaTypeRefer
	case *ast.ParensExp:
		return LuaTypeRefer
	case *ast.TableAccessExp:
		return LuaTypeRefer
	case *ast.FuncCallExp:
		return LuaTypeRefer
	}

	return LuaTypeAll
}

// 把LuaType转换成annotateast.type的字符串
func GetAnnTypeFromLuaType(lua_type LuaType) string {
	switch lua_type {
	case LuaTypeNumber, LuaTypeInter, LuaTypeFloat:
		return "number"
	case LuaTypeString:
		return "string"
	case LuaTypeBool:
		return "boolean"
	case LuaTypeNil:
		return "nil"
	case LuaTypeFunc:
		return "function"
	case LuaTypeTable:
		return "table"
	case LuaTypeRefer:
		return "LuaTypeRefer"
	}

	return "any"
}

// GetExpTypeString 获取ext的类型字符串
func GetExpTypeString(referExp ast.Exp) string {
	luaType := LuaTypeNil
	switch (referExp).(type) {
	case *ast.IntegerExp:
		luaType = LuaTypeInter
	case *ast.FloatExp:
		luaType = LuaTypeFloat
	case *ast.StringExp:
		luaType = LuaTypeString
	case *ast.BinopExp:
		luaType = LuaTypeBool
	}

	return GetLuaTypeString(luaType, referExp)
}

// GetLuaTypeString 获取每个类型的字符串
func GetLuaTypeString(luaType LuaType, referExp ast.Exp) string {
	// 尽量推导下，变量的类型
	expType := GetExpType(referExp)
	if expType != LuaTypeAll {
		luaType = expType
	}

	sufStr := ""
	switch exp := (referExp).(type) {
	case *ast.IntegerExp:
		sufStr = strconv.FormatInt(exp.Val, 10)
	case *ast.FloatExp:
		sufStr = strconv.FormatFloat(exp.Val, 'E', -1, 64)
	case *ast.StringExp:
		sufStr = exp.Str
	}

	switch luaType {
	case LuaTypeNil:
		return "any"
	case LuaTypeBool:
		return "boolean"
	case LuaTypeInter:
		if sufStr == "" {
			return "number"
		}
		return "number" + " = " + sufStr
	case LuaTypeFloat:
		if sufStr == "" {
			return "number"
		}
		return "number" + " = " + sufStr
	case LuaTypeString:
		if sufStr == "" {
			return "string"
		}
		return "string" + " = \"" + sufStr + "\""
	case LuaTypeFunc:
		return "function"
	case LuaTypeTable:
		return "table"
	}

	return "any"
}

// GetExpName 递归获取名称
func GetExpName(node ast.Exp) string {
	switch exp := node.(type) {
	case *ast.NilExp:
		return "#nil"
	case *ast.FalseExp:
		return "#flase"
	case *ast.TrueExp:
		return "#true"
	case *ast.IntegerExp:
		return "#int"
	case *ast.FloatExp:
		return "#float"
	case *ast.StringExp:
		// 如果是字符串，直接返回，前面不增加任何前缀
		return exp.Str
	case *ast.ParensExp:
		return GetExpName(exp.Exp)
	case *ast.VarargExp:
		return "#vararg"
	case *ast.NameExp:
		return "!" + exp.Name
	case *ast.FuncDefExp:
		//log.Error("GetExpName ast.FuncDefExp error")
		return "#errror"
	case *ast.TableConstructorExp:
		return "#table"
	case *ast.UnopExp:
		return "#astUnopExp"
	case *ast.BinopExp:
		return "#astBinopExp"
	case *ast.TableAccessExp:
		return GetTableAccessName(exp)
	case *ast.FuncCallExp:
		return "#funcall"
	}
	return "#other"
}

// GetTableAccessName 拼接table的名称
func GetTableAccessName1(tabExp *ast.TableAccessExp) string {
	strPre := GetExpName1(tabExp.PrefixExp)
	strKey := GetExpName1(tabExp.KeyExp)
	if strings.Contains(strKey, ".") || strings.Contains(strKey, "()") {
		strKey = "#other"
	}
	return strPre + "." + strKey
}

func GetExpName1(node ast.Exp) string {
	switch exp := node.(type) {
	case *ast.NilExp:
		return "#nil"
	case *ast.FalseExp:
		return "#flase"
	case *ast.TrueExp:
		return "#true"
	case *ast.IntegerExp:
		return "#int"
	case *ast.FloatExp:
		return "#float"
	case *ast.StringExp:
		// 如果是字符串，直接返回，前面不增加任何前缀
		return exp.Str
	case *ast.ParensExp:
		return GetExpName1(exp.Exp)
	case *ast.VarargExp:
		return "#vararg"
	case *ast.NameExp:
		return "!" + exp.Name
	case *ast.FuncDefExp:
		//log.Error("GetExpName ast.FuncDefExp error")
		return "#errror"
	case *ast.TableConstructorExp:
		return "#table"
	case *ast.UnopExp:
		return "#astUnopExp"
	case *ast.BinopExp:
		return "#astBinopExp"
	case *ast.TableAccessExp:
		return GetTableAccessName1(exp)
	case *ast.FuncCallExp:
		if exp.NameExp != nil {
			return GetExpName1(exp.PrefixExp) + "." + exp.NameExp.Str + "()"
		}
		return GetExpName1(exp.PrefixExp) + "()"
		//return "#funcall"
	}
	return "#other"
}

// GetTablePreName 递归获取table
func GetTablePreName(node ast.Exp) string {
	switch exp := node.(type) {
	case *ast.StringExp:
		// 如果是字符串，直接返回，前面不增加任何前缀
		return exp.Str
	case *ast.ParensExp:
		return GetTablePreName(exp.Exp)
	case *ast.NameExp:
		return "!" + exp.Name
	case *ast.TableAccessExp:
		strPre := GetTablePreName(exp.PrefixExp)
		strKey := GetTablePreName(exp.KeyExp)

		strKey = strings.ReplaceAll(strKey, ".", "$")
		return strPre + "." + strKey
	}
	return "$other"
}

// GetTableLocList 获取table所有成员和子key的loc list
func GetTableLocList(node ast.Exp) (locList []lexer.Location) {
	switch exp := node.(type) {
	case *ast.StringExp:
		locList = append(locList, exp.Loc)
	case *ast.NameExp:
		locList = append(locList, exp.Loc)
	case *ast.ParensExp:
		locList = append(locList, GetTableLocList(exp.Exp)...)
	case *ast.TableAccessExp:
		locList = append(locList, GetTableLocList(exp.PrefixExp)...)
		locList = append(locList, GetTableLocList(exp.KeyExp)...)
	default:
		var loc lexer.Location
		locList = append(locList, loc)
	}

	return locList
}

// GetTableConstuctorKeyStr table构造时候判断是否有重复key，提取key的字符串
// 目前这判断了key值为整型，字符串，引用其他值等
func GetTableConstuctorKeyStr(node ast.Exp, parentLoc lexer.Location) (strKey string, strShow string, loc lexer.Location) {
	strKey = ""
	strShow = ""
	switch exp := node.(type) {
	case *ast.IntegerExp:
		strShow = strconv.FormatInt(exp.Val, 10)
		strKey = "#int" + strShow
		loc = parentLoc
	case *ast.StringExp:
		// 如果是字符串，直接返回，前面不增加任何前缀
		strShow = exp.Str
		strKey = exp.Str
		loc = exp.Loc
	case *ast.NameExp:
		// 如果是引用的变量，前面增加!
		strKey = fmt.Sprintf("!%s", exp.Name)
		strShow = exp.Name
		loc = exp.Loc
	}

	return strKey, strShow, loc
}

// StrRemovePreG 如果变量含有 !_G.的前缀，表示一定为全局空间的变量，切除_G.
func StrRemovePreG(str string) (flagG bool, strRet string) {
	if strings.HasPrefix(str, "!_G.") {
		flagG = true
		strRet = str[4:]
	} else {
		flagG = false
		strRet = str
	}

	return
}

// StrRemoveSigh 判断变量是否为！的前缀，如果是切除掉前缀
func StrRemoveSigh(str string) (flag bool, strRet string) {
	if strings.HasPrefix(str, "!") {
		flag = true
		strRet = str[1:]
	} else {
		flag = false
		strRet = str
	}

	return
}

// IsVarargOrFuncCall 判断exp是否为可变参数或是函数调用
func IsVarargOrFuncCall(exp ast.Exp) bool {
	switch (exp).(type) {
	case *ast.VarargExp, *ast.FuncCallExp:
		return true
	}
	return false
}

// IsOneValueType 判断exp是否为单值类型 如：整数，字符串等 不完全 区别于函数返回值可能有多个值
func IsOneValueType(exp ast.Exp) bool {
	switch (exp).(type) {
	case *ast.NameExp,
		*ast.StringExp,
		*ast.LuajitNum,
		*ast.FloatExp,
		*ast.IntegerExp,
		*ast.FalseExp,
		*ast.TrueExp,
		*ast.NilExp:
		return true
	}
	return false
}

// IsVararg 判断是否是可变参数
func IsVararg(exp ast.Exp) bool {
	switch exp.(type) {
	case *ast.VarargExp:
		return true
	}
	return false
}

// JudgeIgnoreAssinVale 判断是否要忽略此赋值语句的定义
// 例如这样的语句：a = a
// a = a or
func JudgeIgnoreAssinVale(valExp ast.Exp, expExp ast.Exp) (strVar string, line int) {
	strVar, line = judgeAssignSelf(valExp, expExp)
	return
}

// judgeAssignSelf 根据传人的var = exp, 判断是否为 a = a
// 或者为 _G.b = _G.b
func judgeAssignSelf(valExp ast.Exp, expExp ast.Exp) (strVar string, line int) {
	strVar = ""
	flag := false
	line = 0
	switch exp := expExp.(type) {
	case *ast.NameExp:
		flag = true
		line = exp.Loc.EndLine
	case *ast.TableAccessExp:
		flag = true
		line = exp.Loc.EndLine
	}
	if !flag {
		return
	}

	switch valSubExp := valExp.(type) {
	case *ast.NameExp:
		strName := valSubExp.Name
		strExpName := GetExpName(expExp)
		if len(strExpName) > 1 && strExpName[0] == '!' {
			if strExpName[1:] == strName {
				strVar = strName
			}
		}

		return
	case *ast.TableAccessExp:
		tableExp, ok := expExp.(*ast.TableAccessExp)
		if !ok {
			return
		}

		varPreStr := GetExpName(valSubExp.PrefixExp)
		if varPreStr != "!_G" {
			return
		}

		binOpPreStr := GetExpName(tableExp.PrefixExp)
		if binOpPreStr != "!_G" {
			return
		}

		varKeyStr := GetExpName(valSubExp.KeyExp)
		if !JudgeSimpleStr(varKeyStr) {
			return
		}

		binOpKeyStr := GetExpName(tableExp.KeyExp)
		if varKeyStr != binOpKeyStr {
			return
		}

		subStringExp, ok := valSubExp.KeyExp.(*ast.StringExp)
		if !ok {
			return
		}

		strVar = subStringExp.Str
		line = subStringExp.Loc.StartLine
	}

	return
}

// JoinSimpleStr 字符串数组，进行拼接，如果每个成员为简单的字符串
func JoinSimpleStr(strArray []string) string {
	strFind := ""
	for _, strOne := range strArray {
		if !JudgeSimpleStr(strOne) {
			break
		}

		if strFind == "" {
			strFind = strOne
		} else {
			strFind = strFind + "." + strOne
		}
	}

	return strFind
}

// JudgeSimpleStr 判断是否为简单的字符串，没有包含 . # ! 字符串
func JudgeSimpleStr(strName string) bool {
	if strings.Contains(strName, "!") || strings.Contains(strName, "#") || strings.Contains(strName, ".") {
		return false
	}

	return true
}

// GetSimpleValue 判断是否为简单变量的引用
// 传人的为 !a ，获得的变量为a
func GetSimpleValue(strName string) string {
	if len(strName) <= 1 {
		return ""
	}

	if strName[0] != '!' {
		return ""
	}

	if strName == "!_G" {
		return ""
	}

	strName = strName[1:]
	if !JudgeSimpleStr(strName) {
		return ""
	}

	return strName
}

// GetTableStrTwoStr 传人一个tableStr， 切分出两个字符串
func GetTableStrTwoStr(strTable string) (strOne string, strTwo string) {
	strOne = ""
	strTwo = ""
	if strings.Contains(strTable, "#") {
		return
	}

	trTableArry := strings.Split(strTable, ".")
	if len(trTableArry) != 2 {
		return
	}

	strOne = trTableArry[0]
	strTwo = trTableArry[1]
	if !JudgeSimpleStr(strTwo) {
		strOne = ""
		strTwo = ""
		return
	}

	if !strings.HasPrefix(strOne, "!") {
		strOne = ""
		strTwo = ""
		return
	}

	strOne = strOne[1:]
	return
}

// GetAllExpSimpleNotValue 获取Exp 所有的not a 变量
// notNum 为非的次数
func GetAllExpSimpleNotValue(node ast.Exp, notNum int, strArr *[]string) {
	switch exp := node.(type) {
	case *ast.UnopExp:
		// not
		if exp.Op == lexer.TkOpNot {
			notNum = notNum + 1
			if notNum%2 == 1 {
				strExpName := GetExpName(exp.Exp)
				strValueName := GetSimpleValue(strExpName)
				if strValueName != "" {
					*strArr = append(*strArr, strValueName)
				} else {
					GetAllExpSimpleNotValue(exp.Exp, notNum, strArr)
				}
			}
		}
	case *ast.BinopExp:
		// or
		if exp.Op == lexer.TkOpOr {
			GetAllExpSimpleNotValue(exp.Exp1, notNum, strArr)
			GetAllExpSimpleNotValue(exp.Exp2, notNum, strArr)
		} else if exp.Op == lexer.TkOpAnd {
			// and
			GetAllExpSimpleNotValue(exp.Exp1, notNum, strArr)
			GetAllExpSimpleNotValue(exp.Exp2, notNum, strArr)
		} else if exp.Op == lexer.TkOpEq {
			strExpName := GetExpName(exp.Exp1)
			strValueName := GetSimpleValue(strExpName)
			exp2NilFlag := false
			switch (exp.Exp2).(type) {
			case *ast.NilExp:
				exp2NilFlag = true
			}

			if strValueName != "" && exp2NilFlag {
				notNum = notNum + 1
				if notNum%2 == 1 {
					*strArr = append(*strArr, strValueName)
				}
			}
		} else if exp.Op == lexer.TkOpNe {
			strExpName := GetExpName(exp.Exp1)
			strValueName := GetSimpleValue(strExpName)
			exp2NilFlag := false
			switch (exp.Exp2).(type) {
			case *ast.NilExp:
				exp2NilFlag = true
			}

			if strValueName != "" && exp2NilFlag {
				if notNum%2 == 1 {
					*strArr = append(*strArr, strValueName)
				}
			}
		}
	}
}

// 比较两个exp是否内容重复，只针对部分类型，
func CompExp(node1 ast.Exp, node2 ast.Exp) bool {
	switch exp1 := node1.(type) {
	case *ast.NilExp:
		_, ok := node2.(*ast.NilExp)
		return ok
	case *ast.FalseExp:
		_, ok := node2.(*ast.FalseExp)
		return ok
	case *ast.TrueExp:
		_, ok := node2.(*ast.TrueExp)
		return ok
	case *ast.IntegerExp:
		exp2, ok := node2.(*ast.IntegerExp)
		return ok && exp1.Val == exp2.Val
	case *ast.FloatExp:
		exp2, ok := node2.(*ast.FloatExp)
		return ok && math.Abs(exp1.Val-exp2.Val) < 0.000001
	case *ast.StringExp:
		exp2, ok := node2.(*ast.StringExp)
		return ok && exp1.Str == exp2.Str
	case *ast.ParensExp:
		exp2, ok := node2.(*ast.ParensExp)
		return ok && CompExp(exp1.Exp, exp2.Exp)
	case *ast.VarargExp:
		_, ok := node2.(*ast.VarargExp)
		return ok
	case *ast.NameExp:
		exp2, ok := node2.(*ast.NameExp)
		return ok && exp1.Name == exp2.Name
	case *ast.FuncDefExp:
	case *ast.TableConstructorExp:
		return false //这两种暂不做比较
	case *ast.UnopExp:
		exp2, ok := node2.(*ast.UnopExp)
		if ok && exp1.Op == exp2.Op {
			return CompExp(exp1.Exp, exp2.Exp)
		} else {
			return false
		}
	case *ast.BinopExp:
		exp2, ok := node2.(*ast.BinopExp)
		if ok && exp1.Op == exp2.Op {
			return CompExp(exp1.Exp1, exp2.Exp1) && CompExp(exp1.Exp2, exp2.Exp2)
		} else {
			return false
		}
	case *ast.TableAccessExp:
		exp2, ok := node2.(*ast.TableAccessExp)
		return ok && CompExp(exp1.PrefixExp, exp2.PrefixExp) && CompExp(exp1.KeyExp, exp2.KeyExp)
	case *ast.FuncCallExp:
		exp2, ok := node2.(*ast.FuncCallExp)
		if !ok {
			return false
		}

		if !CompExp(exp1.PrefixExp, exp2.PrefixExp) {
			return false
		}

		if (exp1.NameExp == nil && exp2.NameExp != nil) ||
			(exp1.NameExp != nil && exp2.NameExp == nil) {
			return false
		}

		if (exp1.NameExp != nil && exp2.NameExp != nil) &&
			!CompExp(exp1.NameExp, exp2.NameExp) {
			return false
		}

		if len(exp1.Args) == len(exp2.Args) {
			for i := 0; i < len(exp1.Args); i++ {
				if !CompExp(exp1.Args[i], exp2.Args[i]) {
					return false
				}
			}
		} else {
			return false
		}

		return true
	}

	return false
}

// GetAllExpSimpleValue 获取if 条件里面所有的 if a then, 提前里面的a
// 或是if a ~= nil then 提前里面的a
func GetAllExpSimpleValue(node ast.Exp, strArr *[]string) {
	switch exp := node.(type) {
	case *ast.BinopExp:
		// or
		if exp.Op == lexer.TkOpOr {
			GetAllExpSimpleValue(exp.Exp1, strArr)
			GetAllExpSimpleValue(exp.Exp2, strArr)
		} else if exp.Op == lexer.TkOpAnd {
			// and
			GetAllExpSimpleValue(exp.Exp1, strArr)
			GetAllExpSimpleValue(exp.Exp2, strArr)
		} else if exp.Op == lexer.TkOpEq {
			strExpName := GetExpName(exp.Exp1)
			strValueName := GetSimpleValue(strExpName)
			exp2NilFlag := false
			switch (exp.Exp2).(type) {
			case *ast.NilExp:
				exp2NilFlag = true
			}

			if strValueName != "" && !exp2NilFlag {
				*strArr = append(*strArr, strValueName)
			}
		} else if exp.Op == lexer.TkOpNe {
			strExpName := GetExpName(exp.Exp1)
			strValueName := GetSimpleValue(strExpName)
			exp2NilFlag := false
			switch (exp.Exp2).(type) {
			case *ast.NilExp:
				exp2NilFlag = true
			}

			if strValueName != "" && exp2NilFlag {
				*strArr = append(*strArr, strValueName)
			}
		}

	case *ast.NameExp:
		strName := exp.Name
		*strArr = append(*strArr, strName)
	}
}

// GetTableExpKeyStrVec 获取table定义时候，所有的字符串key值
func GetTableExpKeyStrVec(node ast.Exp) (tableKeyVec []TableKeyStruct) {
	switch tableExp := (node).(type) {
	case *ast.TableConstructorExp:
		valLen := len(tableExp.ValExps)
		for i, keyExp := range tableExp.KeyExps {
			if keyExp == nil {
				continue
			}

			strDetail := ""
			if valLen > i {
				strDetail = GetExpTypeString(tableExp.ValExps[i])
			}

			switch keyStrValue := keyExp.(type) {
			case *ast.StringExp:
				tableKeyVec = append(tableKeyVec, TableKeyStruct{
					StrKey:    keyStrValue.Str,
					Loc:       keyStrValue.Loc,
					StrDetail: strDetail,
				})
			}
		}
	}

	return tableKeyVec
}

// GetTableExpKeyStrLoc 获取table定义时候，获取指定子key的位置信息
// 为table的子key
func GetTableExpKeyStrLoc(node ast.Exp, strSubKey string) (flag bool, loc lexer.Location) {
	tableVec := GetTableExpKeyStrVec(node)
	for _, oneKey := range tableVec {
		if oneKey.StrKey == strSubKey {
			return true, oneKey.Loc
		}
	}

	return false, loc
}

// 获取table定义时候，所有的字符串key值
func GetTableConstructorExpKeyStrVec(node ast.TableConstructorExp) (tableKeyVec []TableKeyStruct) {
	tableExp := (node)

	valLen := len(tableExp.ValExps)
	for i, keyExp := range tableExp.KeyExps {
		if keyExp == nil {
			continue
		}

		strDetail := ""
		if valLen > i {
			strDetail = GetExpTypeString(tableExp.ValExps[i])
		}

		switch keyStrValue := keyExp.(type) {
		case *ast.StringExp:
			tableKeyVec = append(tableKeyVec, TableKeyStruct{
				StrKey:    keyStrValue.Str,
				Loc:       keyStrValue.Loc,
				StrDetail: strDetail,
			})
		}
	}

	return tableKeyVec
}

// 获取table定义时候，获取指定子key的位置信息
// 为table的子key
func GetTableConstructorExpKeyStrLoc(node ast.TableConstructorExp, strSubKey string) (flag bool, loc lexer.Location) {
	tableVec := GetTableConstructorExpKeyStrVec(node)
	for _, oneKey := range tableVec {
		if oneKey.StrKey == strSubKey {
			return true, oneKey.Loc
		}
	}

	return false, loc
}

// GetExpSubKey 判断传人的字符串是否符合！开头的，或是!G开头的
func GetExpSubKey(str string) string {
	if strings.Contains(str, "#") {
		return ""
	}

	strArry := strings.Split(str, ".")
	arrLen := len(strArry)
	if arrLen == 1 {
		oneStr := strArry[0]
		if strings.HasPrefix(oneStr, "!") {
			return oneStr[1:]
		}
		return ""
	} else if arrLen == 2 {
		oneStr := strArry[0]
		if oneStr == "!_G" {
			return strArry[1]
		}

		return ""
	}

	return ""
}

// GetBinopExpType 获取BinopExp指向的类型，如果是 or 类型，判断后面的类型
func GetBinopExpType(exp *ast.BinopExp) LuaType {
	oneType := GetExpType(exp.Exp1)
	if oneType != LuaTypeAll {
		return oneType
	}

	twoType := GetExpType(exp.Exp2)
	if twoType != LuaTypeAll {
		return twoType
	}

	return LuaTypeRefer
}

// GetUnopExpType 获取一元表达式的类型
func GetUnopExpType(exp *ast.UnopExp) LuaType {
	oneType := GetExpType(exp.Exp)
	if oneType != LuaTypeAll {
		return oneType
	}

	return LuaTypeRefer
}

// GetMakeTableConstructorExp 获取构造的时候的TableConstructorExp
func GetMakeTableConstructorExp(valExp ast.Exp) (tableNode *ast.TableConstructorExp) {
	switch exp := valExp.(type) {
	case *ast.TableConstructorExp:
		// 构造的为table
		return exp
	case *ast.BinopExp:
		// 为二元表达式 a or {}
		// 为二元表达式 a or {}
		if exp.Op == lexer.TkOpOr || exp.Op == lexer.TkOpAdd {
			firstTalbeExp1 := GetMakeTableConstructorExp(exp.Exp1)
			if firstTalbeExp1 != nil {
				return firstTalbeExp1
			}

			firstTalbeExp2 := GetMakeTableConstructorExp(exp.Exp2)
			if firstTalbeExp2 != nil {
				return firstTalbeExp2
			}
		}

		return nil
	}

	return nil
}

// GetTablePrefixLoc 获取table prefix的位置信息
func GetTablePrefixLoc(node *ast.TableAccessExp) (loc lexer.Location) {
	loc = node.Loc
	switch exp := node.PrefixExp.(type) {
	case *ast.StringExp:
		loc = exp.Loc
	case *ast.NameExp:
		loc = exp.Loc
	case *ast.ParensExp:
		loc = exp.Loc
	}

	return loc
}

// GetTableKeyLoc 获取table key的位置信息
func GetTableKeyLoc(node *ast.TableAccessExp) (loc lexer.Location) {
	loc = node.Loc
	switch exp := node.KeyExp.(type) {
	case *ast.StringExp:
		loc = exp.Loc
	case *ast.NameExp:
		loc = exp.Loc
	case *ast.ParensExp:
		loc = exp.Loc
	}

	return loc
}

// GetExpLoc 获取table key的位置信息
func GetExpLoc(node ast.Exp) (loc lexer.Location) {
	switch exp := node.(type) {
	case *ast.StringExp:
		loc = exp.Loc
	case *ast.NameExp:
		loc = exp.Loc
	case *ast.ParensExp:
		loc = exp.Loc
	case *ast.FuncDefExp:
		loc = exp.Loc
	case *ast.TableConstructorExp:
		loc = exp.Loc
	case *ast.BinopExp:
		loc = exp.Loc
	case *ast.UnopExp:
		loc = exp.Loc
	case *ast.VarargExp:
		loc = exp.Loc
	case *ast.TableAccessExp:
		loc = exp.Loc
	case *ast.FuncCallExp:
		loc = exp.Loc
	case *ast.TrueExp:
		loc = exp.Loc
	case *ast.FalseExp:
		loc = exp.Loc
	case *ast.FloatExp:
		loc = exp.Loc
	case *ast.IntegerExp:
		loc = exp.Loc
	}

	return loc
}

// ChangeFuncSelfToReferVar 冒号 函数，self语法进行转换
// 判断是否为这样的在冒号函数内, self.b 这样的 self要进行转换为b，统一起来
// a = {}
// function a:test1()
//	 self.b = 3  -- 传人的为self.b
// end
func ChangeFuncSelfToReferVar(fi *FuncInfo, varStruct *DefineVarStruct) {
	firstColonFunc := fi.FindFirstColonFunc()
	if firstColonFunc == nil {
		return
	}
	if !firstColonFunc.IsColon {
		return
	}

	if firstColonFunc.RelateVar == nil {
		return
	}

	if len(varStruct.StrVec) < 1 {
		return
	}

	if varStruct.StrVec[0] == "self" {
		// self进行转换
		//varStruct.StrVec[0] = firstColonFunc.RelateVar.StrName
		strArray := strings.Split(firstColonFunc.RelateVar.StrName, ".")
		if len(strArray) == 1 {
			varStruct.StrVec[0] = firstColonFunc.RelateVar.StrName
		} else {
			strArray = append(strArray, varStruct.StrVec[1:]...)
			varStruct.StrVec = strArray
		}
		// 通过 scope 寻找 table name 定义的位置，肯定在 ':' 函数之前
		relateVar, _ := firstColonFunc.MainScope.FindLocVar(varStruct.StrVec[0], firstColonFunc.Loc)
		if relateVar != nil {
			varStruct.PosLine = relateVar.Loc.StartLine
			varStruct.PosCh = relateVar.Loc.StartColumn
		}
	}
}

// ChangeSelfToVarComplete 冒号 函数，self语法进行转换
func ChangeSelfToVarComplete(fi *FuncInfo, completeVar *CompleteVarStruct) {
	firstColonFunc := fi.FindFirstColonFunc()
	if firstColonFunc == nil {
		return
	}
	if !firstColonFunc.IsColon {
		return
	}

	if firstColonFunc.RelateVar == nil {
		return
	}

	if (*completeVar).StrVec[0] == "self" {
		strArray := strings.Split(firstColonFunc.RelateVar.StrName, ".")
		if len(strArray) == 1 {
			(*completeVar).StrVec[0] = firstColonFunc.RelateVar.StrName
		} else {
			strArray = append(strArray, completeVar.StrVec[1:]...)
			completeVar.StrVec = strArray
		}
	}
}

// IsCompleteNeedShow 判断查找的字符串是否满足显示, 主要判断是否包含需要过滤的字符串
func IsCompleteNeedShow(strName string, completeVar *CompleteVarStruct) bool {
	if !completeVar.FilterCharacterFlag {
		return true
	}

	for _, v := range strName {
		if v == completeVar.FilterOneChar || v == completeVar.FilterTwoChar {
			return true
		}
	}

	return false

}

// GetSubMapStrKey 查找globalInfo，是否SubMaps包含指定的strTableKey
func GetSubMapStrKey(subMaps map[string]*VarInfo, strName string, strTableKey string, line int, charactor int) (firstStr string, secondStr string) {
	if subMaps == nil {
		return "", ""
	}

	for strOne, subVar := range subMaps {
		findLoc := subVar.Loc
		validFlag := true

		if strOne != strTableKey {
			validFlag = false
		}
		if findLoc.StartLine != line || findLoc.EndLine != line {
			validFlag = false
		}
		if findLoc.EndColumn+1 < charactor {
			validFlag = false
		}
		if findLoc.StartColumn >= 1 && findLoc.StartColumn-1 >= charactor {
			validFlag = false
		}

		if validFlag {
			//log.Debug("strOne=%s", strName)
			return strName, ""
		}

		// 判断这个strTableKey，是否为指向的三级key，例如下面的c就是三级的key，前面的两级为a.b
		// 例如 a = {
		//			b = {
		//				c = 1
		//			}
		//      }
		// 不太像三级的table key

		tableVec := GetTableExpKeyStrVec(subVar.ReferExp)
		for _, oneKey := range tableVec {
			findLoc := oneKey.Loc
			validFlag = true
			if oneKey.StrKey != strTableKey {
				validFlag = false
			}
			if findLoc.StartLine != line || findLoc.EndLine != line {
				validFlag = false
			}
			if findLoc.EndColumn+1 < charactor {
				validFlag = false
			}
			if findLoc.StartColumn >= 1 && findLoc.StartColumn-1 >= charactor {
				validFlag = false
			}

			if validFlag {
				return strName, strOne
			}
		}
	}

	return "", ""
}

// GetVarSubGlobalVar 获取变量的子成员变量
func GetVarSubGlobalVar(varInfo *VarInfo, strSubkey string) (subVar *VarInfo) {
	if varInfo == nil {
		return
	}

	// 判断是否有效的
	subMaps := varInfo.SubMaps
	if subMaps == nil {
		return
	}

	findVar := subMaps[strSubkey]
	subVar = findVar
	return subVar
}

// IsSameExp 判断两个Exp是否指向的一样
func IsSameExp(oneExp ast.Exp, twoExp ast.Exp) bool {
	switch exp1 := oneExp.(type) {
	case *ast.NilExp:
		exp2, ok := twoExp.(*ast.NilExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.FalseExp:
		exp2, ok := twoExp.(*ast.FalseExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.TrueExp:
		exp2, ok := twoExp.(*ast.TrueExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.IntegerExp:
		exp2, ok := twoExp.(*ast.IntegerExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.FloatExp:
		exp2, ok := twoExp.(*ast.FloatExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.StringExp:
		exp2, ok := twoExp.(*ast.StringExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.ParensExp:
		exp2, ok := twoExp.(*ast.ParensExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.VarargExp:
		exp2, ok := twoExp.(*ast.VarargExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.NameExp:
		exp2, ok := twoExp.(*ast.NameExp)
		if !ok {
			return false
		}

		return exp1 == exp2

	case *ast.FuncDefExp:
		exp2, ok := twoExp.(*ast.FuncDefExp)
		if !ok {
			return false
		}

		return exp1 == exp2

	case *ast.TableConstructorExp:
		exp2, ok := twoExp.(*ast.TableConstructorExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.UnopExp:
		exp2, ok := twoExp.(*ast.UnopExp)
		if !ok {
			return false
		}

		return exp1 == exp2

	case *ast.BinopExp:

		exp2, ok := twoExp.(*ast.BinopExp)
		if !ok {
			return false
		}

		return exp1 == exp2

	case *ast.TableAccessExp:
		exp2, ok := twoExp.(*ast.TableAccessExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	case *ast.FuncCallExp:
		exp2, ok := twoExp.(*ast.FuncCallExp)
		if !ok {
			return false
		}

		return exp1 == exp2
	}

	return true
}

// IsReferExpEmpty 判断赋值的Exp是否指向的ReferExp为empty
// leftExp 为赋值表达式左边的变量
// rightExp 为赋值表达式右边的变量
func IsReferExpEmpty(leftExp ast.Exp, rightExp ast.Exp, ignoreLeft bool) bool {
	// 首先leftExp是否为简单的字符串，或是_G.a，例如 a 或_G.a
	leftStr := GetExpName(leftExp)
	if !ignoreLeft {
		if GetExpSubKey(leftStr) == "" {
			return false
		}
	}

	switch exp1 := rightExp.(type) {
	case *ast.NilExp:
		// 1) 如果右边直接为nil，直接返回true
		return true

	case *ast.TableConstructorExp:
		// 2) 如果右边直接为{}, 直接返回true
		if len(exp1.KeyExps) == 0 && len(exp1.ValExps) == 0 {
			return true
		}

	case *ast.BinopExp:
		// 3) 如果整个表达式为 a = a or {} 或是 a = a or nil, 返回true
		if exp1.Op != lexer.TkOpOr {
			return false
		}
		oneExp := exp1.Exp1
		twoExp := exp1.Exp2

		oneStr := GetExpName(oneExp)
		if leftStr != oneStr {
			return false
		}

		switch exp2 := twoExp.(type) {
		case *ast.NilExp:
			// 1) 如果右边直接为nil，直接返回true
			return true

		case *ast.TableConstructorExp:
			// 2) 如果右边直接为{}, 直接返回true
			if len(exp2.KeyExps) == 0 && len(exp2.ValExps) == 0 {
				return true
			}
		}

		return false
	}

	return false
}

// IsLocalReferExpEmpty 判断局部变量的定义的时候，定义的Exp是否指向的ReferExp为empty
// strName 为局部变量定义的名字
// rightExp 为赋值表达式右边的变量
func IsLocalReferExpEmpty(strName string, rightExp ast.Exp) bool {
	switch exp1 := rightExp.(type) {
	case *ast.NilExp:
		// 1) 如果右边直接为nil，直接返回true
		return true

	case *ast.TableConstructorExp:
		// 2) 如果右边直接为{}, 直接返回true
		if len(exp1.KeyExps) == 0 && len(exp1.ValExps) == 0 {
			return true
		}

	case *ast.BinopExp:
		// 3) 如果整个表达式为 a = a or {} 或是 a = a or nil, 返回true
		if exp1.Op != lexer.TkOpOr {
			return false
		}
		oneExp := exp1.Exp1
		twoExp := exp1.Exp2

		oneStr := GetExpName(oneExp)
		if GetSimpleValue(oneStr) != strName {
			return false
		}

		switch exp2 := twoExp.(type) {
		case *ast.NilExp:
			// 1) 如果右边直接为nil，直接返回true
			return true

		case *ast.TableConstructorExp:
			// 2) 如果右边直接为{}, 直接返回true
			if len(exp2.KeyExps) == 0 && len(exp2.ValExps) == 0 {
				return true
			}
		}

		return false
	}

	return false
}

// CompleteFilePathToPreStr 给定的文件名，截取前面.部分的字符串， 例如 test.lua ，返回test
func CompleteFilePathToPreStr(pathFile string) (preStr string) {
	// 完整路径提前前缀
	// 字符串中，查找第一个.
	seperateIndex := strings.Index(pathFile, ".")
	if seperateIndex < 0 {
		return ""
	}

	preStr = pathFile[0:seperateIndex]
	return preStr
}

// JudgeReferSuffixFlag 判断引入其他文件的方式，是不要引入的路径要包含其他文件的后缀
func JudgeReferSuffixFlag(referType ReferType, referTypeStr string) bool {
	if referType == ReferTypeDofile || referType == ReferTypeLoadfile {
		return true
	}

	if referType == ReferTypeRequire {
		return false
	}

	if referType != ReferTypeFrame {
		return true
	}

	isSuffixFlag := GConfig.IsImportSuffixFlag(referTypeStr)
	return isSuffixFlag
}

// GetSimpleExpType 获取exp的简单可以推导的类型
func GetSimpleExpType(node ast.Exp) LuaType {
	switch exp := node.(type) {
	case *ast.NilExp:
		return LuaTypeNil
	case *ast.FalseExp:
		return LuaTypeBool
	case *ast.TrueExp:
		return LuaTypeBool
	case *ast.IntegerExp:
		return LuaTypeInter
	case *ast.FloatExp:
		return LuaTypeFloat
	case *ast.StringExp:
		return LuaTypeString
	case *ast.FuncDefExp:
		return LuaTypeAll
	case *ast.TableConstructorExp:
		return LuaTypeAll

	case *ast.BinopExp:
		opValue := exp.Op
		oneType := GetExpType(exp.Exp1)
		twoType := GetExpType(exp.Exp2)

		if opValue == lexer.TkOpOr || opValue == lexer.TkOpAdd {
			if oneType == LuaTypeAll || oneType == LuaTypeRefer {
				return LuaTypeAll
			}

			if twoType == LuaTypeAll || twoType == LuaTypeRefer {
				return LuaTypeAll
			}
		}

		if oneType != LuaTypeAll && oneType != LuaTypeRefer {
			return oneType
		}

		if twoType != LuaTypeAll && twoType != LuaTypeRefer {
			return twoType
		}

		if opValue == lexer.TkOpAdd || opValue == lexer.TkOpSub || opValue == lexer.TkOpMinus ||
			opValue == lexer.TkOpMul || opValue == lexer.TkOpDiv || opValue == lexer.TkOpMod ||
			opValue == lexer.TkOpShr || opValue == lexer.TkOpShl || opValue == lexer.TkOpIdiv ||
			opValue == lexer.TkOpPow || opValue == lexer.TkOpWave {
			return LuaTypeFloat
		}

		if opValue == lexer.TkOpConcat {
			return LuaTypeString
		}

		return LuaTypeRefer

	case *ast.UnopExp:
		oneType := GetExpType(exp.Exp)
		if oneType != LuaTypeAll {
			return oneType
		}

		if exp.Op == lexer.TkOpNen || exp.Op == lexer.TkOpUnm {
			return LuaTypeInter
		}

		if exp.Op == lexer.TkOpNot {
			return LuaTypeBool
		}
	case *ast.NameExp:
		return LuaTypeRefer
	case *ast.ParensExp:
		return GetSimpleExpType(exp.Exp)
	}

	return LuaTypeAll
}
