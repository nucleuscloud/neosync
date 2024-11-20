import { Badge } from '@/components/ui/badge';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { ReactElement } from 'react';

interface Props {
  isPrimaryKey: boolean;
  foreignKey: [boolean, string[]];
  virtualForeignKey: [boolean, string[]];
  isUnique: boolean;
}

export default function ConstraintsCell(props: Props): ReactElement {
  const { isPrimaryKey, foreignKey, virtualForeignKey, isUnique } = props;
  const [isForeignKey, fkCols] = foreignKey;
  const [isVirtualForeignKey, vfkCols] = virtualForeignKey;
  return (
    <span className="max-w-[500px] truncate font-medium">
      <div className="flex flex-col lg:flex-row items-start gap-1">
        {isPrimaryKey && (
          <div>
            <Badge
              variant="outline"
              className="text-xs bg-blue-100 text-gray-800 cursor-default dark:bg-blue-200 dark:text-gray-900"
            >
              Primary Key
            </Badge>
          </div>
        )}
        {isForeignKey && (
          <div>
            <TooltipProvider>
              <Tooltip delayDuration={200}>
                <TooltipTrigger>
                  <Badge
                    variant="outline"
                    className="text-xs bg-orange-100 text-gray-800 cursor-default dark:dark:text-gray-900 dark:bg-orange-300"
                  >
                    Foreign Key
                  </Badge>
                </TooltipTrigger>
                <TooltipContent>
                  {fkCols.map((col) => `Primary Key: ${col}`).join('\n')}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        )}
        {isVirtualForeignKey && (
          <div>
            <TooltipProvider>
              <Tooltip delayDuration={200}>
                <TooltipTrigger>
                  <Badge
                    variant="outline"
                    className="text-xs bg-orange-100 text-gray-800 cursor-default dark:dark:text-gray-900 dark:bg-orange-300"
                  >
                    Virtual Foreign Key
                  </Badge>
                </TooltipTrigger>
                <TooltipContent>
                  {vfkCols.map((col) => `Primary Key: ${col}`).join('\n')}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        )}
        {isUnique && (
          <div>
            <Badge
              variant="outline"
              className="text-xs bg-purple-100 text-gray-800 cursor-default dark:bg-purple-300 dark:text-gray-900"
            >
              Unique
            </Badge>
          </div>
        )}
      </div>
    </span>
  );
}
