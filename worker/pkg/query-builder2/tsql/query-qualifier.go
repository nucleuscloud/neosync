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
	currentTable      string   // stores most recent table found in from clause
	inSearchCondition bool     // tracks when we enter where clause in the tree
	sqlStack          []string // rebuilds new sql string
	Errors            []string
}

func newTSqlListener() *tsqlListener {
	return &tsqlListener{}
}

// builds sql string
func (l *tsqlListener) sqlString() string {
	return strings.TrimSpace(strings.Join(l.sqlStack, ""))
}

// adds string to sql stack
func (l *tsqlListener) push(str string) {
	l.sqlStack = append(l.sqlStack, str)
}

// removes last element in sql stack
func (l *tsqlListener) pop() string {
	if len(l.sqlStack) < 1 {
		l.Errors = append(l.Errors, "stack is empty unable to pop")
		return ""
	}

	result := l.sqlStack[len(l.sqlStack)-1]
	l.sqlStack = l.sqlStack[:len(l.sqlStack)-1]

	return result
}

// creates new tree token with given text
func (l *tsqlListener) setToken(startToken, stopToken antlr.Token, text string) *antlr.CommonToken {
	sourcePair := startToken.GetSource()
	tokenType := startToken.GetTokenType()

	startIndex := startToken.GetStart()
	stopIndex := stopToken.GetStop()
	channel := startToken.GetChannel()

	newToken := antlr.NewCommonToken(sourcePair, tokenType, channel, startIndex, stopIndex)
	newToken.SetText(text)
	return newToken
}

/*
 the following are parser events that are activated by walking the parsed tree
 renaming these functions will break the parser
 available listeners can be found at https://github.com/nucleuscloud/go-antlrv4-parser/blob/main/tsql/tsqlparser_base_listener.go
*/

// parser listener event when we enter where clause
func (l *tsqlListener) EnterSearch_condition(ctx *parser.Search_conditionContext) {
	l.inSearchCondition = true
}

// parser listener event when we exit where clause
func (l *tsqlListener) ExitSearch_condition(ctx *parser.Search_conditionContext) {
	l.inSearchCondition = false
}

// parser listener event when we enter select statement
func (l *tsqlListener) EnterSelect_statement(ctx *parser.Select_statementContext) {
	// important so we don't process select columns
	l.inSearchCondition = false
}

// sets current table found in from clause
func (l *tsqlListener) EnterTable_sources(ctx *parser.Table_sourcesContext) {
	table := ctx.GetText()
	l.currentTable = qualifyTableName(table)
}

// sets current table if alias found
func (l *tsqlListener) EnterTable_alias(ctx *parser.Table_aliasContext) {
	l.currentTable = ctx.GetText()
}

// rebuilds sql string from tree
// adds terminal node text to sql stack and adds appropriate spacing
func (l *tsqlListener) VisitTerminal(node antlr.TerminalNode) {
	if node.GetSymbol().GetTokenType() != antlr.TokenEOF {
		text := node.GetText()
		if text == "," {
			// add space after commas
			l.pop()
			l.push(text)
			l.push(" ")
		} else if text == "." {
			// remove space before periods
			// should be table.column not table . column
			l.pop()
			l.push(text)
		} else {
			// add space after each node text
			l.push(text)
			l.push(" ")
		}
	}
}

// update table name and add qualifiers
func (l *tsqlListener) EnterFull_table_name(ctx *parser.Full_table_nameContext) {
	if !l.inSearchCondition {
		// ignore any table names not in where clause
		return
	}
	// creates new token with table name
	newToken := l.setToken(ctx.GetStart(), ctx.GetStop(), ensureQuoted(l.currentTable))
	ctx.RemoveLastChild()
	ctx.AddTokenNode(newToken)
}

// updates column name
// add table name if missing
func (l *tsqlListener) EnterFull_column_name(ctx *parser.Full_column_nameContext) {
	if !l.inSearchCondition {
		// ignore any table names not in where clause
		return
	}

	var text string
	if ctx.Full_table_name() == nil || ctx.Full_table_name().GetText() == "" {
		text = fmt.Sprintf("%s.%s", ensureQuoted(l.currentTable), parseColumnName(ctx.GetText()))
	} else {
		text = parseColumnName(ctx.GetText())
	}

	newToken := l.setToken(ctx.GetStart(), ctx.GetStop(), text)
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
	// add error listener
	errorListener := newTSqlErrorListener()
	p.AddErrorListener(errorListener)

	listener := newTSqlListener()
	tree := p.Tsql_file()
	// walk tree and listen to events
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	if len(errorListener.Errors) > 0 {
		return "", fmt.Errorf("SQL parsing errors: %s", strings.Join(errorListener.Errors, "; "))
	}
	if len(listener.Errors) > 0 {
		return "", fmt.Errorf("SQL building errors: %s", strings.Join(listener.Errors, "; "))
	}

	return listener.sqlString(), nil
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
