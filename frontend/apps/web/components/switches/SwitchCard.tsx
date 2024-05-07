import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { ReactElement } from 'react';

interface Props {
  isChecked: boolean;
  onCheckedChange: (value: boolean) => void;
  title: string;
  // provide a component that will show up to the right of the title
  postTitle?: ReactElement;

  description?: string;
}

export default function SwitchCard(props: Props): ReactElement {
  const { isChecked, onCheckedChange, title, description, postTitle } = props;
  return (
    <div className="flex flex-row items-center justify-between rounded-lg border p-4 dark:border dark:border-gray-700 shadow-sm">
      <div className="space-y-0.5">
        <div className="flex flex-col md:flex-row gap-2 items-center">
          <Label className="text-sm">{title}</Label>
          {postTitle}
        </div>
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
      </div>
      <Switch checked={isChecked} onCheckedChange={onCheckedChange} />
    </div>
  );
}
