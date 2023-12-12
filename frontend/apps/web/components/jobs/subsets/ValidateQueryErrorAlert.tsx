import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { CheckSqlQueryResponse } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  resp?: CheckSqlQueryResponse;
}

export default function ValidateQueryErrorAlert(
  props: Props
): ReactElement | null {
  const { resp } = props;
  if (!resp || resp.isValid) {
    return null;
  }

  return (
    <div>
      <Alert variant="destructive">
        <AlertTitle>Invalid SQL Query</AlertTitle>
        <AlertDescription>
          {resp.erorrMessage ? resp.erorrMessage : 'unknown error message'}
        </AlertDescription>
      </Alert>
    </div>
  );
}
