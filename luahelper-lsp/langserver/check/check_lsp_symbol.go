package check

import (
	"bytes"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/check/results"
	"luahelper-lsp/langserver/log"
	"reflect"
	"runtime"
	"sort"
	"unicode"
)

const (
	// 最终返回的最大结果
	maxSymbols = 200
	// 最低分
	minScore = -1000
	// 最大的候选查询字符
	maxInputSize = 127
	// 最大输入的查询字符
	maxPatternSize = 63
)

// FindFileAllSymbol 查找指定文件的所有符号
func (a *AllProject) FindFileAllSymbol(strFile string) (symbolVec []common.FileSymbolStruct) {
	fileStruct := a.getVailidCacheFileStruct(strFile)
	if fileStruct == nil {
		log.Error("FindFileAllSymbol error, not find file=%s", strFile)
		return
	}

	// 1) 找到对应的文件进行查找所有的当前符号
	symbolVec = fileStruct.FileResult.FindAllSymbol()

	// 2) 获取文件的注解类型信息, 用于显示当前文档的符号
	annotateSymbolVec := a.GetAnnotateFileSymbolStruct(strFile)
	symbolVec = append(symbolVec, annotateSymbolVec...)
	return
}

// FindWorkspaceAllSymbol 查找工程内所有全局符号
func (a *AllProject) FindWorkspaceAllSymbol(strContent string) (symbolVec []common.FileSymbolStruct) {
	resultSort := &resultSorter{
		results: make([]scoredSymbol, 0),
	}

	// 获取需要查询的文件列表
	fileList := make([]string, 0, len(a.fileStructMap))
	for fileName, fileStruct := range a.fileStructMap {
		if fileStruct.GetFileHandleErr() != results.FileHandleOk {
			continue
		}
		fileList = append(fileList, fileName)
	}

	handleAllFilesSymbols(strContent, a, resultSort, fileList)

	sort.Sort(resultSort)
	log.Debug("handle workspace symbols, query all %d files, find all %d symbols", len(a.fileStructMap), len(resultSort.results))
	if len(resultSort.results) > maxSymbols {
		resultSort.results = resultSort.results[:maxSymbols]
	}

	for _, oneResult := range resultSort.results {
		symbolVec = append(symbolVec, *oneResult.fileSymbol)
	}

	return symbolVec
}

// RuneRole specifies the role of a rune in the context of an input.
type RuneRole byte

const (
	// RNone specifies a rune without any role in the input (i.e., whitespace/non-ASCII).
	RNone RuneRole = iota
	// RSep specifies a rune with the role of segment separator.
	RSep
	// RTail specifies a rune which is a lower-case tail in a word in the input.
	RTail
	// RUCTail specifies a rune which is an upper-case tail in a word in the input.
	RUCTail
	// RHead specifies a rune which is the first character in a word in the input.
	RHead
)

// RuneRoles detects the roles of each byte rune in an input string and stores it in the output
// slice. The rune role depends on the input type. Stops when it parsed all the runes in the string
// or when it filled the output. If output is nil, then it gets created.
func RuneRoles(str string, reuse []RuneRole) []RuneRole {
	var output []RuneRole
	if cap(reuse) < len(str) {
		output = make([]RuneRole, 0, len(str))
	} else {
		output = reuse[:0]
	}

	prev, prev2 := rtNone, rtNone
	for i := 0; i < len(str); i++ {
		r := rune(str[i])

		role := RNone

		curr := rtLower
		if str[i] <= unicode.MaxASCII {
			curr = runeType(rt[str[i]] - '0')
		}

		if curr == rtLower {
			if prev == rtNone || prev == rtPunct {
				role = RHead
			} else {
				role = RTail
			}
		} else if curr == rtUpper {
			role = RHead

			if prev == rtUpper {
				// This and previous characters are both upper case.

				if i+1 == len(str) {
					// This is last character, previous was also uppercase -> this is UCTail
					// i.e., (current char is C): aBC / BC / ABC
					role = RUCTail
				}
			}
		} else if curr == rtPunct {
			switch r {
			case '.', ':':
				role = RSep
			}
		}
		if curr != rtLower {
			if i > 1 && output[i-1] == RHead && prev2 == rtUpper && (output[i-2] == RHead || output[i-2] == RUCTail) {
				// The previous two characters were uppercase. The current one is not a lower case, so the
				// previous one can't be a HEAD. Make it a UCTail.
				// i.e., (last char is current char - B must be a UCTail): ABC / ZABC / AB.
				output[i-1] = RUCTail
			}
		}

		output = append(output, role)
		prev2 = prev
		prev = curr
	}
	return output
}

