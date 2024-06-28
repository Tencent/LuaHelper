package common

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/annotation/annotateparser"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"luahelper-lsp/langserver/log"
	"sort"
)

// 注解的中间格式

// OneClassInfo 定义的一个Class信息
type OneClassInfo struct {
	LastLine   int                                        // 这块class所在定义的最后行数
	ClassState *annotateast.AnnotateClassState            // Class的信息
	FieldMap   map[string]*annotateast.AnnotateFieldState // 所有关联的field成员，key值为filed的名称

	// 这个定义的class是否关联到了变量，关联的含义是，例如下面的例子
	//---@class A
	//local b
	//那么A的Class类型关联给了b变量，如果在该文件内对b变量赋值了子成员，例如下面的，那么A的Class的类型也有子成员a
	//b.a = 1
	//这里是关联的第一轮遍历的变量指针，默认为nil
	RelateVar *VarInfo

	LuaFile string // 这个结构所在的lua文件名
}

// FragmentClassInfo 单个块所对应的class信息, 一个注释块，允许有多个 FragmentClassInfo
type FragmentClassInfo struct {
	ClassList []*OneClassInfo // 多个class信息
}

// OneAliasInfo 单个alias信息
type OneAliasInfo struct {
	AliasState *annotateast.AnnotateAliasState // 这个块对应的多个alias信息
	RelateVar  *VarInfo                        //这里是关联的第一轮遍历的变量指针，默认为nil
	LuaFile    string                          // 这个结构所在的lua文件名
}

// FragmentAliasInfo 单个块所对应的alias信息, 一个注释块，允许有多个 FragmentAliasInfo
type FragmentAliasInfo struct {
	AliasList []*OneAliasInfo // 这个块对应的多个alias信息
}

// FragmentTypeInfo 单个块所对应的type信息, 一个注释块，允许有多个 AnnotateTypeState
type FragmentTypeInfo struct {
	LastLine    int                // 这块Type所在定义的最后行数
	TypeList    []annotateast.Type // 每个AnnotateTypeState的TypeList拼接在这里面
	CommentList []string           // 所有的类型的注释
	ConstList   []bool             // 是否标记常量
	EnumList    []bool             // 是否标记枚举
}

// FragementParamInfo 单个块所对应的所有参数信息, 一个注释块，允许有多个 AnnotateParamState
type FragementParamInfo struct {
	LastLine  int                               // 这块参数所在定义的最后行数
	ParamList []*annotateast.AnnotateParamState // 所有的参数信息
}

// FragementReturnInfo 单个块对应的所有返回信息， 一个注释块，允许有多个 AnnotateReturnState
type FragementReturnInfo struct {
	ReturnTypeList []annotateast.Type // 每个AnnotateReturnState的ReturnTypeList拼接在这里面
	CommentList    []string           // 每一个返回类型的comment
}

// FragementVarargInfo vararg信息
type FragementVarargInfo struct {
	VarargInfo *annotateast.AnnotateVarargState
}

// OneGenericInfo 单个泛型信息
type OneGenericInfo struct {
	Name         string
	NameLoc      lexer.Location
	ParentName   string
	ParamNameLoc lexer.Location
}

// FragementGenericInfo 泛型信息
type FragementGenericInfo struct {
	GenericInfoList []OneGenericInfo
}

// FragementOverloadInfo 函数重载信息
type FragementOverloadInfo struct {
	OverloadList []*annotateast.AnnotateOverloadState
}

// FragementInfo 单个注释块转成的结构
type FragementInfo struct {
	LastLine     int   // 最后一行
	LineVec      []int // 所有有效行的列表
	ClassInfo    *FragmentClassInfo
	AliasInfo    *FragmentAliasInfo
	TypeInfo     *FragmentTypeInfo
	ParamInfo    *FragementParamInfo
	ReturnInfo   *FragementReturnInfo
	VarargInfo   *FragementVarargInfo
	GenericInfo  *FragementGenericInfo
	OverloadInfo *FragementOverloadInfo
}

// GetFirstOneClassInfo 获取注释代码段第一个ClassInfo
func (fr *FragementInfo) GetFirstOneClassInfo() *OneClassInfo {
	if fr.ClassInfo == nil {
		return nil
	}

	if len(fr.ClassInfo.ClassList) == 0 {
		return nil
	}

	return fr.ClassInfo.ClassList[0]
}

