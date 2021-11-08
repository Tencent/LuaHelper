package common

import (
	"io/ioutil"
	"luahelper-lsp/langserver/filefolder"
	"luahelper-lsp/langserver/log"
	"luahelper-lsp/langserver/pathpre"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// DirManager 所有的目录管理
type DirManager struct {
	// 插件前端传人的工程的根目录
	vSRootDir string

	// 插件前端传人的所有次级目录，目前VScode支持多目录文件夹
	subDirVec []string

	// luahelper.json文件包含的相对路径，如果没有luahelper.json，该值默认为./
	configRelativeDir string

	// 为主工程真实的目录，为VSRootDir与ConfigRelativeDir值的拼接组合
	mainDir string

	// 插件前端配置的读取Lua标准库等lua文件的文件夹
	clientExtLuaPath string
}

// create default dir manager
func createDirManager() *DirManager {
	dirManager := &DirManager{
		vSRootDir:         "",
		subDirVec:         []string{},
		configRelativeDir: "./",
		mainDir:           "",
		clientExtLuaPath:  "",
	}

	return dirManager
}

// SetClientPluginPath Set client Plugin path
func (d *DirManager) SetClientPluginPath(path string) {
	if path == "" {
		return
	}

	strTmpDir, err := filepath.Abs(path + "/server/meta")
	if err != nil {
		log.Error("set path err, path=%s", path+"/server/meta")
		return
	}
	strTmpDir = strings.Replace(strTmpDir, "\\", "/", -1)

	d.clientExtLuaPath = strTmpDir
}

// GetClientExtLuaPath get client ext lua path
func (d *DirManager) GetClientExtLuaPath() string {
	return d.clientExtLuaPath
}

func (d *DirManager) setConfigRelativeDir(baseDir string) {
	d.configRelativeDir = baseDir
}

// SetVSRootDir set VSCode root dir
func (d *DirManager) SetVSRootDir(vsRootDir string) {
	d.vSRootDir = vsRootDir
}

// GetMainDir 获取主的目录
func (d *DirManager) GetMainDir() string {
	return d.mainDir
}

// GetVsRootDir 获取主的目录
func (d *DirManager) GetVsRootDir() string {
	return d.vSRootDir
}

// InitMainDir 初始化
func (d *DirManager) InitMainDir() {
	vsRootDir := d.vSRootDir
	if vsRootDir == "" {
		return
	}

	strMainDir := ""
	if strings.HasPrefix(d.configRelativeDir, ".") {
		if d.configRelativeDir != "./" {
			strMainDir = vsRootDir + d.configRelativeDir
		} else {
			strMainDir = vsRootDir
		}
	} else {
		strMainDir = d.configRelativeDir
	}

	strMainDir, _ = filepath.Abs(strMainDir)
	// 最后得到的转换路径为g:/luaproject
	strMainDir = pathpre.GeConvertPathFormat(strMainDir)
	strMainDir = strings.Replace(strMainDir, "./", "/", -1)

	log.Debug("vsRootDir=%s, strMainDir=%s", vsRootDir, strMainDir)

	d.mainDir = strMainDir
}

// SetSubDirs set sub dir vec
func (d *DirManager) SetSubDirs(subList []string) {
	d.subDirVec = subList
}

// IsInWorkspace 判断当前文件是否在当前工作区间下
// filepath 为文件的绝对路径
func (d *DirManager) IsInWorkspace(filepath string) bool {
	if strings.HasPrefix(filepath, d.mainDir) {
		return true
	}
	for _, subDir := range d.subDirVec {
		if strings.HasPrefix(filepath, subDir) {
			return true
		}
	}
	return false
}

// GetMainDirFileList 根据目录获取所有需要分析的lua文件
func (d *DirManager) GetMainDirFileList() (fileList []string) {
	fileList = d.GetDirFileList(d.mainDir, true)
	return fileList
}

// GetSubDirsFileList get sub dir vec file list
func (d *DirManager) GetSubDirsFileList() (fileList []string) {
	for _, subDir := range d.subDirVec {
		subFileList := d.GetDirFileList(subDir, false)
		fileList = append(fileList, subFileList...)
	}
	return fileList
}

// GetPathFileList 获取指定目录lua文件列表
func (d *DirManager) GetPathFileList(path string) (fileList []string) {
	if path == "" {
		return fileList
	}

	fileList = d.GetDirFileList(path, false)
	return fileList
}

// ParallelRun 并行获取目录文件列表的对象
type ParallelRun struct {
	sem chan struct{}

	wg sync.WaitGroup
}

func NewRun(maxPar int) *ParallelRun {
	return &ParallelRun{
		sem: make(chan struct{}, maxPar),
	}
}

func (r *ParallelRun) Add() {
	r.wg.Add(1)
}

func (r *ParallelRun) Done() {
	r.wg.Done()
}

func (r *ParallelRun) Acquire() {
	r.sem <- struct{}{}
}

func (r *ParallelRun) Release() {
	<-r.sem
}

func (r *ParallelRun) Wait() error {
	r.wg.Wait()

	return nil
}

// GetDirFileList 获取所有的文件列表
func (d *DirManager) GetDirFileList(path string, ignoreFlag bool) (fileList []string) {
	if path == "" {
		log.Error("GetDirFileList path is empty.")
		return fileList
	}

	time1 := time.Now()
	corNum := runtime.NumCPU()
	run := NewRun(corNum)

	run.Add()

	// 遍历文件的chan结果
	fileChan := make(chan string)
	go getAllFile(run, path, path, ignoreFlag, fileChan)

	go func() {
		run.Wait()
		close(fileChan)
	}()

	for strName := range fileChan {
		fileList = append(fileList, strName)
	}

	ftime := time.Since(time1).Milliseconds()
	log.Debug("GetDirFileList all size=%d, cost time=%d(ms)", len(fileList), ftime)
	return fileList
}

func dirents(run *ParallelRun, dir string) []os.FileInfo {
	run.Acquire()
	defer run.Release()

	rd, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Error("ReadDir pathname=%s, err=%s", dir, err.Error())
	}

	return rd
}

