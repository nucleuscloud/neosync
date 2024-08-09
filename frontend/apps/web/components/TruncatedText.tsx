import { cn } from '@/libs/utils';
import { ReactElement } from 'react';

interface Props {
  text: string;

  parentClassName?: string;
  truncatedContainerClassName?: string;
  hoveredContainerClassName?: string;
}

export default function TruncatedText(props: Props): ReactElement {
  const {
    text,
    parentClassName,
    truncatedContainerClassName,
    hoveredContainerClassName,
  } = props;
  return (
    <div className={cn('group relative max-w-[200px]', parentClassName)}>
      <div
        className={cn(
          'truncate transition-opacity duration-300 group-hover:opacity-0 font-medium',
          truncatedContainerClassName
        )}
      >
        {text}
      </div>
      <div
        className={cn(
          'absolute top-0 left-0 w-full font-medium opacity-0 transition-opacity duration-300 group-hover:opacity-100 pointer-events-none group-hover:pointer-events-auto',
          hoveredContainerClassName
        )}
      >
        {text}
      </div>
    </div>
  );
}
