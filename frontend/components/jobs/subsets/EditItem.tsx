import ButtonText from '@/components/ButtonText';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { CheckSqlQueryResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { getErrorMessage } from '@/util/util';
import { ReactElement, useState } from 'react';
import ValidateQueryBadge from './ValidateQueryBadge';
import ValidateQueryErrorAlert from './ValidateQueryErrorAlert';
import { TableRow } from './subset-table/column';

interface Props {
  item?: TableRow;
  onItem(item?: TableRow): void;
  onSave(): void;
  onCancel(): void;
  connectionId: string;
}
export default function EditItem(props: Props): ReactElement {
  const { item, onItem, onSave, onCancel, connectionId } = props;
  const [validateResp, setValidateResp] = useState<
    CheckSqlQueryResponse | undefined
  >();

  function onWhereChange(value: string): void {
    if (!item) {
      return;
    }
    onItem({ ...item, where: value });
  }

  async function onValidate(): Promise<void> {
    try {
      const resp = await validateSql(
        connectionId,
        `select * from ${item?.schema}.${item?.table} WHERE ${item?.where};`
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

  function onCancelClick(): void {
    setValidateResp(undefined);
    onCancel();
  }
  function onSaveClick(): void {
    setValidateResp(undefined);
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
  connectionId: string,
  query: string
): Promise<CheckSqlQueryResponse> {
  const queryParams = new URLSearchParams({
    query,
  });
  const res = await fetch(
    `/api/connections/${connectionId}/check-query?${queryParams.toString()}`,
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
