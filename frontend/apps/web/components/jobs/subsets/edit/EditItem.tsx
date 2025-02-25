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
import {
  CheckSqlQueryResponse,
  CheckSqlQueryResponseSchema,
  ConnectionDataService,
  ConnectionService,
  GetTableRowCountResponse,
} from '@neosync/sdk';
import { ReactElement, useEffect, useMemo, useState } from 'react';
import ValidateQueryErrorAlert from '../SubsetErrorAlert';
import { SubsetTableRow } from '../SubsetTable/Columns';
import ValidateQueryBadge from '../ValidateQueryBadge';
import {
  isSubsetRowCountSupported,
  isSubsetValidationSupported,
  ValidSubsetConnectionType,
} from '../utils';
import WhereEditor from './WhereEditor';

interface Props {
  item?: SubsetTableRow;
  onItem(item?: SubsetTableRow): void;
  onSave(): void;
  onCancel(): void;
  connectionId: string;
  connectionType: ValidSubsetConnectionType;
  columns: string[];
}
export default function EditItem(props: Props): ReactElement<any> {
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
  const [rowCountError, setRowCountError] = useState<string>();

  const showRowCountButton = useMemo(
    () => isSubsetRowCountSupported(connectionType),
    [connectionType]
  );
  const showValidateButton = useMemo(
    () => isSubsetValidationSupported(connectionType),
    [connectionType]
  );

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
    if (
      connectionType === 'pgConfig' ||
      connectionType === 'mysqlConfig' ||
      connectionType === 'mssqlConfig'
    ) {
      let queryString = '';
      if (connectionType === 'pgConfig' || connectionType === 'mssqlConfig') {
        queryString = `select * from "${item?.schema}"."${item?.table}" WHERE ${item?.where};`;
      } else if (connectionType === 'mysqlConfig') {
        queryString = `select * from \`${item?.schema}\`.\`${item?.table}\` WHERE ${item?.where};`;
      }

      try {
        const resp = await validateSql({
          id: connectionId,
          query: queryString,
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
      <WhereEditor
        whereClause={item?.where ?? ''}
        onWhereChange={onWhereChange}
        columns={columns}
      />
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

function showSchema(connectionType: ValidSubsetConnectionType | null): boolean {
  return (
    connectionType === 'pgConfig' ||
    connectionType === 'mysqlConfig' ||
    connectionType === 'mssqlConfig'
  );
}
