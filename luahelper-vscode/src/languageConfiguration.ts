import { LanguageConfiguration, IndentAction } from "vscode";

export class LuaLanguageConfiguration implements LanguageConfiguration {
    public onEnterRules = [
        {
			action: { indentAction: IndentAction.None, appendText: "---@" },
			beforeText: /( *---@class)|( *---@field)|( *---@param)|( *---@return)/,
        }
    ];

    public wordPattern = /((?<=')[^']+(?='))|((?<=")[^"]+(?="))|(-?\d*\.\d\w*)|([^\`\~\!\@\$\^\&\*\(\)\=\+\[\{\]\}\\\|\;\:\'\"\,\.\<\>\/\s]+)/g;
}