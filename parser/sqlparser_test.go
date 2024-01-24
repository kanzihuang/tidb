package parser

import (
	db_sql "database/sql"
	testify_require "github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestSplitStatement(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		output string
		rem    string
		err    string
	}{
		{
			input:  "select * from tbl",
			output: "select * from tbl",
		}, {
			input:  "select * from tbl; ",
			output: "select * from tbl;",
		}, {
			input:  "select * from tbl; select * from tbl2;",
			output: "select * from tbl;",
			rem:    " select * from tbl2;",
		}, {
			input:  "select * from /* comment */ tbl;",
			output: "select * from /* comment */ tbl;",
		}, {
			input:  "select * from /* comment ; */ tbl;",
			output: "select * from /* comment ; */ tbl;",
		}, {
			input:  "select * from tbl where semi = ';';",
			output: "select * from tbl where semi = ';';",
		}, {
			input:  "-- select * from tbl",
			output: "",
		}, {
			input:  " ",
			output: "",
		}, {}, {
			input:  ";",
			output: "",
		}, {
			input: "",
		}, {
			name:   "Trailing ;",
			input:  "select 1 from a; update a set b = 2;",
			output: "select 1 from a;",
			rem:    " update a set b = 2;",
		}, {
			name:   "No trailing ;",
			input:  "select 1 from a; update a set b = 2",
			output: "select 1 from a;",
			rem:    " update a set b = 2",
		}, {
			name:   "Trailing whitespace",
			input:  "select 1 from a; update a set b = 2    ",
			output: "select 1 from a;",
			rem:    " update a set b = 2    ",
		}, {
			name:   "Trailing whitespace and ;",
			input:  "select 1 from a; update a set b = 2   ;   ",
			output: "select 1 from a;",
			rem:    " update a set b = 2   ;",
		}, {
			name:   "Handle ForceEOF statements",
			input:  "set character set utf8; select 1 from a",
			output: "set character set utf8;",
			rem:    " select 1 from a",
		}, {
			name:   "Semicolin inside a string",
			input:  "set character set ';'; select 1 from a",
			output: "set charset ';';",
			rem:    "select 1 from a",
			err:    "Unknown character set: ';'",
		}, {
			name:   "Partial DDL",
			input:  "create table a; select 1 from a",
			output: "create table a;",
			rem:    " select 1 from a",
		}, {
			name:  "Partial DDL",
			input: "create table a; select 1 from",
			err:   "near \"\"",
		}, {
			name:   "Partial DDL",
			input:  "create table a ignore me this is garbage; select 1 from a",
			output: "create table a;",
			rem:    "select 1 from a",
			err:    "near \"me this is garbage;",
		}, {
			input:  "select * from table1;--comment;\nselect * from table2;",
			output: "select * from table1;--comment;",
			rem:    "select * from table2",
			err:    "near \"--comment;",
		}, {
			input: "CREATE TABLE `total_data` (`id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id', " +
				"`region` varchar(32) NOT NULL COMMENT 'region name, like zh; th; kepler'," +
				"`data_size` bigint NOT NULL DEFAULT '0' COMMENT 'data size;'," +
				"`createtime` datetime NOT NULL DEFAULT NOW() COMMENT 'create time;'," +
				"`comment` varchar(100) NOT NULL DEFAULT '' COMMENT 'comment'," +
				"PRIMARY KEY (`id`))",
			output: "CREATE TABLE `total_data` (`id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id', " +
				"`region` varchar(32) NOT NULL COMMENT 'region name, like zh; th; kepler'," +
				"`data_size` bigint NOT NULL DEFAULT '0' COMMENT 'data size;'," +
				"`createtime` datetime NOT NULL DEFAULT NOW() COMMENT 'create time;'," +
				"`comment` varchar(100) NOT NULL DEFAULT '' COMMENT 'comment'," +
				"PRIMARY KEY (`id`))",
		},
	}

	p := NewTokenizer()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sql, rem, err := p.SplitStatement(tc.input)
			if len(tc.err) > 0 {
				testify_require.Error(t, err)
				testify_require.Contains(t, err.Error(), tc.err)
				return
			}
			testify_require.NoError(t, err)
			testify_require.Equal(t, tc.output, sql)
			testify_require.Equal(t, tc.rem, rem)
		})
	}
}

