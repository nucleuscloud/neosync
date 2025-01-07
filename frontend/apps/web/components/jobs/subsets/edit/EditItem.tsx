import { ConnectionConfigCase } from '@/app/(mgmt)/[account]/connections/util';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import { Editor, useMonaco } from '@monaco-editor/react';
import {
  CheckSqlQueryResponse,
  CheckSqlQueryResponseSchema,
  ConnectionDataService,
  ConnectionService,
  GetTableRowCountResponse,
} from '@neosync/sdk';
import { editor } from 'monaco-editor';
import { useTheme } from 'next-themes';
import { ReactElement, useEffect, useRef, useState } from 'react';
import ValidateQueryErrorAlert from '../SubsetErrorAlert';
import { SubsetTableRow } from '../SubsetTable/Columns';
import ValidateQueryBadge from '../ValidateQueryBadge';
import {
  isSubsetRowCountSupported,
  isSubsetValidationSupported,
  ValidSubsetConnectionType,
} from '../utils';

interface Props {
  item?: SubsetTableRow;
  onItem(item?: SubsetTableRow): void;
  onSave(): void;
  onCancel(): void;
  connectionId: string;
  connectionType: ValidSubsetConnectionType;
  columns: string[];
}
export default function EditItem(props: Props): ReactElement {
  const {
    item,
    onItem,
    onSave,
    onCancel,
    connectionId,
    connectionType,
    columns,
  } = props;
  const [validateResp, setValidateResp] = useState<
    CheckSqlQueryResponse | undefined
  >();
  const [tableRowCountResp, setTableRowCountResp] = useState<
    GetTableRowCountResponse | undefined
  >();
  const [calculatingRowCount, setCalculatingRowCount] = useState(false);
  const { resolvedTheme } = useTheme();
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const [rowCountError, setRowCountError] = useState<string>();

  const monaco = useMonaco();

  const showRowCountButton = isSubsetRowCountSupported(connectionType);
  const showValidateButton = isSubsetValidationSupported(connectionType);

  useEffect(() => {
    if (monaco) {
      const provider = monaco.languages.registerCompletionItemProvider('sql', {
        triggerCharacters: [' ', '.'], // Trigger autocomplete on space and dot

        provideCompletionItems: (model, position) => {
          const textUntilPosition = model.getValueInRange({
            startLineNumber: 1,
            startColumn: 1,
            endLineNumber: position.lineNumber,
            endColumn: position.column,
          });

          const columnSet = new Set<string>(columns);

          // Check if the last character or word should trigger the auto-complete
          if (!shouldTriggerAutocomplete(textUntilPosition)) {
            return { suggestions: [] };
          }

          const word = model.getWordUntilPosition(position);

          const range = {
            startLineNumber: position.lineNumber,
            startColumn: word.startColumn,
            endLineNumber: position.lineNumber,
            endColumn: word.endColumn,
          };

          const suggestions = Array.from(columnSet).map((name) => ({
            label: name, // would be nice if we could add the type here as well?
            kind: monaco.languages.CompletionItemKind.Field,
            insertText: name,
            range: range,
          }));

          return { suggestions: suggestions };
        },
      });
      /* disposes of the instance if the component re-renders, otherwise the auto-compelte list just keeps appending the column names to the auto-complete, so you get liek 20 'address' entries for ex. then it re-renders and then it goes to 30 'address' entries
       */
      return () => {
        provider.dispose();
      };
    }
  }, [monaco, columns]);

  function onWhereChange(value: string): void {
    if (!item) {
      return;
    }
    onItem({ ...item, where: value });
  }

  useEffect(() => {
    setTableRowCountResp(undefined);
    setValidateResp(undefined);
  }, [item]);

  const { mutateAsync: validateSql } = useMutation(
    ConnectionService.method.checkSqlQuery
  );
  const { mutateAsync: getRowCountByTable } = useMutation(
    ConnectionDataService.method.getTableRowCount
  );

  async function onValidate(): Promise<void> {
    if (connectionType === 'pgConfig' || connectionType === 'mysqlConfig') {
      const pgString = `select * from "${item?.schema}"."${item?.table}" WHERE ${item?.where};`;
      const mysqlString = `select * from \`${item?.schema}\`.\`${item?.table}\` WHERE ${item?.where};`;

      try {
        const resp = await validateSql({
          id: connectionId,
          query: connectionType === 'mysqlConfig' ? mysqlString : pgString,
        });
        setValidateResp(resp);
      } catch (err) {
        setValidateResp(
          create(CheckSqlQueryResponseSchema, {
            isValid: false,
            erorrMessage: getErrorMessage(err),
          })
        );
      }
    }
  }

  async function onGetRowCount(): Promise<void> {
    try {
      setTableRowCountResp(undefined);
      setCalculatingRowCount(true);
      const resp = await getRowCountByTable({
        connectionId: connectionId,
        schema: item?.schema,
        table: item?.table,
        whereClause: item?.where,
      });
      setTableRowCountResp(resp);
      setRowCountError('');
    } catch (err) {
      setCalculatingRowCount(false);
      console.error(err);
      setRowCountError(getErrorMessage(err));
    } finally {
      setCalculatingRowCount(false);
    }
  }

  function onCancelClick(): void {
    setValidateResp(undefined);
    setTableRowCountResp(undefined);
    onCancel();
  }
  function onSaveClick(): void {
    setValidateResp(undefined);
    setTableRowCountResp(undefined);
    onSave();
  }

  // options for the sql editor
  const editorOptions = {
    minimap: { enabled: false },
    roundedSelection: false,
    scrollBeyondLastLine: false,
    readOnly: !item,
    renderLineHighlight: 'none' as const,
    overviewRulerBorder: false,
    overviewRulerLanes: 0,
    lineNumbers: !item || item.where == '' ? ('off' as const) : ('on' as const),
  };

  const constructWhere = (value: string) => {
    if (item?.where && !value.startsWith('WHERE ')) {
      return `WHERE ${value}`;
    } else if (!item?.where) {
      return '';
    }
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col md:flex-row justify-between gap-2 md:gap-0">
        <div className="flex flex-row gap-4">
          {item?.schema && (
            <div
              className={cn(
                'flex flex-row gap-2 items-center',
                showSchema(connectionType) ? undefined : 'hidden'
              )}
            >
              <span className="font-semibold tracking-tight">Schema</span>
              <Badge
                className="px-4 py-2 dark:border-gray-700"
                variant={item?.schema ? 'outline' : 'secondary'}
              >
                {item?.schema ?? ''}
              </Badge>
            </div>
          )}
          {item?.table && (
            <div className="flex flex-row gap-2 items-center">
              <span className="font-semibold tracking-tight">Table</span>
              <Badge
                className="px-4 py-2 dark:border-gray-700"
                variant={item?.table ? 'outline' : 'secondary'}
              >
                {item?.table ?? ''}
              </Badge>
            </div>
          )}
          <div className="flex flex-row items-center">
            <ValidateQueryBadge resp={validateResp} />
          </div>
        </div>
      </div>
      <div className="flex flex-col items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <Editor
          height="60px"
          width="100%"
          language="sql"
          value={constructWhere(item?.where ?? '')}
          theme={resolvedTheme === 'dark' ? 'vs-dark' : 'light'}
          onChange={(e) => onWhereChange(e?.replace('WHERE ', '') ?? '')}
          options={editorOptions}
          onMount={(editor) => {
            editorRef.current = editor;
            editor.focus();
          }}
        />
      </div>
      <ValidateQueryErrorAlert
        validateResp={validateResp}
        rowCountError={rowCountError}
      />
      <div className="flex justify-between gap-4">
        <Button
          type="button"
          variant="secondary"
          disabled={!item}
          onClick={() => onCancelClick()}
        >
          <ButtonText text="Cancel" />
        </Button>
        <div className="flex flex-row gap-2">
          {showRowCountButton && (
            <>
              <TooltipProvider>
                <Tooltip delayDuration={200}>
                  <TooltipTrigger asChild>
                    <Button
                      type="button"
                      variant="outline"
                      disabled={!item?.where}
                      onClick={() => onGetRowCount()}
                    >
                      {calculatingRowCount ? (
                        <Spinner className="text-black dark:text-white" />
                      ) : (
                        <ButtonText text={'Row Count'} />
                      )}
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>
                      Attempts to run a SQL COUNT(*) statement against the
                      source connection for the table with the included WHERE
                      clause
                    </p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
              {tableRowCountResp && tableRowCountResp.count >= 0 ? (
                <Badge
                  variant="darkOutline"
                  className="dark:bg-gray-800 dark:border-gray-800"
                >
                  {tableRowCountResp.count.toString()}
                </Badge>
              ) : null}
            </>
          )}
          {showValidateButton && (
            <TooltipProvider>
              <Tooltip delayDuration={200}>
                <TooltipTrigger asChild>
                  <Button
                    type="button"
                    variant="outline"
                    disabled={!item}
                    onClick={() => onValidate()}
                  >
                    <ButtonText text="Validate" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>
                    Attempts to run a SQL PREPARE statement against the source
                    connection for the table with the included WHERE clause
                  </p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          )}
          <TooltipProvider>
            <Tooltip delayDuration={200}>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  disabled={!item}
                  onClick={() => {
                    const editor = editorRef.current;
                    editor?.setValue('');
                    onSaveClick();
                  }}
                >
                  <ButtonText text="Apply" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>
                  Applies changes to table only, click Save below to fully
                  submit changes
                </p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>
    </div>
  );
}

function showSchema(connectionType: ConnectionConfigCase | null): boolean {
  return connectionType === 'pgConfig' || connectionType === 'mysqlConfig';
}

function shouldTriggerAutocomplete(text: string): boolean {
  const trimmedText = text.trim();
  const textSplit = trimmedText.split(/\s+/);
  const lastSignificantWord = trimmedText.split(/\s+/).pop()?.toUpperCase();
  const triggerKeywords = ['SELECT', 'WHERE', 'AND', 'OR', 'FROM'];

  if (textSplit.length == 2 && textSplit[0].toUpperCase() == 'WHERE') {
    /* since we pre-pend the 'WHERE', we want the autocomplete to show up for the first letter typed
     which would come through as 'WHERE a' if the user just typed the letter 'a'
     so the when we split that text, we check if the length is 2 (as a way of checking if the user has only typed one letter or is still on the first word) and if it is and the first word is 'WHERE' which it should be since we pre-pend it, then show the auto-complete */
    return true;
  } else {
    return (
      triggerKeywords.includes(lastSignificantWord || '') ||
      triggerKeywords.some((keyword) => trimmedText.endsWith(keyword + ' '))
    );
  }
}
