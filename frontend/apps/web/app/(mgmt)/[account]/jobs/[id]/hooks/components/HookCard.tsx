import DeleteConfirmationDialog from '@/components/DeleteConfirmationDialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { getErrorMessage } from '@/util/util';
import { PlainMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import { JobHook, JobHookConfig } from '@neosync/sdk';
import { deleteJobHook, updateJobHook } from '@neosync/sdk/connectquery';
import { ClockIcon, Pencil1Icon, TrashIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { toast } from 'sonner';
import { EditHookForm } from './EditHookForm';

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
          <div className="flex items-start justify-between">
            <div className="space-y-2 flex-grow">
              {/* Header Section */}
              <div className="flex items-center justify-between">
                <h3 className="font-medium text-lg">{hook.name}</h3>
              </div>

              {/* Description */}
              {hook.description && (
                <p className="text-sm text-gray-600 dark:text-gray-400">
                  {hook.description}
                </p>
              )}

              {/* Metadata Row */}
              <div className="flex items-center gap-4 pt-2">
                {/* Timing */}
                <div className="flex items-center gap-1 text-sm text-gray-600 dark:text-gray-400">
                  <ClockIcon className="flex-shrink-0" />
                  <p>{hookTiming}</p>
                </div>

                {/* Priority Badge */}
                <Badge variant="secondary">P{hook.priority}</Badge>

                <Badge variant={hook.enabled ? 'success' : 'outline'}>
                  {hook.enabled ? 'Enabled' : 'Disabled'}
                </Badge>
                {hook.config?.config.case && (
                  <Badge variant="secondary">{hook.config?.config.case}</Badge>
                )}
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-start gap-1 ml-4">
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

interface EditHookButtonProps {
  onEdited(): void;
  hook: JobHook;
  jobConnectionIds: string[];
}

function EditHookButton(props: EditHookButtonProps): ReactElement {
  const { hook, onEdited, jobConnectionIds } = props;
  const { mutateAsync: updateHook } = useMutation(updateJobHook);
  const [open, setOpen] = useState(false);

  async function onUpdate(
    values: Partial<PlainMessage<JobHook>>
  ): Promise<void> {
    try {
      await updateHook({ id: hook.id, ...values });
      toast.success('Successfully updated job hook!');
      onEdited();
      setOpen(false);
    } catch (err) {
      console.error(err);
      toast.error('Unable to update job hook', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" type="button">
          <Pencil1Icon />
        </Button>
      </DialogTrigger>
      <DialogContent
        onPointerDownOutside={(e) => e.preventDefault()}
        className="lg:max-w-4xl"
      >
        <DialogHeader>
          <DialogTitle>Edit Job Hook: {hook.name}</DialogTitle>
          <DialogDescription>
            Change any of the available job hook settings.
          </DialogDescription>
        </DialogHeader>
        <EditHookForm
          key={hook.id}
          hook={hook}
          onSubmit={onUpdate}
          jobConnectionIds={jobConnectionIds}
        />
      </DialogContent>
    </Dialog>
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
