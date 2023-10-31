'use client';

import { handleTransformerMetadata } from '@/app/transformers/EditTransformerOptions';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { PageProps } from '@/components/types';
import { Button } from '@/components/ui/button';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
} from '@/components/ui/command';
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
import { Select, SelectContent, SelectTrigger } from '@/components/ui/select';
import { toast } from '@/components/ui/use-toast';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { cn } from '@/libs/utils';
import {
  CreateCustomTransformerRequest,
  CreateCustomTransformerResponse,
  Transformer,
  TransformerConfig,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { yupResolver } from '@hookform/resolvers/yup';
import { CheckIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { handleCustomTransformerForm } from './CustomTransformerForms/HandleCustomTransformersForm';
import {
  CREATE_CUSTOM_TRANSFORMER_SCHEMA,
  CreateCustomTransformerSchema,
} from './schema';

export default function NewTransformer({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';

  console.log('id', id);

  const [base, setBase] = useState<Transformer>(new Transformer());
  const [openBaseSelect, setOpenBaseSelect] = useState(false);

  const form = useForm<CreateCustomTransformerSchema>({
    resolver: yupResolver(CREATE_CUSTOM_TRANSFORMER_SCHEMA),
    defaultValues: {
      name: '',
      source: '',
      description: '',
      config: { config: { case: '', value: {} } },
    },
  });

  const router = useRouter();
  const account = useAccount();

  async function onSubmit(
    values: CreateCustomTransformerSchema
  ): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const transformer = await createNewTransformer(account.id, values);
      toast({
        title: 'Successfully created transformer!',
        variant: 'success',
      });
      if (transformer.transformer?.id) {
        router.push(`/transformers/${transformer.transformer?.id}`);
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

  return (
    <OverviewContainer
      Header={<PageHeader header="Create a new Transformer" />}
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="source"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Source Transformer</FormLabel>
                <FormControl>
                  <Select
                    open={openBaseSelect}
                    onOpenChange={setOpenBaseSelect}
                  >
                    <SelectTrigger className="w-[1000px]">
                      {base.value ? base.value : 'Select a transformer'}
                    </SelectTrigger>
                    <SelectContent>
                      <Command className="overflow-auto">
                        <CommandInput placeholder="Search transformers..." />
                        <CommandEmpty>No transformers found.</CommandEmpty>
                        <CommandGroup className="overflow-auto h-[200px]">
                          {transformers.map((t, index) => (
                            <CommandItem
                              key={`${t.value}-${index}`}
                              onSelect={(value: string) => {
                                field.onChange;
                                setBase(
                                  transformers.find(
                                    (item) => item.value == value
                                  ) ?? new Transformer()
                                );
                                form.setValue(
                                  'config.config.case',
                                  transformers.find(
                                    (item) => item.value == value
                                  )?.config?.config.case ?? ''
                                );
                                form.setValue('source', value);
                                setOpenBaseSelect(false);
                              }}
                              value={t.value}
                              defaultValue={'passthrough'}
                            >
                              <CheckIcon
                                className={cn(
                                  'mr-2 h-4 w-4',
                                  base.value == t.value
                                    ? 'opacity-100'
                                    : 'opacity-0'
                                )}
                              />
                              {t.value}
                            </CommandItem>
                          ))}
                        </CommandGroup>
                      </Command>
                    </SelectContent>
                  </Select>
                </FormControl>
                <FormDescription>
                  The source transformer to clone.
                </FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          {base.value && (
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
                      The unique name of the transformer.
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
                          placeholder="Transformer description"
                          {...field}
                          className="w-[1000px]"
                        />
                      </FormControl>
                      <FormDescription>
                        The Transformer decription.
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </div>
          )}
          <div className="w-[1000px]">
            {handleCustomTransformerForm(base.value)}
          </div>
          <div className="flex flex-row justify-end">
            <Button type="submit">Next</Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

async function createNewTransformer(
  accountId: string,
  formData: CreateCustomTransformerSchema
): Promise<CreateCustomTransformerResponse> {
  const body = new CreateCustomTransformerRequest({
    accountId: accountId,
    name: formData.name,
    description: formData.description,
    type: handleTransformerMetadata(formData.source).type,
    source: formData.source,
    transformerConfig: formData.config as TransformerConfig,
  });

  const res = await fetch(`/api/transformers/custom`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
    },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return CreateCustomTransformerResponse.fromJson(await res.json());
}
