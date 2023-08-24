import { ReactNode } from 'react';
import { Separator } from '../ui/separator';

interface Props {
  header: string;
  description: string;

  extraHeading?: ReactNode;
  leftIcon?: ReactNode;
}

export default function PageHeader(props: Props) {
  const { header, description, extraHeading, leftIcon } = props;
  return (
    <div className="flex flex-col my-4 gap-2">
      <div className="flex flex-row justify-between">
        <div className="flex flex-row items-center gap-1">
          {leftIcon ? leftIcon : null}
          <h1 className="text-2xl font-bold tracking-tight">{header}</h1>
        </div>
        {extraHeading ? <div>{extraHeading}</div> : null}
      </div>
      <h3>{description}</h3>
      <div className="my-4">
        <Separator />
      </div>
    </div>
  );
}
