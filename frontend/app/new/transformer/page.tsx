'use client';

import { handleTransformerMetadata } from '@/app/transformers/EditTransformerOptions';
import CustomEmailTransformerForm from '@/app/transformers/[id]/components/CustomEmailTransformerForm';
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
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { toast } from '@/components/ui/use-toast';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { cn } from '@/libs/utils';
import {
  CreateCustomTransformerRequest,
  CreateCustomTransformerResponse,
  Transformer,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { toTransformerConfigOptions } from '@/yup-validations/transformers';
import { yupResolver } from '@hookform/resolvers/yup';
import { CheckIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import { DEFINE_NEW_TRANSFORMER_SCHEMA, DefineNewTransformer } from './schema';

export default function NewTransformer(): ReactElement {
  const [base, setBase] = useState<string>('Select a source Transformer');
  const [openBaseSelect, setOpenBaseSelect] = useState(false);
  const [selectedTransformer, setSelectedTransformer] = useState<Transformer>();

  const form = useForm<DefineNewTransformer>({
    resolver: yupResolver(DEFINE_NEW_TRANSFORMER_SCHEMA),
    defaultValues: {
      name: '',
      base: '',
      description: '',
      type: '',
      transformerConfig: {},
    },
  });

  const router = useRouter();
  const account = useAccount();

  async function onSubmit(values: DefineNewTransformer): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const transformer = await createNewTransformer(account.id, values);
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

  console.log('values', form.getValues());

  return (
    <OverviewContainer
      Header={<PageHeader header="Create a new Transformer" />}
    >
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="base"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Source Transformer</FormLabel>
                <FormControl>
                  <Select
                    open={openBaseSelect}
                    onOpenChange={setOpenBaseSelect}
                  >
                    <SelectTrigger className="w-[1000px]">
                      <SelectValue placeholder={base} />
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
                                setBase(value);
                                form.setValue(
                                  'type',
                                  handleTransformerMetadata(value).type
                                );
                                form.setValue('base', value);
                                setSelectedTransformer(
                                  transformers.find(
                                    (item) => item.value == t.value
                                  )
                                );
                                setOpenBaseSelect(false);
                              }}
                              value={t.value}
                              defaultValue={'passthrough'}
                            >
                              <CheckIcon
                                className={cn(
                                  'mr-2 h-4 w-4',
                                  base == t.value ? 'opacity-100' : 'opacity-0'
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
          {base != 'Select a source Transformer' && (
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
            {handleNewTransformerForm(base, selectedTransformer)}
          </div>
          <div className="flex flex-row justify-end">
            <Button type="submit">Next</Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

function handleNewTransformerForm(
  name: string,
  transformer: Transformer | undefined
): ReactElement {
  switch (name) {
    case 'email':
      return (
        <CustomEmailTransformerForm
          transformer={transformer ?? new Transformer()}
        />
      );
    // case 'uuid':
    //   return (
    //     <UuidTransformerForm index={index} setIsSheetOpen={setIsSheetOpen} />
    //   );
    // case 'first_name':
    //   return (
    //     <FirstNameTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'last_name':
    //   return (
    //     <LastNameTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'full_name':
    //   return (
    //     <FullNameTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'phone_number':
    //   return (
    //     <PhoneNumberTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'int_phone_number':
    //   return (
    //     <IntPhoneNumberTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'random_string':
    //   return (
    //     <RandomStringTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'random_int':
    //   return (
    //     <RandomIntTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'random_float':
    //   return (
    //     <RandomFloatTransformerForm
    //       index={index}
    //       setIsSheetOpen={setIsSheetOpen}
    //     />
    //   );
    // case 'gender':
    //   return (
    //     <GenderTransformerForm index={index} setIsSheetOpen={setIsSheetOpen} />
    //   );
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}

async function createNewTransformer(
  accountId: string,
  formData: DefineNewTransformer
): Promise<CreateCustomTransformerResponse> {
  const body = new CreateCustomTransformerRequest({
    accountId: accountId,
    name: formData.name,
    description: formData.description,
    type: formData.type,
    transformerConfig: toTransformerConfigOptions({
      value: formData.base ?? 'passthrough',
      config: formData.transformerConfig,
    }).config,
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
