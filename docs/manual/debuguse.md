## 插件调试的方法
### 1 插件前端的调试
* a) 需要安装npm
* b) npm安装vsce， 命令：npm install -g vsce
* c) 命令行进入luahelper-vscode目录，执行npm install
* d) VSCode 打开luahelper-vscode客户端目录，直接调试即可

前端调试的时候，会新创建一个VSCode终端，在对应的终端里面打开lua工程文件夹就可以调试前端。

### 2 插件后端的调试
后端是go语言写的，如果是用VSCode打开后端程序，需要打开目录：luahelper-lsp
   
#### 2.1 VSCode安装go的插件
 在插件市场搜索Go，安装排名第一的插件即可

 ![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/debug/go.png)
 
#### 2.2 VSCode打开luahelper-lsp目录，按F5调试即可

#### 2.3 设置插件前端连接Go后端
插件前端与插件后端协议的格式Json RPC，如下图所示
![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/running.png)

插件前端连接后端有两种方式：
* a) 通过管道的方式，插件前端直接拉取后端exe程序
* b) 通过socket的方式

为了使插件前端通过socket连接的后端，前端插件的设置如下：
（选设置-》Lua Helper—》 Project:Lsp (勾选socket rpc)

 ![avatar](https://raw.githubusercontent.com/Tencent/LuaHelper/master/images/debug/socket.png)
 
 
 #### 2.4 VSCode重新打开Lua工程
 
通过socket连接上后，就可以调试go的后端程序，在go的后端就可以调试go的代码