type runeType byte

const (
	rtNone runeType = iota
	rtPunct
	rtLower
	rtUpper
)

const rt = "00000000000000000000000000000000000000000000001122222222221000000333333333333333333333333330000002222222222222222222222222200000"

// LastSegment returns the substring representing the last segment from the input, where each
// byte has an associated RuneRole in the roles slice. This makes sense only for inputs of Symbol
// or Filename type.
func LastSegment(input string, roles []RuneRole) string {
	// Exclude ending separators.
	end := len(input) - 1
	for end >= 0 && roles[end] == RSep {
		end--
	}
	if end < 0 {
		return ""
	}

	start := end - 1
	for start >= 0 && roles[start] != RSep {
		start--
	}

	return input[start+1 : end+1]
}

// ToLower transforms the input string to lower case, which is stored in the output byte slice.
// The lower casing considers only ASCII values - non ASCII values are left unmodified.
// Stops when parsed all input or when it filled the output slice. If output is nil, then it gets
// created.
func ToLower(input string, reuse []byte) []byte {
	output := reuse
	if cap(reuse) < len(input) {
		output = make([]byte, len(input))
	}

	for i := 0; i < len(input); i++ {
		r := rune(input[i])
		if r <= unicode.MaxASCII {
			if 'A' <= r && r <= 'Z' {
				r += 'a' - 'A'
			}
		}
		output[i] = byte(r)
	}
	return output[:len(input)]
}

type scoreVal int

func (s scoreVal) val() int {
	return int(s) >> 1
}

func (s scoreVal) prevK() int {
	return int(s) & 1
}

func score(val int, prevK int /*0 or 1*/) scoreVal {
	return scoreVal(val<<1 + prevK)
}

// Matcher 所有查询的结构
type Matcher struct {
	pattern       string
	patternLower  []byte // lower-case version of the pattern
	patternShort  []byte // first characters of the pattern
	caseSensitive bool   // set if the pattern is mix-cased

	patternRoles []RuneRole // the role of each character in the pattern
	roles        []RuneRole // the role of each character in the tested string

	scores [maxInputSize + 1][maxPatternSize + 1][2]scoreVal

	scoreScale float32

	lastCandidateLen     int // in bytes
	lastCandidateMatched bool

	// Here we save the last candidate in lower-case. This is basically a byte slice we reuse for
	// performance reasons, so the slice is not reallocated for every candidate.
	lowerBuf [maxInputSize]byte
	rolesBuf [maxInputSize]RuneRole
}

func (m *Matcher) bestK(i, j int) int {
	if m.scores[i][j][0].val() < m.scores[i][j][1].val() {
		return 1
	}
	return 0
}

// NewMatcher returns a new fuzzy matcher for scoring candidates against the provided pattern.
func NewMatcher(pattern string) *Matcher {
	if len(pattern) > maxPatternSize {
		pattern = pattern[:maxPatternSize]
	}

	m := &Matcher{
		pattern:      pattern,
		patternLower: ToLower(pattern, nil),
	}

	for i, c := range m.patternLower {
		if pattern[i] != c {
			m.caseSensitive = true
			break
		}
	}

	if len(pattern) > 3 {
		m.patternShort = m.patternLower[:3]
	} else {
		m.patternShort = m.patternLower
	}

	m.patternRoles = RuneRoles(pattern, nil)

	if len(pattern) > 0 {
		maxCharScore := 4
		m.scoreScale = 1 / float32(maxCharScore*len(pattern))
	}

	return m
}

