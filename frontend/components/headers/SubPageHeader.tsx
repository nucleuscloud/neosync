import { ReactNode } from 'react';
import { Separator } from '../ui/separator';

interface Props {
  header: string;
  description: string;
  extraHeading?: ReactNode;
}

export default function SubPageHeader(props: Props) {
  const { header, description, extraHeading } = props;
  return (
    <div>
      <div className="flex flex-row flex-wrap justify-between space-y-4 md:space-y-0">
        <div className="space-y-0.5">
          <h2 className="text-xl font-semibold tracking-tight">{header}</h2>
          <p className="text-muted-foreground">{description}</p>
        </div>
        {extraHeading ? <div>{extraHeading}</div> : null}
      </div>
      <Separator className="my-6" />
    </div>
  );
}
