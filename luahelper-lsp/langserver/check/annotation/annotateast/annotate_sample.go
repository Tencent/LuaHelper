package annotateast

// 分析注释功能

// 自带的type类型有下面的
// nil
// boolean
// number
// function
// userdata
// thread
// any
// void
// self

// 1) type 用法的完整格式
// ---@type MY_TYPE[|OTHER_TYPE] [@comment]
// 可能是MY_TYPE 也有可能是OTHER_TYPE
// Since lua is dynamic-typed, a variable may be of different types
// use | to list all possible types

// 描述多个的时候
// ---@type MY_TYPE[|OTHER_TYPE], MY_TYPE[|OTHER_TYPE] [@comment] [@comment]

// 2) class 的用法
// ---@class MY_TYPE[:PARENT_TYPE] [@comment]

// 多重继承的用法
// ---@class MY_TYPE[:PARENT_TYPE], MY_TYPE[:PARENT_TYPE] [@comment]

// 3) field 是class的成员，一般在@class的后面
// ---@field [public|protected|private] field_name FIELDLTYPE[|OTHER_TYPE] [@comment]
// class 允许有多个field成员，每个field成员占用一行

// 4) fun 类型（函数的类型）
// fun(param1:PARAM_TYPE1 [,param2:PARAM_TYPE2]): RETURN_TYPE1[, RETURN_TYPE2]

// 5) return 函数的返回值
// ---@return RETURN_TYPE[|OTHER_TYPE] [@comment1]
// 5.1) 多个返回值，每个返回值之间用，分隔
// ---@return RETURN_TYPE1[|OTHER_TYPE], RETURN_TYPE2[|OTHER_TYPE] [@comment1] [@comment2]

// 5.2）如果有多个返回值，也可以分多行表示
// ---@return RETURN_TYPE[|OTHER_TYPE] [@comment]
// ---@return RETURN_TYPE[|OTHER_TYPE] [@comment]

// 函数返回值的用法
// ---@param one string @参数1是表示string1
// ---@param two number @参数2是表示number1
// ---@return number, number, string
// function getRes(one, two) end

// 6) overload 函数重载的支持
/*
---@overload fun(list:table):string
---@overload fun(list:table, sep:string):string
---@overload fun(list:table, sep:string, i:number):string
---@param list table
---@param sep string
---@param i number
---@param j number
---@return string
function table.concat(list, sep, i, j) end
*/

// 7 @param 函数参数
//---@param param_name MY_TYPE[|other_type] [@comment]
// 7.1 例子
//---@param car Car
//function setCar(car) end

// 7.2 例子
//---@param car Car
//setCallback(function(car) end)

// 7.3 例子
//---@param car Car
//for k, car in ipairs(list) end

// 8 array type
//---@type MY_TYPE[] @comment
//二义性表达如下
//---@type string | number[]  表示类型为string或是number的列表
//---@type (string | number)[] 表示类型为string的列表或是number的列表
//---@type string[] | number[] 表示类型为string的列表或是number的列表

// 9 table type
//---@type table<KEY_TYPE[, KEY_OTHER_TYPE], VALUE_TYPE[, VAULE_OTHER_TYPE]>

// 10 alias
//---@alias NEW_NAME TYPE

//---@alias Handler fun(type: string, data: any):void
//---@param handler Handler
//function addHandler(handler)
//end

// 11 generic 泛型
// ---@generic T1 [: PARENT_TYPE] [, T2 [: PARENT_TYPE]]

//---@generic T : Transport, K
//---@param param1 T
//---@param param2 K
//---@return T
//function test(param1, param2)

// 12 enum 枚举段类型【一个枚举段中定义的变量的值不能重复】
//---@enum start @comment  表示枚举段的开始
//---@enum end @comment  表示枚举段的结束
