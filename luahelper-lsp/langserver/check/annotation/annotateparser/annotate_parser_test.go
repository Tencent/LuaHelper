package annotateparser

import (
	"luahelper-lsp/langserver/check/annotation/annotateast"
	"luahelper-lsp/langserver/check/compiler/lexer"
	"testing"
)

func TestCorrectParse(t *testing.T) {
	commentInfo1 := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@type string",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@type string, number",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-@type string, number, number",
				Line: 3,
				Col:  0,
			},
			{
				Str:  "-@type string[]",
				Line: 4,
				Col:  0,
			},
			{
				Str:  "-@type string[], number",
				Line: 5,
				Col:  0,
			},
			{
				Str:  "-@type table < number , number>",
				Line: 6,
				Col:  0,
			},
			{
				Str:  "-@type    fun ( one:string, two: number):one, two",
				Line: 7,
				Col:  0,
			},
			{
				Str:  "-@type string | number[]",
				Line: 8,
				Col:  0,
			},
			{
				Str:  "-@type (string | number)[]",
				Line: 9,
				Col:  0,
			},
			{
				Str:  "-@alias Handler fun(type: string | number, data: any):void",
				Line: 10,
				Col:  0,
			},
			{
				Str:  "-@class A:B, C @fjsofjsofjo",
				Line: 11,
				Col:  0,
			},
			{
				Str:  "-@overload fun(list:table, sep:string, i:number):string|number, bb[]",
				Line: 12,
				Col:  0,
			},
			{
				Str:  "-@field public ssss fun(list:table, sep:string, i:number):string|number, bb[] sfsdfsdf",
				Line: 13,
				Col:  0,
			},
			{
				Str:  "-@param one  string | fun(list:table, sep:string, i:number):string|number, bb[] sfsdfsdf",
				Line: 14,
				Col:  0,
			},
			{
				Str:  "-@return (string | fun(list:table, sep:string, i:number):string|number, bb[]), string sfsdfsdf",
				Line: 15,
				Col:  0,
			},
			{
				Str:  "-@generic T : Transport, K sfsdfsdf asfdjosf",
				Line: 16,
				Col:  0,
			},
			{
				Str:  "-@vararg string",
				Line: 17,
				Col:  0,
			},
		},
	}
	fragent1, errVec := ParseCommentFragment(commentInfo1)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate type fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent1.Stats) != len(commentInfo1.LineVec) {
		t.Fatalf("parser annotate type stats is not equal")
	}
}

func TestAnnotateParserType(t *testing.T) {
	commentInfo1 := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@type string",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@type string @fjosfjo",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-@type ...",
				Line: 3,
				Col:  0,
			},
		},
	}
	fragent1, errVec := ParseCommentFragment(commentInfo1)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate type fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent1.Stats) == 0 {
		t.Fatalf("parser annotate type stats len is 0")
	}
}

func TestAnnotateParserFun(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@type fun(f: integer|function, what?: string):debuginfo",
				Line: 1,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate fun fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) == 0 {
		t.Fatalf("parser annotate type stats len is 0")
	}
}

func TestAnnotateParserRun(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@return ...",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@return integer? code",
				Line: 2,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) == 0 {
		t.Fatalf("parser annotate type stats len is 0")
	}
}

func TestAnnotateParserConst(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@alias dddd \"'bbb'\"",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@alias dddd \"bbb\"",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-@alias dddd '\"bbb\"'",
				Line: 3,
				Col:  0,
			},
			{
				Str:  "-@alias dddd 'bbb'",
				Line: 4,
				Col:  0,
			},
			{
				Str:  "-@alias exitcode2 '\"exit\"' | '\"signal\"'",
				Line: 5,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) != len(commentInfo.LineVec) {
		t.Fatalf("parser annotate type stats is not equal")
	}
}

func TestAnnotateParserAlias1(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@alias dddd",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-| '\"r\"'   # Read data from this program by `file`.",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-| '\"w\"'   # Write data to this program by `file`.",
				Line: 3,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) != 1 {
		t.Fatalf("parser annotate type stats is not equal")
	}

	aliasState, ok := fragent.Stats[0].(*annotateast.AnnotateAliasState)
	if !ok {
		t.Fatalf("parser annotate type stats is alias")
	}

	multiType, ok1 := aliasState.AliasType.(*annotateast.MultiType)
	if !ok1 {
		t.Fatalf("is not multiType")
	}

	if len(multiType.TypeList) != 2 {
		t.Fatalf("TypeList len is err")
	}
}

func TestAnnotateParserAlias2(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@alias dddd '\"a\"'",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-| '\"r\"'   # Read data from this program by `file`.",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-| '\"w\"'   # Write data to this program by `file`.",
				Line: 3,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) != 1 {
		t.Fatalf("parser annotate type stats is not equal")
	}

	aliasState, ok := fragent.Stats[0].(*annotateast.AnnotateAliasState)
	if !ok {
		t.Fatalf("parser annotate type stats is alias")
	}

	multiType, ok1 := aliasState.AliasType.(*annotateast.MultiType)
	if !ok1 {
		t.Fatalf("is not multiType")
	}

	if len(multiType.TypeList) != 3 {
		t.Fatalf("TypeList len is err")
	}
}

func TestAnnotateParserAlias3(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@alias dddd '\"a\"' | '\"p\"' @sfjosf",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-| '\"r\"'   # Read data from this program by `file`.",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-| '\"w\"'   # Write data to this program by `file`.",
				Line: 3,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) != 1 {
		t.Fatalf("parser annotate type stats is not equal")
	}

	aliasState, ok := fragent.Stats[0].(*annotateast.AnnotateAliasState)
	if !ok {
		t.Fatalf("parser annotate type stats is alias")
	}

	multiType, ok1 := aliasState.AliasType.(*annotateast.MultiType)
	if !ok1 {
		t.Fatalf("is not multiType")
	}

	if len(multiType.TypeList) != 4 {
		t.Fatalf("TypeList len is err")
	}
}

func TestAnnotateParserAlias4(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@alias dddd",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@type string",
				Line: 2,
				Col:  0,
			},
			{
				Str:  "-@alias dddd",
				Line: 3,
				Col:  0,
			},
			{
				Str:  "-@type string",
				Line: 4,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) != 2 {
		t.Fatalf("parser annotate type stats is not equal")
	}
}

func TestAnnotateParserCandidate1(t *testing.T) {
	commentInfo := &lexer.CommentInfo{
		LineVec: []lexer.CommentLine{
			{
				Str:  "-@type string | '\"onfis\"'| '\"bbb\"'",
				Line: 1,
				Col:  0,
			},
			{
				Str:  "-@type '\"onfis\"'| '\"bbb\"'",
				Line: 2,
				Col:  0,
			},
		},
	}
	fragent, errVec := ParseCommentFragment(commentInfo)
	if len(errVec) != 0 {
		t.Fatalf("parser annotate return fatal, errstr=%s", errVec[0].ShowStr)
	}
	if len(fragent.Stats) != 2 {
		t.Fatalf("parser annotate type stats is not equal")
	}
}