// GetFirstOneTypeInfo 获取注释代码段的第一个TypeInfo
func (fr *FragementInfo) GetFirstOneTypeInfo() (oneType annotateast.Type, findFlag bool) {
	if fr.TypeInfo == nil {
		return
	}

	if len(fr.TypeInfo.TypeList) == 0 {
		return
	}

	// 找到了
	return fr.TypeInfo.TypeList[0], true
}

// GetParamTypeInfo 获取一个注释段知道参数名的注解类型
func (fr *FragementInfo) GetParamTypeInfo(paramStr string) (oneType annotateast.Type, strComment string) {
	if fr.ParamInfo == nil {
		return
	}

	for _, oneParamState := range fr.ParamInfo.ParamList {
		if oneParamState.Name == paramStr {
			oneType = oneParamState.ParamType
			strComment = oneParamState.Comment
			break
		}
	}

	return oneType, strComment
}

// GetColorLocVec 获取一个注释代码段所有的需要着色的位置集合
func (fr *FragementInfo) GetColorLocVec() (locVec []lexer.Location) {
	// 1) 获取ClassInfo带来的所有位置信息
	if fr.ClassInfo != nil {
		for _, oneClass := range fr.ClassInfo.ClassList {
			// 先放入名称位置
			locVec = append(locVec, oneClass.ClassState.NameLoc)

			// 然后放入继承的父对象位置
			locVec = append(locVec, oneClass.ClassState.ParentLocList...)

			for _, oneFiled := range oneClass.FieldMap {
				typeLocVec := annotateast.GetTypeColorLocVec(oneFiled.FiledType)
				locVec = append(locVec, typeLocVec...)
			}
		}
	}

	// 2) 获取AliasInfo带来的所有位置信息
	if fr.AliasInfo != nil {
		for _, oneAlias := range fr.AliasInfo.AliasList {
			// 先放入名称位置
			locVec = append(locVec, oneAlias.AliasState.NameLoc)

			// 然后获取所有Type可能产生的位置信息
			typeLocVec := annotateast.GetTypeColorLocVec(oneAlias.AliasState.AliasType)
			locVec = append(locVec, typeLocVec...)
		}
	}

	// 3) 获取需要的Type带来的直接位置信息
	if fr.TypeInfo != nil {
		for _, oneType := range fr.TypeInfo.TypeList {
			typeLocVec := annotateast.GetTypeColorLocVec(oneType)
			locVec = append(locVec, typeLocVec...)
		}
	}

	// 4) 获取所有的参数带来的位置信息
	if fr.ParamInfo != nil {
		for _, oneParam := range fr.ParamInfo.ParamList {
			typeLocVec := annotateast.GetTypeColorLocVec(oneParam.ParamType)
			locVec = append(locVec, typeLocVec...)
		}
	}

	// 5) 获取所有的return信息带来的位置信息
	if fr.ReturnInfo != nil {
		for _, oneReturn := range fr.ReturnInfo.ReturnTypeList {
			typeLocVec := annotateast.GetTypeColorLocVec(oneReturn)
			locVec = append(locVec, typeLocVec...)
		}
	}

	// 6) 获取Vararg带来的类型位置信息
	if fr.VarargInfo != nil {
		typeLocVec := annotateast.GetTypeColorLocVec(fr.VarargInfo.VarargInfo.VarargType)
		locVec = append(locVec, typeLocVec...)
	}

	// 7) 获取所有的泛型带来的位置信息
	if fr.GenericInfo != nil {
		for _, oneGeneric := range fr.GenericInfo.GenericInfoList {
			locVec = append(locVec, oneGeneric.NameLoc)

			if oneGeneric.ParentName != "" {
				locVec = append(locVec, oneGeneric.ParamNameLoc)
			}
		}
	}

	// 8) 获取所有的OverloadInfo带来的位置信息
	if fr.OverloadInfo != nil {
		for _, oneOver := range fr.OverloadInfo.OverloadList {
			typeLocVec := annotateast.GetTypeColorLocVec(oneOver.OverFunType)
			locVec = append(locVec, typeLocVec...)
		}
	}

	return locVec
}

type resultSortFragement struct {
	results []*FragementInfo
}

/*
 * sort.Interface methods
 */
