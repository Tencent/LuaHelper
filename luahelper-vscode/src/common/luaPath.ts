import * as os from 'os';


export class LuaPath {
    // 获取设置的luaPath。如果是windows平台，若路径中包含空格，路径用字符串包括。
    // 例如：D:\Program Files\lua-5.4.2_Win64_bin\lua54.exe 替换为 D:\Program Files\"lua-5.4.2_Win64_bin"\lua54.exe
    private getSetLuaPath(strPath: string): string {
        let retStr: string = strPath;
        let platform: string = os.platform();
        if (platform === "linux" || platform === "darwin") {
            return retStr;
        }

        // windows平台特殊处理
        let path = require("path");
        let pathArr = strPath.split(path.sep);
        strPath = pathArr.join('/');
        pathArr = strPath.split("/");
      
        for (var i = 1; i < pathArr.length - 1; i++) {
            if (pathArr[i].indexOf(" ") > 0) {
                pathArr[i] = "\"" + pathArr[i] + "\"";
            }
        }

        let stdPath = pathArr.join('/');
        return stdPath;
    }

    // 获取lua的版本号
    // 返回的版本号为5.1, 5.2, 5.3, 5.4
    private getLuaBinVersion(strCmd: string): string {
        const { execSync } = require('child_process');
        try {
            let out = execSync(strCmd, {}, (err, stdout, stderr) => {
                if (err) {
                    console.log(err);
                    return "";
                }
                console.log(`stdout: ${stdout}`);
            });

            let str = out.toString("utf-8").replace('\n', '');
            if (str.indexOf("Lua 5.1") >= 0) {
                return "5.1";
            }

            if (str.indexOf("Lua 5.2") >= 0) {
                return "5.2";
            }

            if (str.indexOf("Lua 5.3") >= 0) {
                return "5.3";
            }

            if (str.indexOf("Lua 5.4") >= 0) {
                return "5.4";
            }

            return "";
        } catch (e) {
            //console.error(e);
            //throw(e);
            return "";
        }
    }

    // 获取插件自带的lua二进制完整路径
    private getLuaExePathStr(vSCodeExtensionPath: string, luaVersionStr: string): string {
        let path = require("path");
        let pathArr = vSCodeExtensionPath.split(path.sep);
        let stdPath = pathArr.join('/');

        let retStr: string = "";
        let suffixStr: string = "";
        let platform: string = os.platform();
        switch (platform) {
            case "win32":
                let suffixVerStr: string = luaVersionStr.replace(".", "");
                suffixStr = "/debugger/luasocket/win/x64/lua" + luaVersionStr + "/" + "lua" + suffixVerStr + ".exe";
                break;
            case "linux":
                suffixStr = "/debugger/luasocket/linux/lua" + luaVersionStr + "/lua";
                break;
            case "darwin":
                suffixStr = "/debugger/luasocket/mac/lua" + luaVersionStr + "/lua";
                break;
        }
        retStr = stdPath + suffixStr;
        return retStr;
    }

    // 获取插件自带的socket库的cpath路径
    private getLuaCPathStr(vSCodeExtensionPath: string, luaVersionStr: string): string {
        let path = require("path");
        let pathArr = vSCodeExtensionPath.split(path.sep);
        let stdPath = pathArr.join('/');

        let retStr: string = "";
        let suffixStr: string = "";
        let platform: string = os.platform();
        switch (platform) {
            case "win32":
                suffixStr = "/debugger/luasocket/win/x64/lua" + luaVersionStr + "/?.dll";
                break;
            case "linux":
                suffixStr = "/debugger/luasocket/linux/lua" + luaVersionStr + "/?.so";
                break;
            case "darwin":
                suffixStr = "/debugger/luasocket/mac/lua" + luaVersionStr + "/?.so";
                break;
        }
        retStr = stdPath + suffixStr;
        return retStr;
    }

    // 获取用户lua 二进制位置与cpath库的后缀路径
    // strPath 为传入的用户设置的lua二进制位置
    // vSCodeExtensionPath 为插件运行的目录
    public GetLuaExeCpathStr(strPath: string, vSCodeExtensionPath: string): string[] {
        let strVect: string[] = ["", ""];

        let strVer: string = "";
        if (strPath === "") {
            // 用户没有额外设置lua的二进制位置，获取默认的版本Lua版本号
            let luaVersionStr = this.getLuaBinVersion("lua -v");
            if (luaVersionStr === "") {
                // 没有找到默认的lua二进制，利用插件自带的二进制版本
                let newVSCodeExtensionPath: string = this.getSetLuaPath(vSCodeExtensionPath);
                strVect[0] = this.getLuaExePathStr(newVSCodeExtensionPath, "5.4");
                strVer = "5.4";
            } else {
                //  找到了默认的lua 二进制
                strVect[0] = "lua";
                strVer = luaVersionStr;
            }
        } else {
            // 用户设置了lua二进制的位置
            let strNewPath: string = this.getSetLuaPath(strPath);
            let luaVersionStr = this.getLuaBinVersion(strNewPath + " -v");
            if (luaVersionStr === "") {
                // 没有找到默认的lua二进制，利用插件自带的二进制版本
                let newVSCodeExtensionPath: string = this.getSetLuaPath(vSCodeExtensionPath);
                strVect[0] = this.getLuaExePathStr(newVSCodeExtensionPath, "5.4");
                strVer = "5.4";
            } else {
                //  找到
                strVect[0] = strNewPath;
                strVer = luaVersionStr;
            }
        }

        //let newVSCodeExtensionPath: string = this.getSetLuaPath(vSCodeExtensionPath);
        strVect[1] = this.getLuaCPathStr(vSCodeExtensionPath, strVer);

        return strVect;
    }
}