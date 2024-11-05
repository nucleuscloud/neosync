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
import {
  getGenerateIpAddressClassString,
  getGenerateIpAddressVersionString,
} from '@/util/util';
import { PlainMessage } from '@bufbuild/protobuf';
import {
  GenerateIpAddress,
  GenerateIpAddressClass,
  GenerateIpAddressVersion,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props
  extends TransformerConfigProps<
    GenerateIpAddress,
    PlainMessage<GenerateIpAddress>
  > {}

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
                  new GenerateIpAddress({
                    ...value,
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    version: parseInt(newValue, 10),
                  })
                );
              }}
              value={value.version?.toString()}
            >
              <SelectTrigger className="w-[300px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[GenerateIpAddressVersion.V4, GenerateIpAddressVersion.V6].map(
                  (version) => (
                    <SelectItem
                      key={version}
                      className="cursor-pointer"
                      value={version.toString()}
                    >
                      {getGenerateIpAddressVersionString(version)}
                    </SelectItem>
                  )
                )}
              </SelectContent>
            </Select>
          </div>
          <FormErrorMessage message={errors?.version?.message} />
        </div>
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>IPV4 Class</FormLabel>
          <FormDescription>
            Select the class of the IPV4 Address you want to generate. Defaults
            to public.
          </FormDescription>
        </div>
        <div>
          <div className="flex flex-col">
            <div className="justify-end flex">
              <Select
                disabled={
                  isDisabled || value.version == GenerateIpAddressVersion.V6
                }
                onValueChange={(newValue) => {
                  setValue(
                    new GenerateIpAddress({
                      ...value,
                      // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                      class: parseInt(newValue, 10),
                    })
                  );
                }}
                value={value.class?.toString()}
              >
                <SelectTrigger className="w-[300px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {[
                    GenerateIpAddressClass.PUBLIC,
                    GenerateIpAddressClass.LINK_LOCAL,
                    GenerateIpAddressClass.LOOPBACK,
                    GenerateIpAddressClass.MULTICAST,
                    GenerateIpAddressClass.PRIVATE_A,
                    GenerateIpAddressClass.PRIVATE_B,
                    GenerateIpAddressClass.PRIVATE_C,
                  ].map((ipClass) => (
                    <SelectItem
                      key={ipClass}
                      className="cursor-pointer"
                      value={ipClass.toString()}
                    >
                      {getGenerateIpAddressClassString(ipClass)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <FormErrorMessage message={errors?.class?.message} />
          </div>
        </div>
      </div>
    </div>
  );
}
