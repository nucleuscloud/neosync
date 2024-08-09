import { cn } from '@/libs/utils';
import { ReactElement } from 'react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from './ui/tooltip';

interface Props {
  text: string;

  truncatedContainerClassName?: string;
  hoveredContainerClassName?: string;

  delayDuration?: number;
}

export default function TruncatedText(props: Props): ReactElement {
  const {
    text,
    truncatedContainerClassName,
    hoveredContainerClassName,
    delayDuration = 100,
  } = props;
  return (
    <TooltipProvider>
      <Tooltip delayDuration={delayDuration}>
        <TooltipTrigger asChild>
          <div className={cn('relative max-w-[200px]')}>
            <div
              className={cn(
                'truncate font-medium',
                truncatedContainerClassName
              )}
            >
              {text}
            </div>
          </div>
        </TooltipTrigger>
        <TooltipContent>
          <div
            className={cn('relative font-medium', hoveredContainerClassName)}
          >
            {text}
          </div>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
