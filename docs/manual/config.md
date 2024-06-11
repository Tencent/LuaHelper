[TOC]
# 代码检查配置

## 代码检查种类
### 1 语法错误
告警类型：1, 提示前缀 [Warn type:1]</br>
指不满足lua的语法，例如if 语句没有匹配的then等等。
```lua
local a = 1
if a            -- 没有匹配的then，语法错误
end
```

### 2 变量未找到定义
告警类型：2, 提示前缀 [Warn type:2]</br>
lua是比较灵活的语言，使用变量之前没有定义的概念，未定义的变量默认为nil值。下面的代码段，在lua语法内是合法的。

```lua
local a = 1
local b = a1 + 3  --al之前没有定义，这里会为nil，这里会进行告警
print(b)
```
但是上面写法，会容易出现笔误，原本是希望输入a，但是输入了a1。代码在运行期间为报错。变量未找到定义检查，上面会提示a1未找到定义，进行告警。
### 3 全局变量先使用，后定义
告警类型：3, 提示前缀 [Warn type:3]</br>
全局变量在项目中也会出现使用在前，定义在后。

```lua
print(a)    --先使用，这里会告警
a = 1       --后定义
```
### 4 局部变量定义了，未使用
告警类型：4, 提示前缀 [Warn type:4]</br>
```lua
local a = 1 --这里定义了a局部变量，但是后面没有使用到，告警
```
### 5 table定义构造中有重复的key
告警类型：5, 提示前缀 [Warn type:5]
```lua
local a = {
    b = 1,  --首次定义了成员b
    b = 2,  --再次定义了成员b，重复了，告警
}
```
### 6 加载其他的lua文件，未找到文件
告警类型：6, 提示前缀 [Warn type:6]
```lua
local a = require("test") --目录中不存在test.lua或是test库文件，进行告警
```
### 7 赋值语句参数个数不匹配
告警类型：7, 提示前缀 [Warn type:7]
```lua
local a = 1
local b = 2
local c
c = a, b   -- 赋值语句左边只有一个变量，右边赋值了两个值，右边的个数大于左边的，进行告警
```

### 8 局部变量定义参数个数不匹配
告警类型：8, 提示前缀 [Warn type:8]
```lua
local a = 1
local b = 2
local c = a, b   -- 局部变量定义左右只有一个变量，右边赋值了两个值，右边的个数大于左边的，进行告警
```
### 9 goto用法未找到对应的lable标记
告警类型：9, 提示前缀 [Warn type:9]
```lua
local array = {1, 2, 3}
for k, v in pairs(array) do
    if v == 1 then
        goto cont       -- 没有匹配的cont标签值，只有continue，进行告警
    end

    ::continue::
end
```                
### 10 函数调用参数个数大于定义参数的个数
告警类型：10, 提示前缀 [Warn type:10]
```lua
-- add two value
function calcAdd(one, two)
    print(one + two)
end

calcAdd(1, 2, 3)      -- 函数只定义了两个参数，这里调用的参数有三个，进行告警
``` 
### 11 import其他的lua文件，成员变量未定义
告警类型：11, 提示前缀 [Warn type:11]</br>
这个是项目组特有的引入了hive框架，封装import引用另外一个lua文件
```lua
local test = import("test.lua") -- hive框架，import引入了test.lua文件
test.CalcTest() -- 若test.lua不存在全局变量函数 CalcTest，进行告警
``` 
### 12 if not包含的代码块有误
告警类型：12, 提示前缀 [Warn type:12]</br>
```lua
if not ss then
    print(ss.name)  -- ss这里判断为 nil，调用 name成员，进行告警
end
``` 
### 13 函数定义的参数是否重复
告警类型：13, 提示前缀 [Warn type:13]</br>
```lua
function CalcAdd(one, one) -- 函数定义的重复的参数one，进行告警
    print(one)
end
``` 
### 14 二元表达式，左右两边的表达式是否一样
告警类型：14, 提示前缀 [Warn type:14]</br>
当调用or、and、<、<=、>、>=、==、~= 二元表达式，左右两边一样进行告警
```lua
local a = 1
local b = 1
local c = a and a  -- and表达式左右两边一样，都是 a，进行告警
``` 
### 15 or表达式始终为true
告警类型：15, 提示前缀 [Warn type:15]</br>
例如下面的例子：
```lua
local a = 1
a = a or true     -- or表达式右边包含true，表达式结果始终为true
``` 

### 16 and表达式始终为false
告警类型：16, 提示前缀 [Warn type:16]</br>
例如下面的例子：
```lua
local a = 1
a = a and false   -- and表达式右边包含false，表达式结果始终为false
``` 

## 代码检查配置文件
### 配置文件说明
由于Lua需要调用到C或是其他语言导入的符号，这些导入的符号是未定义的，因此需要忽略这些符号的告警。有时，也需要屏蔽分析的文件夹或文件，忽略指定的文件的告警等，这些都需要特定的配置文件。

配置文件的名字是固定的：luahelper.json，在vscode工程目录下。这样有利于一个项目组共用一份相同的配置文件，配置文件也放在项目的git或是svn管理。
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/json.png)

