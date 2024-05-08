'use client';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  generateEmailTypeStringToEnum,
  getGenerateEmailTypeString,
} from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { GenerateEmail, GenerateEmailType } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps } from './util';

interface Props extends TransformerFormProps<GenerateEmail> {}

export default function GenerateEmailForm(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const emailType =
    (existingConfig?.toJson() as any)?.emailType ?? // eslint-disable-line @typescript-eslint/no-explicit-any
    'GENERATE_EMAIL_TYPE_UUID_V4';
  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.generateEmailConfig),
    defaultValues: {
      emailType: emailType,
    },
  });

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name={`emailType`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm gap-4 ">
              <div className="space-y-0.5">
                <FormLabel>Email Type</FormLabel>
                <FormDescription className="w-[90%]">
                  Configure the email type that will be used during
                  transformation.
                </FormDescription>
              </div>
              <FormControl>
                <Select
                  disabled={isReadonly}
                  onValueChange={(value) => {
                    // this is so hacky, but has to be done due to have we are encoding the incoming config and how the enums are converted to their wire-format string type
                    const emailConfig = new GenerateEmail({
                      emailType: parseInt(value, 10),
                    }).toJson();
                    field.onChange((emailConfig as any).emailType); // eslint-disable-line @typescript-eslint/no-explicit-any
                  }}
                  value={generateEmailTypeStringToEnum(field.value).toString()}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {[
                      GenerateEmailType.UUID_V4,
                      GenerateEmailType.FULLNAME,
                    ].map((emailType) => (
                      <SelectItem
                        key={emailType}
                        className="cursor-pointer"
                        value={emailType.toString()}
                      >
                        {getGenerateEmailTypeString(emailType)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div className="flex justify-end">
          <Button
            type="button"
            disabled={isReadonly}
            onClick={(e) => {
              form.handleSubmit((values) => {
                onSubmit(GenerateEmail.fromJson(values));
              })(e);
            }}
          >
            Save
          </Button>
        </div>
      </Form>
    </div>
  );
}
