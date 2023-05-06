package check

import (
	"fmt"
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"luahelper-lsp/langserver/log"
)

func (a *AllProject) isFieldOfClass(fieldName string, className string) bool {
	createTypeList, flag := a.createTypeMap[className]
	if !flag || len(createTypeList.List) == 0 ||
		createTypeList.List[0].ClassInfo == nil ||
		createTypeList.List[0].ClassInfo.ClassState == nil {
		return false
	}

	//只取第一个，如果有多个，后续会报警
	_, ok := createTypeList.List[0].ClassInfo.FieldMap[fieldName]
	if ok {
		return true
	}

	// 所有父类型也处理下，再次递归获取
	for _, strParent := range createTypeList.List[0].ClassInfo.ClassState.ParentNameList {
		if a.isFieldOfClass(fieldName, strParent) {
			return true
		}
	}

	return false
}

// 获取注解class
func (a *AllProject) IsMemberOfAnnotateClassByVar(strMemName string, strVarName string, varInfo *common.VarInfo) (isMember bool, className string) {
	isMember = false
	className = ""

	if varInfo == nil {
		return
	}

	//1 找到变量定义处的前一行的注解
	annotateFile := a.getAnnotateFile(varInfo.FileName)
	if annotateFile == nil {
		log.Error("GetAnnotateClassByVar annotateFile is nil, file=%s", varInfo.FileName)
		return
	}

	fragmentInfo := annotateFile.GetLineFragementInfo(varInfo.Loc.StartLine - 1)

	classNameVec := []string{}
	//2 取出class name
	if varInfo.IsParam || varInfo.IsForParam {
		// 这里判断是否为函数参数的注解
		// 函数参数的注解，会额外的用下面的注解类型
		// ---@param one class
		if fragmentInfo != nil &&
			fragmentInfo.ParamInfo != nil {

			for i := 0; i < len(fragmentInfo.ParamInfo.ParamList); i++ {
				paramLine := fragmentInfo.ParamInfo.ParamList[i]

				//找到对应的那行---@param one class 再获取class
				if paramLine.Name == strVarName {
					classNameVec = annotateast.GetAllNormalStrList(paramLine.ParamType)
					break
				}
			}
		}
	} else {
		// 非函数参数 一般是 ---@type
		if fragmentInfo != nil &&
			fragmentInfo.TypeInfo != nil {

			for i := 0; i < len(fragmentInfo.TypeInfo.TypeList); i++ {
				typeOne := fragmentInfo.TypeInfo.TypeList[i]
				retVec := annotateast.GetAllNormalStrList(typeOne)
				classNameVec = append(classNameVec, retVec...)
			}
		}
	}

	if len(classNameVec) == 0 {
		return
	}

	for _, classNameOne := range classNameVec {
		if len(className) > 0 {
			className = fmt.Sprintf("%s|", className)
		}
		className = fmt.Sprintf("%s%s", className, classNameOne)
	}

	//3 根据className 查找注解的class信息 只取第一个，如果有多个，后续会报警
	for _, classNameOne := range classNameVec {
		if a.isFieldOfClass(strMemName, classNameOne) {
			return true, className
		}
	}

	return false, className
}

// 获取注解class
func (a *AllProject) IsMemberOfAnnotateClassByLoc(strFile string, strFieldNamelist []string, lineForGetAnnotate int) (isMemberMap map[string]bool, className string) {
	isMemberMap = map[string]bool{}
	className = ""

	// 1) 获取文件对应的annotateFile
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		log.Error("IsMemberOfAnnotateClassByLoc annotateFile is nil, file=%s", strFile)
		return
	}

	classNameVec := []string{}
	fragmentInfo := annotateFile.GetLineFragementInfo(lineForGetAnnotate)
	if fragmentInfo != nil &&
		fragmentInfo.TypeInfo != nil {
		for i := 0; i < len(fragmentInfo.TypeInfo.TypeList); i++ {
			typeOne := fragmentInfo.TypeInfo.TypeList[i]
			retVec := annotateast.GetAllNormalStrList(typeOne)
			classNameVec = append(classNameVec, retVec...)
		}
	}

	if len(classNameVec) <= 0 {
		return
	}

	for _, classNameOne := range classNameVec {
		if len(className) > 0 {
			className = fmt.Sprintf("%s|", className)
		}
		className = fmt.Sprintf("%s%s", className, classNameOne)
	}

	//3 根据className 查找注解的class信息 只取第一个，如果有多个，后续会报警
	for _, classNameOne := range classNameVec {
		for _, strMemName := range strFieldNamelist {
			isMemberMap[strMemName] = a.isFieldOfClass(strMemName, classNameOne)
		}
	}

	return isMemberMap, className
}

// 获取注解class
func (a *AllProject) GetFieldAnnotateType(strFile string, lineForGetAnnotate int, retFieldTypeMap map[string][]string) {
	//retFieldTypeMap = map[string]annotateast.Type{}

	// 1) 获取文件对应的annotateFile
	annotateFile := a.getAnnotateFile(strFile)
	if annotateFile == nil {
		log.Error("GetFieldAnnotateType annotateFile is nil, file=%s", strFile)
		return
	}

	classNameVec := []string{}
	fragmentInfo := annotateFile.GetLineFragementInfo(lineForGetAnnotate)
	if fragmentInfo != nil &&
		fragmentInfo.TypeInfo != nil {
		for i := 0; i < len(fragmentInfo.TypeInfo.TypeList); i++ {
			typeOne := fragmentInfo.TypeInfo.TypeList[i]
			retVec := annotateast.GetAllNormalStrList(typeOne)
			classNameVec = append(classNameVec, retVec...)
		}
	}

	if len(classNameVec) <= 0 {
		return
	}

	// for _, classNameOne := range classNameVec {
	// 	if len(className) > 0 {
	// 		className = fmt.Sprintf("%s|", className)
	// 	}
	// 	className = fmt.Sprintf("%s%s", className, classNameOne)
	// }

	//3 根据className 查找注解的class信息 只取第一个，如果有多个，后续会报警
	for _, classNameOne := range classNameVec {
		a.GetFieldTypeOfClass(classNameOne, retFieldTypeMap)
	}

	return
}

func (a *AllProject) GetFieldTypeOfClass(className string, retFieldType map[string][]string) {

	createTypeList, flag := a.createTypeMap[className]
	if !flag || len(createTypeList.List) == 0 ||
		createTypeList.List[0].ClassInfo == nil ||
		createTypeList.List[0].ClassInfo.ClassState == nil {
		return
	}

	//只取第一个，如果有多个，后续会报警
	for fieldName, fieldState := range createTypeList.List[0].ClassInfo.FieldMap {
		//retFieldType[fieldName] = append(retFieldType[fieldName], fieldState.FiledType)

		retVec := annotateast.GetAllNormalStrList(fieldState.FiledType)
		retFieldType[fieldName] = append(retFieldType[fieldName], retVec...)
	}

	// 所有父类型也处理下，再次递归获取
	for _, strParent := range createTypeList.List[0].ClassInfo.ClassState.ParentNameList {
		a.GetFieldTypeOfClass(strParent, retFieldType)
	}

	return
}
