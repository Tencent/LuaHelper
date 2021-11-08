
import * as vscode from 'vscode';
import { AnnotatorType } from './notifications';
import { LanguageClient } from 'vscode-languageclient/node';
import * as notifications from "./notifications";

let D_Global_Var: vscode.TextEditorDecorationType;
let D_Global_Func: vscode.TextEditorDecorationType;
let D_Annotate_Type: vscode.TextEditorDecorationType;

function createDecoration(key: string, config: vscode.DecorationRenderOptions | undefined = undefined): vscode.TextEditorDecorationType {
    let color = vscode.workspace.getConfiguration("luahelper").get(key);
    config = config || {};
    if (typeof (color) === 'string') {
        config.light = { color: color };
        config.dark = { color: color };
    }
    return vscode.window.createTextEditorDecorationType(config);
}

function updateDecorations() {
    D_Global_Var = createDecoration("colors.Global Field Color");
    D_Global_Func = createDecoration("colors.Global Fun Color");
    D_Annotate_Type = createDecoration("colors.Type(annotation) Color");
}

export function onDidChangeConfiguration(client: LanguageClient) {
    updateDecorations();
}

let timeoutToReqAnn: NodeJS.Timer;

export function requestAnnotators(editor: vscode.TextEditor, client: LanguageClient) {
    let openFlagConfig = vscode.workspace.getConfiguration("luahelper.colors", null).get("Enable");
    var openFlag = false;
    if (openFlagConfig !== undefined) {
        openFlag = <boolean><any>openFlagConfig;
    }

    // 若没有开启开关，退出
    if (!openFlag) {
        return;
    }

    clearTimeout(timeoutToReqAnn);
    timeoutToReqAnn = setTimeout(() => {
        requestAnnotatorsImpl(editor, client);
    }, 250);
}

function requestAnnotatorsImpl(editor: vscode.TextEditor, client: LanguageClient) {
    if (!D_Global_Var) {
        updateDecorations();
    }

    let params: notifications.AnnotatorParams = { uri: editor.document.uri.toString() };
    client.sendRequest<notifications.IAnnotator[]>("luahelper/getVarColor", params).then(list => {
        let map: Map<AnnotatorType, vscode.Range[]> = new Map();
        map.set(AnnotatorType.GlobalVar, []);
        map.set(AnnotatorType.GlobalFunc, []);

        if (list !== undefined && list !== null) {
            list.forEach(data => {
                let uri = vscode.Uri.parse(data.uri);
                vscode.window.visibleTextEditors.forEach((editor) => {
                    let docUri = editor.document.uri;
                    if (uri.path.toLowerCase() === docUri.path.toLowerCase()) {
                        var list = map.get(data.annotatorType);
                        if (list === undefined) {
                            list = data.ranges;
                        } else {
                            list = list.concat(data.ranges);
                        }
                        map.set(data.annotatorType, list);
                    }
                });
            });
        }
        map.forEach((v, k) => {
            updateAnnotators(editor, k, v);
        });
    });
}

function updateAnnotators(editor: vscode.TextEditor, type: AnnotatorType, ranges: vscode.Range[]) {
    switch (type) {
        case AnnotatorType.GlobalVar:
            editor.setDecorations(D_Global_Var, ranges);
            break;
        case AnnotatorType.GlobalFunc:
            editor.setDecorations(D_Global_Func, ranges);
            break;
        case AnnotatorType.AnnotateType:
            editor.setDecorations(D_Annotate_Type, ranges);
            break;
    }
}