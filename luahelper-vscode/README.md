
# LuaHelper Guide

[![version](https://vsmarketplacebadges.dev/version-short/yinfei.luahelper.png)](https://marketplace.visualstudio.com/items?itemName=yinfei.luahelper)
[![rating](https://vsmarketplacebadges.dev/rating-short/yinfei.luahelper.png)](https://marketplace.visualstudio.com/items?itemName=yinfei.luahelper)
[![installs](https://vsmarketplacebadges.dev/installs-short/yinfei.luahelper.png)](https://marketplace.visualstudio.com/items?itemName=yinfei.luahelper)
[![GitHub stars](https://img.shields.io/github/stars/Tencent/LuaHelper.png?style=flat-square&label=github%20stars)](https://github.com/Tencent/LuaHelper)
[![GitHub Contributors](https://img.shields.io/github/contributors/Tencent/LuaHelper.png?style=flat-square)](https://github.com/Tencent/LuaHelper/graphs/contributors)

LuaHelper is a High-performance lua plugin, Language Server Protocol for lua.
- [Github:](https://github.com/Tencent/LuaHelper)  https://github.com/Tencent/LuaHelper
- [Issues:](https://github.com/Tencent/LuaHelper/issues)  https://github.com/Tencent/LuaHelper/issues

![logo](https://raw.githubusercontent.com/Tencent/LuaHelper/master/docs/images/logo.png)
## Introduction

Lua is very popular in game development because of its simple syntax and flexible use. However, its ecology is not perfect, and IDE  tools and support are few, which affects Lua's development efficiency and quality. LuaHelper complies with Microsoft Language Server Protocol and is a cross-platform Lua code editing and testing tool developed in go language.
Compared with other Lua plugins currently on the market, it has the following **improvements**:

- [X] 1.Coroutine development, real-time detection, millisecond output detection results
- [X] 2.Support large-scale Lua projects, perfectly support editing and testing of 1000+ file project 
- [X] 3.Comprehensive error type detection, including: grammar detection, semantic detection 
- [X] 4.Various types of reference search, including: multi-file reference search, multi-layer reference search 
- [X] 5.Rich configurable items, including: multiple alarm information configurations, ignorable file settings 
- [X] 6.Low memory consumption, low-performance machines can still run smoothly

--------------------------------------------------------------------------------------------------------------------
Lua因其语法简单、使用灵活，在游戏开发中十分流行。但其生态并不完善，IDE开发工具及配套支持较少，一定程度上影响了Lua的开发效率及质量。LuaHelper遵从微软Language Server Protocol协议，是采用go语言开发的一种跨平台Lua代码编辑及检测工具。

相较目前市面其他Lua插件，具有以下**改进**：

- [X] 1.协程开发，实时检测，毫秒级输出检测结果
- [X] 2.支持大型Lua项目，完美支持1000+文件项目工程的编辑与检测
- [X] 3.全面的错误类型检测，包括：语法检测、语义检测
- [X] 4.多种类引用查找，包括：多文件引用查找、多层引用查找
- [X] 5.丰富的可配置项，包括：多种告警信息配置、可忽略文件设定
- [X] 6.内存消耗低，低性能机器仍可流畅运行

## Documentation
* [MainPage [项目主页]](https://github.com/Tencent/LuaHelper "MainPage [项目主页]") |
[Background [项目背景]](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/introduction.md "Background [项目背景介绍]") 
* [Configuration [检查配置]](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/config.md "Configuration [检查配置]")
* [Debug Principle [调试原理]](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/debugPrinciple.md "Debug Principle [调试原理]") |
[Debug Use [接入调试方法]](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/usedebug.md "Debug Use [接入调试方法]") | [Debug and Run Single Lua File [单文件调试与运行]](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/debugsinglefile.md "Debug and Run Sigle Lua File [单文件调试与运行]")

## Feature Summary

### Code Editing
* [Defintion Find [定义跳转]](#DefintionFind)
* [Find All References [引用查找]](#FindAllReferences)
* [Document Symbols [文件符号表查询]](#DocumentSymbols)
* [Workspace Symbols [工程符号表查询]](#WorkspaceSymbols)
* [Auto Code Completion [自动代码补全]](#AutoCodeCompletion)
* [Format Code [代码格式化]](#FormatCode)
* [Hover [代码悬停]](#Hover)
* [Hightlight Global Var [全局变量着色]](#HightlightGlobalVar)

### Code Detection
* [Syntax Check [语法检测]](#SyntaxCheck)
* [Semantic Check [语义检测]](#SemanticCheck)
* [Quick Analysis [快速增量分析]](#QuickAnalysis)

### Debugger
* [Debug Attach [调试连接其他进程]](#DebugAttach)
* [Debug Single Lua File [调试单lua文件]](#DebugSingleLuaFile)
* [Run Single Lua File [运行单lua文件]](#RunSingleLuaFile)

## Feature Detail
###  <span id="DefintionFind">Defintion Find/定义跳转</span>
**支持局部、全局文件定义查询跳转**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/GotoDefinition.gif)

###  <span id="FindAllReferences">Find All References/引用查找</span>
**支持基于作用域的各类型引用查找**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/FindReferences.gif)

###  <span id="DocumentSymbols">Document Symbols/文件符号表查询</span>
**支持文件域符号表查询，在搜索栏输入`@`**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/DocmentSymbol.gif)

###  <span id="WorkspaceSymbols">Workspace Symbols/工程符号表查询</span>
**支持工程域符号表查询，在搜索栏输入`#`**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/WorkspaceSymbol.gif)

###  <span id="AutoCodeCompletion">Auto Code Completion/自动代码补全</span>
**支持变量、函数的自动输入提示**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/CodeCompletion.gif)

###  <span id="FormatCode">Format Code/代码格式化</span>
**支持代码格式化**</br>
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/Format.gif)

###  <span id="Hover">Hover/代码悬停</span>
**支持代码悬停提示**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/Hover.gif)

###  <span id="HightlightGlobalVar">Hightlight Global Var/全局变量着色</span>
**支持全局变量高亮着色**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/GlobalColor.gif)

###  <span id="SyntaxCheck">Syntax Check/语法检测</span>
**提供丰富的语法错误检测类型**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/SyntaxCheck.gif)

###  <span id="SemanticCheck">Semantic Check/语义检测</span>
**支持多种类型的语义检测**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/SemanticCheck.gif)

###  <span id="QuickAnalysis">Quick Analysis/快速增量分析</span>
**支持增量变化分析，分析结果诊断输出**
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/RealTimeCheck.gif)

###  <span id="DebugAttach">Debug Attach/调试连接其他进程</span>
[Debug Detail Use/调试详细使用](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/usedebug.md)</br>
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/debug/debugattach.png)

###  <span id="DebugSingleLuaFile">Debug Single Lua File/调试单lua文</span>
[Debug Single Lua File [调试单文件]](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/debugsinglefile.md "DebugSigle Lua File [调试单文件]")
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/debug/debugfilerun.png)

###  <span id="RunSingleLuaFile">Run Single Lua File/运行单lua文件</span>
[Run Single Lua File [运行单文件]](https://github.com/Tencent/LuaHelper/blob/master/docs/manual/debugsinglefile.md "Run Sigle Lua File [运行单文件]")
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/debug/runonefile.gif)

## Installation
**App Market Installation**
* Click the Vs Code application market icon 
* Search luahelper in the input box 
* Click to install Lua Helper

**应用市场安装**
* 点击Vs Code应用市场图标
* 在输入框中搜索 luahelper
* 点击安装Lua Helper

![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/Install.gif)

 -  若应有市场异常，点击[应用市场](https://marketplace.visualstudio.com/items?itemName=yinfei.luahelper&ssr=false#overview)，尝试重新安装

## Acknowledgements
* [luago-books](https://github.com/zxh0/luago-book), go语言生成lua的AST，修改了源码（对AST的每个节点增加了列的属性，同时也优化了性能）。
* [LuaFormatter](https://github.com/Koihik/LuaFormatter), c++写的Lua代码格式化库，性能较高。
* [LuaPanda](https://github.com/Tencent/LuaPanda), 集成了LuaPanda的调试组件，LuaPanda的作者stuartwang也给我们提供了很多帮助。
* [EmmyLua](https://github.com/EmmyLua), 作者阿唐对我们整个插件的实现提供很多帮助和建议。

## Developer
 [yinfei](https://github.com/yinfei8),  [handsomeli](https://github.com/badboylikeit), richardzha

## Contribution
 [nattygui](https://github.com/nattygui)

## Support
If you have any questions, please refer to [FAQ](#FAQ). If you have any questions, please use [issues](https://github.com/Tencent/LuaHelper/issues). We will follow and reply.
如有问题先参阅 [FAQ](./docs/manual/FAQ.md) ，如有问题建议使用 [issues](https://github.com/Tencent/LuaHelper/issues) ，我们会关注和回复。

Email：yvanfyin@tencent.com; handsomeli@tencent.com; richardzha@tencent.com

QQ群：747590892