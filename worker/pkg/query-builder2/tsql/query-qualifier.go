package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/nucleuscloud/go-antlrv4-parser/tsql"
)

type tSqlErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []string
}

func newTSqlErrorListener() *tSqlErrorListener {
	return &tSqlErrorListener{
		Errors: []string{},
	}
}

func (l *tSqlErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol any, line, column int, msg string, e antlr.RecognitionException) {
	errorMessage := fmt.Sprintf("line %d:%d %s", line, column, msg)
	l.Errors = append(l.Errors, errorMessage)
}

type tsqlListener struct {
	*parser.BaseTSqlParserListener
	currentTable      string
	inSearchCondition bool
	sqlStack          []string
	Errors            []string
}

func newTSqlListener() *tsqlListener {
	return &tsqlListener{}
}

func (l *tsqlListener) SqlString() string {
	return strings.TrimSpace(strings.Join(l.sqlStack, ""))
}

func (l *tsqlListener) Push(str string) {
	l.sqlStack = append(l.sqlStack, str)
}

func (l *tsqlListener) Pop() string {
	if len(l.sqlStack) < 1 {
		l.Errors = append(l.Errors, "stack is empty unable to pop")
		return ""
	}

	result := l.sqlStack[len(l.sqlStack)-1]
	l.sqlStack = l.sqlStack[:len(l.sqlStack)-1]

	return result
}

// EnterSearch_condition is called when production search_condition is entered.
func (l *tsqlListener) EnterSearch_condition(ctx *parser.Search_conditionContext) {
	l.inSearchCondition = true
}

// ExitSearch_condition is called when production search_condition is exited.
func (l *tsqlListener) ExitSearch_condition(ctx *parser.Search_conditionContext) {
	l.inSearchCondition = false
}

// EnterSelect_statement is called when production select_statement is entered.
func (l *tsqlListener) EnterSelect_statement(ctx *parser.Select_statementContext) {
	l.inSearchCondition = false
}

// sets current table
func (l *tsqlListener) EnterTable_sources(ctx *parser.Table_sourcesContext) {
	table := ctx.GetText()
	l.currentTable = qualifyTableName(table)
}

// EnterTable_alias is called when production table_alias is entered.
func (l *tsqlListener) EnterTable_alias(ctx *parser.Table_aliasContext) {
	l.currentTable = ctx.GetText()
}

func (l *tsqlListener) VisitTerminal(node antlr.TerminalNode) {
	if node.GetSymbol().GetTokenType() != antlr.TokenEOF {
		text := node.GetText()
		if text == "," {
			l.Pop()
			l.Push(text)
			l.Push(" ")
		} else if text == "." {
			l.Pop()
			l.Push(text)
		} else {
			l.Push(text)
			l.Push(" ")
		}
	}
}

func (l *tsqlListener) SetToken(startToken, stopToken antlr.Token, text string) *antlr.CommonToken {
	sourcePair := startToken.GetSource()
	tokenType := startToken.GetTokenType()

	startIndex := startToken.GetStart()
	stopIndex := stopToken.GetStop()
	channel := startToken.GetChannel()

	newToken := antlr.NewCommonToken(sourcePair, tokenType, channel, startIndex, stopIndex)
	newToken.SetText(text)
	return newToken
}

// update table name and add qualifiers
func (l *tsqlListener) EnterFull_table_name(ctx *parser.Full_table_nameContext) {
	if !l.inSearchCondition {
		return
	}
	newToken := l.SetToken(ctx.GetStart(), ctx.GetStop(), ensureQuoted(l.currentTable))
	ctx.RemoveLastChild()
	ctx.AddTokenNode(newToken)
}

// updates column name
// add table name if missing
func (l *tsqlListener) EnterFull_column_name(ctx *parser.Full_column_nameContext) {
	if !l.inSearchCondition {
		return
	}

	var text string
	if ctx.Full_table_name() == nil || ctx.Full_table_name().GetText() == "" {
		text = fmt.Sprintf("%s.%s", ensureQuoted(l.currentTable), parseColumnName(ctx.GetText()))
	} else {
		text = parseColumnName(ctx.GetText())
	}

	newToken := l.SetToken(ctx.GetStart(), ctx.GetStop(), text)
	ctx.RemoveLastChild()
	ctx.AddTokenNode(newToken)
}

func QualifyWhereCondition(sql string) (string, error) {
	inputStream := antlr.NewInputStream(sql)

	// create the lexer
	lexer := parser.NewTSqlLexer(inputStream)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// create the parser
	p := parser.NewTSqlParser(tokens)
	errorListener := newTSqlErrorListener()
	p.AddErrorListener(errorListener)

	listener := newTSqlListener()
	tree := p.Tsql_file()
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	if len(errorListener.Errors) > 0 {
		return "", fmt.Errorf("SQL parsing errors: %s", strings.Join(errorListener.Errors, "; "))
	}
	if len(listener.Errors) > 0 {
		return "", fmt.Errorf("SQL building errors: %s", strings.Join(listener.Errors, "; "))
	}

	return listener.SqlString(), nil
}

func parseColumnName(colText string) string {
	split := strings.Split(colText, ".")
	if len(split) == 1 {
		return ensureQuoted(split[0])
	}
	if len(split) == 2 {
		return ensureQuoted(split[1])
	}
	return ensureQuoted(colText)
}

func ensureQuoted(str string) string {
	if strings.HasPrefix(str, `"`) && strings.HasSuffix(str, `"`) {
		return str
	}
	return fmt.Sprintf("%q", str)
}

// adds quotes around schema and table
func qualifyTableName(table string) string {
	if strings.HasPrefix(table, `"`) && strings.HasSuffix(table, `"`) {
		return table
	}
	split := strings.Split(table, ".")
	qualifiedName := []string{}
	for _, piece := range split {
		qualifiedName = append(qualifiedName, fmt.Sprintf("%q", piece))
	}
	return strings.Join(qualifiedName, ".")
}
