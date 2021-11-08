# 调试器原理

## 调试器协议
调试器的实现采用了微软推出的[Debug Adapter Protocol](https://microsoft.github.io/debug-adapter-protocol/)，它把开发工具（IDE或Iditor）与不同语言的debugger功能解耦， 并把两者之间的通信的方式抽象为通用协议（JSON格式）。通过调试适配器协议，可以为开发工具实现通用调试器，该调试器可以通过调试适配器与不同的调试器进行通信。调试适配器可以在多个开发工具中重复使用，从而大大减少了在不同工具中支持新调试器的工作。

![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/debug/debugprinciple.png)


调试的过程涉及了三个角色：
* 开发工具（IDE、Editor)。
* 调试适配器协议(Debug Adapter Protocol)，中间沟通作用。协议采用JSON格式，通用性较强。
* 后端调试器(Debugger)，负责调试主体功能的实现。

调试适配器初始化之后，就可以接受启动调试的请求。请求主要分为两类：

* launch请求：调试适配器以调试模式启动程序，然后开始与其通信。
* attach请求：调试适配器连接到已经运行的程序。 最终用户在这里负责启动和终止程序。

## 实现架构
LuaPanda调试的架构采用了attach请求模式，attach的作用是附加到运行Lua代码的可执行程序上，调试器前端VSCode工程与运行的lua进程（被调试的进程）通过socket通信。下图是LuaHelper集成了LuaPanda调试模块后的调试架构图。

![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/debug/debugstruct.png)

整个调试功能分为两大主体：
* VSCode打开的Lua工程：上图中左边框。
* 运行Lua工程的进程：上图中左边框。

整个调试的流程如下：
* 1)打开Lua工程，激活LuaHelper插件。
* 2)Lua工程中主动引入LuaPanda.lua文件，引入的代码片段为：
    ```lua
    require("LuaPanda").start("127.0.0.1", 8818);
    ```
* 3)VSCode打开的Lua工程，F5主动开启调试功能，监听socket连接。
* 4)主动执行Lua工程，当代码中运行到引入的LuaPanda.lua文件时，会主动socket链接前端调试的工程。
* 5)运行Lua工程的进程与前端调试的工程attach成功，开始调试功能。
* 6)前端调试工程设置好断点，通知给运行Lua工程的进程，运行Lua工程的进程设置好hook信息。
* 7)运行Lua工程的进程执行到预先设置的断点处，从lua虚拟机获取到的调试信息，发送给前端工程展示。