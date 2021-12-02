import { window, StatusBarItem, StatusBarAlignment } from 'vscode';
import { LanguageClient } from 'vscode-languageclient/node';
import * as notifications from "./notifications";
import * as vscode from 'vscode';

let statusBar: StatusBarItem;

export class OnlinePeople {
    private timeoutToReqAnn: NodeJS.Timer;

    public Start(client: LanguageClient) {
        if (!statusBar) {
            statusBar = window.createStatusBarItem(StatusBarAlignment.Right);
        }

        if (this.timeoutToReqAnn) {
            clearInterval(this.timeoutToReqAnn);
        }

        sendRequest(statusBar, client);

        // 每分钟触发一次
        this.timeoutToReqAnn = setInterval(function () {
            sendRequest(statusBar, client);
        }, 60000);
    }


    // 销毁对象和自由资源
    dispose() {
        statusBar.dispose();
        if (this.timeoutToReqAnn) {
            clearInterval(this.timeoutToReqAnn);
        }
    }
}

function sendRequest(statusBar: StatusBarItem, client: LanguageClient) {
    let params: notifications.GetOnlineParams = { Req: 0 };
    client.sendRequest<notifications.GetOnlineReturn>("luahelper/getOnlineReq", params).then(onelineReturn => {
        let showConfig = vscode.workspace.getConfiguration("luahelper.base", null).get("showOnline");
        var openFlag = false;
        if (showConfig !== undefined) {
            openFlag = <boolean><any>showConfig;
        }

        if (openFlag === true) {
            if (onelineReturn.Num === 0) {
                statusBar.text = "LuaHelper";
                statusBar.tooltip = "LuaHelper";
                statusBar.show();
            } else {
                if (vscode.env.language === "zh-cn" || vscode.env.language === "zh-tw") {
                    statusBar.text = "LuaHelper 在线人数 : " + onelineReturn.Num;
                    statusBar.tooltip = "LuaHelper online People: " + onelineReturn.Num;
                } else {
                    statusBar.text = "LuaHelper : " + onelineReturn.Num;
                    statusBar.tooltip = "LuaHelper Online People: " + onelineReturn.Num;
                }
                statusBar.show();
            }
        } else {
            statusBar.text = "LuaHelper";
            statusBar.tooltip = "LuaHelper";
            statusBar.show();
        }
    });
}
