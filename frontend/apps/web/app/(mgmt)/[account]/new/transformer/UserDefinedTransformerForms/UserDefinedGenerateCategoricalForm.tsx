'use client';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { GenerateCategorical } from '@neosync/sdk';
import { ReactElement } from 'react';

interface Props {
  value: GenerateCategorical;
  setValue(value: GenerateCategorical): void;
  isDisabled?: boolean;
}

export default function UserDefinedGenerateCategoricalForm(
  props: Props
): ReactElement {
  const { value, setValue, isDisabled } = props;

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="space-y-0.5">
        <FormLabel>Categories</FormLabel>
        <FormDescription>
          Provide a list of comma-separated string values that you want to
          randomly select from.
        </FormDescription>
      </div>
      <div className="flex flex-col items-start w-[300px]">
        <Input
          value={value.categories}
          type="string"
          onChange={(e) =>
            setValue(
              new GenerateCategorical({ ...value, categories: e.target.value })
            )
          }
          disabled={isDisabled}
        />
      </div>
    </div>
  );
}
