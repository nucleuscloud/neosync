import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { CheckSqlQueryResponse } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  validateResp?: CheckSqlQueryResponse;
  rowCountError?: string;
}

export default function ValidateQueryErrorAlert(
  props: Props
): ReactElement<any> | null {
  const { validateResp, rowCountError } = props;

  const hasSqlError =
    validateResp && !validateResp.isValid && validateResp.erorrMessage;
  const hasRowCountError = rowCountError && rowCountError.trim() !== '';

  if (!hasSqlError && !hasRowCountError) {
    return null;
  }

  return (
    <div className="flex flex-col gap-2">
      {!validateResp?.isValid && (
        <Alert variant="destructive">
          <AlertTitle>Invalid SQL Query</AlertTitle>
          <AlertDescription>
            {validateResp?.erorrMessage
              ? validateResp?.erorrMessage
              : 'unknown error message'}
          </AlertDescription>
        </Alert>
      )}
      {rowCountError && (
        <Alert variant="destructive">
          <AlertTitle>Unable to get table row count</AlertTitle>
          <AlertDescription>
            {rowCountError ? rowCountError : 'unknown error message'}
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