func (s *resultSortFragement) Len() int { return len(s.results) }
func (s *resultSortFragement) Less(i, j int) bool {
	iscore, jscore := s.results[i].LastLine, s.results[j].LastLine
	return iscore < jscore
}
func (s *resultSortFragement) Swap(i, j int) {
	s.results[i], s.results[j] = s.results[j], s.results[i]
}

type resultSortEnumVec struct {
	enumVec []*annotateast.AnnotateEnumState
}

func (s *resultSortEnumVec) Len() int { return len(s.enumVec) }
func (s *resultSortEnumVec) Less(i, j int) bool {
	iscore, jscore := s.enumVec[i].EnumLoc.StartLine, s.enumVec[j].EnumLoc.StartLine
	return iscore < jscore
}
func (s *resultSortEnumVec) Swap(i, j int) {
	s.enumVec[i], s.enumVec[j] = s.enumVec[j], s.enumVec[i]
}

// CreateTypeInfo 所有产生一个新类型的结构
// 注释什么class 或 alias时候，会产生一个新的类型结构
type CreateTypeInfo struct {
	LastLine  int           // 这个结构定义的最后行数
	ClassInfo *OneClassInfo // 当对应的为class时候，指向的指针
	AliasInfo *OneAliasInfo // 当对应的为alias时候，指向的指针
}

// GetFileNameAndLoc 获取出现的lua文件名及其位置新
func (ci *CreateTypeInfo) GetFileNameAndLoc() (luaFile string, loc lexer.Location) {
	if ci.AliasInfo != nil {
		luaFile = ci.AliasInfo.LuaFile
		loc = ci.AliasInfo.AliasState.NameLoc
		return
	}

	if ci.ClassInfo != nil {
		luaFile = ci.ClassInfo.LuaFile
		loc = ci.ClassInfo.ClassState.NameLoc
	}

	return
}

// CreateTypeList 同文件中为了处理同名，存放为列表结构
type CreateTypeList struct {
	List []*CreateTypeInfo
}

// IsRepeateTypeInfo 判断传入的createTypeInfo是否已经存在了
func (cl *CreateTypeList) IsRepeateTypeInfo(createTypeInfo *CreateTypeInfo) bool {
	if len(cl.List) == 0 {
		return false
	}

	for _, oneCreateType := range cl.List {
		if oneCreateType == createTypeInfo {
			// 找到了相同的
			return true
		}
	}

	return false
}

// EnumFragment 枚举的注释块段，包括开始与结束
type EnumFragment struct {
	StartEnum *annotateast.AnnotateEnumState // 枚举段的开始
	EndEnum   *annotateast.AnnotateEnumState // 枚举段的结束
}

// AnnotateFile 单个文件生成的核心注解信息
type AnnotateFile struct {
	FragementMap    map[int]*FragementInfo    // 文件对应的所有注释块信息
	CreateTypeMap   map[string]CreateTypeList // 文件管理的所有定义新产生的类信息
	sortFragement   *resultSortFragement      // 用于根据行号排序的内部结构
	LuaFile         string                    // 这个文件对应的lua名称
	checkErrVec     []CheckError              // 注解检测到的错误信息
	EnumFragmentVec []EnumFragment            // 所有的枚举段落
	IsEnumType      bool                      // 是否有枚举类型的type定义信息
}

// CreateAnnotateFile 创建文件的所有注解信息
func CreateAnnotateFile(luaFile string) *AnnotateFile {
	return &AnnotateFile{
		FragementMap:  map[int]*FragementInfo{},
		CreateTypeMap: map[string]CreateTypeList{},
		sortFragement: &resultSortFragement{
			results: []*FragementInfo{},
		},
		LuaFile:    luaFile,
		IsEnumType: false,
	}
}

// PushTypeDefineError 插入一个注解类型的错误
func (af *AnnotateFile) PushTypeDefineError(errStr string, errLoc lexer.Location) {
	if GConfig.IsIgnoreErrorFile(af.LuaFile, CheckErrorAnnotate) {
		return
	}

	oneCheckErr := CheckError{
		ErrType: CheckErrorAnnotate,
		ErrStr:  errStr,
		Loc:     errLoc,
	}

	af.checkErrVec = append(af.checkErrVec, oneCheckErr)
}

