import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { JobHook, JobHookConfig } from '@neosync/sdk';
import { ClockIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import EditHookButton from './EditHookButton';
import RemoveHookButton from './RemoveHookButton';

interface Props {
  hook: JobHook;
  jobConnectionIds: string[];
  onDeleted(): void;
  onEdited(): void;
}

export default function HookCard(props: Props): ReactElement {
  const { hook, onDeleted, onEdited, jobConnectionIds } = props;
  const hookTiming = getHookTiming(hook.config ?? new JobHookConfig());
  return (
    <div id={`jobhook-${hook.id}`}>
      <Card>
        <CardContent className="p-4">
          <div className="flex justify-between">
            <div className="space-y-2 flex-grow">
              {/* Header Section */}
              <div className="flex items-center justify-between">
                <h3 className="font-medium text-lg">{hook.name}</h3>
              </div>

              <p className="text-sm text-gray-600 dark:text-gray-400">
                {hook.description}
              </p>

              <div className="flex items-center gap-4 pt-2">
                <div className="flex items-center gap-1 text-sm text-gray-600 dark:text-gray-400">
                  <ClockIcon className="flex-shrink-0" />
                  <p>{hookTiming}</p>
                </div>

                <Badge variant="secondary">
                  <span>P</span>
                  {hook.priority}
                </Badge>
                <Badge variant={hook.enabled ? 'success' : 'outline'}>
                  {hook.enabled ? 'Enabled' : 'Disabled'}
                </Badge>
                {hook.config?.config.case && (
                  <Badge variant="secondary">{hook.config?.config.case}</Badge>
                )}
              </div>
            </div>

            <div className="flex items-center gap-2">
              <EditHookButton
                hook={hook}
                onEdited={onEdited}
                jobConnectionIds={jobConnectionIds}
              />
              <RemoveHookButton hook={hook} onDeleted={onDeleted} />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function getHookTiming(config: JobHookConfig): string | undefined {
  switch (config.config.case) {
    case 'sql': {
      return config.config.value.timing?.timing.case;
    }
  }
  return undefined;
}