// Score returns the score returned by matching the candidate to the pattern.
// This is not designed for parallel use. Multiple candidates must be scored sequentially.
// Returns a score between 0 and 1 (0 - no match, 1 - perfect match).
func (m *Matcher) Score(candidate string) float32 {
	if len(candidate) > maxInputSize {
		candidate = candidate[:maxInputSize]
	}
	lower := ToLower(candidate, m.lowerBuf[:])
	m.lastCandidateLen = len(candidate)

	if len(m.pattern) == 0 {
		// Empty patterns perfectly match candidates.
		return 1
	}

	if m.match(candidate, lower) {
		sc := m.computeScore(candidate, lower)
		if sc > minScore/2 && !m.poorMatch() {
			m.lastCandidateMatched = true
			if len(m.pattern) == len(candidate) {
				// Perfect match.
				return 1
			}

			if sc < 0 {
				sc = 0
			}
			normalizedScore := float32(sc) * m.scoreScale
			if normalizedScore > 1 {
				normalizedScore = 1
			}

			return normalizedScore
		}
	}

	m.lastCandidateMatched = false
	return 0
}

func (m *Matcher) match(candidate string, candidateLower []byte) bool {
	i, j := 0, 0
	for ; i < len(candidateLower) && j < len(m.patternLower); i++ {
		if candidateLower[i] == m.patternLower[j] {
			j++
		}
	}
	if j != len(m.patternLower) {
		return false
	}

	// The input passes the simple test against pattern, so it is time to classify its characters.
	// Character roles are used below to find the last segment.
	m.roles = RuneRoles(candidate, m.rolesBuf[:])

	return true
}

func (m *Matcher) computeScore(candidate string, candidateLower []byte) int {
	pattLen, candLen := len(m.pattern), len(candidate)

	for j := 0; j <= len(m.pattern); j++ {
		m.scores[0][j][0] = minScore << 1
		m.scores[0][j][1] = minScore << 1
	}
	m.scores[0][0][0] = score(0, 0) // Start with 0.

	segmentsLeft, lastSegStart := 1, 0
	for i := 0; i < candLen; i++ {
		if m.roles[i] == RSep {
			segmentsLeft++
			lastSegStart = i + 1
		}
	}

	// A per-character bonus for a consecutive match.
	consecutiveBonus := 2
	wordIdx := 0 // Word count within segment.
	for i := 1; i <= candLen; i++ {

		role := m.roles[i-1]
		isHead := role == RHead

		if isHead {
			wordIdx++
		} else if role == RSep && segmentsLeft > 1 {
			wordIdx = 0
			segmentsLeft--
		}

		var skipPenalty int
		if i == 1 || (i-1) == lastSegStart {
			// Skipping the start of first or last segment.
			skipPenalty++
		}

		for j := 0; j <= pattLen; j++ {
			// By default, we don't have a match. Fill in the skip data.
			m.scores[i][j][1] = minScore << 1

			// Compute the skip score.
			k := 0
			if m.scores[i-1][j][0].val() < m.scores[i-1][j][1].val() {
				k = 1
			}

			skipScore := m.scores[i-1][j][k].val()
			// Do not penalize missing characters after the last matched segment.
			if j != pattLen {
				skipScore -= skipPenalty
			}
			m.scores[i][j][0] = score(skipScore, k)

			if j == 0 || candidateLower[i-1] != m.patternLower[j-1] {
				// Not a match.
				continue
			}
			pRole := m.patternRoles[j-1]

			if role == RTail && pRole == RHead {
				if j > 1 {
					// Not a match: a head in the pattern matches a tail character in the candidate.
					continue
				}
				// Special treatment for the first character of the pattern. We allow
				// matches in the middle of a word if they are long enough, at least
				// min(3, pattern.length) characters.
				if !bytes.HasPrefix(candidateLower[i-1:], m.patternShort) {
					continue
				}
			}

			// Compute the char score.
			var charScore int
			// Bonus 1: the char is in the candidate's last segment.
			if segmentsLeft <= 1 {
				charScore++
			}
			// Bonus 2: Case match or a Head in the pattern aligns with one in the word.
			// Single-case patterns lack segmentation signals and we assume any character
			// can be a head of a segment.
			if candidate[i-1] == m.pattern[j-1] || role == RHead && (!m.caseSensitive || pRole == RHead) {
				charScore++
			}

			// Penalty 1: pattern char is Head, candidate char is Tail.
			if role == RTail && pRole == RHead {
				charScore--
			}
			// Penalty 2: first pattern character matched in the middle of a word.
			if j == 1 && role == RTail {
				charScore -= 4
			}

			// Third dimension encodes whether there is a gap between the previous match and the current
			// one.
			for k := 0; k < 2; k++ {
				sc := m.scores[i-1][j-1][k].val() + charScore

				isConsecutive := k == 1 || i-1 == 0 || i-1 == lastSegStart
				if isConsecutive {
					// Bonus 3: a consecutive match. First character match also gets a bonus to
					// ensure prefix final match score normalizes to 1.0.
					// Logically, this is a part of charScore, but we have to compute it here because it
					// only applies for consecutive matches (k == 1).
					sc += consecutiveBonus
				}
				if k == 0 {
					// Penalty 3: Matching inside a segment (and previous char wasn't matched). Penalize for the lack
					// of alignment.
					if role == RTail || role == RUCTail {
						sc -= 3
					}
				}

				if sc > m.scores[i][j][1].val() {
					m.scores[i][j][1] = score(sc, k)
				}
			}
		}
	}

	result := m.scores[len(candidate)][len(m.pattern)][m.bestK(len(candidate), len(m.pattern))].val()

	return result
}

