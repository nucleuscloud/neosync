import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { toast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { CheckSqlQueryResponse, GetTableRowCountResponse } from '@neosync/sdk';
import { ReactElement, useEffect, useState } from 'react';
import ValidateQueryBadge from './ValidateQueryBadge';
import ValidateQueryErrorAlert from './ValidateQueryErrorAlert';
import { TableRow } from './subset-table/column';

interface Props {
  item?: TableRow;
  onItem(item?: TableRow): void;
  onSave(): void;
  onCancel(): void;
  connectionId: string;
  dbType: string;
}
export default function EditItem(props: Props): ReactElement {
  const { item, onItem, onSave, onCancel, connectionId, dbType } = props;
  const [validateResp, setValidateResp] = useState<
    CheckSqlQueryResponse | undefined
  >();
  const [tableRowCountResp, setTableRowCountResp] = useState<
    GetTableRowCountResponse | undefined
  >();
  const [calculatingRowCount, setCalculatingRowCount] = useState(false);
  const { account } = useAccount();

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
    } catch (err) {
      setCalculatingRowCount(false);
      console.error(err);
      toast({
        title: 'Unable to get table row count.',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
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
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  variant="secondary"
                  disabled={!item}
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
                  disabled={!item || !item.where}
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
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  type="button"
                  disabled={!item}
                  onClick={() => onSaveClick()}
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
        <Textarea
          disabled={!item}
          placeholder={
            !!item
              ? 'Add a table filter here'
              : 'Click edit on a row above to change the where clause'
          }
          value={item?.where ?? ''}
          onChange={(e) => onWhereChange(e.currentTarget.value)}
        />
      </div>
      <div>
        <Textarea
          placeholder="Where clause preview"
          disabled={true}
          value={buildSelectQuery(item?.where)}
        />
      </div>
      <ValidateQueryErrorAlert resp={validateResp} />
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

function buildSelectQuery(whereClause?: string): string {
  if (!whereClause) {
    return '';
  }
  return `WHERE ${whereClause};`;
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
