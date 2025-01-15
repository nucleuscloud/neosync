import { FormDescription, FormLabel } from '@/components/ui/form';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { ColumnRemovalStrategy } from '@/yup-validations/jobs';
import { ExternalLinkIcon } from '@radix-ui/react-icons';
import Link from 'next/link';
import { ReactElement } from 'react';

interface Props {
  value: ColumnRemovalStrategy;
  setValue(strategy: ColumnRemovalStrategy): void;
}

export default function ColumnRemovalStrategyOptionsForm(
  props: Props
): ReactElement {
  const { value, setValue } = props;

  return (
    <div className="flex flex-col gap-2">
      <FormLabel>Column Removal Strategy</FormLabel>
      <div className="flex flex-row gap-1">
        <FormDescription>
          Choose how to handle situations when columns that were previously
          mapped no longer exist in the source database schema.
        </FormDescription>
        <Link
          href="https://docs.neosync.dev/guides/column-removal-strategies"
          target="_blank"
          className="hover:underline inline-flex gap-1 flex-row items-center tracking-tight text-[0.8rem] text-muted-foreground"
        >
          Learn More
          <ExternalLinkIcon className="w-3 h-3" />
        </Link>
      </div>
      <RadioGroup
        onValueChange={(newval) => {
          setValue(newval as ColumnRemovalStrategy);
        }}
        value={value}
      >
        <StrategyRadioItem
          value="halt"
          label="Halt - Stop the job if a column in your mapping is missing from the source database"
        />
        <StrategyRadioItem
          value="continue"
          label="Continue - Proceed with the job and ignore any mapped columns that are no longer present in the source database"
        />
      </RadioGroup>
    </div>
  );
}

interface StrategyRadioItemProps {
  value: ColumnRemovalStrategy;
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