配置文件的格式是json的，文件名为：luahelper.json，完整的配置格式如下：
```json
{
    "BaseDir":"./",
    "ShowWarnFlag":          1,
    "ReferMatchPathFlag":    0,
    "IgnoreFileNameVarFlag": 0,
    "ProjectFiles":[
    
    ],
    "IgnoreModules":["hive", "import"],
    "IgnoreFileVars": [
        {
            "File": "port/gm.lua",
            "Vars": ["bb", "cc"]
        },
         {
            "File": "port/ss.lua",
            "Vars": ["aa", "dd"]
        }
    ],
    "IgnoreReadFiles": [
        "config/port.lua",
		"config/act.lua"
    ],
    "IgnoreErrorTypes": [
        4
	],
	"IgnoreFileOrFolder": [
		"port/on.*lua",
		"tests/",
		"one.lua"
	],
    "IgnoreFileErr": [
       "port/add.lua",
       "common/test.lua"
	],
    "IgnoreFileErrTypes": [
        {
            "File": "port/bbb.lua",
            "Types": [4]
        },
        {
            "File": "port/ss.lua",
            "Types": [4, 5]
        }
    ],
    "ProtocolVars": []
}
``` 

每一项配置的具体说明
* "BaseDir":"./", </br>
表示加载的的lua文件目录，这个是相对目录，相对于配置文件。 </br>
按需要配置成所需要的，笔者的项目中，配置的为："BaseDir":"./bin/",

* "ShowWarnFlag":          1,</br>
表示是否开启所有检测告警，1为开启，0为关闭


* "ReferMatchPathFlag":    0,</br>
表示引入其他文件时，是否完整匹配路径，1为是，0为否</br>
笔者后台项目中，该值设置为1，默认都设置为0即可。

* "IgnoreFileNameVarFlag": 0,</br>
表示是否忽略lua文件同名的变量，1为是，0为否</br>
笔者所在的客户端组中，例如有test.lua文件，由于框架的原因，需要忽略test变量。项目中文件较多，需要忽略的变量也较多，因此设置了一个开关。默认设置为0即可。

* "ProjectFiles":[],</br>
表示项目入口文件，默认设置为空</br>
笔者所在的后台项目组中，检查的粒度更加严格，有项目的概念，该入口文件类似于C++工程的main.cpp文件，通过该文件会按顺序加载其他的lua文件。

* "IgnoreModules":["hive", "import"],</br>
为检查项目中，整体忽略未定义的变量，这里按需求配置。</br>
lua工程经常需要调用C++等其他宿主语言导入的符号，这里需要忽略指定的符号。</br>
上面的含义：是整体上忽略"hive"和"import"变量找不到定义的告警（告警类型：2）。

* IgnoreFileVars:[],
    ```json
    "IgnoreFileVars":[
        {
            "File": "port/gm.lua",
            "Vars": ["bb", "cc"]
        },
         {
            "File": "port/ss.lua",
            "Vars": ["aa", "dd"]
        }
    ],
    ``` 
  指定的文件中，忽略指定的未定义变量。上面的配置的含义是：</br>
  "port/gm.lua"文件, 忽略找不到"bb"和"cc"变量定义的告警（告警类型：2）。</br>
  "port/ss.lua"文件, 忽略找不到"aa"和"dd"变量定义的告警（告警类型：2）。
  
* IgnoreReadFiles:[],
    ```json
    "IgnoreReadFiles":[
        "config/port.lua",
        "config/act.lua"
    ],
    ``` 
    忽略读取不到指定的文件告警（告警类型：6）。</br>
    上面的含义是，忽略读取不到config/port.lua和config/act.lua文件的告警。
    
* IgnoreErrorTypes:[4],</br>
  整体忽略指定类型的告警，可以忽略多项。</br>
  填写的值为整型，为上面的代码检查种类：1-14</br>
  例如上面填写的4，表示忽略告警类型4(局部变量定义了，未使用)。

* IgnoreFileOrFolder:[],
    ```json
    "IgnoreFileOrFloder": [
    	"port/on.*lua",
    	"tests/",
    	"one.lua"
    ],
    ``` 
    忽略分析指定的文件或文件夹，支持正则。忽略分析的文件，不会提供各种编辑辅助，也不会进行代码检查。</br>
    上面配置的含义为：</br>
    忽略分析port文件夹下，以on开头的lua文件</br>
    忽略分析tests文件夹</br>
    忽略分析one.lua文件</br>
    
* "IgnoreFileErr": [],
    ```json
    "IgnoreFileErr": [
        "port/add.lua",
        "common/test.lua"
    ],
    ``` 	
    忽略指定文件的所有告警。</br>
    上面的含义是: "port/add.lua"和"common/test.lua"文件所有告警将会忽略。

* "IgnoreFileErrTypes": [],
    ```json
    "IgnoreFileErrTypes": [
        {
            "File": "port/bbb.lua",
            "Types": [4]
        },
        {
            "File": "port/ss.lua",
            "Types": [4, 5]
        }
    ],
    ```
    忽略指定文件指定类型的告警，上面的含义是：</br>
    "port/bbb.lua"文件，忽略类型为：4的告警。</br>
    "port/ss.lua"文件，忽略类型为：4、5的告警。
    
* "ProtocolVars": []</br>
   为笔者后台项目定制的协议前缀提示，默认可以忽略。

* "ReferFrameFiles": []</br>
   为笔者后台项目定制。

* "PathSeparator": "."</br>
   当require或其他方式引入lua文件时，路径分割符，默认为"."
   ```lua
   local test = require("common.test")  -- 路径分隔符为., 实际对应的文件为：commone/test.lua
   local log = require("common/log")    -- 路径分隔符为., 实际对应的文件为：commone/log.lua
   ```
   
### 配置文件模板下载
#### 后台项目
  利用到了hive和import引入文件框架</br>
  [luahelper.json模板](./../jsonconfig/server/luahelper.json "后台luahelper.json")
#### 客户端项目
  [luahelper.json模板](./../jsonconfig/client/luahelper.json "客户端luahelper.json")
