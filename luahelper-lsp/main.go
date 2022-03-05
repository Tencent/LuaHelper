package main

import (
	"flag"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"

	"luahelper-lsp/langserver"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"

	"github.com/yinfei8/jrpc2/channel"
)

func main() {
	modeFlag := flag.Int("mode", 0, "mode type, 0 is run cmd, 1 is local rpc, 2 is socket rpc")
	logFlag := flag.Int("logflag", 0, "0 is not open log, 1 is open log")
	localpath := flag.String("localpath", "", "local project path")
	flag.Parse()

	// 是否开启日志
	enableLog := false
	if *logFlag == 1 {
		enableLog = true
	}

	// socket rpc时，默认开启日志，方便定位问题
	if *modeFlag == 2 {
		enableLog = true
	}

	// 开启日志时，才开启pprof
	if enableLog {
		go func() {
			http.ListenAndServe("localhost:6060", nil)
		}()
	}

	// 初始化为日志系统
	log.InitLog(enableLog)

	common.GlobalConfigDefautInit()
	common.GConfig.IntialGlobalVar()

	//*modeFlag = 0
	if *modeFlag == 1 {
		cmdRPC()
	} else if *modeFlag == 2 {
		socketRPC()
	} else if *modeFlag == 0 {
		runLocalDiagnostices(*localpath)
	}
}

//cmd 的方式运行rpc
func cmdRPC() {
	log.Debug("local stat running ....")
	Server := langserver.CreateServer()
	Server.Start(channel.Header("")(os.Stdin, os.Stdout))

	log.Debug("Server started ....")
	if err := Server.Wait(); err != nil {
		log.Debug("Server exited: %v", err)
	}
	log.Debug("Server exited return")
}

// 网络的方式运行rpc
func socketRPC() {
	log.Debug("socket running ....")
	lst, err := net.Listen("tcp", "localhost:7778")
	if err != nil {
		log.Error("Listen: %v", err)
		return
	}

	var wg sync.WaitGroup
	for {
		conn, err := lst.Accept()
		log.Debug("accept new conn ....")
		Server := langserver.CreateServer()
		if err != nil {
			if channel.IsErrClosing(err) {
				err = nil
			} else {
				log.Error("Error accepting new connection: %v", err)
			}
			wg.Wait()
			log.Error("Error accepting new connection: %v", err)
			return
		}
		ch := channel.Header("")(conn, conn)
		wg.Add(1)
		go func() {
			defer wg.Done()
			Server.Start(ch)
			if err := Server.Wait(); err != nil && err != io.EOF {
				log.Debug("Server exited: %v", err)
			}

			log.Debug("Server exited11: %v", err)
		}()
	}
}

func runLocalDiagnostices(localpath string) {
	log.Debug("local Diagnostices running ....")
	lspServer := langserver.CreateLspServer()
	lspServer.RunLocalDiagnostices(localpath)
	log.Debug("local Diagnostices exited ")
}
