import { ReactNode } from 'react';
import { Separator } from '../ui/separator';

interface Props {
  header: string;
  description?: string;

  extraHeading?: ReactNode;
  leftIcon?: ReactNode;
}

export default function PageHeader(props: Props) {
  const { header, description, extraHeading, leftIcon } = props;
  return (
    <div className="flex flex-col gap-2 py-10">
      <div className="flex flex-row justify-between">
        <div className="flex flex-row items-center gap-1">
          {leftIcon ? leftIcon : null}
          <h1 className="text-2xl font-bold tracking-tight">{header}</h1>
        </div>
        {extraHeading ? <div>{extraHeading}</div> : null}
      </div>
      {description ? (
        <h3 className="text-muted-foreground">{description}</h3>
      ) : null}
      <div className="my-4">
        <Separator />
      </div>
    </div>
  );
}