func (m *Matcher) poorMatch() bool {
	if len(m.pattern) < 2 {
		return false
	}

	i, j := m.lastCandidateLen, len(m.pattern)
	k := m.bestK(i, j)

	var counter, len int
	for i > 0 {
		take := (k == 1)
		k = m.scores[i][j][k].prevK()
		if take {
			len++
			if k == 0 && len < 3 && m.roles[i-1] == RTail {
				// Short match in the middle of a word
				counter++
				if counter > 1 {
					return true
				}
			}
			j--
		} else {
			len = 0
		}
		i--
	}
	return false
}

// resultSorter is a utility struct for collecting, filtering, and
// sorting symbol results.
type resultSorter struct {
	m       *Matcher
	results []scoredSymbol
}

// scoredSymbol is a symbol with an attached search relevancy score.
// It is used internally by resultSorter.
type scoredSymbol struct {
	score      float32
	fileSymbol *common.FileSymbolStruct
}

/*
 * sort.Interface methods
 */
func (rs *resultSorter) Len() int { return len(rs.results) }
func (rs *resultSorter) Less(i, j int) bool {
	return rs.results[i].score > rs.results[j].score
}
func (rs *resultSorter) Swap(i, j int) {
	rs.results[i], rs.results[j] = rs.results[j], rs.results[i]
}

// 根据当前varInfo 中的信息 赋予需要返回的symbol信息
func addSymbolInfo(oneSymbol *common.FileSymbolStruct, varInfo *common.VarInfo, isglobal bool) {
	if varInfo.ExtraGlobal != nil {
		if varInfo.ExtraGlobal.GFlag {
			oneSymbol.ContainerName = "_G"
		} else if varInfo.ExtraGlobal.StrProPre != "" {
			oneSymbol.ContainerName = varInfo.ExtraGlobal.StrProPre
		} else {
			oneSymbol.ContainerName = "global"
		}
	} else {
		if isglobal {
			oneSymbol.ContainerName = "global"
		} else {
			oneSymbol.ContainerName = "local"
		}
	}

	if varInfo.ReferFunc != nil {
		oneSymbol.Kind = common.IKFunction
	}
}

// collect is a thread-safe method that will record the passed-in
// symbol in the list of results if its score > 0.
func (rs *resultSorter) collect(strName, fileName, prefix string, isglobal bool, varInfo *common.VarInfo) {

	if prefix != "" {
		strName = prefix + "." + strName
	}

	score := rs.m.Score(strName)

	if score < 0 {
		return
	}

	oneSymbol := &common.FileSymbolStruct{
		Name:     strName,
		Kind:     common.IKVariable,
		Loc:      varInfo.Loc,
		FileName: fileName,
	}

	addSymbolInfo(oneSymbol, varInfo, isglobal)

	sc := scoredSymbol{score, oneSymbol}
	rs.results = append(rs.results, sc)
}

