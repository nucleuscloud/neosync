import { Badge } from '@/components/ui/badge';
import { GetTableRowCountResponse } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  resp?: GetTableRowCountResponse;
}

export default function TableRowCountBadge(props: Props): ReactElement | null {
  const { resp } = props;
  if (!resp) {
    return null;
  }
  console.log(resp.count);
  return (
    <Badge
      // variant={resp.isValid ? 'success' : 'destructive'}
      className="cursor-default px-4 py-2"
    >
      {19}
    </Badge>
  );
}
