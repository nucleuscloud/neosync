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
  CommandList,
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
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import {
  convertTransformerConfigSchemaToTransformerConfig,
  convertTransformerConfigToForm,
} from '@/yup-validations/jobs';
import { useMutation, useQuery } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import {
  GenerateBool,
  SystemTransformer,
  TransformerConfig,
  TransformerSource,
} from '@neosync/sdk';
import {
  createUserDefinedTransformer,
  getSystemTransformers,
  isTransformerNameAvailable,
  validateUserJavascriptCode,
} from '@neosync/sdk/connectquery';
import { CheckIcon } from '@radix-ui/react-icons';
import { useRouter, useSearchParams } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement, useEffect, useState } from 'react';
import { Controller, useForm } from 'react-hook-form';
import { UserDefinedTransformerForm } from './UserDefinedTransformerForms/UserDefinedTransformerForm';
import {
  CreateUserDefinedTransformerFormContext,
  CreateUserDefinedTransformerFormValues,
} from './schema';

function getTransformerSource(sourceStr: string): TransformerSource {
  const sourceNum = parseInt(sourceStr, 10);
  if (isNaN(sourceNum) || !TransformerSource[sourceNum]) {
    return TransformerSource.UNSPECIFIED;
  }
  return sourceNum as TransformerSource;
}

export default function NewTransformer(): ReactElement {
  const { account } = useAccount();

  const { data, isLoading } = useQuery(getSystemTransformers);
  const transformers = data?.transformers ?? [];

  const transformerQueryParam = useSearchParams().get('transformer');
  const transformerSource = getTransformerSource(
    transformerQueryParam ?? TransformerSource.UNSPECIFIED.toString()
  );
  const { mutateAsync: isTransformerNameAvailableAsync } = useMutation(
    isTransformerNameAvailable
  );
  const { mutateAsync: isJavascriptCodeValid } = useMutation(
    validateUserJavascriptCode
  );

  const [openBaseSelect, setOpenBaseSelect] = useState(false);
  const posthog = usePostHog();

  const form = useForm<
    CreateUserDefinedTransformerFormValues,
    CreateUserDefinedTransformerFormContext
  >({
    resolver: yupResolver(CreateUserDefinedTransformerFormValues),
    mode: 'onChange',
    defaultValues: {
      name: '',
      source: transformerSource,
      config: convertTransformerConfigToForm(
        new TransformerConfig({
          config: { case: 'generateBoolConfig', value: new GenerateBool() },
        })
      ),
      description: '',
    },
    context: {
      accountId: account?.id ?? '',
      isTransformerNameAvailable: isTransformerNameAvailableAsync,
      isJavascriptCodeValid: isJavascriptCodeValid,
    },
  });

  const router = useRouter();
  const { mutateAsync } = useMutation(createUserDefinedTransformer);

  async function onSubmit(
    values: CreateUserDefinedTransformerFormValues
  ): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const transformer = await mutateAsync({
        accountId: account.id,
        name: values.name,
        description: values.description,
        source: values.source,
        transformerConfig: convertTransformerConfigSchemaToTransformerConfig(
          values.config
        ),
      });
      posthog.capture('New Transformer Created', {
        source: values.source,
        sourceName: transformers.find((t) => t.source === values.source)?.name,
      });
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

  const formSource = form.watch('source');

  const base =
    transformers.find((t) => t.source === formSource) ??
    new SystemTransformer();

  const configCase = form.watch('config.case');

  useEffect(() => {
    if (
      isLoading ||
      base.source === TransformerSource.UNSPECIFIED ||
      configCase ||
      !transformerQueryParam
    ) {
      return;
    }

    form.setValue('config', convertTransformerConfigToForm(base.config));
  }, [isLoading, base.source, configCase, transformerQueryParam]);

  return (
    <OverviewContainer
      Header={<PageHeader header="Create a New Transformer" />}
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
                        <CommandList>
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
                                  setOpenBaseSelect(false);
                                }}
                                value={t.name}
                              >
                                <CheckIcon
                                  className={cn(
                                    'mr-2 h-4 w-4',
                                    base.name === t.name
                                      ? 'opacity-100'
                                      : 'opacity-0'
                                  )}
                                />
                                {t.name}
                              </CommandItem>
                            ))}
                          </CommandGroup>
                        </CommandList>
                      </Command>
                    </SelectContent>
                  </Select>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          {formSource != null && formSource !== 0 && (
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
            <FormField
              control={form.control}
              name="config"
              render={({ field }) => (
                <FormItem>
                  <FormControl>
                    <UserDefinedTransformerForm
                      value={convertTransformerConfigSchemaToTransformerConfig(
                        field.value
                      )}
                      setValue={(newValue) => {
                        field.onChange(
                          convertTransformerConfigToForm(newValue)
                        );
                      }}
                      disabled={false}
                      errors={form.formState.errors}
                    />
                  </FormControl>
                </FormItem>
              )}
            ></FormField>
          </div>
          <div className="flex flex-row justify-end">
            <Button type="submit" disabled={!form.formState.isValid}>
              Submit
            </Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}
