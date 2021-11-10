package langserver

import (
	"runtime"
)

// OnlineReport 向中心服务器，需要上报统计的信息
type OnlineReport struct {
	ClientType         string `json:"clientType"`         // 客户端类型
	FileNumber         int    `json:"fileNumber"`         // 工程lua文件的数量
	OsType             string `json:"osType"`             // 客户端操作系统类型
	ClientVer          string `json:"clientVer"`          // 客户端的版本号，例如0.2.1
	FirstReport        int    `json:"firstReport"`        // 客户端是否首次打开插件上报， 1为是
	CostMsTime         int    `json:"costMsTime"`         // 初次加载的耗时时间毫秒
	WorkspaceFolderNum int    `json:"workspaceFolderNum"` // 用户多文件夹数
}

// OnlineReportReturn 中心服务的返回
type OnlineReportReturn struct {
	Num int `json:"Num"` // 所有在线的人数
}

// SetOnlineReportParam 连接初始化时，同步客户端的类型，以及这个工程的所包含的lua文件数量
func (l *LspServer) SetOnlineReportParam(clientType string, fileNumber int, costMsTime int, workspaceFolderNum int) {
	l.onlineReport.ClientType = clientType
	l.onlineReport.FileNumber = fileNumber
	l.onlineReport.CostMsTime = costMsTime
	l.onlineReport.WorkspaceFolderNum = workspaceFolderNum
}

// SetLuaFileNumber 更新工程lua文件的数量
func (l *LspServer) SetLuaFileNumber(fileNumber int) {
	l.onlineReport.FileNumber = fileNumber
}

// GetOnlineReportData 获取向中心服统计的数据
func (l *LspServer) GetOnlineReportData() OnlineReport {
	return l.onlineReport
}

// SetFirstReportFlag 设置是否第一次上报标记
func (l *LspServer) SetFirstReportFlag(flag int) {
	l.onlineReport.FirstReport = flag
}

// SetReportOtherInfo 设置上报的其他信息
func (l *LspServer) SetReportOtherInfo() {
	l.onlineReport.OsType = runtime.GOOS
}