// ClearCheckError 清除校验的错误，注解语法的错误保留
func (af *AnnotateFile) ClearCheckError() {
	j := 0
	for _, val := range af.checkErrVec {
		if val.AnnotateSyntaxFlag {
			af.checkErrVec[j] = val
			j++
		}
	}

	af.checkErrVec = af.checkErrVec[:j]
}

// RankCheckError 排序错误的告警信息, 语法错误排在最前面
func (af *AnnotateFile) RankCheckError() {
	sort.Sort(af)
}

// Len 注解错误排序相关的获取长度的函数
func (af *AnnotateFile) Len() int { return len(af.checkErrVec) }

// Less 注解错误排序比较的函数
func (af *AnnotateFile) Less(i, j int) bool {
	oneLoc := af.checkErrVec[i].Loc
	twoLoc := af.checkErrVec[j].Loc
	if oneLoc.StartLine == twoLoc.StartLine {
		return oneLoc.StartColumn < twoLoc.StartColumn
	}

	return oneLoc.StartLine < twoLoc.StartLine
}

// Swap 注解错误排序较好的函数
func (af *AnnotateFile) Swap(i, j int) {
	af.checkErrVec[i], af.checkErrVec[j] = af.checkErrVec[j], af.checkErrVec[i]
}

// PushTypeDuplicateError 插入一个注解重复类型的错误
func (af *AnnotateFile) PushTypeDuplicateError(errStr string, errLoc lexer.Location, relateVec []RelateCheckInfo) {
	if GConfig.IsIgnoreErrorFile(af.LuaFile, CheckErrorAnnotate) {
		return
	}

	oneCheckErr := CheckError{
		ErrType: CheckErrorAnnotate,
		ErrStr:  errStr,
		Loc:     errLoc,
	}

	oneCheckErr.RelateVec = relateVec

	af.checkErrVec = append(af.checkErrVec, oneCheckErr)
}