// 递归获取目录下面所有的lua文件
// dirStr 传人的子目录的原始路径
// ignoreFlag 表示是否需要判断忽略文件
func getAllFile(run *ParallelRun, pathname string, dirStr string, ignoreFlag bool, fileChan chan<- string) {
	defer run.Done()

	g := GConfig
	rd := dirents(run, pathname)

	for _, fi := range rd {
		// 首先判断是否linux软链接
		symlinkFlag := false
		if fi.Mode()&os.ModeSymlink != 0 {
			symlinkFlag = true
		}

		if fi.IsDir() || symlinkFlag {
			strName := fi.Name()
			if strName == ".svn" || strName == ".git" {
				continue
			}

			completeStr := pathname
			if !strings.HasSuffix(pathname, "/") {
				completeStr += "/"
			}

			completeStr += strName + "/"
			suffixStr := completeStr[len(dirStr)+1:]

			if ignoreFlag && g.isIgnoreFloder(suffixStr) {
				continue
			}

			run.Add()
			go getAllFile(run, completeStr, dirStr, ignoreFlag, fileChan)

			// 如果是单纯的文件夹，后面不用再判断是否为文件
			if !symlinkFlag {
				continue
			}
		}

		completeStr := pathname
		if !strings.HasSuffix(pathname, "/") {
			completeStr += "/"
		}
		completeStr += fi.Name()

		if !g.IsHandleAsLua(completeStr) {
			continue
		}

		suffixStr := completeStr[len(dirStr)+1:]
		if ignoreFlag && g.isIgnoreFile(suffixStr) {
			continue
		}

		fileChan <- completeStr
	}
}

// MatchCompleteReferFile match complete refer file
// curFile is cur file
func (d *DirManager) MatchCompleteReferFile(curFile string, referStr string) (referFile string) {
	matchDir := d.matchBestDir(curFile)
	if matchDir == "" {
		matchDir = d.mainDir
	}

	// 1) 如果是包含了后缀，查看路径下面是否直接有, 直接存在返回true
	strPath := d.GetCompletePath(matchDir, referStr)
	g := GConfig
	if g.FileExistCache(strPath) {
		// lua文件存在，正常
		referFile = strPath
	}

	return
}

// MatchAllDirReferFile 所有目录下匹配路径
func (d *DirManager) MatchAllDirReferFile(curFile string, referStr string) (referFile string) {
	if d.mainDir == "" {
		return
	}

	g := GConfig
	strPath := d.GetCompletePath(d.mainDir, referStr)
	if g.FileExistCache(strPath) {
		// lua文件存在，正常
		referFile = strPath
		return
	}

	for _, subDir := range d.subDirVec {
		strPath := d.GetCompletePath(subDir, referStr)
		if g.FileExistCache(strPath) {
			// lua文件存在，正常
			referFile = strPath
			return
		}
	}

	return referFile
}

func (d *DirManager) matchBestDir(curFile string) (matchDir string) {
	if d.mainDir == "" {
		return
	}

	bestLen := 0
	if strings.HasPrefix(curFile, d.mainDir) {
		bestLen = len(d.mainDir)
		matchDir = d.mainDir
	}

	if d.clientExtLuaPath != "" && strings.HasPrefix(curFile, d.clientExtLuaPath) {
		bestLen = len(d.clientExtLuaPath)
		matchDir = d.clientExtLuaPath
	}

	for _, subDir := range d.subDirVec {
		if !strings.HasPrefix(curFile, subDir) {
			continue
		}

		if len(subDir) > bestLen {
			bestLen = len(subDir)
			matchDir = subDir
		}
	}

	return matchDir
}

