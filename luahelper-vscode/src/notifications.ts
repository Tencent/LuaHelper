import * as vscode from 'vscode';

export interface AnnotatorParams {
    uri: string;
}

export enum AnnotatorType {
    GlobalVar,
    GlobalFunc,
    AnnotateType,
}

export interface IAnnotator {
    uri: string;
    ranges: vscode.Range[];
    annotatorType: AnnotatorType;
}

// 启动插件加载工程所花费的时间
export interface IProgressReport {
    state: number; // 通知的状态，0为加载目录, 1为加载文件中，2为加载完毕
    text: string;  // 显示的字符串
}

export interface GetOnlineParams {
    Req: number;// 参数无意义
}


export interface GetOnlineReturn {
    Num: number;// 所有在线的人数
}