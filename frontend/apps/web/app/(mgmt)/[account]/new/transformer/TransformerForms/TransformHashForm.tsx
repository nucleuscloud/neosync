'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { getTransformHashTypeString } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import {
  TransformHash,
  TransformHash_HashType,
  TransformHashSchema,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<TransformHash> {}

export default function TransformHashForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
      <div className="space-y-0.5 w-[80%]">
        <FormLabel>Email Type</FormLabel>
        <FormDescription>
          Select the algorithm you want to use to hash the data.
        </FormDescription>
      </div>
      <div className="flex flex-col">
        <div className="justify-end flex">
          <Select
            disabled={isDisabled}
            onValueChange={(newValue) => {
              setValue(
                create(TransformHashSchema, {
                  ...value,
                  // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                  algo: parseInt(newValue, 10),
                })
              );
            }}
            value={value.algo?.toString()}
          >
            <SelectTrigger className="w-[300px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {[
                TransformHash_HashType.SHA256,
                TransformHash_HashType.SHA512,
                TransformHash_HashType.SHA1,
                TransformHash_HashType.MD5,
              ].map((algo) => (
                <SelectItem
                  key={algo}
                  className="cursor-pointer"
                  value={algo.toString()}
                >
                  {getTransformHashTypeString(algo)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <FormErrorMessage message={errors?.algo?.message} />
      </div>
    </div>
  );
}