// analysisAnnotateFragement 分析单个块的注解结构
func (af *AnnotateFile) analysisAnnotateFragement(lastLine int, annotateFragment *annotateast.AnnotateFragment) {
	oneClassInfo := &OneClassInfo{
		ClassState: nil,
		FieldMap:   map[string]*annotateast.AnnotateFieldState{},
		RelateVar:  nil,
		LuaFile:    af.LuaFile,
		LastLine:   lastLine,
	}

	classInfo := FragmentClassInfo{
		ClassList: []*OneClassInfo{},
	}

	aliasInfo := FragmentAliasInfo{
		AliasList: []*OneAliasInfo{},
	}

	typeInfo := FragmentTypeInfo{
		TypeList: []annotateast.Type{},
	}

	paramInfo := FragementParamInfo{
		ParamList: []*annotateast.AnnotateParamState{},
	}

	returnInfo := FragementReturnInfo{
		ReturnTypeList: []annotateast.Type{},
		CommentList:    []string{},
	}

	genericInfo := FragementGenericInfo{
		GenericInfoList: []OneGenericInfo{},
	}

	overloadInfo := FragementOverloadInfo{
		OverloadList: []*annotateast.AnnotateOverloadState{},
	}

	varargInfo := FragementVarargInfo{
		VarargInfo: nil,
	}

	fragmentInfo := &FragementInfo{
		LastLine: lastLine,
	}

	// 每个语句进行分析
	for index, oneState := range annotateFragment.Stats {
		fragmentInfo.LineVec = append(fragmentInfo.LineVec, annotateFragment.Lines[index])

		switch state := oneState.(type) {
		case *annotateast.AnnotateAliasState:
			aliasInfo.AliasList = append(aliasInfo.AliasList, &OneAliasInfo{
				AliasState: state,
				LuaFile:    af.LuaFile,
			})

		case *annotateast.AnnotateClassState:
			if oneClassInfo.ClassState == nil {
				oneClassInfo.ClassState = state
			} else {
				classInfo.ClassList = append(classInfo.ClassList, oneClassInfo)

				// 一个注释块有两个class信息, 前面的class信息被插入进去
				oneClassInfo = &OneClassInfo{
					ClassState: state,
					FieldMap:   map[string]*annotateast.AnnotateFieldState{},
					RelateVar:  nil,
					LuaFile:    af.LuaFile,
					LastLine:   lastLine,
				}
			}
		case *annotateast.AnnotateFieldState:
			if oneClassInfo.ClassState != nil {
				oneClassInfo.FieldMap[state.Name] = state
			} else {
				log.Debug("before ClassState is nil, filed=%s", state.Name)
			}

		case *annotateast.AnnotateParamState:
			paramInfo.ParamList = append(paramInfo.ParamList, state)

		case *annotateast.AnnotateTypeState:
			for i, subType := range state.ListType {
				typeInfo.TypeList = append(typeInfo.TypeList, subType)
				typeInfo.CommentList = append(typeInfo.CommentList, state.Comment)
				typeInfo.ConstList = append(typeInfo.ConstList, state.ListConst[i])
				typeInfo.EnumList = append(typeInfo.EnumList, state.ListEnum[i])
			}
		case *annotateast.AnnotateReturnState:
			returnInfo.ReturnTypeList = append(returnInfo.ReturnTypeList, state.ReturnTypeList...)
			for i := 0; i < len(state.ReturnTypeList); i++ {
				returnInfo.CommentList = append(returnInfo.CommentList, state.Comment)
			}

		case *annotateast.AnnotateGenericState:
			for index, name := range state.NameList {
				oneGenericInfo := OneGenericInfo{
					Name:    name,
					NameLoc: state.NameLocList[index],
				}

				if len(state.ParentNameList) > index {
					oneGenericInfo.ParentName = state.ParentNameList[index]
					oneGenericInfo.ParamNameLoc = state.ParentLocList[index]
				}

				genericInfo.GenericInfoList = append(genericInfo.GenericInfoList, oneGenericInfo)
			}

		case *annotateast.AnnotateOverloadState:
			overloadInfo.OverloadList = append(overloadInfo.OverloadList, state)

		case *annotateast.AnnotateVarargState:
			varargInfo.VarargInfo = state
		}
	}

	// 1) class注释段
	if oneClassInfo.ClassState != nil {
		classInfo.ClassList = append(classInfo.ClassList, oneClassInfo)
	}
	if len(classInfo.ClassList) > 0 {
		fragmentInfo.ClassInfo = &classInfo
	}

	// 2) alias段
	if len(aliasInfo.AliasList) > 0 {
		fragmentInfo.AliasInfo = &aliasInfo
	}

	// 3) type段
	if len(typeInfo.TypeList) > 0 {
		fragmentInfo.TypeInfo = &typeInfo
	}

	// 4) 参数注释段
	if len(paramInfo.ParamList) > 0 {
		fragmentInfo.ParamInfo = &paramInfo
	}

	// 5) 函数返回段
	if len(returnInfo.ReturnTypeList) > 0 {
		fragmentInfo.ReturnInfo = &returnInfo
	}

	// 6) 泛型段
	if len(genericInfo.GenericInfoList) > 0 {
		fragmentInfo.GenericInfo = &genericInfo
	}

	// 7) overload 函数重载段
	if len(overloadInfo.OverloadList) > 0 {
		fragmentInfo.OverloadInfo = &overloadInfo
	}

	// 8) 可变参数段
	if varargInfo.VarargInfo != nil {
		fragmentInfo.VarargInfo = &varargInfo
	}

	af.FragementMap[lastLine] = fragmentInfo
	af.sortFragement.results = append(af.sortFragement.results, fragmentInfo)
}

// GetErrorVec 获取所有的注解错误信息
func (af *AnnotateFile) GetErrorVec() (checkErrVec []CheckError) {
	return af.checkErrVec
}

// 这个文件插入一个新的Type符号
func (af *AnnotateFile) insertNewType(name string, oneTypeInfo *CreateTypeInfo) {
	if typeList, ok := af.CreateTypeMap[name]; ok {
		typeList.List = append(typeList.List, oneTypeInfo)
		af.CreateTypeMap[name] = typeList
	} else {
		oneTypeList := CreateTypeList{
			List: []*CreateTypeInfo{
				oneTypeInfo,
			},
		}

		af.CreateTypeMap[name] = oneTypeList
	}
}

