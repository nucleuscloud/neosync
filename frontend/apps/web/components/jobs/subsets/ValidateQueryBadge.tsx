import { Badge } from '@/components/ui/badge';
import { CheckSqlQueryResponse } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  resp?: CheckSqlQueryResponse;
}

export default function ValidateQueryBadge(props: Props): ReactElement | null {
  const { resp } = props;
  if (!resp) {
    return null;
  }
  const text = resp.isValid ? 'VALID' : 'INVALID';
  return (
    <Badge
      variant={resp.isValid ? 'success' : 'destructive'}
      className="cursor-default px-4 py-2 h-6"
    >
      {text}
    </Badge>
  );
}
