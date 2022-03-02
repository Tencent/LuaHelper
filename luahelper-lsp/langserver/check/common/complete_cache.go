package common

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
)

/*
* 缓冲代码补全的内容
* 默认情况下，代码补全的时候，后台会给客户端推送所有的 lsp.CompletionItem[] 列表
* 这种方式，当补全的列表特别多的时候，后台需要计算出每一个补全的详细信息，比较费时；因为需要计算每一个的Detail、Documentation

* 后面优化为，代码详细补全分为两阶段
* 第一阶段：后台会推送lsp.CompletionItem[] 列表，这个是一个简要的列表
* 第二阶段：当客户端选择每一个候选词时，客户端再会发送completionItem/resolve，此时后台再给详细的补全信息

 */

// CacheKind 缓存的类型
type CacheKind int

const (
	// CKindNormal 为普通的，不需要获取其他信息，直接透传信息
	CKindNormal CacheKind = 1

	// CKindVar 缓冲的变量信息，需要通过具体的变量来获取所有的补全信息
	CKindVar CacheKind = 2

	// CkindClassField 注解系统class 的field域
	CkindClassField CacheKind = 3

	// CkindAnnotateType 注解系统类型系统
	CkindAnnotateType CacheKind = 4
)

// OneCompleteData 为后台缓存的一个代码补全信息
type OneCompleteData struct {
	Label          string
	Kind           ItemKind
	InsetText      string
	Detail         string // 详细信息
	Documentation  string
	LuaFile        string                          // 补全出现的文件名
	VarInfo        *VarInfo                        // 补全的变量名称
	ExpandVarInfo  *VarInfo                        // 扩展的变量名称
	CacheKind      CacheKind                       // 缓存的类型
	FieldState     *annotateast.AnnotateFieldState // 提示为注解的class信息
	FieldColonFlag annotateast.FieldColonType      // 当为FieldState时候，是否为：函数
	CreateTypeInfo *CreateTypeInfo
}

// CompleteCache 缓存所有的补全信息
type CompleteCache struct {
	index            int                 // 当前的index
	maxNum           int                 // 上次提示的最大数量
	excludeNum       int                 // 上次提示时，排除的最大数量
	dataList         []OneCompleteData   // 所有的缓存信息，用列表存储就ok
	existMap         map[string]int      // 为已经存在的map，防止重复
	excludeMap       map[string]struct{} // 冒号语法需要排除的map
	colonFlag        bool                // 代码补全最后是否为冒号语法
	beforeHashtag    bool                //  补全的词前面是否包含#
	clearParamQuotes bool                // 补全时候，是否要清除候选词的引号
	completeVar      *CompleteVarStruct  // 缓存的输入代码补全的结构
}

// CreateCompleteCache 创建一个代码补全缓存
func CreateCompleteCache() *CompleteCache {
	cache := &CompleteCache{
		index:            0,
		maxNum:           1,
		excludeNum:       1,
		dataList:         []OneCompleteData{},
		existMap:         map[string]int{},
		excludeMap:       map[string]struct{}{},
		colonFlag:        false,
		clearParamQuotes: false,
		completeVar:      nil,
	}

	return cache
}

// ResertData resert cache data
func (cache *CompleteCache) ResertData() {
	cache.index = 0
	cache.dataList = make([]OneCompleteData, 0, cache.maxNum)
	cache.existMap = make(map[string]int, cache.maxNum)
	cache.excludeMap = make(map[string]struct{}, cache.excludeNum)
	cache.colonFlag = false
	cache.beforeHashtag = false
	cache.clearParamQuotes = false
	cache.completeVar = nil
}

// SetCompleteVar set completeVar
func (cache *CompleteCache) SetCompleteVar(completeVar *CompleteVarStruct) {
	cache.completeVar = completeVar
}

// GetCompleteVar get completeVar
func (cache *CompleteCache) GetCompleteVar() *CompleteVarStruct {
	return cache.completeVar
}

