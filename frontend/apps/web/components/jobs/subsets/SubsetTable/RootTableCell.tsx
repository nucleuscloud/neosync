import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ReactElement } from 'react';

interface Props {
  isRootTable: boolean;
}
export default function RootTableCell(props: Props): ReactElement {
  const { isRootTable } = props;
  return (
    <div className="flex justify-center pr-4">
      {isRootTable && (
        <TooltipProvider>
          <Tooltip delayDuration={200}>
            <TooltipTrigger asChild>
              <div>
                <Badge variant="outline">Root</Badge>
              </div>
            </TooltipTrigger>
            <TooltipContent className="max-w-xs px-2 text-center mx-auto">
              This is a Root table that only has foreign key references to
              children tables. Subsetting this table will subset all of
              it&apos;s children tables.
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      )}
    </div>
  );
}
