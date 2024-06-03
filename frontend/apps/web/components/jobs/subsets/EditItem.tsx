import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { getErrorMessage } from '@/util/util';
import { Editor, useMonaco } from '@monaco-editor/react';
import { CheckSqlQueryResponse, GetTableRowCountResponse } from '@neosync/sdk';
import { editor } from 'monaco-editor';
import { useTheme } from 'next-themes';
import { ReactElement, useEffect, useRef, useState } from 'react';
import ValidateQueryErrorAlert from './SubsetErrorAlert';
import ValidateQueryBadge from './ValidateQueryBadge';
import { TableRow } from './subset-table/column';

interface Props {
  item?: TableRow;
  onItem(item?: TableRow): void;
  onSave(): void;
  onCancel(): void;
  connectionId: string;
  dbType: string;
  columns: string[];
}
export default function EditItem(props: Props): ReactElement {
  const { item, onItem, onSave, onCancel, connectionId, dbType, columns } =
    props;
  const [validateResp, setValidateResp] = useState<
    CheckSqlQueryResponse | undefined
  >();
  const [tableRowCountResp, setTableRowCountResp] = useState<
    GetTableRowCountResponse | undefined
  >();
  const [calculatingRowCount, setCalculatingRowCount] = useState(false);
  const { account } = useAccount();
  const { resolvedTheme } = useTheme();
  const editorRef = useRef<editor.IStandaloneCodeEditor | null>(null);
  const [rowCountError, setRowCountError] = useState<string>();

  const monaco = useMonaco();

  useEffect(() => {
    if (monaco) {
      // Register a completion item provider for the SQL language
      const provider = monaco.languages.registerCompletionItemProvider('sql', {
        triggerCharacters: [' ', '.'], // Trigger autocomplete on space and dot

        provideCompletionItems: (model, position) => {
          const textUntilPosition = model.getValueInRange({
            startLineNumber: 1,
            startColumn: 1,
            endLineNumber: position.lineNumber,
            endColumn: position.column,
          });

          console.log('columns', columns);

          const columnSet = new Set<string>(columns);

          // Check if the last character or word indicates a context where columns might be needed
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
            label: name, // would be nice if we could add the type here as well
            kind: monaco.languages.CompletionItemKind.Field,
            insertText: name,
            range: range,
          }));

          console.log('suggestions', suggestions);

          return { suggestions: suggestions };
        },
      });
      // disposes of the instance if the component re-renders
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

  async function onValidate(): Promise<void> {
    const pgSting = `select * from "${item?.schema}"."${item?.table}" WHERE ${item?.where};`;
    const mysqlString = `select * from \`${item?.schema}\`.\`${item?.table}\` WHERE ${item?.where};`;

    try {
      const resp = await validateSql(
        account?.id ?? '',
        connectionId,
        dbType == 'mysql' ? mysqlString : pgSting
      );
      setValidateResp(resp);
    } catch (err) {
      setValidateResp(
        new CheckSqlQueryResponse({
          isValid: false,
          erorrMessage: getErrorMessage(err),
        })
      );
    }
  }

  async function onGetRowCount(): Promise<void> {
    try {
      setTableRowCountResp(undefined);
      setCalculatingRowCount(true);
      const resp = await getTableRowCount(
        account?.id ?? '',
        connectionId,
        item?.schema ?? '',
        item?.table ?? '',
        item?.where
      );
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
      <div className="flex flex-row justify-between">
        <div className="flex flex-row gap-4">
          <div className="flex flex-row gap-2 items-center">
            <span className="font-semibold tracking-tight">Schema</span>
            <Badge
              className="px-4 py-2 dark:border-gray-700"
              variant={item?.schema ? 'outline' : 'secondary'}
            >
              {item?.schema ?? ''}
            </Badge>
          </div>
          <div className="flex flex-row gap-2 items-center">
            <span className="font-semibold tracking-tight">Table</span>
            <Badge
              className="px-4 py-2 dark:border-gray-700"
              variant={item?.table ? 'outline' : 'secondary'}
            >
              {item?.table ?? ''}
            </Badge>
          </div>
          <div className="flex flex-row items-center">
            <ValidateQueryBadge resp={validateResp} />
          </div>
        </div>
        <div className="flex flex-row gap-4">
          <TooltipProvider>
            <Tooltip delayDuration={200}>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="secondary"
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
                  Attempts to run a SQL COUNT(*) statement against the source
                  connection for the table with the included WHERE clause
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
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="secondary"
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
          <Button
            type="button"
            variant="secondary"
            disabled={!item}
            onClick={() => onCancelClick()}
          >
            <ButtonText text="Cancel" />
          </Button>
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
      <div>
        <div className="flex flex-col items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
          {!item ? (
            <div className="h-[60px] w-full text-gray-400 dark:text-gray-600 text-sm justify-center flex">
              Click the edit button on the table that you want to subset and add
              a table filter here. For example, country = &apos;US&apos;
            </div>
          ) : (
            <Editor
              height="60px"
              width="100%"
              language="sql"
              value={constructWhere(item?.where ?? '')}
              theme={resolvedTheme === 'dark' ? 'vs-dark' : 'light'}
              onChange={(e) => onWhereChange(e?.replace('WHERE ', '') ?? '')}
              options={editorOptions}
            />
          )}
        </div>
      </div>
      <ValidateQueryErrorAlert
        validateResp={validateResp}
        rowCountError={rowCountError}
      />
    </div>
  );
}

async function validateSql(
  accountId: string,
  connectionId: string,
  query: string
): Promise<CheckSqlQueryResponse> {
  const queryParams = new URLSearchParams({
    query,
  });
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}/check-query?${queryParams.toString()}`,
    {
      method: 'GET',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CheckSqlQueryResponse.fromJson(await res.json());
}

async function getTableRowCount(
  accountId: string,
  connectionId: string,
  schema: string,
  table: string,
  where?: string
): Promise<GetTableRowCountResponse> {
  const queryParams = new URLSearchParams({
    schema,
    table,
  });
  if (where) {
    queryParams.set('where', where);
  }
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}/get-table-row-count?${queryParams.toString()}`,
    {
      method: 'GET',
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetTableRowCountResponse.fromJson(await res.json());
}

function shouldTriggerAutocomplete(text: string): boolean {
  const lastSignificantWord = text
    .replace(/\s+/g, ' ')
    .trim()
    .split(' ')
    .pop()
    ?.toUpperCase();
  const triggerKeywords = ['SELECT', 'WHERE', 'AND', 'OR', 'FROM'];
  return triggerKeywords.includes(lastSignificantWord || '');
}
