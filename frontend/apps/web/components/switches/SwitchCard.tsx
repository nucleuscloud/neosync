import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { cn } from '@/libs/utils';
import { ReactElement } from 'react';

interface Props {
  isChecked: boolean;
  onCheckedChange: (value: boolean) => void;
  title: string;
  // provide a component that will show up to the right of the title
  postTitle?: ReactElement;
  description?: string;
  containerClassName?: string;
  titleClassName?: string;
}

export default function SwitchCard({
  isChecked,
  onCheckedChange,
  title,
  description,
  postTitle,
  containerClassName,
  titleClassName,
}: Props): ReactElement {
  return (
    <div
      className={cn(
        'flex flex-row items-center justify-between rounded-lg border p-4 shadow-sm',
        'dark:border-gray-700',
        containerClassName
      )}
    >
      <div className="space-y-0.5">
        <div className="flex flex-col md:flex-row gap-2 items-center">
          <Label className={cn('text-sm font-medium', titleClassName)}>
            {title}
          </Label>
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
