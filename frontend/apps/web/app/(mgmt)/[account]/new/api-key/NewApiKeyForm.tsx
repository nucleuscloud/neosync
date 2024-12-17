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
import { cn } from '@/libs/utils';
import { getErrorMessage } from '@/util/util';
import { ApiKeyFormValues } from '@/yup-validations/apikey';
import { timestampFromMs } from '@bufbuild/protobuf/wkt';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { createAccountApiKey } from '@neosync/sdk/connectquery';
import { CalendarIcon } from '@radix-ui/react-icons';
import { PopoverTrigger } from '@radix-ui/react-popover';
import { addDays } from 'date-fns';
import { useRouter } from 'next/navigation';
import { usePostHog } from 'posthog-js/react';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';

export interface ApiKeyValueSessionStore {
  keyValue: string;
}

export default function NewApiKeyForm(): ReactElement {
  const { account } = useAccount();
  const router = useRouter();
  const form = useForm<ApiKeyFormValues>({
    mode: 'onChange',
    resolver: yupResolver(ApiKeyFormValues),
    defaultValues: {
      name: '',
      expiresAtSelect: '7',
      expiresAt: startOfDay(addDays(new Date(), 7)),
    },
  });
  const posthog = usePostHog();
  const { mutateAsync } = useMutation(createAccountApiKey);

  async function onSubmit(values: ApiKeyFormValues): Promise<void> {
    if (!account) {
      return;
    }
    try {
      const apiKey = await mutateAsync({
        accountId: account.id,
        expiresAt: timestampFromMs(values.expiresAt.getTime()),
        name: values.name,
      });
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
      } else {
        router.push(`/${account?.name}/settings/api-keys`);
      }
      posthog.capture('New API Key Created');
      toast.success('Successfully created API key!');
    } catch (err) {
      console.error(err);
      toast.error('Unable to create api key', {
        description: getErrorMessage(err),
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
