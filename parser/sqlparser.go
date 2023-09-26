package parser

import (
	"errors"
	"fmt"
	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/parser/test_driver"
	goio "io"
	"reflect"
	"regexp"
	"strconv"
)

var (
	ErrInvalidBuffer       = errors.New("sqlparser: invalid buffer in Tokenizer")
	ErrInvalidBufferOffset = errors.New("sqlparser: invalid offset of buffer")
	ErrInvalidBufferSize   = errors.New("sqlparser: invalid size of buffer")
	ErrNoStatements        = errors.New("sqlparser: no statements in reader")
	ErrUnknown             = errors.New("sqlparser: unknown error")
)

const (
	maxParseErrorLength = 2000
)

type Tokenizer struct {
	Parser *Parser
	buffer *buffer
	cache  []ast.StmtNode
}

func NewTokenizer(opts ...ParserOption) *Tokenizer {
	p := &Tokenizer{
		Parser: New(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

type ParserOption func(parser *Tokenizer)

func WithBuffer(buffer *buffer) ParserOption {
	return func(parser *Tokenizer) {
		parser.buffer = buffer
	}
}

func (p *Tokenizer) popFromCache() (ast.StmtNode, bool) {
	if len(p.cache) == 0 {
		return nil, false
	}
	stmt := p.cache[0]
	p.cache = p.cache[1:]
	return stmt, true
}

func (p *Tokenizer) ParseNext() (ast.StmtNode, error) {
	if stmt, ok := p.popFromCache(); ok {
		return stmt, nil
	}
	for {
		nodes, err := p.parseReader()
		if err == goio.EOF {
			if len(nodes) == 0 {
				break
			}
		} else if err != nil {
			return nil, err
		}
		if len(nodes) > 0 {
			p.cache = nodes
			break
		}
	}
	stmt, ok := p.popFromCache()
	if !ok {
		return nil, ErrNoStatements
	}
	return stmt, nil
}

func (p *Tokenizer) parseReader() (stmt []ast.StmtNode, err error) {
	if p.buffer == nil || len(p.buffer.buf) == 0 {
		return nil, ErrInvalidBuffer
	}
	var eof bool
	sql, err := p.buffer.ReadString()
	if err == goio.EOF {
		eof = true
		if len(sql) == 0 {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	nodes, _, err := p.Parser.Parse(sql, "", "")
	stmtStartPos, lastScanOffset, reflectErr := p.getParserOffset()
	if reflectErr != nil {
		return nil, errors.New(fmt.Sprintf("failed to get parser offset: %v\n", reflectErr))
	}

	if err != nil {
		parserResult, reflectErr := p.getParserResult()
		if reflectErr != nil {
			return nil, errors.New(fmt.Sprintf("failed to get parser result: %v\n", reflectErr))
		}

		if !eof && lastScanOffset+getErrorLength(err)-stmtStartPos > len(p.buffer.buf)/16 {
			// 增加缓冲区长度，忽略错误重新解析
			withBufferSize(len(p.buffer.buf) * 2)(p.buffer)
			p.buffer.Reset(stmtStartPos)
			return parserResult, nil
		}
		if len(parserResult) > 0 {
			p.buffer.Reset(stmtStartPos)
			return parserResult, nil
		}
		return nil, err
	}
	if eof || stmtStartPos < p.buffer.size {
		if len(nodes) == 0 {
			return nil, ErrNoStatements
		}
		p.buffer.Reset(stmtStartPos)
		return nodes, nil
	}
	if num := len(nodes); num > 1 {
		p.buffer.Reset(stmtStartPos - len(nodes[num-1].OriginalText()))
		return nodes[:num-1], nil
	}
	return nil, ErrUnknown
}

func (p *Tokenizer) getParserResult() ([]ast.StmtNode, error) {
	nodes, err := getFieldPointer(p.Parser, "result")
	if err != nil {
		return nil, err
	}
	return *nodes.(*[]ast.StmtNode), nil
}

func getFieldPointer(obj interface{}, fieldName string) (interface{}, error) {
	val := reflect.ValueOf(obj)
	if val.Kind() != reflect.Pointer {
		return nil, errors.New("value kind is not Pointer: " + val.Kind().String())
	}
	elem := val.Elem()
	if elem.Kind() != reflect.Struct {
		return nil, errors.New("elem kind is not Struct: " + val.Kind().String())
	}
	field := elem.FieldByName(fieldName)
	if field.Kind() == reflect.Invalid {
		return nil, errors.New("field name is not found: " + fieldName)
	}
	return reflect.NewAt(field.Type(), field.Addr().UnsafePointer()).Interface(), nil
}

func (p *Tokenizer) getParserOffset() (stmtStartPos, lastScanOffset int, err error) {
	lexer, err := getFieldPointer(p.Parser, "lexer")
	if err != nil {
		return 0, 0, err
	}
	pos, err := getFieldPointer(lexer, "stmtStartPos")
	if err != nil {
		return 0, 0, err
	}
	off, err := getFieldPointer(lexer, "lastScanOffset")
	if err != nil {
		return 0, 0, err
	}

	return *(pos.(*int)), *(off.(*int)), nil
}

// ParseOneStmt parses a query and returns an ast.StmtNode.
// The query must have one statement, otherwise ErrSyntax is returned.
func (p *Tokenizer) ParseOneStmt(sql, charset, collation string) (ast.StmtNode, error) {
	return p.Parser.ParseOneStmt(sql, charset, collation)
}

// Parse parses a query string to raw ast.StmtNode.
// If charset or collation is "", default charset and collation will be used.
func (p *Tokenizer) Parse(sql, charset, collation string) (stmt []ast.StmtNode, warns []error, err error) {
	return p.Parser.Parse(sql, charset, collation)
}

// ParseSQL parses a query string to raw ast.StmtNode.
func (p *Tokenizer) ParseSQL(sql string, params ...ParseParam) (stmt []ast.StmtNode, warns []error, err error) {
	return p.Parser.ParseSQL(sql, params...)
}

// SplitStatement returns the first sql statement up to either a; or EOF
// and the remainder from the given buffer
func (p *Tokenizer) SplitStatement(blob string) (left string, right string, err error) {
	pieces, err := p.SplitStatementToPieces(blob)
	if err != nil {
		return "", "", err
	}
	switch length := len(pieces); length {
	case 2:
		right = pieces[1]
		fallthrough
	case 1:
		left = pieces[0]
		fallthrough
	case 0:
		break
	default:
		return "", "", errors.New("the number of statements is great than 2")
	}
	return left, right, nil
}

// SplitStatementToPieces split raw sql statement that may have multi sql pieces to sql pieces
// returns the sql pieces blob contains; or error if sql cannot be parsed
func (p *Tokenizer) SplitStatementToPieces(blob string) (pieces []string, err error) {
	tokenizer, _, err := p.Parse(blob, "", "")
	if err != nil {
		return nil, err
	}
	pieces = make([]string, 0, len(tokenizer))
	for _, node := range tokenizer {
		pieces = append(pieces, node.Text())
	}
	return pieces, nil
}

func getErrorLength(err error) int {
	info := err.Error()
	if len(info) > maxParseErrorLength {
		matched := regexpErrorLength.FindStringSubmatch(info)
		if length, err := strconv.Atoi(matched[len(matched)-1]); err == nil && length > len(info) {
			return length
		}
	}
	return len(info)
}

var regexpErrorLength = regexp.MustCompile("\\(total length ([0-9]+)\\)$")
