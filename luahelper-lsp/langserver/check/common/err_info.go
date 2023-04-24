package common

// CheckErrorType 检查错误的类型
type CheckErrorType int

const (
	_ CheckErrorType = iota

	// CheckErrorSyntax 语法分析阶段出错，构造ast树有问题，无法继续进行,重大错误
	CheckErrorSyntax = 1

	// CheckErrorNoDefine 变量未找到的定义
	CheckErrorNoDefine = 2

	// CheckErrorCycleDefine 因循环依赖或是加载顺序导致的变量未定义
	CheckErrorCycleDefine = 3

	// CheckErrorLocalNoUse 定义了的局部变量未使用
	CheckErrorLocalNoUse = 4

	// CheckErrorTableDuplicateKey table中是否有重复的key
	CheckErrorTableDuplicateKey = 5

	// CheckErrorNoFile 指定加载文件，但是没有发现文件
	CheckErrorNoFile = 6

	// CheckErrorAssignParamNum 赋值语句时候，参数个数不匹配，例如 a,b=c,d,e；后面的赋值个数多余前面的
	CheckErrorAssignParamNum = 7

	// CheckErrorLocalParamNum 局部变量定义时候，参数个数不匹配，例如 local a = b,c；后面的参数多余前面的
	CheckErrorLocalParamNum = 8

	// CheckErrorGotoLabel  goto时找不到对应的label标记
	CheckErrorGotoLabel = 9

	// CheckErrorCallParam 函数调用参数数量出错
	CheckErrorCallParam = 10

	// CheckErrorImportVar 引用import其他lua文件，例如 local a = import("one.lua")；调用a.bbb， 其中bbb在引用的one.lua文件中未定义
	CheckErrorImportVar = 11

	// CheckErrorNotIfVar 前面 if not aa then，后面调用了aa.bb的变量
	CheckErrorNotIfVar = 12

	// CheckErrorDuplicateParam 函数定义的参数重复了
	CheckErrorDuplicateParam = 13

	// CheckErrorDuplicateExp 二元表达式 两边的表达式是否一样，一样进行告警
	CheckErrorDuplicateExp = 14

	// CheckErrorOrAlwaysTrue 二元表达式 or 永远为true，例如 a = a or true ，始终为true
	CheckErrorOrAlwaysTrue = 15

	// CheckErrorAndAlwaysFalse 二元表达式 and 永远为false，例如 a = a and false, 始终为false
	CheckErrorAndAlwaysFalse = 16

	// CheckErrorNoUseAssign 定义了的局部变量未使用, 后面只是简短的赋值
	CheckErrorNoUseAssign = 17

	// CheckErrorAnnotate 注解系统引入的错误
	CheckErrorAnnotate = 18

	// 分支条件重复
	CheckErrorDuplicateIf = 19

	// 变量自身赋值
	CheckErrorSelfAssign = 20

	// 浮点数做等于或不等于判断
	CheckErrorFloatEq = 21

	// 错误的成员变量
	CheckErrorClassField = 22

	// 给常量赋值
	CheckErrorConstAssign = 23

	// 函数调用参数类型出错
	CheckErrorCallParamType = 24

	// 函数返回值类型出错
	CheckErrorFuncRetErr = 25

	// 赋值表达式改变了类型
	CheckErrorAssignType = 26

	// 二元运算表达式类型不同
	CheckErrorBinopType = 27

	// local函数未调用(可选，由于lua中可以隐式调用函数，所以不一定准确，全局函数，rpc函数更是如此)
	CheckErrorLocFuncNotCall = 28

	// CheckErrorMax
	CheckErrorMax = 29
)
