'use client';
import { UserDefinedTransformerForm } from '@/app/(mgmt)/[account]/new/transformer/UserDefinedTransformerForms/UserDefinedTransformerForm';
import {
  UPDATE_USER_DEFINED_TRANSFORMER,
  UpdateUserDefinedTransformer,
} from '@/app/(mgmt)/[account]/new/transformer/schema';
import { useAccount } from '@/components/providers/account-provider';
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
import { Input } from '@/components/ui/input';
import { Select, SelectTrigger, SelectValue } from '@/components/ui/select';
import { toast } from '@/components/ui/use-toast';
import { getErrorMessage, getTransformerSourceString } from '@/util/util';
import {
  convertTransformerConfigSchemaToTransformerConfig,
  convertTransformerConfigToForm,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  UpdateUserDefinedTransformerRequest,
  UpdateUserDefinedTransformerResponse,
  UserDefinedTransformer,
} from '@neosync/sdk';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { Controller, useForm } from 'react-hook-form';

interface Props {
  currentTransformer: UserDefinedTransformer;
  onUpdated(transformer: UserDefinedTransformer): void;
}

export default function UpdateUserDefinedTransformerForm(
  props: Props
): ReactElement {
  const { currentTransformer, onUpdated } = props;
  const { account } = useAccount();

  const form = useForm<UpdateUserDefinedTransformer>({
    mode: 'onChange',
    resolver: yupResolver(UPDATE_USER_DEFINED_TRANSFORMER),
    values: {
      name: currentTransformer?.name ?? '',
      description: currentTransformer?.description ?? '',
      id: currentTransformer?.id ?? '',
      config: convertTransformerConfigToForm(currentTransformer.config),
    },
    context: { name: currentTransformer?.name, accountId: account?.id ?? '' },
  });

  async function onSubmit(values: UpdateUserDefinedTransformer): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const transformer = await updateCustomTransformer(
        account?.id ?? '',
        currentTransformer?.id ?? '',
        values
      );
      toast({
        title: 'Successfully updated transformer!',
        variant: 'success',
      });
      if (transformer.transformer) {
        onUpdated(transformer.transformer);
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to update transformer',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          name="source"
          render={() => (
            <FormItem>
              <FormLabel>Source Transformer</FormLabel>
              <FormDescription>The system transformer source.</FormDescription>
              <FormControl>
                <Select disabled={true}>
                  <SelectTrigger>
                    <SelectValue
                      placeholder={getTransformerSourceString(
                        currentTransformer.source
                      )}
                    />
                  </SelectTrigger>
                </Select>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <div>
          <Controller
            control={form.control}
            name="name"
            render={({ field: { onChange, ...field } }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormDescription>
                  The unique name of the Transformer.
                </FormDescription>
                <FormControl>
                  <Input
                    placeholder="Transformer Name"
                    {...field}
                    onChange={async ({ target: { value } }) => {
                      onChange(value);
                      await form.trigger('name');
                    }}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <div className="pt-10">
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormDescription>The Transformer decription.</FormDescription>
                  <FormControl>
                    <Input placeholder="Transformer Name" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>
        <div>
          <UserDefinedTransformerForm value={currentTransformer?.source} />
        </div>
        <div className="flex flex-row justify-between">
          <NextLink href={`/${account?.name}/transformers?tab=ud`}>
            <Button type="button">Back</Button>
          </NextLink>
          <Button type="submit">Save</Button>
        </div>
      </form>
    </Form>
  );
}

async function updateCustomTransformer(
  accountId: string,
  transformerId: string,
  formData: UpdateUserDefinedTransformer
): Promise<UpdateUserDefinedTransformerResponse> {
  const body = new UpdateUserDefinedTransformerRequest({
    transformerId: transformerId,
    name: formData.name,
    description: formData.description,
    transformerConfig: convertTransformerConfigSchemaToTransformerConfig(
      formData.config
    ),
  });

  const res = await fetch(
    `/api/accounts/${accountId}/transformers/user-defined`,
    {
      method: 'PUT',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(body),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateUserDefinedTransformerResponse.fromJson(await res.json());
}
