
import * as vscode from 'vscode';
import * as Net from 'net';
import { Tools } from './common/tools';
import { DebugLogger } from './common/logManager';
import { LuaDebugSession } from './debug/luaDebug';
import {LuaPath} from './common/luaPath';

// debug启动时的配置项处理
export class LuaConfigurationProvider implements vscode.DebugConfigurationProvider {
    private _server?: Net.Server;
    private static RunFileTerminal;
    resolveDebugConfiguration(folder: vscode.WorkspaceFolder | undefined, config: vscode.DebugConfiguration, token?: vscode.CancellationToken): vscode.ProviderResult<vscode.DebugConfiguration> {
        // if launch.json is missing or empty
        if (!config.type && !config.name) {
            const editor = vscode.window.activeTextEditor;
            if (editor && editor.document.languageId === 'lua') {
                vscode.window.showInformationMessage('请先正确配置launch文件!');
                config.type = 'LuaHelper-Debug';
                config.name = 'LuaHelper';
                config.request = 'launch';
            }
        }

        // 不调试而直接运行当前文件
        if (config.noDebug) {
            // 获取活跃窗口
            let retObject = Tools.getVSCodeAvtiveFilePath();
            if (retObject["retCode"] !== 0) {
                DebugLogger.DebuggerInfo(retObject["retMsg"]);
                return;
            }
            let filePath = retObject["filePath"];

            if (LuaConfigurationProvider.RunFileTerminal) {
                LuaConfigurationProvider.RunFileTerminal.dispose();
            }
            LuaConfigurationProvider.RunFileTerminal = vscode.window.createTerminal({
                name: "Run Lua File (LuaHelper)",
                env: {},
            });

            // 把路径加入package.path
            let path = require("path");
            let pathCMD = "'";
            let pathArr = Tools.VSCodeExtensionPath.split(path.sep);
            let stdPath = pathArr.join('/');
            pathCMD = pathCMD + stdPath + "/debugger/?.lua;";
            pathCMD = pathCMD + config.packagePath.join(';');
            pathCMD = pathCMD + "'";

            let luaPath:LuaPath = new LuaPath();
            let luaPathStr:string = "";
            if (config.luaPath && config.luaPath !== '') {
              luaPathStr = config.luaPath;
            }
            let strVect:string[] = luaPath.GetLuaExeCpathStr(luaPathStr, Tools.VSCodeExtensionPath);
            
            // 路径socket的路径加入到package.cpath中
            let cpathCMD = "'";
            //cpathCMD = cpathCMD + stdPath + GetLuasocketPath();
            cpathCMD = cpathCMD + config.packagePath.join(';');
            cpathCMD = cpathCMD + strVect[1] + ";";
            cpathCMD = cpathCMD + "'";
            cpathCMD = " package.cpath = " + cpathCMD + ".. package.cpath; ";

            //拼接命令
            pathCMD = " \"package.path = " + pathCMD + ".. package.path; ";
        
            let doFileCMD = filePath;
            let runCMD = pathCMD + cpathCMD + "\" " + doFileCMD;

            let LuaCMD = strVect[0] + " -e ";

            LuaConfigurationProvider.RunFileTerminal.show();
            //兼容terminal启动耗时较长的情况，延迟发送命令
            setTimeout(() => LuaConfigurationProvider.RunFileTerminal.sendText(LuaCMD + runCMD, true), 2000);
            return;
        }

        // 关于打开调试控制台的自动设置
        if (config.name === "LuaHelper-DebugFile") {
            if (!config.internalConsoleOptions) {
                config.internalConsoleOptions = "neverOpen";
            }
        } else {
            if (!config.internalConsoleOptions) {
                config.internalConsoleOptions = "openOnSessionStart";
            }
        }

         // rootFolder 固定为 ${workspaceFolder}, 用来查找本项目的launch.json.
         config.rootFolder = '${workspaceFolder}';

         if (!config.TempFilePath) {
             config.TempFilePath = '${workspaceFolder}';
         }
 
         // 开发模式设置
         if( config.DevelopmentMode !== true ){
             config.DevelopmentMode = false;
         }

         if(!config.program){
            config.program = '';
        }

        if(config.packagePath == undefined){
            config.packagePath = [];
        }
        
        if(config.truncatedOPath == undefined){
            config.truncatedOPath = "";
        }

        if(config.distinguishSameNameFile == undefined){
            config.distinguishSameNameFile = false;
        }

        if(config.dbCheckBreakpoint == undefined){
            config.dbCheckBreakpoint = false;
        }

        if(!config.args){
            config.args = new Array<string>();
        }

        if(config.autoPathMode == undefined){
            // 默认使用自动路径模式
            config.autoPathMode = true;
        }
        
        if (!config.cwd) {
            config.cwd = '${workspaceFolder}';
        }

        if (!config.luaFileExtension) {
            config.luaFileExtension = '';
        }else{
            // luaFileExtension 兼容 ".lua" or "lua"
            let firseLetter = config.luaFileExtension.substr(0, 1);
            if(firseLetter === '.'){
                config.luaFileExtension =  config.luaFileExtension.substr(1);
            }
        }

        config.VSCodeExtensionPath = Tools.VSCodeExtensionPath;

        if (config.stopOnEntry == undefined) {
            config.stopOnEntry = true;
        }

        if (config.pathCaseSensitivity == undefined) {
            config.pathCaseSensitivity = false;
        }

        if (config.connectionPort == undefined) {
            config.connectionPort = 8818;
        }

        if (config.logLevel == undefined) {
            config.logLevel = 1;
        }

        if (config.autoReconnect != true) {
            config.autoReconnect = false;
        }

        if (config.updateTips == undefined) {
            config.updateTips = true;
        }

        if (config.useCHook == undefined) {
            config.useCHook = true;
        }

        if (config.isNeedB64EncodeStr == undefined) {
            config.isNeedB64EncodeStr = true;
        }

        if (config.VSCodeAsClient == undefined) {
            config.VSCodeAsClient = false;
        }

        if (config.connectionIP == undefined) {
            config.connectionIP = "127.0.0.1";
        }

        if (!this._server) {
            this._server = Net.createServer(socket => {
                const session = new LuaDebugSession();
                session.setRunAsServer(true);
                session.start(<NodeJS.ReadableStream>socket, socket);
            }).listen(0);
        }
        // make VS Code connect to debug server instead of launching debug adapter
        let addressInfo = this._server.address() as Net.AddressInfo;
        config.debugServer = addressInfo.port;
        return config;
    }

    dispose() {
        if (this._server) {
            this._server.close();
        }
    }
}