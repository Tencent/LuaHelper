package langserver

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"luahelper-lsp/langserver/log"
)

// 客户端获取当前查找在线人数，查找会去中心服上查询所有在线人数

// GetOnlineParams 全局颜色配置
type GetOnlineParams struct {
	Req int `json:"Req"` // 参数无意义
}

// GetOnlineReturn 在线的返回
type GetOnlineReturn struct {
	Num int `json:"Num"` // 所有在线的人数
}

// 插件在线人数
var onlinePeopleNum int = 0

// GetOnlineReq 获取当前所有在线人数的接口
func (l *LspServer) GetOnlineReq(ctx context.Context, vs GetOnlineParams) (onlineReturn GetOnlineReturn, err error) {
	log.Debug("GetOnlineReq num=%d", onlinePeopleNum)
	onlineReturn.Num = onlinePeopleNum
	return
}

// UDPReportOnline 上报协程
func (l *LspServer) UDPReportOnline() {
	log.Debug("UDPReportOnline \n")
	conn, err := net.Dial("udp", "42.194.136.76:7778") //打开监听端口
	if err != nil {
		log.Error("conn fail...")
		return
	}
	
	defer conn.Close()
	log.Debug("client connect server successed \n")

	//创建协程，收取udp的在线回包数据
	go handleRecv(conn)

	l.SetReportOtherInfo()

	for {
		if l.enableReport {
			user := l.GetOnlineReportData()
			info, err := json.Marshal(user)
			if err != nil {
				log.Error("report Data err=%s", err.Error())
			} else {
				conn.Write(info)
			}
			l.SetFirstReportFlag(0)
		}

		// 每120秒上报一次
		time.Sleep(time.Second * 120)
	}
}

// 处理udp的回包，获取在线人数
func handleRecv(conn net.Conn) {
	var reportReturn OnlineReportReturn
	data := make([]byte, 2048)
	for {
		msgRead, err := conn.Read(data) //将读取的字节流赋值给msg_read和err
		if msgRead == 0 || err != nil { //如果字节流为0或者有错误
			continue
		}

		err1 := json.Unmarshal(data[0:msgRead], &reportReturn)
		if err1 != nil {
			log.Debug("handleRecv error:%s", err1.Error())
		} else {
			log.Debug("handleRecv ok, num=%d", reportReturn.Num)
			onlinePeopleNum = reportReturn.Num
		}
	}
}
