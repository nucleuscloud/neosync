import { Button } from '@/components/ui/button';
import { ReactElement } from 'react';

interface Props {
  onClick(): void;
}

export default function ExportJobMappingsButton(props: Props): ReactElement {
  const { onClick } = props;
  return (
    <div>
      <Button type="button" variant="outline" onClick={onClick}>
        Export
      </Button>
    </div>
  );
}