// 这个文件的所有注释块依据行号，提取这个文件的所有符号
func (af *AnnotateFile) generateNewType() {
	for _, fragment := range af.sortFragement.results {
		// 1) 先判断是否有class信息域
		if fragment.ClassInfo != nil {
			for _, oneClass := range fragment.ClassInfo.ClassList {
				oneTypeInfo := &CreateTypeInfo{
					LastLine:  fragment.LastLine,
					ClassInfo: oneClass,
				}

				af.insertNewType(oneClass.ClassState.Name, oneTypeInfo)
			}
		}

		// 2) 判断是否存在alias信息域
		if fragment.AliasInfo != nil {
			for _, oneAlias := range fragment.AliasInfo.AliasList {
				oneTypeInfo := &CreateTypeInfo{
					LastLine:  fragment.LastLine,
					AliasInfo: oneAlias,
				}

				af.insertNewType(oneAlias.AliasState.Name, oneTypeInfo)
			}
		}
	}
}

// 所有排序的变量
type resultSortVar struct {
	results []*VarInfo
}

/*
 * sort.Interface methods
 */
func (s *resultSortVar) Len() int { return len(s.results) }
func (s *resultSortVar) Less(i, j int) bool {
	if s.results[i].Loc.StartLine == s.results[j].Loc.StartLine {
		// 首先判断两者为父子关系
		oneVar := s.results[i]
		twoVar := s.results[j]

		if oneVar.HasDeepSubVarInfo(twoVar) {
			return false
		}

		if twoVar.HasDeepSubVarInfo(oneVar) {
			return true
		}

		return s.results[i].Loc.StartColumn < s.results[j].Loc.StartColumn
	}

	return s.results[i].Loc.StartLine < s.results[j].Loc.StartLine
}

func (s *resultSortVar) Swap(i, j int) {
	s.results[i], s.results[j] = s.results[j], s.results[i]
}

// 判断varInfo是否要与紧跟着前面的注释段里面的class或是alias关联起来
// 返回值表示这次是否有关联到
func relateVarBeforeFragementNewType(varInfo *VarInfo, fragmentInfo *FragementInfo) bool {
	if fragmentInfo.ClassInfo != nil {
		for _, oneClassInfo := range fragmentInfo.ClassInfo.ClassList {
			if oneClassInfo.RelateVar != nil {
				// 已经关联过了
				continue
			}

			// 这个class关联到变量，返回
			oneClassInfo.RelateVar = varInfo

			return true
		}
	}

	if fragmentInfo.AliasInfo != nil {
		for _, oneAliasInfo := range fragmentInfo.AliasInfo.AliasList {
			if oneAliasInfo.RelateVar != nil {
				// 已经关联过了
				continue
			}

			// 这个alias关联到变量，返回
			oneAliasInfo.RelateVar = varInfo
			return true
		}
	}

	return false
}

// 所有排序的变量
type resultSortCreate struct {
	// 所有变量的列表
	results []*CreateTypeInfo

	// 查找变量的行数
	varLine int
}

func calcCreateTypeScore(createType *CreateTypeInfo, varLine int) (score int) {
	// 排序规则
	// 1) 如果createType在变量行号前面，行号越靠后的，优先级越高
	// 2) 如果createType在变量行号后面，行号越小的，优先越高
	if createType.LastLine <= varLine {
		// 此时分数是正的
		score = createType.LastLine
	} else {
		// 此时分数是负的
		score = varLine - createType.LastLine
	}

	return score
}

/*
 * sort.Interface methods
 */
func (s *resultSortCreate) Len() int { return len(s.results) }
func (s *resultSortCreate) Less(i, j int) bool {
	scoreI := calcCreateTypeScore(s.results[i], s.varLine)
	scoreJ := calcCreateTypeScore(s.results[j], s.varLine)

	return scoreI < scoreJ
}

func (s *resultSortCreate) Swap(i, j int) {
	s.results[i], s.results[j] = s.results[j], s.results[i]
}

// 通过名称在这个文件查找对应的CreateType
// 可能存在多个名称的，查找有个排序的顺序，先找行号前面的，后找行号后面的
func (af *AnnotateFile) findBestMatchCreateType(name string, line int) (createType *CreateTypeInfo) {
	createTypeList, ok := af.CreateTypeMap[name]
	if !ok {
		return
	}

	if len(createTypeList.List) == 0 {
		return
	}

	resultCreate := resultSortCreate{
		results: []*CreateTypeInfo{},
		varLine: line,
	}

	resultCreate.results = append(resultCreate.results, createTypeList.List...)
	sort.Sort(&resultCreate)

	bestCreateType := resultCreate.results[0]
	if bestCreateType.AliasInfo != nil {
		// 查找这个alias是否已经关联了变量
		if bestCreateType.AliasInfo.RelateVar == nil {
			createType = bestCreateType
			return
		}

		return
	}

	if bestCreateType.ClassInfo != nil {
		// 查找这个class是否已经关联了变量
		if bestCreateType.ClassInfo.RelateVar == nil {
			createType = bestCreateType
			return
		}

		return
	}

	return createType
}

