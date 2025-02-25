import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ReactElement } from 'react';
import { handleDataTypeBadge } from '../SchemaTable/util';

interface Props {
  value: string;
}

export default function DataTypeCell(props: Props): ReactElement<any> {
  const { value } = props;
  return (
    <TooltipProvider>
      <Tooltip delayDuration={200}>
        <TooltipTrigger asChild>
          <div>
            <Badge variant="outline" className="max-w-[100px]">
              <span className="truncate block overflow-hidden">
                {handleDataTypeBadge(value)}
              </span>
            </Badge>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <p>{value}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