// 根据query的内容 进行查找
func (rs *resultSorter) getQuerySymbols(fileResult *results.FileResult) {
	fileName := fileResult.Name
	// 查找全局变量
	for strName, oneVar := range fileResult.GlobalMaps {
		if oneVar.ExtraGlobal.GFlag {
			rs.collect(strName, fileName, "_G", true, oneVar)
			// 若前缀包含 且是表结构，则返回所有信息， 否则按照条件进行查找
			// rs.getVarmapsSymbols(oneVar.SubMaps, fileName, "_G", false, true, false)
		} else {
			rs.collect(strName, fileName, "", true, oneVar)
			// 若前缀包含 且是表结构，则返回所有信息， 否则按照条件进行查找
			rs.getVarmapsSymbols(oneVar.SubMaps, fileName, strName, true, false)
		}
	}
	// 搜索协议前缀符号
	rs.getProtocolMapsSymbols(fileResult.ProtocolMaps, fileName)

	// 搜索local 变量
	// c. 搜索全部局部变量
	// 定义当前文件下的scope索引
	curFileExcludeScopes := make(map[*common.ScopeInfo]bool)
	// c.1 查找所有的globalScope
	gSopes := fileResult.FindGMapsScopes()
	for _, scope := range gSopes {
		curFileExcludeScopes[scope] = true
	}

	// c.2 再查找mainScope下的局部变量（函数和变量）
	mainScope := fileResult.MainFunc.MainScope
	rs.getLocVarMapsSymbols(mainScope.LocVarMap, fileName, false, false)

	// c.3 最后查找所有的subScopes（只查找函数）
	scopes := make([]*common.ScopeInfo, 0, len(mainScope.SubScopes))
	scopes = append(scopes, mainScope.SubScopes...)
	for len(scopes) != 0 {
		tempScopes := []*common.ScopeInfo{}
		for _, scope := range scopes {
			if curFileExcludeScopes[scope] {
				delete(curFileExcludeScopes, scope)
				continue
			}

			rs.getLocVarMapsSymbols(scope.LocVarMap, fileName, false, true)
			tempScopes = append(tempScopes, scope.SubScopes...)
		}
		scopes = tempScopes
	}
}

// 协议变量下满足的符号信息
func (rs *resultSorter) getProtocolMapsSymbols(protocolMaps map[string]*common.VarInfo, fileName string) {
	if protocolMaps == nil {
		return
	}
	for strName, oneVar := range protocolMaps {
		for {
			rs.collect(strName, fileName, oneVar.ExtraGlobal.StrProPre, true, oneVar)
			if oneVar.ExtraGlobal.Prev == nil {
				break
			}
			oneVar = oneVar.ExtraGlobal.Prev
		}
	}
}

// 局部变量下满足的符号信息
func (rs *resultSorter) getLocVarMapsSymbols(locVarMap map[string]*common.VarInfoList, fileName string, hasPrefix,
	onlyFunc bool) {
	for strName, oneVar := range locVarMap {
		varLen := len(oneVar.VarVec)
		if varLen <= 0 {
			continue
		}

		vars := oneVar.VarVec[varLen-1]

		if !hasPrefix {
			if onlyFunc && vars.SubMaps == nil && vars.ReferFunc == nil {
				continue
			}
			rs.collect(strName, fileName, "", false, vars)

			// 获取submap下的符号
			rs.getVarmapsSymbols(vars.SubMaps, fileName, strName, false, onlyFunc)

		} else {
			for strSubName, oneVar := range vars.SubMaps {
				rs.collect(strSubName, fileName, strName, false, oneVar)
			}
		}
	}
}

// 获取当前varInfoMaps下满足查询条件的symbols, 用于 getQuerySymbols 中指定查找过程
func (rs *resultSorter) getVarmapsSymbols(varmaps map[string]*common.VarInfo, fileName, strName string,
	isglobal, onlyFunc bool) {
	for strSubName, subOneVar := range varmaps {
		if onlyFunc && subOneVar.ReferFunc == nil {
			continue
		}
		rs.collect(strSubName, fileName, strName, isglobal, subOneVar)
	}
}

// 获取文件注解参数的符号
func (a *AllProject) getAnnotateFileSymbol(rs *resultSorter, fileName string) {
	annotateSymbolVec := a.GetAnnotateFileSymbolStruct(fileName)
	if len(annotateSymbolVec) == 0 {
		return
	}

	for _, symbol := range annotateSymbolVec {
		strName := symbol.Name
		score := rs.m.Score(strName)
		if score < 0 {
			continue
		}

		oneSymbol := &common.FileSymbolStruct{
			Name:          strName,
			Kind:          symbol.Kind,
			Loc:           symbol.Loc,
			FileName:      fileName,
			ContainerName: "annotate class",
		}
		if symbol.Kind == common.IKAnnotateAlias {
			oneSymbol.ContainerName = "annotate alias"
		}

		sc := scoredSymbol{score, oneSymbol}
		rs.results = append(rs.results, sc)
	}
}

