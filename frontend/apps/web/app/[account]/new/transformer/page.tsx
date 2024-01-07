'use client';

import FormError from '@/components/FormError';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
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
import { getErrorMessage } from '@/util/util';
import {
  convertTransformerConfigSchemaToTransformerConfig,
  convertTransformerConfigToForm,
} from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  CreateUserDefinedTransformerRequest,
  CreateUserDefinedTransformerResponse,
  SystemTransformer,
  TransformerConfig,
} from '@neosync/sdk';
import { CheckIcon } from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { handleUserDefinedTransformerForm } from './UserDefinedTransformerForms/HandleUserDefinedTransformersForm';
import {
  CREATE_USER_DEFINED_TRANSFORMER_SCHEMA,
  CreateUserDefinedTransformerSchema,
} from './schema';

export default function NewTransformer(): ReactElement {
  const [base, setBase] = useState<SystemTransformer>(
    new SystemTransformer({})
  );
  const [openBaseSelect, setOpenBaseSelect] = useState(false);
  const { account } = useAccount();
  const searchParams = useSearchParams();
  const transformerNameToClone = searchParams.get('transformerToClone');

  const { data } = useGetSystemTransformers();
  const transformers = data?.transformers ?? [];
  const transformerToClone = data?.transformers.find(
    (item: SystemTransformer) => item.source == transformerNameToClone
  );
  const isAutoFill = transformerToClone ? true : false;

  const form = useForm<CreateUserDefinedTransformerSchema>({
    resolver: yupResolver(CREATE_USER_DEFINED_TRANSFORMER_SCHEMA),
    mode: 'onChange',
    defaultValues: {
      name: '',
      source: '',
      type: '',
      config: convertTransformerConfigToForm(new TransformerConfig()),
      description: '',
    },
    context: { accountId: account?.id ?? '' },
  });

  const router = useRouter();

  useEffect(() => {
    if (transformerToClone) {
      form.setValue('name', 'name1');
      form.setValue('source', transformerToClone.source);
      form.setValue(
        'config',
        convertTransformerConfigToForm(transformerToClone.config)
      );
      form.setValue('description', transformerToClone.description);
      setBase(transformerToClone);
      setOpenBaseSelect(false);
    }
  }, []);

  async function onSubmit(
    values: CreateUserDefinedTransformerSchema
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
        router.push(
          `/${account?.name}/transformers/${transformer.transformer?.id}`
        );
      } else {
        router.push(`/${account?.name}/transformers`);
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

  return (
    <OverviewContainer
      Header={<PageHeader header="Create a new Transformer" />}
      containerClassName="px-12 md:px-24 lg:px-32"
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="source"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Source Transformer</FormLabel>
                <FormDescription>
                  The system transformer to clone.
                </FormDescription>
                <FormControl>
                  <Select
                    open={openBaseSelect}
                    onOpenChange={setOpenBaseSelect}
                  >
                    <SelectTrigger>
                      {base.name ? base.name : 'Select a transformer'}
                    </SelectTrigger>
                    <SelectContent>
                      <Command className="overflow-auto">
                        <CommandInput placeholder="Search transformers..." />
                        <CommandEmpty>No transformers found.</CommandEmpty>
                        <CommandGroup className="overflow-auto h-[200px]">
                          {transformers.map((t) => (
                            <CommandItem
                              key={`${t.source}`}
                              onSelect={() => {
                                field.onChange(t.source);
                                form.setValue(
                                  'config',
                                  convertTransformerConfigToForm(t.config)
                                );
                                form.setValue('type', t.dataType);
                                setBase(t ?? new SystemTransformer({}));
                                setOpenBaseSelect(false);
                              }}
                              value={t.name}
                            >
                              <CheckIcon
                                className={cn(
                                  'mr-2 h-4 w-4',
                                  base.name == t.name
                                    ? 'opacity-100'
                                    : 'opacity-0'
                                )}
                              />
                              {t.name}
                            </CommandItem>
                          ))}
                        </CommandGroup>
                      </Command>
                    </SelectContent>
                  </Select>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          {form.getValues('source') && (
            <div>
              <Controller
                control={form.control}
                name="name"
                render={({ field: { onChange, ...field } }) => (
                  <FormItem>
                    <FormLabel>Name</FormLabel>
                    <FormDescription>
                      The unique name of the transformer.
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
                    <FormError
                      errorMessage={form.formState.errors.name?.message ?? ''}
                    />
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
                      <FormDescription>
                        The Transformer description.
                      </FormDescription>
                      <FormControl>
                        <Input
                          placeholder="Transformer description"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </div>
          )}
          <div>
            {handleUserDefinedTransformerForm(form.getValues('source'))}
          </div>
          <div className="flex flex-row justify-end">
            <Button
              type="submit"
              disabled={!isAutoFill ?? !form.formState.isValid}
            >
              Submit
            </Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

async function createNewTransformer(
  accountId: string,
  formData: CreateUserDefinedTransformerSchema
): Promise<CreateUserDefinedTransformerResponse> {
  const body = new CreateUserDefinedTransformerRequest({
    accountId: accountId,
    name: formData.name,
    description: formData.description,
    type: formData.type,
    source: formData.source,
    transformerConfig: convertTransformerConfigSchemaToTransformerConfig(
      formData.config
    ),
  });

  const res = await fetch(
    `/api/accounts/${accountId}/transformers/user-defined`,
    {
      method: 'POST',
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
  return CreateUserDefinedTransformerResponse.fromJson(await res.json());
}
