# 源码介绍及安装

## 1 整体介绍
微软Language Server Protocol协议，是一种被用于编辑器或集成开发环境与支持比如自动补全，定义跳转，查找所有引用等语言特性的语言服务器之间的一种协议。


早期，没有LSP协议，如果要为语言编写不同编辑器下的插件，每种编辑器下都需要单独开发。
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/Nolsp.png)

后面，微软推出了LSP协议，把插件的开发分为前端和后端，后端可以使用其他语言，且可以适配不同的前端。那么不同编辑器下插件，可以利用同一个后端，极大的减少插件的开发量。同时，插件的后端可以采用任何语言，不局限于前端的语言。前端与后端通信，采用Json协议格式。如下图所示。
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/Lsp.png)

## 2 源码介绍
本插件，前端是利用Typescript语言，后端是采用Go语言。
源码目录如下：
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/src.png)

luahelper-lsp目录，是后端go程序目录。</br>
luahelper-vscode目录，是前端typescript程序目录。

## 3 源码编译的方法
### 3.1 后端go程序编译方法
 * **a）首先把源码 ..\LuaHelper\luahelper-lsp目录添加到GoPath环境变量中**
 * **b）进入到 ..\LuaHelper\luahelper-lsp\bin目录，执行go编译命令：**</br>
     go build -ldflags "-w -s" lualsp </br>
     如果是windows环境，会编译出lualsp.exe</br>
     如果是linux环境，会编译出lualsp, 二进制更改名称为: linuxlualsp</br>
     如果是mac环境，会编译出lualsp, 二进制更改名称为: maclualsp</br>
 * **c) 把编译出来的go二进制（总共三个环境二进制），拷贝到插件前端目录中：**</br>
     ..\LuaHelper\luahelper-vscode\server ,目录如下图所示：
   ![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/clientexe.png)
 * **d) 前端插件编译的时候，会把..\LuaHelper\luahelper-vscode\server目录下的二进制打包到插件中。**</br>
 插件运行时，会根据不同的平台加载相应的二进制执行。
 

### 3.2 前端插件的编译方法
* **a) 先安装npm包管理**</br>
  npm完整的安装方法：</br>
  https://www.jianshu.com/p/03a76b2e7e00</br>
  查看所有的npm配置 </br>
  npm 设置代理，代理设置方法，在默认的配置文件中(zhangsan为windows用户名）: </br>
  C:\Users\zhangsan</br>
  文件：.npmrc</br>
  .npmrc文件内增加下面的内容：</br>
  http-proxy=http://127.0.0.1:12639</br>
  https-proxy=http://127.0.0.1:12639</br>
* **b) 全局安装vsce**</br>
  安装命令为：npm install -g vsce</br>
* **c) 进来插件前端源码目录，安装相应依赖包**</br>
  目录为：..\LuaHelper\luahelper-vscode
  运行命令: npm install </br>
* **d) 编译源码** </br>
  目录为：..\LuaHelper\luahelper-vscode</br>
  运行命令：vsce package</br>
  当前目录下为生成对应的vsix文件