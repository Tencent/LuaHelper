# Change Log

## 0.2.21 (Mar 7, 2023)
+ 修复了部分用户反馈的bug
+ 增加了类型的检查
+ 后台二进制采用go1.20.1版本编译
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.19 (July 29, 2022)
+ 合入外网提交的todo highlight on multi dashes(5m1Ly)
+ 修复允许设置对大表的分析 ([issues:112](https://github.com/Tencent/LuaHelper/issues/112))
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.17 (July 19, 2022)
+ 针对大工程require引起卡顿问题，进行了优化
+ 优化了self的推导([issues:96](https://github.com/Tencent/LuaHelper/issues/96))
+ 增强了函数参数不匹配的校验
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.16 (June 1, 2022)
+ 增加了试验的类型告警项
+ 优化了self的推导([issues:96](https://github.com/Tencent/LuaHelper/issues/96))
+ 修复了外网一处字符串崩溃([issues:102](https://github.com/Tencent/LuaHelper/issues/102))
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.15 (Apr 9, 2022)
+ 对table的hover继续进行了优化
+ 修复了外网一处死循环崩溃
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.14 (Apr 6, 2022)
+ 修复了外网一处崩溃
+ 对table的hover展示进行了优化
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.12 (Mar 29, 2022)
+ 后台二进制采用go1.18版本编译
+ 优化了部分推导的类型
+ 增加了一个设置开关
+ 修复了外网的崩溃([issues:88](https://github.com/Tencent/LuaHelper/issues/88))
+ 优化了插件被激活的情况([issues:89](https://github.com/Tencent/LuaHelper/issues/89))
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.11 (Mar 7, 2022)
+ 注解功能增加了alias功能
+ 增加显示了函数的参数类型与返回值
+ 优化了基于历史记录的代码补全功能
+ 修复了外网的崩溃([issues:80](https://github.com/Tencent/LuaHelper/issues/80))
+ 正式环境移除了pprof性能监控([issues:82](https://github.com/Tencent/LuaHelper/issues/82))
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.10 (Dec 14, 2021)
+ 优化了多文件扩展全局table(感谢：ggshily的合入)
+ 增强了对原表的支持([issues:61](https://github.com/Tencent/LuaHelper/issues/61))
+ 优化了对function的snippet提示
+ 新增了几类告警
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.9 (Dec 3, 2021)
+ 修复require中的路径跳转功能失效的bug
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.8 (Dec 2, 2021)
+ snippet关键词进行了优化
+ 增强了# 特殊字符串的代码补全
+ 代码补全提示table时，会展开table的成员信息
+ hover一个table时，会详细信息table的信息
+ 函数参数个数不匹配告警进行了优化
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.7 (Dec 1, 2021)
+ define查找文件路径时进行了优化
+ 增强了函数参数个数不匹配的检测
+ 合并外网Pull Request
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.6 (Nov 23, 2021)
+ require语法功能增强
+ 增加require其他lua文件的路径分隔符
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.2.5 (Nov 19, 2021)
+ 修复Linux平台下格式化失效的bug
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.4 (Nov 16, 2021)
+ 修复外网Issues47 https://github.com/Tencent/LuaHelper/issues/47
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.3 (Nov 16, 2021)
+ 修复外网Issues46 https://github.com/Tencent/LuaHelper/issues/46
+ 修复外网Issues47 https://github.com/Tencent/LuaHelper/issues/47
+ 修复了textDocument/documentSymbol 显示函数位置信息不准确的bug
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.2 (Nov 12, 2021)
+ 修复外网Issues42 https://github.com/Tencent/LuaHelper/issues/42
+ 修复外网Issues44 https://github.com/Tencent/LuaHelper/issues/44
+ 完善了require语法
+ [Github](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
  
## 0.2.1 (Nov 10, 2021)
+ 正式对外开发，发布0.2.1版本
+ [Repository](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper

## 0.1.66 (Nov 8, 2021)
+ 修复了VSCode图片链接地址
  
## 0.1.65 (Oct 28, 2021)
+ 修复Linux平台下插件不能运行的bug
  
## 0.1.64 (Oct 27, 2021)
+ 修复插件不能与其他插件共存的bug

## 0.1.63 (Oct 27, 2021)
+ 增加lua5.1、lua5.2、lua5.4调试功能
+ 添加插件格式化的配置
+ 添加设置忽略文件与文件夹功能

## 0.1.61 (Oct 22, 2021)
+ 更新icon图标

## 0.1.60 (Oct 22, 2021)
+ 后台修复一处崩溃

## 0.1.58 (Jul 23, 2021)
+ 解决LuaJit 整型词法解析的bug

## 0.1.57 (Jul 8, 2021)
+ 修复全局_ENV环境表

## 0.1.56 (Jul 6, 2021)
+ 对系统模块的提示和补全优化

## 0.1.55 (June 11, 2021)
+ 修复了格式化bug

## 0.1.54 (June 8, 2021)
+ 优化了大工程检查的性能
+ 代码补全、hover预留展示优化
+ 注解颜色进行了优化


## 0.1.52 (Apr 29, 2021)
+ 修复查找引用的bug
+ 修复self关键字的bug


## 0.1.51 (Apr 12, 2021)
+ 修复查找定义的不准确的bug
+ 增加注解系统的颜色


## 0.1.50 (Apr 8, 2021)
+ 修复了for语句的注解bug


## 0.1.49 (Apr 8, 2021)
+ 对注解功能进行了增强
+ 去掉了注解变量的着色
+ 修复了一处崩溃


## 0.1.48 (Feb 16, 2021)
+ 增强了关键字thenend、doend代码段补全
+ 修复了一处注解bug

## 0.1.47 (Feb 13, 2021)
+ 修复插件一处崩溃
+ 优化注解alias功能

## 0.1.46 (Dec 22, 2020)
+ 优化self变量的代码补全
+ 优化增量更新模块
+ 函数参数个数检查bug修复

## 0.1.45 (Dec 22, 2020)
+ 修复部分情况下代码跳转bug
+ 修复增量更新代码时，增加检测的bug

## 0.1.44 (Dec 14, 2020)
+ 修复插件的两处崩溃

## 0.1.43 (Dec 14, 2020)
+ 注解功能支持配置文件自动推导类型完善

## 0.1.42 (Dec 12, 2020)
+ 修复全局变量着色的bug

## 0.1.41 (Dec 9, 2020)
+ 修复了gbk编码导致的代码提示乱码
+ 注解功能支持配置文件自动推导类型

## 0.1.40 (Dec 3, 2020)
+ 修复了注解空字符串引起的崩溃
+ 优化了代码补全提示速度

## 0.1.39 (Nov 27, 2020)
+ 修复：优化了路径自动补全的问题
+ 注解针对客户端的使用，进行了定制的支持


## 0.1.38 (Nov 23, 2020)
+ 修复：语法无法跳转和无法代码补全bug

## 0.1.37 (Nov 20, 2020)
+ 全局文档查找，支持模糊查找
+ 底层代码补全、查找定义进一步完善优化
+ 注解支持：冒号写法

## 0.1.36 (Nov 17, 2020)
+ 查找定义：为打开文件时进行优化
+ 查找其他的文件bug进行修复


## 0.1.35 (Nov 16, 2020)
+ 代码补全时，lua路径提示支持模糊匹配


## 0.1.34 (Nov 16, 2020)
+ 注解类型支持以数字开头的名称
+ lua函数返回值类型推导增强

## 0.1.33 (Nov 13, 2020)
+ 注解提示fun时候，显示函数颜色
+ 修复了一处崩溃

## 0.1.32 (Nov 13, 2020)
+ 支持多目录工程
+ 修复了部分注解的bug
+ 增强了快速输入注解的功能
+ lua原表支持增强


## 0.1.31 (Oct 17, 2020)
+ 去掉了默认插件加载的耗时显示
+ 修复了注解table class的bug
  
## 0.1.30 (Oct 15, 2020)
+ 修复_G前缀的lua table无法提示子成员的bug


## 0.1.29 (Oct 15, 2020)
+ 增强多层table的定义
+ 修复单文档内的符号重复问题


## 0.1.28 (Oct 12, 2020)
+ 注解类型系统功能增强
+ 一处崩溃修复

## 0.1.27 (Oct 9, 2020)
+ 注解类型系统进一步完善，增强了提示
+ 注解类型系统增加了告警


## 0.1.26 (Oct 4, 2020)
+ 原表代码提示进行了增加
+ 修复了诊断告警的bug
+ 完善了统计信息

## 0.1.25 (Oct 3, 2020)
+ 注解系统进行了完善
+ 代码提示进行优化
+ 修复了崩溃

## 0.1.22 (Set 24, 2020)
+ 完善补全输入引入文件的路径提示
+ 代码补全性能优化
+ 增加部分注解功能

## 0.1.21 (Set 12, 2020)
+ 修复vscode升级到1.49.0版本后，导致的诊断错误无法清除bug

## 0.1.20 (Set 11, 2020)
+ 支持lua文件拖动到工程目录中
+ 修复TextDocumentHighlight着色功能的bug
+ 支持lua5.4的语法
+ 代码整理及优化

## 0.1.18 (August 19, 2020)
+ windows平台下，目录存在链接接问题修复

## 0.1.17 (August 19, 2020)
+ linux平台下，目录存在链接问题修复

## 0.1.16 (August 18, 2020)
+ 增加了rename批量替换变量的功能
+ 增加了TextDocumentHighlight选择单纯变色功能
+ 优化当前文档查找符号功能
+ 优化全局符号搜索功能
+ 增强了linux平台下，目录存在外链接的处理
+ 优化部分告警

## 0.1.14 (July 28, 2020)
+ 增强了引入lua文件的方式
+ 修复配置忽略的文件无法生效的bug

## 0.1.12 (July 21, 2020)
+ 代码格式化调整
+ 代码补全遇到函数返回值优化

## 0.1.11 (July 15, 2020)
+ 新增了其他类型的文件关联成.lua类型的功能
+ 新增了快捷生成函数注释
+ 新增了提示函数参数时候，分别提示函数参数的信息

## 0.1.10 (July 9, 2020)
+ 完善了局部变量定义了，未使用的告警
+ 修复了关键字not没有着色的问题

## 0.1.8 (July 8, 2020)
+ 集成了LuaPanda的调试组件
+ 新增了or表达式永远为true的告警
+ 新增了and表达式永远为false的告警
+ 开启了局部变量定义了未使用的告警
+ 增强了查找定义的实现
+ 修复了部分用户反馈的bug

## 0.1.6 (June 12, 2020)
+ 变量的定义能递归跟踪到引用的其他变量
+ 查找引用功能增强了
+ 代码补全增强了功能，且修复了一个小bug

## 0.1.4 (June 2, 2020)
+ 修复了一些小bug

## 0.1.3 (June 2, 2020)
+ 没有项目配置文件时，增加默认的检查项

## 0.1.2 (May 29, 2020)

+ 增加require引用一个文件，查找定义
+ 增加require引用一个文件，查找引用
+ 增加require引用一个文件，代码智能提示

## 0.1.1 (May 27, 2020)

+ 完善了文档
+ 修复了全局变量着色的bug


## 0.1.0 (May 22, 2020)

+ 实现了代码编辑辅助等功能
+ 实现了工程级代码检查功能