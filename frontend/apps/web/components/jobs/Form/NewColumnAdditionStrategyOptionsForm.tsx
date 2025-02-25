import { FormDescription, FormLabel } from '@/components/ui/form';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { NewColumnAdditionStrategy } from '@/yup-validations/jobs';
import { ExternalLinkIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

interface Props {
  value: NewColumnAdditionStrategy;
  setValue(strategy: NewColumnAdditionStrategy): void;
  disableAutoMap?: boolean;
}

export default function NewColumnAdditionStrategyOptionsForm(
  props: Props
): ReactElement {
  const { value, setValue, disableAutoMap } = props;

  return (
    <div className="flex flex-col gap-2">
      <FormLabel>New Column Addition Strategy</FormLabel>
      <div className="flex flex-row gap-1">
        <FormDescription>
          Determine what happens when a new column is detected during a job run.
        </FormDescription>
        <Link
          href="https://docs.neosync.dev/guides/new-column-addition-strategies"
          target="_blank"
          className="hover:underline inline-flex gap-1 flex-row items-center tracking-tight text-[0.8rem] text-muted-foreground"
        >
          Learn More
          <ExternalLinkIcon className="w-3 h-3" />
        </Link>
      </div>
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
        {!disableAutoMap && (
          <StrategyRadioItem
            value="automap"
            label="AutoMap - Automatically generate a new value"
          />
        )}
        <StrategyRadioItem
          value="continue"
          label="Continue - Ignores new columns; may fail if column doesn't have default"
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
