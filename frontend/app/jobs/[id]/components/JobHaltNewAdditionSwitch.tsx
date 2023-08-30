import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/components/ui/use-toast';
import { ReactElement, useState } from 'react';

interface Props {
  isHalted: boolean;
  mutate: () => {};
}

export default function JobHaltNewAdditionSwitch(props: Props): ReactElement {
  const { mutate } = props;
  const { toast } = useToast();
  const [halt, setHalt] = useState(props.isHalted);
  async function changeSwitch(value: boolean): Promise<void> {
    try {
      await updateJobHaltSwitch(value);
      setHalt(value);
      toast({
        description: 'Halt Job on new column addition updated!',
      });
      mutate();
    } catch (e) {
      toast({
        variant: 'destructive',
        description: 'Failed to update halt job on new column addition',
      });
    }
  }
  return (
    <div className="w-96">
      <div className="flex flex-row items-center justify-between rounded-lg border p-4">
        <div className="space-y-0.5">
          <Label className="text-base">Halt Job on new column addition</Label>
          <p className="text-sm text-muted-foreground">
            Stops job runs if new column is detected
          </p>
        </div>
        <Switch checked={halt} onCheckedChange={changeSwitch} />
      </div>
    </div>
  );
}

async function updateJobHaltSwitch(value: boolean): Promise<void> {
  const res = await fetch(`/api/job/update-halt-on-new-addition`, {
    method: 'POST',
    body: JSON.stringify({
      haltOnNewColumnAddition: value,
    }),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  await res.json();
}
