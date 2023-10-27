'use client';
import { handleCustomTransformerForm } from '@/app/new/transformer/page';
import {
  DEFINE_NEW_TRANSFORMER_SCHEMA,
  DefineNewTransformer,
} from '@/app/new/transformer/schema';
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
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import {
  CustomTransformer,
  UpdateCustomTransformerRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { toTransformerConfigOptions } from '@/yup-validations/transformers';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';

interface Props {
  defaultTransformerValues: CustomTransformer | undefined;
}

export default function UpdateTransformerForm(props: Props): ReactElement {
  const { defaultTransformerValues } = props;

  const form = useForm<DefineNewTransformer>({
    resolver: yupResolver(DEFINE_NEW_TRANSFORMER_SCHEMA),
    defaultValues: {
      name: '',
      base: '',
      description: '',
      type: '',
      transformerConfig: {},
    },
    values: {
      name: defaultTransformerValues?.name ?? '',
      base: defaultTransformerValues?.source ?? '',
      description: defaultTransformerValues?.description ?? '',
      type: defaultTransformerValues?.type ?? '',
      transformerConfig: defaultTransformerValues ?? {},
    },
  });

  const router = useRouter();
  const account = useAccount();

  async function onSubmit(values: DefineNewTransformer): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const transformer = await updateCustomTransformer(account.id, values);
      if (transformer.transformerId) {
        router.push(`/transformers/${transformer.transformerId}`);
      } else {
        router.push(`/transformers`);
      }
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create transformer',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  const { data } = useGetSystemTransformers();
  const transformers = data?.transformers ?? [];

  const val = transformers.find(
    (item) => item.value == defaultTransformerValues?.source ?? ''
  );

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="base"
          render={() => (
            <FormItem>
              <FormLabel>Source Transformer</FormLabel>
              <FormControl>
                <Select disabled={true}>
                  <SelectTrigger className="w-[1000px]">
                    <SelectValue
                      placeholder={String(
                        defaultTransformerValues?.source ?? ''
                      )}
                    />
                  </SelectTrigger>
                </Select>
              </FormControl>
              <FormDescription>
                The source transformer to clone.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />
        <div>
          <FormField
            control={form.control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input
                    placeholder="Transformer Name"
                    {...field}
                    className="w-[1000px]"
                  />
                </FormControl>
                <FormDescription>
                  The unique name of the Transformer.
                </FormDescription>
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
                  <FormControl>
                    <Input
                      placeholder="Transformer Name"
                      {...field}
                      className="w-[1000px]"
                    />
                  </FormControl>
                  <FormDescription>The Transformer decription.</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>
        <div className="w-[1000px]">{handleCustomTransformerForm(val)}</div>
        <div className="flex flex-row justify-end">
          <Button type="submit">Save</Button>
        </div>
      </form>
    </Form>
  );
}

async function updateCustomTransformer(
  accountId: string,
  formData: DefineNewTransformer
): Promise<UpdateCustomTransformerRequest> {
  const body = new UpdateCustomTransformerRequest({
    transformerId: accountId,
    name: formData.name,
    description: formData.description,
    transformerConfig: toTransformerConfigOptions({
      value: formData.base ?? 'passthrough',
      config: formData.transformerConfig,
    }).config,
  });

  const res = await fetch(`/api/transformers/custom`, {
    method: 'PUT',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return UpdateCustomTransformerRequest.fromJson(await res.json());
}
