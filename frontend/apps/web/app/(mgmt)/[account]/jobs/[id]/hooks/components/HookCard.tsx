import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { JobHook } from '@neosync/sdk';
import { ClockIcon, Pencil1Icon, TrashIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';

interface Props {
  hook: JobHook;
}

export default function HookCard(props: Props): ReactElement {
  const { hook } = props;
  return (
    <div id={`jobhook-${hook.id}`}>
      <Card
      // className={`${hook.enabled ? 'bg-white' : 'bg-gray-50'} transition-colors`}
      >
        <CardContent className="p-4">
          <div className="flex items-start justify-between">
            <div className="space-y-2 flex-grow">
              {/* Header Section */}
              <div className="flex items-center justify-between">
                <h3 className="font-medium text-lg">{hook.name}</h3>
                {/* <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant={hook.active ? 'default' : 'outline'}
                    onClick={() => onToggleActive(hook.id)}
                  >
                    {hook.active ? 'Active' : 'Inactive'}
                  </Button>
                </div> */}
              </div>

              {/* Description */}
              {hook.description && (
                <p className="text-sm text-gray-600">{hook.description}</p>
              )}

              {/* URL */}
              {/* <div className="text-sm text-gray-500 font-mono break-all">
                {hook.url}
              </div> */}

              {/* Metadata Row */}
              <div className="flex items-center gap-4 pt-2">
                {/* Timing */}
                <div className="flex items-center gap-1 text-sm text-gray-600">
                  <ClockIcon className="flex-shrink-0" />
                  {/* <code className="px-1 bg-gray-100 rounded">
                    {hook.timing}
                  </code> */}
                </div>

                {/* Priority Badge */}
                <Badge
                  variant="secondary"
                  // className={getPriorityStyles(hook.priority)}
                >
                  {hook.priority} priority
                </Badge>
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-start gap-1 ml-4">
              <Button
                type="button"
                size="sm"
                variant="ghost"
                className="text-gray-500 hover:text-gray-700"
                // onClick={() => onEdit(hook)}
              >
                <Pencil1Icon />
              </Button>
              <Button
                type="button"
                size="sm"
                variant="ghost"
                className="text-red-500 hover:text-red-700"
                // onClick={() => onDelete(hook.id)}
              >
                <TrashIcon />
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
