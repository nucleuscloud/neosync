import { Separator } from '../ui/separator';

interface Props {
  header: string;
  description: string;
}

export default function SubPageHeader(props: Props) {
  const { header, description } = props;
  return (
    <div>
      <div className="space-y-0.5">
        <h2 className="text-xl font-semibold tracking-tight">{header}</h2>
        <p className="text-muted-foreground">{description}</p>
      </div>
      <Separator className="my-6" />
    </div>
  );
}
