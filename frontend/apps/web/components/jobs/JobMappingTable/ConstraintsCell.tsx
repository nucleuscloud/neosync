import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/libs/utils';
import { ReactElement } from 'react';

interface Props {
  isPrimaryKey: boolean;
  foreignKey: [boolean, string[]];
  virtualForeignKey: [boolean, string[]];
  isUnique: boolean;
}

export default function ConstraintsCell(props: Props): ReactElement<any> {
  const { isPrimaryKey, foreignKey, virtualForeignKey, isUnique } = props;
  const [isForeignKey, fkCols] = foreignKey;
  const [isVirtualForeignKey, vfkCols] = virtualForeignKey;
  return (
    <span className="max-w-[500px] truncate font-medium">
      <div className="flex flex-col lg:flex-row items-start gap-1">
        {isPrimaryKey && (
          <div>
            <HoverBadge
              badgeText="P"
              tooltipContent="Primary Key"
              badgeClassName="bg-blue-100 dark:bg-blue-200"
            />
          </div>
        )}
        {isForeignKey && (
          <div>
            <HoverBadge
              badgeText="FK"
              tooltipContent={`FK: ${fkCols
                .map((col) => `Primary Key: ${col}`)
                .join('\n')}`}
              badgeClassName="bg-orange-100 dark:bg-orange-300"
            />
          </div>
        )}
        {isVirtualForeignKey && (
          <div>
            <HoverBadge
              badgeText="VFK"
              tooltipContent={`VFK: ${vfkCols
                .map((col) => `Primary Key: ${col}`)
                .join('\n')}`}
              badgeClassName="bg-orange-100 dark:bg-orange-300"
            />
          </div>
        )}
        {isUnique && (
          <div>
            <HoverBadge
              badgeText="U"
              tooltipContent="Unique Key"
              badgeClassName="bg-red-100 dark:bg-red-300"
            />
          </div>
        )}
      </div>
    </span>
  );
}

interface HoverBadgeProps {
  badgeText: string;
  tooltipContent: string;
  badgeClassName?: string;
}

function HoverBadge(props: HoverBadgeProps): ReactElement<any> {
  const { badgeText, tooltipContent, badgeClassName } = props;
  return (
    <TooltipProvider>
      <Tooltip delayDuration={200}>
        <TooltipTrigger>
          <Badge
            variant="outline"
            className={cn(
              'text-xs cursor-default text-gray-800 dark:text-gray-900',
              badgeClassName
            )}
          >
            {badgeText}
          </Badge>
        </TooltipTrigger>
        <TooltipContent>{tooltipContent}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
