# 单文件调试与运行
### 1背景
在做lua开发过程中，有时希望在一个独立文件中测试函数执行结果。单文件的执行/调试也是为了方便这种场景。

### 2调试配置
首先需要设置调试配置，调试配置参考前文。然后选择下图中LuaHelper-DebugFile选项
![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/debug/debugfile.png)


### 3调试单个文件
调试选项切换LuaPanda-DebugFile, 代码编辑窗口切换到待调试文件，按F5快捷键进行调试。这种模式下，不用引入LuaPanda.lua文件和luasocket库。

![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/debug/debugfilerun.png)


### 4单文件运行
调试选项切换LuaPanda-DebugFile, 代码编辑窗口切换到待运行的文件，按ctrl + F5 快捷键运行lua文件。这种模式下，也不用引入LuaPanda.lua文件和luasocket库。

![avatar](https://raw.githubusercontent.com/yinfei8/LuaHelper/master/images/debug/runonefile.gif)