// 判断varInfo是否要与紧跟着前面的注释段里面的type域关联起来
// 返回值表示这次是否有关联到
func (af *AnnotateFile) relateVarBeforeFragementType(varInfo *VarInfo, fragmentInfo *FragementInfo) bool {
	if fragmentInfo.TypeInfo == nil {
		return false
	}

	// 需要判断顺序
	for index, astType := range fragmentInfo.TypeInfo.TypeList {
		varIndex := (int)(varInfo.VarIndex)
		if (index + 1) != varIndex {
			continue
		}

		// 尝试关联这个astType，获取这个类型对应的
		strName := annotateast.TraverseOneType(astType)
		if strName == "" {
			break
		}

		bestCreateType := af.findBestMatchCreateType(strName, varInfo.Loc.StartColumn)
		if bestCreateType == nil {
			break
		}

		if bestCreateType.AliasInfo != nil {
			bestCreateType.AliasInfo.RelateVar = varInfo
			return true
		}

		if bestCreateType.ClassInfo != nil {
			bestCreateType.ClassInfo.RelateVar = varInfo
			return true
		}
	}

	return false
}

// RelateTypeVarInfo 这个文件新产生的符号，关联到第一阶段中的变量信息
// globalMaps 为lua文件所产生的所有全局信息
// mainScope 为lua文件所对应的主scope信息，通过scope可以遍历到所有的局部变量
func (af *AnnotateFile) RelateTypeVarInfo(globalMaps map[string]*VarInfo, mainScope *ScopeInfo) {
	needRelateNum := 0
	for _, typeList := range af.CreateTypeMap {
		needRelateNum = needRelateNum + len(typeList.List)
	}

	// 没有需要被关联的符号
	if needRelateNum == 0 {
		return
	}

	// 所有的变量放入的resultSortVar结构中
	resultVar := resultSortVar{
		results: []*VarInfo{},
	}
	for _, varInfo := range globalMaps {
		resultVar.results = append(resultVar.results, varInfo)
		pushSubMapVars(varInfo, &(resultVar.results))
	}
	mainScope.GetAllVarInfos(&(resultVar.results))

	// 按坐标信息进行排序，坐标小的变量放在前面
	sort.Sort(&resultVar)

	// 判断是否要新产生的TypeVarInfo关联到这个对应的VarInfo
	for _, varInfo := range resultVar.results {
		// 1) 首先判断前面一行是否直接的定义新类型，有那么直接关联
		lastLine := varInfo.Loc.StartLine - 1
		fragmentInfo, ok := af.FragementMap[lastLine]
		if !ok {
			// 前面没有任何代码段，返回
			continue
		}

		if relateVarBeforeFragementNewType(varInfo, fragmentInfo) {
			// 与前面的注释段反向关联到了
			needRelateNum--
			continue
		}

		// 2) 判断前面的注释段，是否有显示的用@type 关联类型
		if af.relateVarBeforeFragementType(varInfo, fragmentInfo) {
			// 这个变量关联到了
			needRelateNum--
			continue
		}
	}
}

