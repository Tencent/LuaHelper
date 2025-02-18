import * as vscode from 'vscode';
import * as path from "path";
import * as child_process from "child_process";
import * as fs from "fs";

function needShowFormatErr(): boolean {
    let formatErrConfig = vscode.workspace.getConfiguration("luahelper.base", null).get("formatErrShow");
    var formatErrshow = true;
    if (formatErrConfig !== undefined) {
        formatErrshow = <boolean><any>formatErrConfig;
    }
    return formatErrshow;
}

// 获取整型配置参数的字符串值
// itemStr 为配置项的名称
function getFormatIntStr(itemStr: string): string {
    let value: number | undefined = vscode.workspace.getConfiguration("luahelper.format", null).get(itemStr);
    if (value === undefined) {
        return "";
    }

    itemStr = itemStr.replace(/_/g, "-");
    let strResult: string = "--" + itemStr + "=" + String(value);
    return strResult;
}

// 获取字符串配置参数的字符串值
// itemStr 为配置项的名称
function getFormatStrStr(itemStr: string): string {
    let value: string | undefined = vscode.workspace.getConfiguration("luahelper.format", null).get(itemStr);
    if (value === undefined) {
        return "";
    }

    itemStr = itemStr.replace(/_/g, "-");
    let strResult: string = "--" + itemStr + "=" + String(value);
    return strResult;
}

// 获取bool配置参数的字符串值
// itemStr 为配置项的名称
function getFormatBoolStr(itemStr: string): string {
    let value: boolean | undefined = vscode.workspace.getConfiguration("luahelper.format", null).get(itemStr);
    if (value === undefined) {
        return "";
    }

    itemStr = itemStr.replace(/_/g, "-");
    let strResult: string = "";
    if (value) {
        strResult = "--" + itemStr;
    } else {
        strResult = "--no-" + itemStr;
    }

    return strResult;
}

// 获取格式化配置的参数
function getFormatConfigStrTable(fileName: string): string[] {

    if (vscode.workspace.workspaceFolders) {
        for (let i = 0; i < vscode.workspace.workspaceFolders.length; i++) {
            const folder = vscode.workspace.workspaceFolders[i];
            if (fileName.startsWith(folder.uri.fsPath)) {
                const c = path.join(folder.uri.fsPath, ".lua-format");
                if (fs.existsSync(c)){
                    return ["-c", c];
                }
            }
        }
    }

    let strAllResult: string[] = [];
    let oneStr: string = "";
    oneStr = getFormatIntStr("column_limit");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatIntStr("indent_width");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("use_tab");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatIntStr("tab_width");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatIntStr("continuation_indent_width");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("keep_simple_control_block_one_line");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("keep_simple_function_one_line");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatIntStr("spaces_before_call");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("keep_simple_block_one_line");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("align_args");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("break_after_functioncall_lp");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("break_before_functioncall_rp");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("align_parameter");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("chop_down_parameter");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("break_after_functiondef_lp");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("break_before_functiondef_rp");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("align_table_field");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("break_after_table_lb");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("break_before_table_rb");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("chop_down_table");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("chop_down_kv_table");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatStrStr("table_sep");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("extra_sep_at_table_end");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("break_after_operator");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("double_quote_to_single_quote");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("single_quote_to_double_quote");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("spaces_inside_functiondef_parens");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("spaces_inside_functioncall_parens");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("spaces_inside_table_braces");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatBoolStr("spaces_around_equals_in_field");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }

    oneStr = getFormatIntStr("line_breaks_after_function_body");
    if (oneStr !== "") {
        strAllResult.push(oneStr);
    }
    return strAllResult;
}

// 获取luafmt二进制的文件的路径
function GetLuaFmtPath(context: vscode.ExtensionContext): string {
    var vscodeRunStr: string = path.resolve(context.extensionPath, "server");
    let binaryPath = "";
    if (process.platform === "win32") {
        binaryPath = path.resolve(vscodeRunStr, "win32", "lua-format");
    } else if (process.platform === "darwin") {
        binaryPath = path.resolve(vscodeRunStr, "darwin", "lua-format");
    } else if (process.platform === "linux") {
        binaryPath = path.resolve(vscodeRunStr, "linux", "lua-format");
    }

    return binaryPath;
}

