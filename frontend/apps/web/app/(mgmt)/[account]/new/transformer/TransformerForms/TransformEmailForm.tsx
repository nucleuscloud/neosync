'use client';
import FormErrorMessage from '@/components/FormErrorMessage';
import { FormDescription, FormLabel } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import {
  getGenerateEmailTypeString,
  getInvalidEmailActionString,
} from '@/util/util';
import { create } from '@bufbuild/protobuf';
import {
  GenerateEmailType,
  InvalidEmailAction,
  TransformEmail,
  TransformEmailSchema,
} from '@neosync/sdk';
import { ReactElement } from 'react';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<TransformEmail> {}

export default function TransformEmailForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  return (
    <div className="flex flex-col w-full space-y-4">
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Preserve Length</FormLabel>
          <FormDescription>
            Set the length of the output email to be the same as the input
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveLength}
          onCheckedChange={(checked) =>
            setValue(
              create(TransformEmailSchema, {
                ...value,
                preserveLength: checked,
              })
            )
          }
          disabled={isDisabled}
        />
        <FormErrorMessage message={errors?.preserveLength?.message} />
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Preserve Domain</FormLabel>
          <FormDescription>
            Preserve the input domain including top level domain to the output
            value. For ex. if the input is john@gmail.com, the output will be
            ij23o@gmail.com
          </FormDescription>
        </div>
        <Switch
          checked={value.preserveDomain}
          onCheckedChange={(checked) =>
            setValue(
              create(TransformEmailSchema, {
                ...value,
                preserveDomain: checked,
              })
            )
          }
          disabled={isDisabled}
        />
        <FormErrorMessage message={errors?.preserveDomain?.message} />
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Excluded Domains</FormLabel>
          <FormDescription>
            Provide a list of comma-separated domains that you want to be
            excluded from the transformer. Do not provide an @ with the domains.{' '}
          </FormDescription>
        </div>
        <div className="min-w-[300px]">
          <Input
            type="string"
            className="min-w-[300px]"
            value={value.excludedDomains.map((d) => d.trim()).join(',')}
            onChange={(e) =>
              setValue(
                create(TransformEmailSchema, {
                  ...value,
                  excludedDomains: e.target.value.split(',').filter((d) => !!d),
                })
              )
            }
            disabled={isDisabled}
          />
        </div>
        <FormErrorMessage message={errors?.excludedDomains?.message} />
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Email Type</FormLabel>
          <FormDescription>
            Configure the email type that will be used during transformation.
          </FormDescription>
        </div>
        <Select
          disabled={isDisabled}
          onValueChange={(newVal) => {
            setValue(
              create(TransformEmailSchema, {
                ...value,
                // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                emailType: parseInt(newVal, 10),
              })
            );
          }}
          value={
            value.emailType?.toString() ?? GenerateEmailType.UUID_V4.toString()
          }
        >
          <SelectTrigger className="w-[300px]">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {[GenerateEmailType.UUID_V4, GenerateEmailType.FULLNAME].map(
              (emailType) => (
                <SelectItem
                  key={emailType}
                  className="cursor-pointer"
                  value={emailType.toString()}
                >
                  {getGenerateEmailTypeString(emailType)}
                </SelectItem>
              )
            )}
          </SelectContent>
        </Select>
        <FormErrorMessage message={errors?.emailType?.message} />
      </div>
      <div className="flex flex-row items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs">
        <div className="space-y-0.5 w-[80%]">
          <FormLabel>Invalid Email Action</FormLabel>
          <FormDescription>
            Configure the invalid email action that will be run in the event the
            system encounters an email that does not conform to RFC 5322.
          </FormDescription>
        </div>
        <div className="flex flex-col">
          <div className="justify-end flex">
            <Select
              disabled={isDisabled}
              onValueChange={(newValue) => {
                setValue(
                  create(TransformEmailSchema, {
                    ...value,
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    invalidEmailAction: parseInt(newValue, 10),
                  })
                );
              }}
              value={
                value.invalidEmailAction?.toString() ??
                InvalidEmailAction.REJECT.toString()
              }
            >
              <SelectTrigger className="w-[300px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {[
                  InvalidEmailAction.GENERATE,
                  InvalidEmailAction.NULL,
                  InvalidEmailAction.PASSTHROUGH,
                  InvalidEmailAction.REJECT,
                ].map((action) => (
                  <SelectItem
                    key={action}
                    className="cursor-pointer"
                    value={action.toString()}
                  >
                    {getInvalidEmailActionString(action)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <FormErrorMessage message={errors?.invalidEmailAction?.message} />
        </div>{' '}
      </div>
    </div>
  );
}