// AnalysisAllComment 这个文件的所有注释进行分析
func (af *AnnotateFile) AnalysisAllComment(commentMap map[int]*lexer.CommentInfo) {
	var enumVec []*annotateast.AnnotateEnumState

	// 1) 遍历所有块的注释，提取有用的注释信息
	for lastLine, commentInfo := range commentMap {
		// 只处理头部注释
		if !commentInfo.HeadFlag {
			continue
		}

		// 注释块解析成注释段落
		annotateFragment, parseErrVec := annotateparser.ParseCommentFragment(commentInfo)

		// 分析单个注释的段落
		af.analysisAnnotateFragement(lastLine, &annotateFragment)

		// 提取所有的枚举段落的开始与结束
		for _, oneState := range annotateFragment.Stats {
			if enumState, ok := oneState.(*annotateast.AnnotateEnumState); ok {
				enumVec = append(enumVec, enumState)
			}
		}

		// 判断是否忽略注解类型告警
		if GConfig.IsGlobalIgnoreErrType(CheckErrorAnnotate) {
			continue
		}

		if GConfig.IsIgnoreErrorFile(af.LuaFile, CheckErrorAnnotate) {
			continue
		}

		// 插入一个词法分析的错误
		for _, oneParseErr := range parseErrVec {
			oneCheckErr := CheckError{
				ErrType:            CheckErrorAnnotate,
				ErrStr:             oneParseErr.ErrStr,
				Loc:                oneParseErr.ErrLoc,
				AnnotateSyntaxFlag: true,
			}
			af.checkErrVec = append(af.checkErrVec, oneCheckErr)
		}
	}

	// 2) 所有块注释信息，依据行号，进行排序
	sort.Sort(af.sortFragement)

	// 3) 按行号遍历所有的注释块信息，生成这个文件内所有产生的新符号
	af.generateNewType()

	resultSortEnumVec := resultSortEnumVec{
		enumVec: enumVec,
	}
	sort.Sort(&resultSortEnumVec)

	// 4) 枚举段落的开始与结束拼接
	af.mergeEnumFragment(resultSortEnumVec.enumVec)

	// 5) 计算是否有枚举类型
	af.calcIsEnumType()
}

func (af *AnnotateFile) calcIsEnumType() {
	for _, fragement := range af.FragementMap {
		if fragement.TypeInfo == nil {
			continue
		}

		for _, enumFlag := range fragement.TypeInfo.EnumList {
			if enumFlag == true {
				af.IsEnumType = true
				return
			}
		}
	}
}

// mergeEnumFragment 枚举段落的开始与结束拼接
func (af *AnnotateFile) mergeEnumFragment(enumVec []*annotateast.AnnotateEnumState) {
	af.EnumFragmentVec = []EnumFragment{}

	if len(enumVec) == 0 {
		return
	}

	var enumStack AnnotateEnumStack
	for _, oneEnum := range enumVec {
		if oneEnum.EnumType == annotateast.EnumTypeStart {
			enumStack.Push(oneEnum)
		} else if oneEnum.EnumType == annotateast.EnumTypeEnd {

			// 枚举段落的结束
			startEnum := enumStack.Pop()
			if startEnum == nil {
				log.Error("enum end not find start, line:%d", oneEnum.EnumLoc.StartLine)
				continue
			}

			af.EnumFragmentVec = append(af.EnumFragmentVec, EnumFragment{
				StartEnum: startEnum,
				EndEnum:   oneEnum,
			})
		}
	}

	if enumStack.Len() > 0 {
		for _, oneEnum := range enumStack.Slice {
			log.Error("enum start not find end, line:%d", oneEnum.EnumLoc.StartLine)
		}
	}
}

// GetLineFragementInfo 获取这个注解文件指定行号的注解块信息
func (af *AnnotateFile) GetLineFragementInfo(lastLine int) (fragment *FragementInfo) {
	fragment = af.FragementMap[lastLine]
	return fragment
}

// GetBestFragementInfo 根据行号，匹配一个FragementInfo
func (af *AnnotateFile) GetBestFragementInfo(line int) *FragementInfo {
	for _, oneFragment := range af.FragementMap {
		for _, oneLine := range oneFragment.LineVec {
			if oneLine == line {
				return oneFragment
			}
		}
	}

	return nil
}

// GetBestCreateTypeInfo 注解文件内根据名称来查找对应的OneClassInfo
// strName 为对应的字符串名称
// lastLine 为变量出现所在的行号
func (af *AnnotateFile) GetBestCreateTypeInfo(strName string, lastLine int) (createType *CreateTypeInfo) {
	createTypeList, flag := af.CreateTypeMap[strName]
	if !flag || len(createTypeList.List) == 0 {
		// 这个文件没有找到对应的符号
		return nil
	}

	resultCreate := resultSortCreate{
		results: []*CreateTypeInfo{},
		varLine: lastLine,
	}

	// 有多个，根据行号进行排序
	resultCreate.results = append(resultCreate.results, createTypeList.List...)
	sort.Sort(&resultCreate)

	// 返回优先级最高的CreateTypeInfo
	return resultCreate.results[0]
}

// IsHasEnumType 判断是否含义枚举类型
func (af *AnnotateFile) IsHasEnumType() bool {
	return af.IsEnumType
}
