import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { ReactElement } from 'react';

interface Props {
  isChecked: boolean;
  onCheckedChange: (value: boolean) => void;
  title: string;
  description?: string;
}

export default function SwitchCard(props: Props): ReactElement {
  const { isChecked, onCheckedChange, title, description } = props;
  return (
    <div className="flex flex-row items-center justify-between rounded-lg border p-4 dark:border dark:border-gray-700 shadow-sm">
      <div className="space-y-0.5">
        <Label className="text-sm">{title}</Label>
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
      </div>
      <Switch checked={isChecked} onCheckedChange={onCheckedChange} />
    </div>
  );
}