// SetClearParamQuotes set clearParamQuotes
func (cache *CompleteCache) SetClearParamQuotes(flag bool) {
	cache.clearParamQuotes = flag
}

// GetClearParamQuotes set clearParamQuotes
func (cache *CompleteCache) GetClearParamQuotes() bool {
	return cache.clearParamQuotes
}

// SetBeforeHashtag set before hash tag
func (cache *CompleteCache) SetBeforeHashtag(flag bool) {
	cache.beforeHashtag = flag
}

// GetBeforeHashtag set before hash tag
func (cache *CompleteCache) GetBeforeHashtag() bool {
	return cache.beforeHashtag
}

// GetDataList 获取所有的数据
func (cache *CompleteCache) GetDataList() []OneCompleteData {
	if len(cache.dataList) > cache.maxNum {
		cache.maxNum = len(cache.dataList)
	}

	if len(cache.excludeMap) > cache.excludeNum {
		cache.excludeNum = len(cache.excludeMap)
	}

	return cache.dataList
}

// GetIndexData 获取所有的数据
func (cache *CompleteCache) GetIndexData(index int) (item OneCompleteData, flag bool) {
	flag = false
	itemLen := len(cache.dataList)
	if index < 0 || index >= itemLen {
		return
	}

	item = cache.dataList[index]
	flag = true
	return
}

// InsertCompleteData insert one data
func (cache *CompleteCache) InsertCompleteData() {

}

// GetIndex get index number
func (cache *CompleteCache) GetIndex() int {
	return cache.index
}

// SetColonFlag 设置最后的补全是否为冒号语法：
func (cache *CompleteCache) SetColonFlag(flag bool) {
	cache.colonFlag = flag
}

// GetColonFlag 获取最后的补全是否为冒号语法：
func (cache *CompleteCache) GetColonFlag() bool {
	return cache.colonFlag
}

// InsertExcludeStr 屏蔽冒号语法的字符串
func (cache *CompleteCache) InsertExcludeStr(strName string) {
	cache.excludeMap[strName] = struct{}{}
}

// ExistStr 判断是否已经存在了补全的信息
func (cache *CompleteCache) ExistStr(strLabel string) bool {
	_, flag := cache.existMap[strLabel]
	return flag
}

// IsExcludeStr 是否为冒号语法排除的
func (cache *CompleteCache) IsExcludeStr(strLabel string) bool {
	_, flag := cache.excludeMap[strLabel]
	return flag
}

// InsertCompleteVar 插入补全缓存类型的CKindVar
func (cache *CompleteCache) InsertCompleteVar(luaFile string, label string, varInfo *VarInfo) {
	oneComplete := OneCompleteData{
		Label:     label,
		LuaFile:   luaFile,
		VarInfo:   varInfo,
		CacheKind: CKindVar,
	}

	if varInfo.ReferFunc != nil {
		oneComplete.Kind = IKFunction
	}

	// 如果为引用的模块
	if varInfo.ReferInfo != nil {
		oneComplete.Kind = IKField
	}

	cache.dataList = append(cache.dataList, oneComplete)
	cache.existMap[label] = len(cache.dataList) - 1
}

// InsertCompleteVar 插入补全缓存类型的CKindVar
// 当为VarInfo时候，是否补充第一个参数为self
func (cache *CompleteCache) InsertCompleteVarInclude(luaFile string, label string, varInfo *VarInfo) {
	oneComplete := OneCompleteData{
		Label:     label,
		LuaFile:   luaFile,
		VarInfo:   varInfo,
		CacheKind: CKindVar,
		//IncludeSelfParam: true,
	}

	if varInfo.ReferFunc != nil {
		oneComplete.Kind = IKFunction
	}

	// 如果为引用的模块
	if varInfo.ReferInfo != nil {
		oneComplete.Kind = IKField
	}

	cache.dataList = append(cache.dataList, oneComplete)
	cache.existMap[label] = len(cache.dataList) - 1
}

