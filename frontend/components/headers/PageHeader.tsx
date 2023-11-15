import { cn } from '@/libs/utils';
import { ReactNode } from 'react';
import { Separator } from '../ui/separator';

interface Props {
  header: string;
  description?: string;

  extraHeading?: ReactNode;
  leftIcon?: ReactNode;

  pageHeaderContainerClassName?: string;
}

export default function PageHeader(props: Props) {
  const {
    header,
    description,
    extraHeading,
    leftIcon,
    pageHeaderContainerClassName,
  } = props;
  return (
    <div
      className={cn(
        'page-header-container flex flex-col gap-2',
        pageHeaderContainerClassName
      )}
    >
      <div className="flex flex-row justify-between">
        <div className="flex flex-row items-center gap-1">
          {leftIcon ? leftIcon : null}
          <h1 className="text-2xl font-bold tracking-tight">{header}</h1>
        </div>
        {extraHeading ? <div>{extraHeading}</div> : null}
      </div>
      {description ? (
        <h3 className="text-muted-foreground text-sm">{description}</h3>
      ) : null}

      <div>
        <Separator />
      </div>
    </div>
  );
}
