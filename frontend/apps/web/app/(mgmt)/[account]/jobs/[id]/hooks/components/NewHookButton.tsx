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
import { PlainMessage } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import { NewJobHook } from '@neosync/sdk';
import { createJobHook } from '@neosync/sdk/connectquery';
import { ReactElement, useState } from 'react';
import { toast } from 'sonner';
import NewHookForm from './NewHookForm';

interface Props {
  jobId: string;
  jobConnectionIds: string[];
  onCreated(): void;
}

export default function NewHookButton(props: Props): ReactElement {
  const { jobId, jobConnectionIds, onCreated } = props;
  const { mutateAsync: createHook } = useMutation(createJobHook);
  const [open, setOpen] = useState(false);

  async function onCreate(
    values: Partial<PlainMessage<NewJobHook>>
  ): Promise<void> {
    try {
      await createHook({
        jobId,
        hook: values,
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
        <Button type="button">New Hook</Button>
      </DialogTrigger>
      <DialogContent
        onPointerDownOutside={(e) => e.preventDefault()}
        className="lg:max-w-4xl"
      >
        <DialogHeader>
          <DialogTitle>Create new Job Hook</DialogTitle>
          <DialogDescription>
            Configure values for a new job hook that will be invoked during a
            job run
          </DialogDescription>
        </DialogHeader>
        <NewHookForm jobConnectionIds={jobConnectionIds} onSubmit={onCreate} />
      </DialogContent>
    </Dialog>
  );
}
