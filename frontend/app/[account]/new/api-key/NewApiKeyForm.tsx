'use client';
import RequiredLabel from '@/components/labels/RequiredLabel';
import { useAccount } from '@/components/providers/account-provider';
import { Button } from '@/components/ui/button';
import { endOfDay, format, startOfDay } from 'date-fns';

import { Calendar } from '@/components/ui/calendar';
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
import { Popover, PopoverContent } from '@/components/ui/popover';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useToast } from '@/components/ui/use-toast';
import { cn } from '@/libs/utils';
import {
  CreateAccountApiKeyRequest,
  CreateAccountApiKeyResponse,
  GetAccountApiKeyResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { getErrorMessage } from '@/util/util';
import { Timestamp } from '@bufbuild/protobuf';
import { yupResolver } from '@hookform/resolvers/yup';
import { CalendarIcon } from '@radix-ui/react-icons';
import { PopoverTrigger } from '@radix-ui/react-popover';
import { addDays } from 'date-fns';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { mutate } from 'swr';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  name: Yup.string().required(),
  expiresAtSelect: Yup.string().oneOf(['7', '30', '60', '90', 'custom']),
  expiresAt: Yup.date().required(),
});
type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

export interface ApiKeyValueSessionStore {
  keyValue: string;
}

export default function NewApiKeyForm(): ReactElement {
  const { account } = useAccount();
  const { toast } = useToast();
  const router = useRouter();
  const form = useForm<FormValues>({
    resolver: yupResolver(FORM_SCHEMA),
    defaultValues: {
      name: '',
      expiresAtSelect: '7',
      expiresAt: startOfDay(addDays(new Date(), 7)),
    },
  });

  async function onSubmit(values: FormValues): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const apiKey = await createNewApiKey(values, account.id);
      if (apiKey.apiKey?.id) {
        if (apiKey.apiKey.keyValue && !!window?.sessionStorage) {
          const storeVal: ApiKeyValueSessionStore = {
            keyValue: apiKey.apiKey.keyValue,
          };
          window.sessionStorage.setItem(
            apiKey.apiKey.id,
            JSON.stringify(storeVal)
          );
        }
        router.push(`/${account?.name}/settings/api-keys/${apiKey.apiKey.id}`);
        mutate(
          `/api/api-keys/account/${apiKey.apiKey.id}`,
          new GetAccountApiKeyResponse({
            apiKey: apiKey.apiKey,
          })
        );
      } else {
        router.push(`/${account?.name}/settings/api-keys`);
      }
      toast({
        title: 'Successfully created API key!',
        variant: 'success',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to create api key',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className="flex flex-col gap-8"
      >
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                Name <RequiredLabel />
              </FormLabel>
              <FormControl>
                <Input placeholder="API Key Name" {...field} />
              </FormControl>
              <FormDescription>
                The unique, friendly name of the api key.
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name="expiresAtSelect"
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                Expiration <RequiredLabel />
              </FormLabel>
              <Select
                onValueChange={(value) => {
                  field.onChange(value);
                  if (value !== 'custom') {
                    form.setValue(
                      'expiresAt',
                      addDays(startOfDay(new Date()), parseInt(value, 10))
                    );
                  }
                }}
                value={field.value}
              >
                <FormControl>
                  <SelectTrigger>
                    <SelectValue placeholder="Select an expiration date" />
                  </SelectTrigger>
                </FormControl>
                <SelectContent>
                  <SelectItem className="cursor-pointer" value="7">
                    7 days
                  </SelectItem>
                  <SelectItem className="cursor-pointer" value="30">
                    30 days
                  </SelectItem>
                  <SelectItem className="cursor-pointer" value="60">
                    60 days
                  </SelectItem>
                  <SelectItem className="cursor-pointer" value="90">
                    90 days
                  </SelectItem>
                  <SelectItem className="cursor-pointer" value="custom">
                    Custom
                  </SelectItem>
                </SelectContent>
              </Select>
              <FormDescription>
                The token will expire on{' '}
                {format(form.getValues().expiresAt, 'PPP')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {form.watch().expiresAtSelect === 'custom' && (
          <FormField
            control={form.control}
            disabled={form.getValues().expiresAtSelect !== 'custom'}
            name="expiresAt"
            render={({ field }) => (
              <FormItem className="flex flex-col">
                <Popover>
                  <PopoverTrigger asChild>
                    <FormControl>
                      <Button
                        variant={'outline'}
                        className={cn(
                          'w-[240px] pl-3 text-left font-normal',
                          !field.value && 'text-muted-foreground'
                        )}
                      >
                        {field.value ? (
                          format(field.value, 'PPP')
                        ) : (
                          <span>Pick a date</span>
                        )}
                        <CalendarIcon className="ml-auto h-4 w-4 opacity-50" />
                      </Button>
                    </FormControl>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={field.value}
                      onSelect={(val) => {
                        field.onChange(startOfDay(val ?? new Date()));
                      }}
                      disabled={(date) =>
                        date < endOfDay(new Date()) ||
                        // must be days instead of years to account for leap year
                        // backend constraints to within 365 days of the current time
                        date > addDays(startOfDay(new Date()), 365)
                      }
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
                <FormMessage />
              </FormItem>
            )}
          />
        )}
        <div className="flex flex-row justify-end">
          <Button type="submit">Submit</Button>
        </div>
      </form>
    </Form>
  );
}

async function createNewApiKey(
  formData: FormValues,
  accountId: string
): Promise<CreateAccountApiKeyResponse> {
  const body = new CreateAccountApiKeyRequest({
    accountId,
    name: formData.name,
    expiresAt: new Timestamp({
      seconds: BigInt(formData.expiresAt.getTime() / 1000),
    }),
  });

  const res = await fetch(`/api/accounts/${accountId}/api-keys`, {
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
  return CreateAccountApiKeyResponse.fromJson(await res.json());
}