// InsertCompleteKey 插入关键字的代码补全
func (cache *CompleteCache) InsertCompleteKey(label string) {
	oneComplete := OneCompleteData{
		Label:     label,
		Kind:      IKKeyword,
		CacheKind: CKindNormal,
	}

	cache.dataList = append(cache.dataList, oneComplete)
	cache.existMap[label] = len(cache.dataList) - 1
}

// InsertCompleteSnippet 插入Snippet
func (cache *CompleteCache) InsertCompleteSnippet(label string) {
	oneComplete := OneCompleteData{
		Label:     label,
		Kind:      IKSnippet,
		CacheKind: CKindNormal,
	}

	cache.dataList = append(cache.dataList, oneComplete)
	cache.existMap[label] = len(cache.dataList) - 1
}

// InsertCompleteSysModuleMem 插入系统模块的成员
func (cache *CompleteCache) InsertCompleteSysModuleMem(label, detail, documentation string, kind ItemKind) {
	cache.InsertCompleteNormal(label, detail, documentation, kind)
}

// InsertCompleteSystemFunc 插入关键字的代码补全
func (cache *CompleteCache) InsertCompleteSystemFunc(label, detail, documentation string) {
	cache.InsertCompleteNormal(label, detail, documentation, IKFunction)
}

// InsertCompleteSystemModule 插入系统模块的补全
func (cache *CompleteCache) InsertCompleteSystemModule(label, detail, documentation string) {
	cache.InsertCompleteNormal(label, detail, documentation, IKField)
}

// InsertCompleteNormal 插入普通的
func (cache *CompleteCache) InsertCompleteNormal(label, detail, documentation string, kind ItemKind) {
	oneComplete := OneCompleteData{
		Label:         label,
		Detail:        detail,
		Documentation: documentation,
		Kind:          kind,
		CacheKind:     CKindNormal,
	}
	cache.dataList = append(cache.dataList, oneComplete)
	cache.existMap[label] = len(cache.dataList) - 1
}

// InsertCompleteNormal 插入普通的
func (cache *CompleteCache) InsertCompleteExpand(label, detail, documentation string, kind ItemKind, expandVar *VarInfo) {
	oneComplete := OneCompleteData{
		Label:         label,
		Detail:        detail,
		Documentation: documentation,
		Kind:          kind,
		CacheKind:     CKindNormal,
		ExpandVarInfo: expandVar,
	}
	cache.dataList = append(cache.dataList, oneComplete)
	cache.existMap[label] = len(cache.dataList) - 1
}

// InsertCompleteInnotateType 插入关键字的代码补全
func (cache *CompleteCache) InsertCompleteInnotateType(label string, creatType *CreateTypeInfo) {
	oneComplete := OneCompleteData{
		Label:          label,
		Kind:           IKAnnotateAlias,
		CacheKind:      CkindAnnotateType,
		CreateTypeInfo: creatType,
	}
	cache.dataList = append(cache.dataList, oneComplete)
	cache.existMap[label] = len(cache.dataList) - 1
}

// InsertCompleteClassField 插入注解系统的class field域
func (cache *CompleteCache) InsertCompleteClassField(luaFile, label string, field *annotateast.AnnotateFieldState,
	colonType annotateast.FieldColonType) {
	oneComplete := OneCompleteData{
		Label:          label,
		LuaFile:        luaFile,
		Kind:           IKVariable, // 暂时定义为变量，可能为函数
		FieldState:     field,
		CacheKind:      CkindClassField,
		FieldColonFlag: colonType,
	}

	if annotateast.GetAllFuncType(field.FiledType) != nil {
		oneComplete.Kind = IKFunction
	}

	if index, flag := cache.existMap[label]; flag {
		// 注解类型代码补全优先级高，如果出现了替换之前的内容
		cache.dataList[index] = oneComplete
	} else {
		cache.dataList = append(cache.dataList, oneComplete)
		cache.existMap[label] = len(cache.dataList) - 1
	}
}
