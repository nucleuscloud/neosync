'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { create } from '@bufbuild/protobuf';
import { GenerateCategorical, GenerateCategoricalSchema } from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateCategorical> {}

export default function GenerateCategoricalForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4 rounded-lg border dark:border-gray-700 p-3">
      <div className="space-y-0.5">
        <FormLabel>Categories</FormLabel>
        <FormDescription>
          Provide a list of comma-separated string values that you want to
          randomly select from.
        </FormDescription>
      </div>
      <div className="flex flex-col items-start">
        <Input
          value={value.categories}
          type="string"
          onChange={(e) =>
            setValue(
              create(GenerateCategoricalSchema, {
                ...value,
                categories: e.target.value,
              })
            )
          }
          disabled={isDisabled}
        />
        <FormErrorMessage message={errors?.categories?.message} />
      </div>
    </div>
  );
}
