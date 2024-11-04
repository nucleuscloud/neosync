import { FormDescription, FormLabel } from '@/components/ui/form';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { NewColumnAdditionStrategy } from '@/yup-validations/jobs';
import { ReactElement } from 'react';

interface Props {
  value: NewColumnAdditionStrategy;
  setValue(strategy: NewColumnAdditionStrategy): void;
}

export default function NewColumnAdditionStrategyOptionsForm(
  props: Props
): ReactElement {
  const { value, setValue } = props;

  return (
    <div className="flex flex-col gap-2">
      <FormLabel>New Column Addition Strategy</FormLabel>
      <FormDescription>
        Determine what happens when a new column is detected during a job run.
      </FormDescription>
      <RadioGroup
        onValueChange={(newval) => {
          setValue(newval as NewColumnAdditionStrategy);
        }}
        value={value}
      >
        <StrategyRadioItem
          value="halt"
          label="Halt - Stop the run if a new column is detected"
        />
        <StrategyRadioItem
          value="automap"
          label="AutoMap - Automatically generate a new value"
        />
        <StrategyRadioItem
          value="continue"
          label="Continue - Proceed without doing anything"
        />
      </RadioGroup>
    </div>
  );
}

interface StrategyRadioItemProps {
  value: NewColumnAdditionStrategy;
  label: string;
}

function StrategyRadioItem(props: StrategyRadioItemProps): ReactElement {
  const { value, label } = props;
  return (
    <div className="flex items-center gap-2">
      <RadioGroupItem value={value} id={value} />
      <Label htmlFor={value}>{label}</Label>
    </div>
  );
}
