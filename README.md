# Lua Helper Guide

## Introduction

Lua因其语法简单、使用灵活，在游戏开发中十分流行。但其生态并不完善，IDE开发工具及配套支持较少，一定程度上影响了Lua的开发效率及质量。LuaHelper遵从微软Language Server Protocol协议，是采用go语言开发的一种跨平台Lua代码编辑及检测工具。具有多种类编辑辅助、检测种类丰富、实时性高、内存占用少等特性。

## Feature Summary

### Code Editing
* [Defintion Find/定义跳转](#DefintionFind)
* [Find All References/引用查找](#FindAllReferences)
* [Document Symbols/文件符号表查询](#DocumentSymbols)
* [Workspace Symbols/工程符号表查询](#WorkspaceSymbols)
* [Auto Code Completion/自动代码补全](#AutoCodeCompletion)
* [Reformat Code/代码格式化](#ReformatCode)
* [Hover/代码悬停](#Hover)
* [Hightlight Global Var/全局变量着色](#HightlightGlobalVar)

### Code Detection
* [Syntax Check/语法检测](#SyntaxCheck)
* [Semantic Check/语义检测](#SemanticCheck)
* [Quick Analysis/快速增量分析](#QuickAnalysis)


## Feature Detail
###  <span id="DefintionFind">Defintion Find/定义跳转</span>
> **支持局部、全局文件定义查询跳转**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/GotoDefinition.gif)

###  <span id="FindAllReferences">Find All References/引用查找</span>
> **支持基于作用域的各类型引用查找**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/FindReferences.gif)

###  <span id="DocumentSymbols">Document Symbols/文件符号表查询</span>
> **支持文件域符号表查询，在搜索栏输入@**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/DocmentSymbol.gif)

###  <span id="WorkspaceSymbols">Workspace Symbols/工程符号表查询</span>
> **支持工程域符号表查询，在搜索栏输入#**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/WorkspaceSymbol.gif)

###  <span id="AutoCodeCompletion">Auto Code Completion/自动代码补全</span>
> **支持变量、函数的自动输入提示**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/CodeCompletion.gif)

###  <span id="ReformatCode">Reformat Code/代码格式化</span>
> **支持代码格式化**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/Format.gif)

###  <span id="Hover">Hover/代码悬停</span>
> **支持代码悬停提示**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/Hover.gif)

###  <span id="HightlightGlobalVar">Hightlight Global Var/全局变量着色</span>
> **支持全局变量高亮着色**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/GlobalColor.gif)

###  <span id="SyntaxCheck">Syntax Check/语法检测</span>
> **提供丰富的语法错误检测类型**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/SyntaxCheck.gif)

###  <span id="SemanticCheck">Semantic Check/语义检测</span>
> **支持多种类型的语义检测**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/SemanticCheck.gif)

###  <span id="QuickAnalysis">Quick Analysis/快速增量分析</span>
> **支持增量变化分析，分析结果诊断输出**
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/RealTimeCheck.gif)


## Installation
插件安装方法：
1. 点击Vs Code应用市场图标
2. 在输入框中搜索 luahelper
3. 安装它

![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/install.png)

## Acknowledgements
* [luago-books](https://github.com/zxh0/luago-book), go语言生成lua的AST，修改了源码（对AST的每个节点增加了列的属性，同时也优化了性能）。
* [LuaFormatter](https://github.com/Koihik/LuaFormatter), c++写的Lua代码格式化库，性能较高。

## Support
如有问题先参阅 [FAQ](./Docs/Manual/FAQ.md) ，如有问题建议使用 [issues](https://github.com/yinfei8/LuaHelper/issues) ，我们会关注和回复。

QQ群：747590892