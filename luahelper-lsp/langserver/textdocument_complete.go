package langserver

import (
	"context"
	"fmt"
	"luahelper-lsp/langserver/check"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/codingconv"
	"luahelper-lsp/langserver/log"
	lsp "luahelper-lsp/langserver/protocol"
	"regexp"
	"strings"
	"unicode"
)

type CompletionItemTmp struct {
	Label string                 `json:"label"`
	Kind  lsp.CompletionItemKind `json:"kind,omitempty"`
	//Detail        string             `json:"detail,omitempty"`
	//Documentation string             `json:"documentation,omitempty"`
	Data interface{} `json:"data,omitempty"`
	//SortText      string             `json:"sortText,omitempty"`
}

type CompletionListTmp struct {
	IsIncomplete bool                `json:"isIncomplete"`
	Items        []CompletionItemTmp `json:"items"`
}

// TextDocumentComplete  代码只能补全（提示）interface{}, error   comList lsp.CompletionListTmp
func (l *LspServer) TextDocumentComplete(ctx context.Context, vs lsp.CompletionParams) (compltionReturn interface{}, err error) {
	l.requestMutex.Lock()
	defer l.requestMutex.Unlock()

	// 判断打开的文件，是否是需要分析的文件
	comResult := l.beginFileRequest(vs.TextDocument.URI, vs.Position)
	if !comResult.result {
		return
	}

	project := l.getAllProject()
	project.ClearCompleteCache()
	strFile := comResult.strFile

	// 1) 判断是否输入的为 --- 注释，用于快捷生成函数定义的注释
	// 输入-时候，传人的为空，特殊处理
	if vs.Context.TriggerCharacter == "" && judgeBeforeCommentHorizontal(comResult.contents, comResult.offset) {
		// 处理快捷生成注解, 以及提升注解系统
		comList, _ := l.handleGenerateComment(strFile, comResult.contents, comResult.offset, (int)(comResult.pos.Line))
		return comList, err
	}

	if vs.Context.TriggerCharacter == "-" {
		// 处理快捷生成注解, 以及提升注解系统
		comList, _ := l.handleGenerateComment(strFile, comResult.contents, comResult.offset, (int)(comResult.pos.Line))
		return comList, err
	}

	// 2) 判断是否输入的为---@ , 用于注解系统
	if vs.Context.TriggerCharacter == "@" {
		// 处理快捷生成---@ 注解的提示
		comList, _ := l.handleGenerateAnnotateArea(strFile, comResult.contents, comResult.offset, (int)(comResult.pos.Line))
		return comList, err
	}

	// 3) 判断这行前面是---@开头的，但是补全注解其他内容
	typeComList, flagType := l.handleGenerateAnnotateType(strFile, comResult.contents, comResult.offset, (int)(comResult.pos.Line))
	if flagType {
		return typeComList, nil
	}

	// 4) 判断是否为提示输入引入其他文件路径补全
	compFileList, flag := l.judgeCompeleteFile(strFile, comResult.contents, comResult.offset)
	if flag {
		return compFileList, nil
	}

	// 5.1) 获取这个代码补全的前缀字符串
	preCompStr, splitByte := getCompeletePreStr(comResult.contents, comResult.offset)
	if preCompStr == "" {
		return
	}

	var compVar common.CompleteVarStruct
	// #字代码补全特殊处理，前面拼接#
	preStr := ""

	paramCandidateType := l.getFuncParamCandidateType(ctx, vs.TextDocument.URI, vs.Position)
	if preCompStr == " " && paramCandidateType == nil {
		return
	}

	// 5.2) 输入的为#代码补全
	if preCompStr == "#" {
		compVar = getDefaultHashtag(comResult)
		preStr = "#"
	} else {
		if (preCompStr == "\"" || preCompStr == "'" || preCompStr == " ") && paramCandidateType != nil {
			if preCompStr == "\"" {
				splitByte = '"'
			} else if preCompStr == "'" {
				splitByte = '\''
			} else {
				splitByte = ' '
			}
			compVar.StrVec = append(compVar.StrVec, preCompStr)
			compVar.OnelyParamQuotesFlag = true
		} else {
			// 5.2) 按照.进行分割字符串
			compVar, flag = getComplelteStruct(preCompStr, (int)(comResult.pos.Line), (int)(comResult.pos.Character))
			if !flag {
				return
			}

			if len(compVar.StrVec) == 1 && !compVar.LastEmptyFlag {
				beforeIndex := comResult.offset - len(preCompStr) - 1
				if beforeIndex >= 0 && comResult.contents[beforeIndex] == '#' {
					compVar.IgnoreKeyWord = true
					preStr = "#"
				}
			}

			// 扫描输入的地方，前面是否包含等于意思是，是否为等于右边的补全
			beforeHasTag := isBeforeHasHashtag(comResult.contents, comResult.offset)
			project.GetCompleteCache().SetBeforeHashtag(beforeHasTag)
		}
	}
	compVar.SplitByte = splitByte
	compVar.ParamCandidateType = paramCandidateType

	project.CodeComplete(strFile, compVar)
	items := l.convertToCompItems(preStr)
	log.Debug("TextDocumentComplete str=%s, veclen=%d", preCompStr, len(items))
	return CompletionListTmp{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

func getDefaultHashtag(comResult commFileRequest) common.CompleteVarStruct {
	return common.CompleteVarStruct{
		PosLine:       (int)(comResult.pos.Line),
		PosCh:         (int)(comResult.pos.Character),
		StrVec:        []string{"#"},
		IsFuncVec:     []bool{false},
		ColonFlag:     false,
		LastEmptyFlag: false,
		IgnoreKeyWord: true,
	}
}

// 处理注解系统的代码补全，注解系统是以---@开头的
func (l *LspServer) handleGenerateAnnotateType(strFile string, contents []byte, offset int,
	posLine int) (comList CompletionListTmp, flag bool) {
	strLine := getPreLineStr(offset, contents)
	if strLine == "" {
		return
	}

	beginIndex := strings.LastIndex(strLine, "---@")
	if beginIndex == -1 {
		return
	}

	// 当前输入的单词，允许包含 _.
	beforeIndex := offset - 1
	for index := offset - 1; index >= 0; index-- {
		ch := contents[index]
		if ch == '\r' || ch == '\n' {
			break
		}

		if ch == '_' || ch == '.' || IsDigit(ch) || IsLetter(ch) {
			beforeIndex = index
			continue
		}
		break
	}

	strWord := string(contents[beforeIndex:offset])
	log.Debug("strWord=%s", strWord)

	flag = true
	annotateStr := strLine[beginIndex+2:]

	project := l.getAllProject()
	project.AnnotateTypeComplete(strFile, annotateStr, strWord, posLine)
	comList.IsIncomplete = false
	comList.Items = l.convertToCompItems("")
	return
}

func idxOfSquareBracketAndQuoteHelp(strLine string, quoteStr string) (squareBracketIdx int, lastQuoteIdx int) {
	if strings.Count(strLine, quoteStr)%2 == 1 {
		lastQuoteIdx = strings.LastIndex(strLine, quoteStr)
		for i := lastQuoteIdx - 1; i >= 0; i-- {
			if strLine[i] == ' ' {
				continue
			} else if strLine[i] == '[' {
				squareBracketIdx = i
				return
			} else {
				squareBracketIdx = -1
				return
			}
		}
	}

	squareBracketIdx = -1
	return
}

// 获取["的位置
func idxOfSquareBracketAndQuote(strLine string) (squareBracketIdx int, lastQuoteIdx int) {
	idx1, idx2 := idxOfSquareBracketAndQuoteHelp(strLine, "\"")
	if idx1 > 0 {
		return idx1, idx2
	}

	return idxOfSquareBracketAndQuoteHelp(strLine, "'")
}

// 判断是否为等于右边的补全
func isBeforeHasHashtag(contents []byte, offset int) (flag bool) {
	// 前面是否包含过空格
	spaceFlag := false
	for i := offset - 1; i >= 0; i = i - 1 {
		ch := contents[i]
		if ch == '\n' || ch == '\r' {
			break
		}

		if ch == ' ' {
			spaceFlag = true
			continue
		}

		if spaceFlag {
			if ch == '=' {
				flag = true
			}
			break
		}

		if ch == '_' || ch == '.' || IsDigit(ch) || IsLetter(ch) {
			continue
		}

		break

	}

	return flag
}

// 获取补全前置的输入字符串
// preStr 返回的为补全前面的字符串
// splitByte 为补全前面字符串的分割字符
func getCompeletePreStr(contents []byte, offset int) (preStr string, splitByte byte) {
	beforeIndex := GetBeforeIndex(contents, offset-1)
	if beforeIndex-1 >= 0 {
		splitByte = contents[beforeIndex-1]
	}

	rangeConents := contents[beforeIndex:offset]
	str := string(rangeConents)

	strLine := getPreLineStr(offset, contents)
	log.Debug("completion str=%s", str)

	// 1) 判断是为数字开头的
	var reIntegerBegin = regexp.MustCompile(`^[0-9]+.+$`)
	if reIntegerBegin.MatchString(str) {
		log.Debug("TextDocumentComplete str=%s is number", str)
		return
	}

	// 1.1) 判断是["
	tagIdx, _ := idxOfSquareBracketAndQuote(strLine)
	if tagIdx > 0 {
		lineOffset := getLineOffset(offset, contents)
		beforeIndex = GetBeforeIndex(contents, lineOffset+tagIdx-1)
		rangeConents = contents[beforeIndex:offset]
		str = string(rangeConents)
	}

	// 2) 判断是否为输入的为字符串里面的
	if strings.Count(strLine, "\"")%2 == 1 || strings.Count(strLine, "'")%2 == 1 {
		if tagIdx == 0 {
			return
		}
	}

	// 3) 判断是否在注释的行里
	// todo 这里是简单判断，需要优化，因为有可能为 "--" 这样的场景
	if strings.Contains(strLine, "--") {
		return
	}

	// 4) 判断是否为字符串链接，切割掉之前的字符串
	lastIndex := strings.LastIndex(str, "..")
	if lastIndex >= 0 {
		subStr := string(str[lastIndex+2:])
		if subStr == "" {
			log.Error("TextDocumentComplete str=%s concat.., subStr=%s", str, subStr)
			return
		}
		str = subStr
	}

	preStr = str
	return
}

// 判断是否为文件目录补全
func (l *LspServer) judgeCompeleteFile(strFile string, contents []byte, offset int) (comList CompletionListTmp, flag bool) {
	comList.IsIncomplete = false

	// 获取当前行的所有内容
	strLine := getPreLineStr(offset, contents)
	if strLine == "" {
		return
	}

	// 1) 定义为匹配的引入其他文件的字符串
	referNameStr := ""
	matchReferStr := ""
	referFileTypes := common.GConfig.GetAllReferFileTypes()
	for _, strOne := range referFileTypes {
		regexpStr := strOne + ` *?(\()? *?[\"|\'][0-9a-zA-Z_/|.]*`
		regRefer := regexp.MustCompile(regexpStr)
		findStrVec := regRefer.FindAllString(strLine, -1)
		if len(findStrVec) == 0 {
			continue
		}

		lastStr := findStrVec[len(findStrVec)-1]
		if !strings.HasSuffix(strLine, lastStr) {
			continue
		}
		referNameStr = strOne
		matchReferStr = lastStr
		break
	}

	if matchReferStr == "" {
		return
	}
	referType := common.StrToReferType(referNameStr)
	if referType == common.ReferNotValid {
		return
	}

	// 2) 定义为引入其他文件，前缀已经输入的文件字符串
	referIndex := strings.Index(matchReferStr, "\"")
	if referIndex == -1 {
		referIndex = strings.Index(matchReferStr, "'")
	}
	if referIndex == -1 {
		return
	}
	preFileStr := matchReferStr[referIndex+1:]
	project := l.getAllProject()
	project.CodeCompleteFile(strFile, referNameStr, referType, preFileStr)
	comList.Items = l.convertToCompItems("")
	flag = true
	return
}

// 字符串进行拆分
func getComplelteStruct(str string, line, character int) (completeVar common.CompleteVarStruct, flag bool) {
	lastEmptyFlag := false
	colonFlag := false
	lastCh := str[len(str)-1]
	strContent := str
	//先判断最后一个符号:
	if lastCh == ':' || lastCh == '.' {
		if len(str) == 1 {
			return
		}

		lastEmptyFlag = true
		if lastCh == ':' {
			colonFlag = true
		}

		strContent = strContent[:len(str)-1]
	} else {
		// 判断前面是否以冒号开头
		for i := len(strContent) - 1; i >= 0; i-- {
			ch := strContent[i]
			if IsDigit(ch) || IsLetter(ch) || ch == ' ' {
				continue
			}

			if ch == ':' {
				// 以冒号分割
				colonFlag = true
				lastEmptyFlag = true
				strContent = strContent[:i]
			}
			break
		}
	}

	// 判断光标的前面是否为 ["b 或是['a
	// 值为1表示为", 值为2表示为 '
	semiColon := 0
	semiColonFlag := false
	for i := len(strContent) - 1; i >= 0; i-- {
		ch := strContent[i]
		if ch == '[' {
			break
		}

		if ch == '\'' {
			if semiColon != 0 {
				semiColon = 0
				break
			}

			semiColon = 1
			continue
		}

		if ch == '"' {
			if semiColon != 0 {
				semiColon = 0
				break
			}
			semiColon = 2
			continue
		}

		if semiColon != 0 {
			if ch == ' ' {
				// 可以容忍空格
				continue
			}

			if ch == '[' {
				semiColonFlag = true
				break
			}
		}
	}

	if semiColonFlag {
		if semiColon == 1 {
			strContent = strContent + "']"
		} else {
			strContent = strContent + "\"]"
		}
	}

	varStruct := check.StrToDefineVarStruct(strContent)
	if !varStruct.ValidFlag {
		return
	}

	flag = true
	completeVar = common.CompleteVarStruct{
		PosLine:       line,
		PosCh:         character,
		StrVec:        varStruct.StrVec,
		IsFuncVec:     varStruct.IsFuncVec,
		ColonFlag:     colonFlag,
		LastEmptyFlag: lastEmptyFlag,
		Exp:           varStruct.Exp,
	}

	if len(varStruct.StrVec) == 1 && varStruct.StrVec[0] != "" {
		oneStr := varStruct.StrVec[0]
		oneChar := oneStr[0]
		completeVar.FilterCharacterFlag = false // 查找的结果，是否过滤指定的字符

		if !IsLetter(oneChar) {
			return
		}

		completeVar.FilterCharacterFlag = true
		completeVar.FilterOneChar = (rune)(oneChar)
		if unicode.IsUpper(completeVar.FilterOneChar) {
			completeVar.FilterTwoChar = unicode.ToLower(completeVar.FilterOneChar)
		} else if unicode.IsLower(completeVar.FilterOneChar) {
			completeVar.FilterTwoChar = unicode.ToUpper(completeVar.FilterOneChar)
		}
	}

	return
}

func judgeBeforeCommentHorizontal(contents []byte, offset int) bool {
	num := 0
	for index := offset - 1; index >= 0; index-- {
		ch := contents[index]
		if ch == '-' {
			num = num + 1
			continue
		}

		break
	}

	if num == 1 || num == 2 || num == 3 {
		return true
	}

	return false
}

// 处理快捷生成函数的注释 type CompletionList struct
func (l *LspServer) handleGenerateComment(strFile string, contents []byte, offset int,
	posLine int) (comList lsp.CompletionList, err error) {
	comList.IsIncomplete = false

	// 1) 向前找到，看能否找到3个---
	beforeIndex := offset - 1
	for index := offset - 1; index >= 0; index-- {
		ch := contents[index]
		if ch == '-' {
			beforeIndex = index
			continue
		}

		break
	}

	if offset-beforeIndex != 1 && offset-beforeIndex != 2 && offset-beforeIndex != 3 {
		log.Debug("handleGenerateComment len is not 3, strFile=%s", strFile)
		return
	}

	// 2) 向后找，判断后面是否还有-
	endIndex := offset
	for index := offset; index < len(contents); index++ {
		ch := contents[index]
		if ch == '-' {
			endIndex = index
			continue
		}
		break
	}
	if endIndex != offset {
		log.Debug("handleGenerateComment after has -, strFile=%s", strFile)
		return
	}

	project := l.getAllProject()
	// 获取这行的函数快捷注释
	completeVecs := project.FuncCommentComplete(strFile, posLine)

	// 获取快捷提示所有的注解前缀
	for _, oneComplete := range completeVecs {
		var oneLspComplete lsp.CompletionItem

		oneLspComplete.Label = "---" + oneComplete.Label
		oneLspComplete.Kind = lsp.TextCompletion
		oneLspComplete.Detail = oneComplete.Detail

		oneLspComplete.Documentation = lsp.MarkupContent{
			Kind:  lsp.Markdown,
			Value: oneComplete.Documentation,
		}

		//oneLspComplete.Documentation = oneComplete.Documentation

		oneLspComplete.InsertText = oneComplete.InsetText
		oneLspComplete.InsertTextFormat = lsp.SnippetTextFormat
		comList.Items = append(comList.Items, oneLspComplete)
	}

	return
}

// 处理快捷生成---@ 注解的提示
func (l *LspServer) handleGenerateAnnotateArea(strFile string, contents []byte, offset int,
	posLine int) (comList CompletionListTmp, err error) {
	comList.IsIncomplete = false

	// 1) 向前找到，看能否找到3个---
	offset = offset - 1
	beforeIndex := offset - 1
	for index := offset - 1; index >= 0; index-- {
		ch := contents[index]
		if ch == '-' {
			beforeIndex = index
			continue
		}

		break
	}

	if offset-beforeIndex != 3 {
		log.Debug("handleGenerateAnnotateArea len is not 3, strFile=%s", strFile)
		return
	}

	project := l.getAllProject()
	project.CompleteAnnotateArea()
	comList.Items = l.convertToCompItems("")
	return
}

func (l *LspServer) convertToCompItems(preStr string) (items []CompletionItemTmp) {
	project := l.getAllProject()
	cacheItem := project.GetCompleteCacheItems()

	items = make([]CompletionItemTmp, len(cacheItem))
	for i := 0; i < len(cacheItem); i++ {
		oneComplete := &(cacheItem[i])
		item := &items[i]
		if preStr == "" {
			item.Label = oneComplete.Label
		} else {
			item.Label = preStr + oneComplete.Label
		}

		item.Kind = lsp.CompletionItemKind(oneComplete.Kind)
		if oneComplete.Kind == common.IKFunction {
			item.Kind = lsp.FunctionCompletion
		} else if oneComplete.Kind == common.IKKeyword {
			item.Kind = lsp.KeywordCompletion
		} else if oneComplete.Kind == common.IKField {
			item.Kind = lsp.FieldCompletion
		} else if oneComplete.Kind == common.IKAnnotateClass {
			item.Kind = lsp.InterfaceCompletion
		} else if oneComplete.Kind == common.IKAnnotateAlias {
			item.Kind = lsp.InterfaceCompletion
		}

		item.Data = float64(i)
	}

	return items
}

// TextDocumentCompleteResolve test
// 当代码补全，客户端预览其中某一个结果时候，提示部分信息
func (l *LspServer) TextDocumentCompleteResolve(ctx context.Context, vs lsp.CompletionItem) (compItem lsp.CompletionItem,
	err error) {
	compItem = vs
	log.Debug("TextDocumentCompleteResolve sss...")
	floatValue, flag := vs.Data.(float64)
	if !flag {
		return
	}

	itemIndex := int(floatValue)
	log.Debug("TextDocumentCompleteResolve ok, index=%d", itemIndex)
	project := l.getAllProject()
	item, luaFileStr, flag := project.GetCompleteCacheIndexItem(itemIndex)
	if !flag {
		log.Debug("TextDocumentCompleteResolve err2, index=%d", itemIndex)
		return
	}

	strDoc := codingconv.ConvertStrToUtf8(item.Documentation)
	strDetail := item.Detail

	beforeHasTag := project.GetCompleteCache().GetBeforeHashtag()

	// json snippet 里面不能包含. 不然任意地方输入 . 时候会激活响应的snippet
	if snippetItem, ok := common.GConfig.CompSnippetMap[vs.Label]; ok {
		compItem.InsertText = snippetItem.InsertText
		compItem.InsertTextFormat = lsp.SnippetTextFormat
		strDetail = snippetItem.Detail
		// 当snippet为function时，若前面有#符号，进行特殊的替换
		if vs.Label == "function" && beforeHasTag {
			compItem.InsertText = "function(${1:})\n\t${0:}\nend"
			strDetail = "function()\n\nend"
		}
	}

	if project.GetCompleteCache().GetClearParamQuotes() {
		compItem.InsertTextFormat = lsp.SnippetTextFormat
		strTemp := strings.ReplaceAll(compItem.Label, "\"", "")
		strTemp = strings.ReplaceAll(strTemp, "'", "")
		compItem.InsertText = strTemp
	}

	strMarkdown := fmt.Sprintf("```%s\n%s\n```", "lua", codingconv.ConvertStrToUtf8(strDetail))
	strMarkdown = fmt.Sprintf("%s\n%s", strMarkdown, strDoc)
	if luaFileStr != "" {
		strMarkdown = fmt.Sprintf("%s\n\r%s", strMarkdown, luaFileStr)
	}

	compItem.Detail = ""
	compItem.Documentation = lsp.MarkupContent{
		Kind:  lsp.Markdown,
		Value: strMarkdown,
	}

	return
}
