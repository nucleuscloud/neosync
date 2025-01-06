import { cn } from '@/libs/utils';
import { ReactElement, ReactNode } from 'react';
import { Badge } from '../ui/badge';
import { Separator } from '../ui/separator';

interface Props {
  header: string;
  leftBadgeValue?: string;
  extraHeading?: ReactNode;
  leftIcon?: ReactNode;
  progressSteps?: ReactElement;
  pageHeaderContainerClassName?: string;
  subHeadings?: ReactNode | ReactNode[];
}

export default function PageHeader(props: Props) {
  const {
    header,
    extraHeading,
    leftIcon,
    pageHeaderContainerClassName,
    leftBadgeValue,
    progressSteps,
    subHeadings: subHeadingsOrSingle,
  } = props;
  const subHeadings = Array.isArray(subHeadingsOrSingle)
    ? subHeadingsOrSingle
    : [subHeadingsOrSingle];
  return (
    <div
      className={cn(
        'page-header-container flex flex-col gap-2',
        pageHeaderContainerClassName
      )}
    >
      <div className="flex flex-col xl:flex-row xl:items-center items-start justify-between gap-2">
        <div className="flex flex-col md:flex-row items-center gap-3">
          {leftIcon ? leftIcon : null}
          <div className="flex flex-row items-center gap-4">
            <h1 className="text-2xl font-bold tracking-tight truncate max-w-[300px] lg:max-w-[600px]">
              {header}
            </h1>
            {leftBadgeValue && (
              <Badge variant="outline" className="text-nowrap">
                {leftBadgeValue}
              </Badge>
            )}
          </div>
        </div>
        {progressSteps && (
          <div className="flex-1 flex justify-center">{progressSteps}</div>
        )}
        {extraHeading ? <div>{extraHeading}</div> : null}
      </div>
      {subHeadings.map((subheading, ind) => (
        <div key={ind} className="text-sm">
          {subheading}
        </div>
      ))}
      <Separator className="dark:bg-gray-600" />
    </div>
  );
}
