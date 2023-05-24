package langserver

import (
	"context"
	"fmt"

	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
)

// RunLocalDiagnostices 运行本地模式，校验错误
func (l *LspServer) RunLocalDiagnostices(localpath string) {
	RootPath := "file://" + localpath
	RootURI := localpath

	pathpre.InitialRootURIAndPath(string(RootURI), string(RootPath))

	log.Debug("Initialize ..., rootDir=%s, rooturl=%s", RootPath, RootURI)

	vscodeRoot := pathpre.VscodeURIToString(string(RootURI))
	dirManager := common.GConfig.GetDirManager()
	dirManager.SetVSRootDir(vscodeRoot)

	initOptions := &InitializationOptions{
		AllEnable: true,
	}

	// 按顺序插入
	checkFlagList := getCheckFlagList(initOptions)

	ctx := context.Background()

	initErr := l.initialCheckProject(ctx, checkFlagList, "local", 0, nil, true, nil, nil)
	if initErr != nil {
		log.Error("initial luahelper err: " + initErr.Error())
		return
	}
	log.Debug("initial luahelper ok")
	project := l.getAllProject()
	if project == nil {
		log.Error("CheckProject is nil")
		return
	}

	fileErrorMap := project.GetAllFileErrorInfo()

	// 保存全局的错误诊断信息
	l.fileErrorMap = fileErrorMap
	if len(fileErrorMap) == 0 {
		log.Debug("GetAllFileErrorInfo is empty..")
		return
	}

	for file, info := range fileErrorMap {
		for _, errinfo := range info {
			fmt.Printf("%v, line=%v, errType=%v, errStr=%s\n", file, errinfo.Loc.StartLine, (int)(errinfo.ErrType),
				errinfo.ErrStr)
		}
	}
}
