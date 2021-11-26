package check

import (
	"fmt"
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/common"
	"strings"
)

// GetLspHoverVarStr 提示信息hover
func (a *AllProject) GetLspHoverVarStr(strFile string, varStruct *common.DefineVarStruct) (lableStr, docStr, luaFileStr string) {
	// 1) 判断是否为系统的函数提示
	//strVecLen := len(varStruct.StrVec)

	var lastSymbol *common.Symbol
	var findList []*common.Symbol
	for {
		symbol, symList := a.FindVarDefine(strFile, varStruct)
		// 原生的变量没有找到, 直接返回
		if symbol == nil {
			return
		}

		if len(symList) == 0 {
			// 没有追踪到，切短下次，继续找
			subLen := len(varStruct.StrVec) - 1
			if subLen == 0 {
				// 子项不能再切分了，退出
				return
			}
			varStruct.StrVec = varStruct.StrVec[0:subLen]
		} else {
			lastSymbol = symList[len(symList)-1]
			findList = symList
			//lastSymbol = symList[0]
			break
		}
	}

	if lastSymbol == nil {
		return
	}

	strOneComment := ""
	strLastBefore := ""
	strLastType := ""
	strPreFirst := ""
	preFirstFlag := false
	dirManager := common.GConfig.GetDirManager()

	for _, oneSymbol := range findList {
		strType, strLable1, strDoc1, strPre, flag := a.getVarHoverInfo(oneSymbol, varStruct)
		if !preFirstFlag {
			strPreFirst = strPre
			preFirstFlag = true
		}

		if flag {
			lableStr = strLable1
			if strPreFirst == "_G." && strings.HasPrefix(strLable1, "function ") {
				//  修复特殊的 _G.function a() 这样的显示
				lableStr = "[_G] " + strLable1
			} else {
				if strPreFirst == "local " && strings.HasPrefix(strLable1, "function _G.") {
					lableStr = strPreFirst + "function " + strings.TrimPrefix(strLable1, "function _G.")
				} else {
					if strings.HasPrefix(strLable1, "_G.") {
						lableStr = strLable1
					} else {
						lableStr = strPreFirst + strLable1
					}
				}
			}

			docStr = strDoc1
			luaFileStr = dirManager.RemovePathDirPre(oneSymbol.FileName)
			return
		}

		if strLastType == "" || strLastType == "any" {
			strLastType = strType
			strLastBefore = strLable1
			luaFileStr = dirManager.RemovePathDirPre(oneSymbol.FileName)
		}

		if strOneComment == "" && strDoc1 != "" {
			strOneComment = strDoc1
		}
	}

	lableStr = strLastBefore
	if strPreFirst == "_G." && strings.HasPrefix(strLastBefore, "function ") {
		//  修复特殊的 _G.function a() 这样的显示
		lableStr = "[_G] " + strLastBefore
	} else {
		if strPreFirst == "local " && strings.HasPrefix(strLastBefore, "function _G.") {
			lableStr = strPreFirst + "function " + strings.TrimPrefix(strLastBefore, "function _G.")
		} else {
			if strings.HasPrefix(strLastBefore, "_G.") {
				lableStr = strLastBefore
			} else {
				lableStr = strPreFirst + strLastBefore
			}
		}
	}

	docStr = strOneComment
	return
}

