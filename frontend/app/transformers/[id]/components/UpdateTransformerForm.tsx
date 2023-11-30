'use client';
import { handleCustomTransformerForm } from '@/app/new/transformer/CustomTransformerForms/HandleCustomTransformersForm';
import {
  UPDATE_CUSTOM_TRANSFORMER,
  UpdateCustomTransformer,
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
import {
  Transformer,
  TransformerConfig,
  UpdateCustomTransformerRequest,
  UpdateCustomTransformerResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { Controller, useForm } from 'react-hook-form';

interface Props {
  currentTransformer: Transformer | undefined;
}

export default function UpdateTransformerForm(props: Props): ReactElement {
  const { currentTransformer } = props;

  const form = useForm<UpdateCustomTransformer>({
    resolver: yupResolver(UPDATE_CUSTOM_TRANSFORMER),
    defaultValues: {
      name: '',
      source: '',
      description: '',
      id: '',
      config: { config: { case: '', value: {} } },
    },
    values: {
      name: currentTransformer?.name ?? '',
      source: currentTransformer?.source ?? '',
      description: currentTransformer?.description ?? '',
      type: currentTransformer?.dataType ?? '',
      id: currentTransformer?.id ?? '',
      config: {
        config: {
          case: currentTransformer?.config?.config.case,
          value: currentTransformer?.config?.config.value ?? {},
        },
      },
    },
    context: { name: currentTransformer?.name },
  });

  const router = useRouter();
  const { account } = useAccount();

  async function onSubmit(values: UpdateCustomTransformer): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const transformer = await updateCustomTransformer(
        currentTransformer?.id ?? '',
        values
      );
      toast({
        title: 'Successfully updated transformer!',
        variant: 'success',
      });
      if (transformer.transformer?.id) {
        router.push(`/transformers`);
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
          control={form.control}
          name="source"
          render={() => (
            <FormItem>
              <FormLabel>Source Transformer</FormLabel>
              <FormDescription>
                The system transformer to clone.
              </FormDescription>
              <FormControl>
                <Select disabled={true}>
                  <SelectTrigger className="w-[1000px]">
                    <SelectValue
                      placeholder={String(currentTransformer?.source ?? '')}
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
                    <Input
                      placeholder="Transformer Name"
                      {...field}
                      className="w-[1000px]"
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>
        <div className="w-[1000px]">
          {handleCustomTransformerForm(currentTransformer?.source)}
        </div>
        <div className="flex flex-row justify-end">
          <Button type="submit">Save</Button>
        </div>
      </form>
    </Form>
  );
}

async function updateCustomTransformer(
  transformerId: string,
  formData: UpdateCustomTransformer
): Promise<UpdateCustomTransformerResponse> {
  const body = new UpdateCustomTransformerRequest({
    transformerId: transformerId,
    name: formData.name,
    description: formData.description,
    transformerConfig: formData.config as TransformerConfig,
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
  return UpdateCustomTransformerResponse.fromJson(await res.json());
}