type symbolsChan struct {
	strfile          string
	sendRunFlag      bool
	sendResultSorter *resultSorter
	sendAllProject   *AllProject

	returnResult []scoredSymbol
}

func handleAllFilesSymbols(pattern string, allProject *AllProject, results *resultSorter, fileList []string) {
	// 定义最终的results 和每次协程需要处理的结果
	resultSorters := make([]*resultSorter, len(fileList))

	//获取本机核心数
	handleFileLen := len(fileList)
	corNum := runtime.NumCPU() + 2
	if handleFileLen < corNum {
		corNum = handleFileLen
	}

	//协程组的通道
	chs := make([]chan symbolsChan, corNum)
	//创建反射对象集合，用于监听
	selectCase := make([]reflect.SelectCase, corNum)
	//开启协程池
	for i := 0; i < corNum; i++ {
		chs[i] = make(chan symbolsChan)
		selectCase[i].Dir = reflect.SelectRecv
		selectCase[i].Chan = reflect.ValueOf(chs[i])
		go goroutineFindSymbols(chs[i])
	}
	//初始化协程
	for i := 0; i < corNum; i++ {
		resultSorters[i] = &resultSorter{
			m:       NewMatcher(pattern),
			results: make([]scoredSymbol, 0),
		}
		chanRequest := symbolsChan{
			strfile:          fileList[i],
			sendRunFlag:      true,
			sendAllProject:   allProject,
			sendResultSorter: resultSorters[i],
		}
		chs[i] <- chanRequest
	}

	//reflect接收数据
	taskDone := 0
	for recvNum := 0; recvNum < handleFileLen; {
		chosen, recv, recvOK := reflect.Select(selectCase)
		if !recvOK {
			log.Error("ch%d error\n", chosen)
			//确保循环退出
			recvNum++
			continue
		}

		recvFindSymbol(results, recv.Interface().(symbolsChan))

		if recvNum+corNum < handleFileLen {
			resultSorters[recvNum+corNum] = &resultSorter{
				m:       NewMatcher(pattern),
				results: make([]scoredSymbol, 0),
			}
			chanRequest := symbolsChan{
				strfile:          fileList[recvNum+corNum],
				sendRunFlag:      true,
				sendResultSorter: resultSorters[recvNum+corNum],
				sendAllProject:   allProject,
			}
			chs[chosen] <- chanRequest
		} else {
			chanRequest := symbolsChan{
				sendRunFlag: false,
			}
			chs[chosen] <- chanRequest
		}
		//任务完成数
		taskDone++
		//确保循环退出
		recvNum++
	}
	if taskDone == handleFileLen-1 {
		log.Debug("all files has send work...")
	}
	//关闭所有的子协程
	for i := 0; i < corNum; i++ {
		close(chs[i])
	}
}

func goroutineFindSymbols(ch chan symbolsChan) {
	for {
		request := <-ch
		if !request.sendRunFlag {
			// 协程停止运行
			break
		}
		a := request.sendAllProject
		fileResult := a.fileStructMap[request.strfile].FileResult
		resultSorter := request.sendResultSorter
		resultSorter.getQuerySymbols(fileResult)

		a.getAnnotateFileSymbol(resultSorter, request.strfile)

		var returnChan symbolsChan
		resultLen := len(resultSorter.results)
		if resultLen == 0 {
			returnChan = symbolsChan{returnResult: nil}
		} else if resultLen <= maxSymbols {
			returnChan = symbolsChan{returnResult: resultSorter.results}
		} else {
			sort.Sort(resultSorter)
			returnChan = symbolsChan{returnResult: resultSorter.results[0:maxSymbols]}
		}
		ch <- returnChan
	}
}

//将每个协程得到的结果整合到最终需要返回的结果中
func recvFindSymbol(results *resultSorter, recvData symbolsChan) {
	if recvData.returnResult == nil {
		return
	}
	results.results = append(results.results, recvData.returnResult...)
}
