package parser

import (
	"fmt"
	testify_require "github.com/stretchr/testify/require"
	"strings"
	"testing"
)

type stmtCase struct {
	name   string
	input  string
	output string
	err    string
}

func (stmt *stmtCase) getOutput() string {
	if len(stmt.output) == 0 {
		return stmt.input
	}
	return stmt.output
}

type stmtsCase []stmtCase

func (stmts stmtsCase) getNumber() int {
	return len(stmts)
}

func (stmts stmtsCase) getName(i int) string {
	if len(stmts) > 0 {
		return stmts[0].name
	}
	return ""
}

func (stmts stmtsCase) concatInput() string {
	var input strings.Builder
	for i := 0; i < len(stmts); i++ {
		input.WriteString(stmts.getInput(i))
	}
	return input.String()
}

func (stmts stmtsCase) getInput(i int) string {
	if i >= len(stmts) {
		return ""
	}
	input := stmts[i].input
	if i < len(stmts)-1 && (len(input) == 0 || input[len(input)-1] != ';') {
		input += ";"
	}
	return input
}

func (stmts stmtsCase) getOutput(i int) string {
	if i >= len(stmts) {
		return ""
	}
	stmt := stmts[i]
	output := stmt.output
	if len(output) == 0 {
		output = stmt.input
	}
	if i < len(stmts)-1 && (len(output) == 0 || output[len(output)-1] != ';') {
		output += ";"
	}
	return output
}

func (stmts stmtsCase) getError() string {
	for i := 0; i < len(stmts); i++ {
		length := len(stmts[i].err)
		if length > 0 {
			err := stmts[i].err
			if i < len(stmts)-1 && err[length-1] == '"' {
				err = fmt.Sprintf("%s;%s", err[:length-1], stmts[i+1].input)
			}
			return err
		}
	}
	return ""
}

func (stmts stmtsCase) filter(valid bool) stmtsCase {
	result := make(stmtsCase, 0, len(stmts))
	for _, stmt := range stmts {
		if valid == (len(stmt.err) == 0) {
			result = append(result, stmt)
		}
	}
	return result
}

func testParseOneStmt(t *testing.T, stmts stmtsCase) {
	testAllCases(t, newStmtsCases(stmts))
}

type stmtsCases []stmtsCase

func newStmtsCases(stmts stmtsCase) stmtsCases {
	cases := make(stmtsCases, 0, len(stmts))
	for _, stmt := range stmts {
		cases = append(cases, stmtsCase{
			stmt,
		})
	}
	return cases
}

func (cases stmtsCases) countStmts() int {
	num := 0
	for _, c := range cases {
		num += len(c)
	}
	return num
}

func (cases stmtsCases) expand() stmtsCase {
	stmts := make(stmtsCase, 0, cases.countStmts())
	for _, c := range cases {
		stmts = append(stmts, c...)
	}
	return stmts
}

func (cases stmtsCases) insert(stmt stmtCase, headOrTail bool) stmtsCases {
	result := make(stmtsCases, 0, len(cases))
	for _, tc := range cases {
		stmts := make(stmtsCase, 0, len(tc)+1)
		if headOrTail {
			stmts = append(stmts, stmt)
			stmts = append(stmts, tc...)
		} else {
			stmts = append(stmts, tc...)
			stmts = append(tc, stmt)
		}
		result = append(result, stmts)
	}
	return result
}

func testSplitStatement(t *testing.T, stmts stmtsCase, second stmtCase) {
	headOrTail := []bool{true, false}
	for _, b := range headOrTail {
		cases := newStmtsCases(stmts)
		cases.insert(second, b)
		testAllCases(t, cases)
	}
}

func testAllCases(t *testing.T, testCases []stmtsCase) {
	for _, stmts := range testCases {
		testParse(t, stmts)
		t.Run(stmts.getName(0), func(t *testing.T) {
			testParse(t, stmts)
		})
	}
}

func testParse(t *testing.T, stmts stmtsCase) {
	p := NewTokenizer()
	firstInput := stmts.getInput(0)
	nodes, _, err := p.Parse(stmts.concatInput(), "", "")
	wantErr := stmts.getError()
	if len(wantErr) > 0 {
		testify_require.Error(t, err, firstInput)
		testify_require.Contains(t, err.Error(), wantErr, firstInput)
		return
	}
	testify_require.NoError(t, err, firstInput)
	for i, node := range nodes {
		testify_require.Equal(t, stmts.getOutput(i), node.Text(), node.OriginalText())
	}
}

func testParseNext(t *testing.T, stmts stmtsCase) {
	p := NewTokenizer(WithBuffer(NewBuffer(strings.NewReader(stmts.concatInput()), withBufferSize(1024))))
	wantErr := stmts.getError()
	for i := 0; i < len(stmts); i++ {
		node, err := p.ParseNext()
		if err == ErrNoStatements {
			break
		} else if err != nil {
			if len(wantErr) > 0 {
				testify_require.Contains(t, err.Error(), wantErr)
			} else {
				testify_require.NoError(t, err)
			}
			return
		}
		testify_require.Equal(t, stmts.getOutput(i), node.Text())
	}
	_, err := p.ParseNext()
	testify_require.Equal(t, ErrNoStatements, err)
	testify_require.Empty(t, wantErr)
}
