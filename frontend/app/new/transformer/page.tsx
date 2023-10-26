'use client';

import CustomEmailTransformerForm from '@/app/transformers/[id]/components/CustomEmailTransformerForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
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
import {
  Select,
  SelectContent,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { cn } from '@/libs/utils';
import { yupResolver } from '@hookform/resolvers/yup';
import { CheckIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import useFormPersist from 'react-hook-form-persist';
import { useSessionStorage } from 'usehooks-ts';
import { DefineFormValues } from '../job/schema';
import { DEFINE_NEW_TRANSFORMER_SCHEMA, DefineNewTransformer } from './schema';

export default function NewTransformer({
  searchParams,
}: PageProps): ReactElement {
  const sessionPrefix = searchParams?.sessionId ?? '';
  const [defaultValues] = useSessionStorage<DefineNewTransformer>(
    `${sessionPrefix}-new-transformer-define`,
    {
      transformerName: '',
      baseTransformer: '',
    }
  );

  const [baseTransformer, setBaseTransformer] = useState<string>('');
  const [openBaseTransformerSelect, setOpenBaseTransformerSelect] =
    useState(false);

  const form = useForm({
    resolver: yupResolver<DefineNewTransformer>(DEFINE_NEW_TRANSFORMER_SCHEMA),
    defaultValues,
  });
  useFormPersist(`${sessionPrefix}-new-job-define`, {
    watch: form.watch,
    setValue: form.setValue,
    storage: window.sessionStorage,
  });

  async function onSubmit(_values: DefineFormValues) {
    console.log('creating transformer');
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
            name="transformerName"
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

          <FormField
            control={form.control}
            name="baseTransformer"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Source Transformer</FormLabel>
                <FormControl>
                  <Select
                    open={openBaseTransformerSelect}
                    onOpenChange={setOpenBaseTransformerSelect}
                  >
                    <SelectTrigger className="w-[1000px]">
                      <SelectValue placeholder="Select a base transformer"></SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <Command className="overflow-auto">
                        <CommandInput placeholder="Search transformers..." />
                        <CommandEmpty>No transformers found.</CommandEmpty>
                        <CommandGroup className="overflow-auto h-[200px]">
                          {transformers.map((t, index) => (
                            <CommandItem
                              key={`${t.value}-${index}`}
                              onSelect={() => {
                                field.onChange;
                                setBaseTransformer(t.value);
                                setOpenBaseTransformerSelect(false);
                              }}
                              value={t.value}
                              defaultValue={'passthrough'}
                            >
                              <CheckIcon
                                className={cn(
                                  'mr-2 h-4 w-4',
                                  baseTransformer == t.value
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
          <div className="w-[1000px]">
            {handleNewTransformerForm(baseTransformer)}
          </div>
          <div className="flex flex-row justify-end">
            <Button type="submit">Next</Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

function handleNewTransformerForm(name: string): ReactElement {
  switch (name) {
    case 'email':
      return <CustomEmailTransformerForm />;
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
