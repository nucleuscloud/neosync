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
import { getGenerateIpAddressVersionString } from '@/util/util';
import { create } from '@bufbuild/protobuf';
import {
  GenerateIpAddress,
  GenerateIpAddressSchema,
  GenerateIpAddressType,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateIpAddress> {}

export default function GenerateIpAddressForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>IP Version</FormLabel>
          <FormDescription>
            Select if you want to generate an IPv4 or IPv6 address.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <Select
              disabled={isDisabled}
              onValueChange={(newValue) => {
                setValue(
                  create(GenerateIpAddressSchema, {
                    ...value,
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    ipType: parseInt(newValue, 10),
                  })
                );
              }}
              value={value.ipType?.toString()}
            >
              <SelectTrigger className="w-[300px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[
                  GenerateIpAddressType.V4_LINK_LOCAL,
                  GenerateIpAddressType.V4_LOOPBACK,
                  GenerateIpAddressType.V4_MULTICAST,
                  GenerateIpAddressType.V4_PRIVATE_A,
                  GenerateIpAddressType.V4_PRIVATE_B,
                  GenerateIpAddressType.V4_PRIVATE_C,
                  GenerateIpAddressType.V4_PUBLIC,
                  GenerateIpAddressType.V6,
                ].map((version) => (
                  <SelectItem
                    key={version}
                    className="cursor-pointer"
                    value={version.toString()}
                  >
                    {getGenerateIpAddressVersionString(version)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <FormErrorMessage message={errors?.ipType?.message} />
        </div>
      </div>
    </div>
  );
}
