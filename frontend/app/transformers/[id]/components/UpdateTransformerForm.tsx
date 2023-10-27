'use client';
import { handleCustomTransformerForm } from '@/app/new/transformer/page';
import {
  DEFINE_NEW_TRANSFORMER_SCHEMA,
  DefineNewTransformer,
} from '@/app/new/transformer/schema';
import { handleTransformerMetadata } from '@/app/transformers/EditTransformerOptions';
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
  CustomTransformer,
  Transformer,
  UpdateCustomTransformerRequest,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';
import { getErrorMessage } from '@/util/util';
import { toTransformerConfigOptions } from '@/yup-validations/transformers';
import { yupResolver } from '@hookform/resolvers/yup';
import { CheckIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';

interface Props {
  defaultTransformerValues: CustomTransformer;
}

export default function UpdateTransformerForm(props: Props): ReactElement {
  const [base, setBase] = useState<Transformer>(new Transformer());
  const [openBaseSelect, setOpenBaseSelect] = useState(false);

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
      name: defaultTransformerValues.name,
      base: defaultTransformerValues.type, //needs to be updated
      description: defaultTransformerValues.description,
      type: defaultTransformerValues.type,
      transformerConfig: defaultTransformerValues,
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

  console.log('values', form.getValues());

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
        <FormField
          control={form.control}
          name="base"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Source Transformer</FormLabel>
              <FormControl>
                <Select open={openBaseSelect} onOpenChange={setOpenBaseSelect}>
                  <SelectTrigger className="w-[1000px]">
                    <SelectValue placeholder="Select a transformer" />
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
                                'type',
                                handleTransformerMetadata(value).type
                              );
                              form.setValue('base', value);
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
        <div className="w-[1000px]">{handleCustomTransformerForm(base)}</div>
        <div className="flex flex-row justify-end">
          <Button type="submit">Next</Button>
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
    type: formData.type,
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
