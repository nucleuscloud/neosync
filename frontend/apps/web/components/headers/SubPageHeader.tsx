import { cn } from '@/libs/utils';
import { ReactNode } from 'react';
import { Separator } from '../ui/separator';

interface Props {
  header: string;
  rightHeaderIcon?: ReactNode;
  description: string;
  extraHeading?: ReactNode;
  subHeadings?: ReactNode | ReactNode[];
}

export default function SubPageHeader(props: Props) {
  const {
    header,
    rightHeaderIcon,
    description,
    extraHeading,
    subHeadings: subHeadingsOrSingle,
  } = props;
  const subHeadings = Array.isArray(subHeadingsOrSingle)
    ? subHeadingsOrSingle
    : [subHeadingsOrSingle];
  return (
    <div className={cn('page-subheader-container flex flex-col gap-2')}>
      <div className="flex flex-col md:flex-row flex-wrap justify-between gap-4 md:gap-0">
        <div className="flex flex-col gap-0.5">
          <div className="flex flex-row gap-2 items-center">
            <h2 className="text-xl font-semibold tracking-tight">{header}</h2>
            {rightHeaderIcon ? rightHeaderIcon : null}
          </div>
          <p className="text-muted-foreground">{description}</p>
        </div>
        {extraHeading ? <div>{extraHeading}</div> : null}
      </div>
      {subHeadings.map((subheading, ind) => (
        <div key={ind} className="text-sm">
          {subheading}
        </div>
      ))}
      <Separator className="dark:bg-gray-700" />
    </div>
  );
}