func (a *AllProject) getVarHoverInfo(symbol *common.Symbol, varStruct *common.DefineVarStruct) (strType string,
	strLable, strDoc, strPre string, findFlag bool) {
	// 1) 首先提取注解类型
	if symbol.AnnotateType != nil {
		str := annotateast.TypeConvertStr(symbol.AnnotateType)
		strLable = varStruct.Str + " : " + str

		// 判断是否关联成number，如果是number类型尝试获取具体的值
		if symbol.VarInfo != nil {
			// strType := symbol.VarInfo.GetVarTypeDetail()
			// if strings.HasPrefix(strType, "number: ") {
			// 	strLable = varStruct.Str + " : number = " + strings.TrimPrefix(strType, "number: ")
			// }
			// strLable = varStruct.Str + " : " + strType

			if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
				strPre = "local "
			}

			if symbol.VarInfo.ExtraGlobal != nil && symbol.VarInfo.ExtraGlobal.GFlag {
				strPre = "_G."
			}
		}

		strDoc = symbol.AnnotateComment
		strDoc = strings.ReplaceAll(strDoc, "\n", "  \n")
		findFlag = true
		return
	}

	// 2) 判断变量类型是否存在
	if symbol.VarInfo == nil {
		return
	}

	strType = ""
	if symbol.VarInfo.ReferInfo != nil {
		strType = symbol.VarInfo.ReferInfo.GetReferComment()
	} else {
		strType = symbol.VarInfo.GetVarTypeDetail()
		referFunc := symbol.VarInfo.ReferFunc
		if referFunc != nil {
			if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
				strPre = "local "
			}

			if symbol.VarInfo.ExtraGlobal != nil && symbol.VarInfo.ExtraGlobal.GFlag {
				strType = "function _G." + referFunc.GetFuncCompleteStr(varStruct.StrVec[len(varStruct.StrVec)-1], true, false)
			} else {
				strType = "function " + referFunc.GetFuncCompleteStr(varStruct.StrVec[len(varStruct.StrVec)-1], true, false)
			}
		}
	}

	strLable = strings.Join(varStruct.StrVec, ".")
	if symbol.VarInfo.ReferFunc == nil {
		// if strings.HasPrefix(strType, "number: ") {
		// 	strType = "number = " + strings.TrimPrefix(strType, "number: ")
		// }

		strLable = strLable + " : " + strType
		if symbol.VarInfo.ExtraGlobal == nil && !symbol.VarInfo.IsMemFlag {
			strPre = "local "
		}

		if symbol.VarInfo.ExtraGlobal != nil && symbol.VarInfo.ExtraGlobal.GFlag {
			strPre = "_G."
		}
	} else {
		strLable = strType
	}

	strDoc = a.GetLineComment(symbol.FileName, symbol.VarInfo.Loc.EndLine)
	//strDoc = getFinalStrComment(strDoc, true)
	strDoc = GetStrComment(strDoc)
	return
}

// 对注释进行一些处理, 提取出注解的特殊的markdown格式
func GetStrComment(strComment string) (str string) {
	if strComment == "" {
		return strComment
	}

	// 之前lua格式的doc
	preLuaStr := ""

	// 删除注释一些空格和其他的多余的格式
	splitStrArr := strings.Split(strComment, "\n")
	annotatFlag := false
	for index, oneStr := range splitStrArr {
		oneStr = strings.TrimLeft(oneStr, " ")
		oneStr = strings.TrimPrefix(oneStr, "-*")
		//oneStr = strings.TrimPrefix(oneStr, "*")
		oneStr = strings.TrimPrefix(oneStr, "-")
		oneStr = strings.TrimLeft(oneStr, " ")

		annotatFlag = false
		if strings.HasPrefix(oneStr, "@class") || strings.HasPrefix(oneStr, "@alias") || strings.HasPrefix(oneStr, "@param") ||
			strings.HasPrefix(oneStr, "@return") || strings.HasPrefix(oneStr, "@type") || strings.HasPrefix(oneStr, "@overload") ||
			strings.HasPrefix(oneStr, "@generic") || strings.HasPrefix(oneStr, "@vararg") || strings.HasPrefix(oneStr, "@version") {
			annotatFlag = true
		}

		if annotatFlag {
			if preLuaStr == "" {
				// 如果之前没有注解，拼接lua
				preLuaStr = fmt.Sprintf("\n```%s\n%s\n", "lua", oneStr)
			} else {
				// 如果之前有注解，拼接这次
				preLuaStr = preLuaStr + oneStr + "\n"
			}

			if index == len(splitStrArr)-1 {
				if preLuaStr != "" {
					str = str + preLuaStr + "```"
				}
			}
		} else {
			if preLuaStr == "" {
				str = str + "  \n" + oneStr
			} else {
				str = str + preLuaStr + "```" + "  \n" + oneStr
				preLuaStr = ""
			}
		}
	}

	// str = str + "\n\r" + oneStr

	return str
}