// IsInDir juget file is in project dir
func (d *DirManager) IsInDir(strFile string) bool {
	if d.mainDir == "" {
		return false
	}

	if strings.HasPrefix(strFile, d.mainDir) {
		return true
	}

	if strings.HasPrefix(strFile, d.clientExtLuaPath) {
		return true
	}

	for _, subDir := range d.subDirVec {
		if strings.HasPrefix(subDir, strFile) {
			return true
		}
	}

	return false
}

// GetCompletePath 传人目录和后面文件名，拼接成完整的路径
func (d *DirManager) GetCompletePath(baseDir string, fileName string) string {
	strPath := baseDir
	if !strings.HasSuffix(baseDir, "/") {
		strPath = strPath + "/"
	}

	strPath = strPath + fileName
	return strPath
}

// GetAllCompleFile 获取输入文件补全的文件
// suffixFlag 表示引入其他文件时候，是否要包含文件的后缀名
func (d *DirManager) GetAllCompleFile(pathname string, referType ReferType,
	suffixFlag bool, tipList *[]OneTipFile) error {
	mainDir := d.mainDir
	if mainDir == "" {
		return nil
	}

	g := GConfig
	rd, err := ioutil.ReadDir(pathname)
	for _, fi := range rd {
		if fi.IsDir() {
			strName := fi.Name()
			if strName == ".svn" || strName == ".git" {
				continue
			}

			completeStr := pathname
			if !strings.HasSuffix(pathname, "/") {
				completeStr += "/"
			}

			completeStr += strName + "/"
			suffixStr := completeStr[len(mainDir)+1:]

			if g.isIgnoreFloder(suffixStr) {
				continue
			}

			d.GetAllCompleFile(completeStr, referType, suffixFlag, tipList)
			continue
		}

		completeStr := pathname
		if !strings.HasSuffix(pathname, "/") {
			completeStr += "/"
		}
		pathBeforeStr := completeStr[len(mainDir)+1:]

		// 判断是否要处理
		completeStr += fi.Name()
		suffixStr := completeStr[len(mainDir)+1:]
		// 需要忽略的文件
		if g.isIgnoreFile(suffixStr) {
			continue
		}

		// 文件为so类型
		soFlag := false
		if referType == ReferTypeRequire && strings.HasSuffix(fi.Name(), ".so") {
			soFlag = true
		}

		if !g.IsHandleAsLua(suffixStr) && !soFlag {
			continue
		}

		// 获取下路径分割符
		pathSeparator := GConfig.GetPathSeparator()
		if pathSeparator != "/" {
			// 分割符进行替换
			pathBeforeStr = strings.Replace(pathBeforeStr, "/", pathSeparator, -1)
		}

		// 文件名的前缀
		fileNamePre := CompleteFilePathToPreStr(fi.Name())
		if fileNamePre == "" {
			continue
		}

		if soFlag {
			oneTipFile := OneTipFile{
				StrName: pathBeforeStr + fileNamePre,
				StrDesc: "so file",
			}

			*tipList = append(*tipList, oneTipFile)
		} else {
			// 非so文件的处理
			if suffixFlag {
				// 包含后缀
				completeStr = pathBeforeStr + fi.Name()
			} else {
				// 不包含后缀
				completeStr = pathBeforeStr + fileNamePre
			}

			oneTipFile := OneTipFile{
				StrName: completeStr,
				StrDesc: "lua file",
			}

			*tipList = append(*tipList, oneTipFile)
		}
	}
	return err
}

type scoredMatchStruct struct {
	candidateStr string
	score        int
}

type resultSorterMatch struct {
	results []scoredMatchStruct
}

/*
 * sort.Interface methods
 */
func (s *resultSorterMatch) Len() int { return len(s.results) }
func (s *resultSorterMatch) Less(i, j int) bool {
	iscore, jscore := s.results[i].score, s.results[j].score
	return iscore > jscore
}
func (s *resultSorterMatch) Swap(i, j int) {
	s.results[i], s.results[j] = s.results[j], s.results[i]
}

