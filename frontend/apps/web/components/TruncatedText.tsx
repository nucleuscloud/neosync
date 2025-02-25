import { cn } from '@/libs/utils';
import { TooltipContentProps } from '@radix-ui/react-tooltip';
import { ReactElement } from 'react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from './ui/tooltip';
interface Props {
  text: string;
  side?: TooltipContentProps['side'];
  align?: TooltipContentProps['align'];
  truncatedContainerClassName?: string;
  hoveredContainerClassName?: string;
  maxWidth?: number;
  delayDuration?: number;
}

export default function TruncatedText(props: Props): ReactElement<any> {
  const {
    text,
    truncatedContainerClassName,
    hoveredContainerClassName,
    delayDuration = 100,
    side = 'top',
    maxWidth = 200,
    align,
  } = props;
  return (
    <TooltipProvider>
      <Tooltip delayDuration={delayDuration}>
        <TooltipTrigger asChild>
          <div style={{ maxWidth: `${maxWidth}px` }} className="relative">
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
        <TooltipContent side={side} align={align}>
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