export class LuaFormatRangeProvider implements vscode.DocumentRangeFormattingEditProvider {
    private context: vscode.ExtensionContext;

    constructor(context: vscode.ExtensionContext) {
        this.context = context;
    }

    public provideDocumentRangeFormattingEdits(document: vscode.TextDocument, range: vscode.Range, options: vscode.FormattingOptions, token: vscode.CancellationToken): Thenable<vscode.TextEdit[]> {
        let data = document.getText();
        data = data.substring(document.offsetAt(range.start), document.offsetAt(range.end));
        let formatErrshow = needShowFormatErr();
        let formatConfigStrTable: string[] = getFormatConfigStrTable(document.fileName);
        return new Promise((resolve, reject) => {
            let binaryPath = GetLuaFmtPath(this.context);
            if (binaryPath === "") {
                return;
            }

            try {
                const args = ["-i"];
                for (let str of formatConfigStrTable) {
                    args.push(str);
                }

                const cmd = child_process.spawn(binaryPath, args, {});
                const result: Buffer[] = [], errorMsg: Buffer[] = [];
                cmd.on('error', err => {
                    if (formatErrshow) {
                        vscode.window.showErrorMessage(`Run lua-format error : '${err.message}'`);
                    }
                    reject(err);
                });
                cmd.stdout.on('data', data => {
                    result.push(Buffer.from(data));
                });
                cmd.stderr.on('data', data => {
                    errorMsg.push(Buffer.from(data));
                });
                cmd.on('exit', code => {
                    const resultStr = Buffer.concat(result).toString();
                    if (code) {
                        if (formatErrshow) {
                            vscode.window.showErrorMessage(`Run lua-format failed with exit code: ${code}`);
                        }
                        reject(new Error(`Run lua-format failed with exit code: ${code}`));
                        return;
                    }
                    if (resultStr.length > 0) {
                        resolve([new vscode.TextEdit(range, resultStr)]);
                    }
                });
                cmd.stdin.write(data);
                cmd.stdin.end();

            } catch (e) {
                console.log("exception");
            }
        });
    }
}

export class LuaFormatProvider implements vscode.DocumentFormattingEditProvider {
    private context: vscode.ExtensionContext;

    constructor(context: vscode.ExtensionContext) {
        this.context = context;
    }

    public provideDocumentFormattingEdits(document: vscode.TextDocument, options: vscode.FormattingOptions, token: vscode.CancellationToken): Thenable<vscode.TextEdit[]> {
        var data = document.getText();

        let formatErrshow = needShowFormatErr();
        let formatConfigStrTable: string[] = getFormatConfigStrTable(document.fileName);
        return new Promise((resolve, reject) => {
            let binaryPath = GetLuaFmtPath(this.context);
            if (binaryPath === "") {
                return;
            }

            try {
                const args = ["-i"];
                for (let str of formatConfigStrTable) {
                    args.push(str);
                }

                const cmd = child_process.spawn(binaryPath, args, {});
                const result: Buffer[] = [], errorMsg: Buffer[] = [];
                cmd.on('error', err => {
                    if (formatErrshow) {
                        vscode.window.showErrorMessage(`Run lua-format error : '${err.message}'`);
                    }
                    reject(err);
                });
                cmd.stdout.on('data', data => {
                    result.push(Buffer.from(data));
                });
                cmd.stderr.on('data', data => {
                    errorMsg.push(Buffer.from(data));
                });
                cmd.on('exit', code => {
                    const resultStr = Buffer.concat(result).toString();
                    if (code) {
                        if (formatErrshow) {
                            vscode.window.showErrorMessage(`Run lua-format failed with exit code: ${code}`);
                        }
                        reject(new Error(`Run lua-format failed with exit code: ${code}`));
                        return;
                    }
                    if (resultStr.length > 0) {
                        const range = document.validateRange(new vscode.Range(0, 0, Infinity, Infinity));
                        resolve([new vscode.TextEdit(range, resultStr)]);
                    }
                });
                cmd.stdin.write(data);
                cmd.stdin.end();
            } catch (e) {
                console.log("exception");
            }
        });
    }
}