// 计算匹配字符串的得分
func calcMatchStrScore(fileName string, referFileName string, condidateStr string) (score int) {
	score = 0

	lastIndex := strings.LastIndex(condidateStr, referFileName)
	if lastIndex == -1 {
		score = -1000000
		return
	}

	// 前面的内容
	preStr := condidateStr[0:lastIndex]

	// 按/进行切分，看有能切分成多少
	splitVec := strings.Split(preStr, "/")

	score = score + len(splitVec)*(-1000)

	// 原生的路径，也进行切分，看匹配程度
	splitOldVec := strings.Split(fileName, "/")
	for index, splitStr := range splitVec {
		if index >= len(splitOldVec) {
			break
		}

		if splitStr == splitOldVec[index] {
			score = score + 10
		} else {
			break
		}
	}

	return
}

// GetBestMatchReferFile lua文件中引入另外一个文件时候，模糊匹配文件
// curFile 为当前的lua文件名
// referFile 为引用的lua文件名
// allFilesMap 为项目中所有包含的lua文件
// 返回值为匹配最合适的文件
func GetBestMatchReferFile(curFile string, referFile string, allFilesMap map[string]struct{}) (findStr string) {
	// 首先判断传人的文件是否带有后缀的
	suffixFlag := false
	seperateIndex := strings.Index(referFile, ".")
	if seperateIndex >= 0 {
		// 文件名中带有 . 分割符，说明是带有完整的后缀的
		suffixFlag = true
	}

	candidateVec := []string{}
	lenReferFile := len(referFile)
	for strFile := range allFilesMap {
		referFileTmp := referFile

		if suffixFlag {
			if len(strFile) > lenReferFile {
				referFileTmp = "/" + referFile
			}

			// 如果是带了后缀，拿后缀的去匹配
			if !strings.HasSuffix(strFile, referFileTmp) {
				continue
			}
		} else {
			// 如果不带后缀，map中的路径名先提取
			preFile := completeFilePathToPreStr(strFile)
			if preFile == "" {
				continue
			}

			if len(preFile) > lenReferFile {
				referFileTmp = "/" + referFile
			}

			if !strings.HasSuffix(preFile, referFileTmp) {
				continue
			}
		}

		// 包含引入的后缀
		candidateVec = append(candidateVec, strFile)
	}

	// 没有找到任何匹配的
	if len(candidateVec) == 0 {
		return
	}

	// 只有一个匹配的，直接返回，没有可挑选的
	if len(candidateVec) == 1 {
		findStr = candidateVec[0]
		return
	}

	matchResults := &resultSorterMatch{
		results: make([]scoredMatchStruct, 0),
	}

	for _, condidateStr := range candidateVec {
		oneResult := scoredMatchStruct{
			candidateStr: condidateStr,
			score:        calcMatchStrScore(curFile, referFile, condidateStr),
		}
		matchResults.results = append(matchResults.results, oneResult)
	}

	sort.Sort(matchResults)
	findStr = matchResults.results[0].candidateStr
	return
}

// IsDirExistWorkspace 判断当前文件夹下是否存在于 当前项目中的mainDir 和 subDirs中
func (d *DirManager) IsDirExistWorkspace(path string) bool {
	if !filefolder.IsDirExist(path) {
		return false
	}

	if d.mainDir != "" && IsSubDir(path, d.mainDir) {
		return true
	}

	for _, dir := range d.subDirVec {
		if IsSubDir(path, dir) {
			return true
		}
	}
	return false
}

// PushOneSubDir 新增一个子文件夹
func (d *DirManager) PushOneSubDir(subDir string) {
	d.subDirVec = append(d.subDirVec, subDir)
}

// RemoveOneSubDir 移除一个子文件夹
func (d *DirManager) RemoveOneSubDir(subDir string) bool {
	// 获取当前文件夹在subDirs 下的索引
	fileIndex := -1
	for index, dir := range d.subDirVec {
		if dir == subDir {
			fileIndex = index
		}
	}

	if fileIndex == -1 {
		return false
	}

	// 若移除的当前workspace 文件夹中包含的子文件夹， 则不需要做任何处理
	if !d.IsDirExistWorkspace(subDir) {
		return false
	}
	d.subDirVec = append(d.subDirVec[0:fileIndex], d.subDirVec[fileIndex+1:]...)

	return true
}

// RemovePathDirPre 移除路径目录的前缀
func (d *DirManager) RemovePathDirPre(strPath string) (strFile string) {
	bestDir := d.matchBestDir(strPath)
	if bestDir == "" {
		return strPath
	}

	strFile = strings.TrimPrefix(strPath, bestDir+"/")
	return strFile
}

// IsSubDir 判断传入的path 是否是另一个文件夹dir 下的子文件夹
func IsSubDir(path string, dir string) bool {
	if path == dir {
		return true
	}

	pathLen := len(path)
	dirLen := len(dir)

	if pathLen >= dirLen {
		return false
	}

	if strings.HasPrefix(dir, path) && dir[pathLen] == '/' {
		return true
	}

	return false
}
