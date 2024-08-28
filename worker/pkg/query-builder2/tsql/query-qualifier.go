package parser

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	parser "github.com/nucleuscloud/neosync/internal/tsql"
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
}

func newTSqlListener() *tsqlListener {
	return &tsqlListener{}
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

// updates column name to include table
func (l *tsqlListener) EnterFull_column_name(ctx *parser.Full_column_nameContext) {
	if !l.inSearchCondition || (ctx.Full_table_name() != nil && ctx.Full_table_name().GetText() != "") {
		return
	}
	originalToken := ctx.GetStart()
	sourcePair := originalToken.GetSource()
	tokenType := originalToken.GetTokenType()

	startIndex := originalToken.GetStart()
	stopIndex := ctx.GetStop().GetStop()
	channel := originalToken.GetChannel()

	newToken := antlr.NewCommonToken(sourcePair, tokenType, channel, startIndex, stopIndex)
	newToken.SetText(fmt.Sprintf("%s.%q", l.currentTable, ctx.GetText()))
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

	return reconstructSQLFromTree(tree), nil
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

type sqlBuilder struct {
	lastText string // keeps track of last terminal node text
}

// reconstructSQLFromTree traverses the parse tree and builds the SQL string with spaces.
func reconstructSQLFromTree(tree antlr.Tree) string {
	var sb strings.Builder
	sqlBuilder := &sqlBuilder{}
	sqlBuilder.walkAndReconstruct(tree, &sb)
	return strings.TrimSpace(sb.String())
}

// walkAndReconstruct recursively traverses the tree and reconstructs the SQL query.
func (s *sqlBuilder) walkAndReconstruct(tree antlr.Tree, sb *strings.Builder) {
	switch node := tree.(type) {
	case antlr.TerminalNode:
		if node.GetSymbol().GetTokenType() == antlr.TokenEOF {
			return
		}
		// append text without adding space after each terminal node
		s.lastText = node.GetText()
		sb.WriteString(node.GetText())
	case antlr.RuleNode:
		for i := 0; i < node.GetChildCount(); i++ {
			child := node.GetChild(i)
			// only add a space before certain nodes to avoid adding spaces around qualified columns
			if i > 0 && shouldAddSpaceBefore(child) && node.GetText() != "." && s.lastText != "." {
				sb.WriteString(" ")
			}
			s.walkAndReconstruct(child, sb)
		}
	}
}

// shouldAddSpaceBefore determines whether a space should be added before the given tree node.
func shouldAddSpaceBefore(tree antlr.Tree) bool {
	if node, ok := tree.(antlr.TerminalNode); ok {
		text := node.GetText()
		if text == "." || text == "," {
			return false
		}
	}
	return true
}
