import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { getErrorMessage } from '@/util/util';
import { useMutation } from '@connectrpc/connect-query';
import { Connection, JobHook, JobService } from '@neosync/sdk';
import { Pencil1Icon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { toast } from 'sonner';
import { EditHookForm } from './EditHookForm';

interface Props {
  onEdited(): void;
  hook: JobHook;
  jobConnections: Connection[];
}

export default function EditHookButton(props: Props): ReactElement {
  const { hook, onEdited, jobConnections } = props;
  const { mutateAsync: updateHook } = useMutation(
    JobService.method.updateJobHook
  );
  const [open, setOpen] = useState(false);

  async function onUpdate(values: Partial<JobHook>): Promise<void> {
    try {
      await updateHook({
        id: hook.id,
        config: values.config,
        description: values.description,
        enabled: values.enabled,
        name: values.name,
        priority: values.priority,
      });
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
        className="lg:max-w-4xl max-h-[85vh] overflow-y-auto"
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
          jobConnections={jobConnections}
        />
      </DialogContent>
    </Dialog>
  );
}
