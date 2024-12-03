import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { JobHook } from '@neosync/sdk';
import { deleteJobHook } from '@neosync/sdk/connectquery';
import { ClockIcon, Pencil1Icon, TrashIcon } from '@radix-ui/react-icons';
import { ReactElement } from 'react';
import { toast } from 'sonner';

interface Props {
  hook: JobHook;
  onDeleted(): void;
}

export default function HookCard(props: Props): ReactElement {
  const { hook, onDeleted } = props;
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
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  {hook.description}
                </p>
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
                  P{hook.priority}
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
                // onClick={() => onEditClicked()}
              >
                <Pencil1Icon />
              </Button>
              <RemoveHookButton hook={hook} onDeleted={onDeleted} />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

interface RemoveHookButtonProps {
  onDeleted(): void;
  hook: Pick<JobHook, 'id' | 'name'>;
}
function RemoveHookButton(props: RemoveHookButtonProps): ReactElement {
  const { hook, onDeleted } = props;
  const { mutateAsync: removeHook } = useMutation(deleteJobHook);

  async function onDelete(): Promise<void> {
    try {
      await removeHook({ id: hook.id });
      toast.success('Successfully removed job hook!');
      onDeleted();
    } catch (err) {
      console.error(err);
      toast.error('Unable to remove job hook', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <DeleteConfirmationDialog
      trigger={
        <Button variant="destructive" type="button">
          <TrashIcon />
        </Button>
      }
      headerText={`Are you sure you want to delete job hook: ${hook.name}?`}
      description="Deleting this hook is irreversable!"
      onConfirm={async () => onDelete()}
    />
  );
}