func TestSQLParser_ParseNext(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []string
		size   int
		err    string
	}{
		{
			name:  "one statement",
			input: "select 1 from t;",
			output: []string{
				"select 1 from t;",
			},
			size: 20,
		},
		{
			name:  "two statement",
			input: "select 1 from t;select 2 from t;",
			output: []string{
				"select 1 from t;",
				"select 2 from t;",
			},
			size: 20,
		},
		{
			name:  "three statement",
			input: "select 1 from t;select 2 from t;select 3 from t;",
			output: []string{
				"select 1 from t;",
				"select 2 from t;",
				"select 3 from t;",
			},
			size: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewTokenizer(WithBuffer(NewBuffer(strings.NewReader(tt.input), withBufferSize(tt.size))))
			for i := 0; i < len(tt.output); i++ {
				node, err := p.ParseNext()
				if err == ErrNoStatements {
					break
				} else if err != nil {
					testify_require.Error(t, err)
					testify_require.Contains(t, err.Error(), tt.err)
					return
				}
				testify_require.NoError(t, err)
				testify_require.Equal(t, tt.output[i], node.Text())
			}
			_, err := p.ParseNext()
			testify_require.Equal(t, ErrNoStatements, err)
		})
	}
}

func TestSQLParser_getParserResult(t *testing.T) {
	p := NewTokenizer()
	input := "select 1 from t;select 1 from"
	output := "select 1 from t;"
	_, _, _ = p.Parse(input, "", "")
	result, err := p.getParserResult()
	testify_require.NoError(t, err)
	testify_require.Equal(t, 1, len(result))
	testify_require.Equal(t, output, result[0].Text())
}

func Test_getFieldPointer(t *testing.T) {
	var (
		id   = 1024
		addr = "Beijing"
		buf  = []byte("buffer")
		time = db_sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	)
	testCase := &struct {
		name      string
		fieldName string
		id        int
		addr      string
		buf       []byte
		time      db_sql.NullTime
		pointer   *db_sql.NullTime
		want      interface{}
		err       error
	}{}

	idPtr, err := getFieldPointer[int](testCase, "id")
	testify_require.NoError(t, err)
	testify_require.Zero(t, *idPtr)
	*idPtr = id
	idPtr, _ = getFieldPointer[int](testCase, "id")
	testify_require.Equal(t, id, *idPtr)

	addrPtr, err := getFieldPointer[string](testCase, "addr")
	testify_require.NoError(t, err)
	testify_require.Zero(t, *addrPtr)
	*addrPtr = addr
	addrPtr, _ = getFieldPointer[string](testCase, "addr")
	testify_require.Equal(t, addr, *addrPtr)

	bufPtr, err := getFieldPointer[[]byte](testCase, "buf")
	testify_require.NoError(t, err)
	testify_require.Zero(t, *bufPtr)
	*bufPtr = buf
	bufPtr, _ = getFieldPointer[[]byte](testCase, "buf")
	testify_require.Equal(t, buf, *bufPtr)

	timePtr, err := getFieldPointer[db_sql.NullTime](testCase, "time")
	testify_require.NoError(t, err)
	testify_require.Zero(t, *timePtr)
	*timePtr = time
	timePtr, _ = getFieldPointer[db_sql.NullTime](testCase, "time")
	testify_require.Equal(t, time, *timePtr)

	pointer, err := getFieldPointer[*db_sql.NullTime](testCase, "pointer")
	testify_require.NoError(t, err)
	testify_require.Zero(t, *pointer)
	*pointer = &time
	pointer, _ = getFieldPointer[*db_sql.NullTime](testCase, "pointer")
	testify_require.Equal(t, &time, *pointer)
}

func TestSQLParser_ParseOneStmt(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		err    error
	}{
		{
			name:  "select from `my-table`",
			input: "select 'a\\nb' from `my-table`",
		},
		{
			name:  "select from `my-tabel1`",
			input: "select 'a\\nb' from `my-table1`",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewTokenizer()
			got, err := p.ParseOneStmt(tt.input, "", "")
			testify_require.Equal(t, tt.err, err)
			if err != nil {
				return
			}
			if len(tt.output) == 0 {
				tt.output = tt.input
			}
			stmt := got.Text()
			testify_require.Equal(t, tt.output, stmt)
		})
	}
}
