import { ReactElement } from 'react';
import { Badge } from './ui/badge';

interface Props {
  tagValue: string;
}

export default function GradientTag(props: Props): ReactElement {
  const { tagValue } = props;

  return (
    <div className="w-full">
      <Badge variant="gradient">{tagValue}</Badge>
    </div>
  );
}
