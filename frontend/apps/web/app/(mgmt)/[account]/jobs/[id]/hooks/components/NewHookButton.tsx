import ButtonText from '@/components/ButtonText';
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
import { Connection, JobService, NewJobHook } from '@neosync/sdk';
import { PlusIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { toast } from 'sonner';
import NewHookForm from './NewHookForm';

interface Props {
  jobId: string;
  jobConnections: Connection[];
  onCreated(): void;
}

export default function NewHookButton(props: Props): ReactElement<any> {
  const { jobId, jobConnections, onCreated } = props;
  const { mutateAsync: createHook } = useMutation(
    JobService.method.createJobHook
  );
  const [open, setOpen] = useState(false);

  async function onCreate(values: Partial<NewJobHook>): Promise<void> {
    try {
      await createHook({
        jobId,
        hook: {
          config: values.config,
          description: values.description,
          enabled: values.enabled,
          name: values.name,
          priority: values.priority,
        },
      });
      toast.success('Successfully created job hook!');
      onCreated();
      setOpen(false);
    } catch (err) {
      console.error(err);
      toast.error('Unable to create job hook', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button type="button">
          <ButtonText leftIcon={<PlusIcon />} text="New Hook" />
        </Button>
      </DialogTrigger>
      <DialogContent
        onPointerDownOutside={(e) => e.preventDefault()}
        className="lg:max-w-4xl max-h-[85vh] overflow-y-auto"
      >
        <DialogHeader>
          <DialogTitle>Create new Job Hook</DialogTitle>
          <DialogDescription>
            Configure values for a new job hook that will be invoked during a
            job run
          </DialogDescription>
        </DialogHeader>
        <NewHookForm jobConnections={jobConnections} onSubmit={onCreate} />
      </DialogContent>
    </Dialog>
  );
}
