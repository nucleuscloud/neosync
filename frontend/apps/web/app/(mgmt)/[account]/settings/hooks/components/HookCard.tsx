import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { AccountHook } from '@neosync/sdk';
import { ReactElement } from 'react';
import EditHookButton from './EditHookButton';
import RemoveHookButton from './RemoveHookButton';

interface Props {
  hook: AccountHook;
  onDeleted(): void;
  onEdited(): void;
}

export default function HookCard(props: Props): ReactElement<any> {
  const { hook, onDeleted, onEdited } = props;

  return (
    <div id={`accounthook-${hook.id}`}>
      <Card>
        <CardContent className="p-4">
          <div className="flex justify-between">
            <div className="flex grow flex-col gap-2">
              {/* Header Section */}
              <div className="flex items-center justify-between">
                <h3 className="font-medium text-lg">{hook.name}</h3>
              </div>

              <p className="text-sm text-gray-600 dark:text-gray-400">
                {hook.description}
              </p>

              <div className="flex items-center gap-4 pt-2 flex-wrap">
                <Badge variant={hook.enabled ? 'success' : 'outline'}>
                  {hook.enabled ? 'Enabled' : 'Disabled'}
                </Badge>
                {hook.config?.config.case && (
                  <Badge variant="secondary">{hook.config?.config.case}</Badge>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <EditHookButton hook={hook} onEdited={onEdited} />
              <RemoveHookButton hook={hook} onDeleted={onDeleted} />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
