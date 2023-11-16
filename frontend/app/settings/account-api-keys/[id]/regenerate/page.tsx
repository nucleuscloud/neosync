'use client';
import { ApiKeyValueSessionStore } from '@/app/new/account-api-key/NewApiKeyForm';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import RequiredLabel from '@/components/labels/RequiredLabel';
import SkeletonForm from '@/components/skeleton/SkeletonForm';
import { PageProps } from '@/components/types';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
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
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useToast } from '@/components/ui/use-toast';
import { useGetAccountApiKey } from '@/libs/hooks/useGetAccountApiKey';
import { cn } from '@/libs/utils';
import {
  GetAccountApiKeyResponse,
  RegenerateAccountApiKeyRequest,
  RegenerateAccountApiKeyResponse,
} from '@/neosync-api-client/mgmt/v1alpha1/api_key_pb';
import { getErrorMessage } from '@/util/util';
import { Timestamp } from '@bufbuild/protobuf';
import { yupResolver } from '@hookform/resolvers/yup';
import { CalendarIcon } from '@radix-ui/react-icons';
import { addDays, addYears, endOfDay, format, startOfDay } from 'date-fns';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { mutate } from 'swr';
import * as Yup from 'yup';

const FORM_SCHEMA = Yup.object({
  expiresAtSelect: Yup.string().oneOf(['7', '30', '60', '90', 'custom']),
  expiresAt: Yup.date().required(),
});
type FormValues = Yup.InferType<typeof FORM_SCHEMA>;

export default function RegenerateAccountApiKey({
  params,
}: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetAccountApiKey(id);
  const { toast } = useToast();
  const router = useRouter();

  const form = useForm<FormValues>({
    resolver: yupResolver(FORM_SCHEMA),
    defaultValues: {
      expiresAtSelect: '7',
      expiresAt: startOfDay(addDays(new Date(), 7)),
    },
  });

  async function onSubmit(values: FormValues): Promise<void> {
    if (!id) {
      return;
    }
    try {
      const updatedApiKey = await regenerateApiKey(values, id);
      if (updatedApiKey.apiKey?.keyValue && !!window?.sessionStorage) {
        const storeVal: ApiKeyValueSessionStore = {
          keyValue: updatedApiKey.apiKey.keyValue,
        };
        window.sessionStorage.setItem(id, JSON.stringify(storeVal));
      }
      router.push(`/settings/account-api-keys/${id}`);
      mutate(
        `/api/api-keys/account/${id}`,
        new GetAccountApiKeyResponse({
          apiKey: updatedApiKey.apiKey,
        })
      );
      toast({
        title: 'Successfully regenerated api key!',
        variant: 'success',
      });
    } catch (err) {
      console.error(err);
      toast({
        title: 'Unable to regenerate api key!',
        description: getErrorMessage(err),
        variant: 'destructive',
      });
    }
  }

  if (!id) {
    return <div>Not Found</div>;
  }
  if (isLoading) {
    return (
      <div>
        <SkeletonForm />
      </div>
    );
  }
  if (!data?.apiKey) {
    return (
      <div className="mt-10">
        <Alert variant="destructive">
          <AlertTitle>{`Error: Unable to retrieve api key`}</AlertTitle>
        </Alert>
      </div>
    );
  }
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header={`Regenerate API Key: ${data.apiKey.name}`}
          description={
            'Submitting this form will generate a new token. Be aware that any scripts or applications using this token will need to be updated.'
          }
        />
      }
      containerClassName="mx-24"
    >
      <Form {...form}>
        <form
          onSubmit={form.handleSubmit(onSubmit)}
          className="flex flex-col gap-8"
        >
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
                  The time that this API key will expire.
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
                          date > addYears(startOfDay(new Date()), 1)
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
    </OverviewContainer>
  );
}

async function regenerateApiKey(
  formData: FormValues,
  keyId: string
): Promise<RegenerateAccountApiKeyResponse> {
  const body = new RegenerateAccountApiKeyRequest({
    id: keyId,
    expiresAt: new Timestamp({
      seconds: BigInt(formData.expiresAt.getTime() / 1000),
    }),
  });

  const res = await fetch(`/api/api-keys/account/${keyId}/regenerate`, {
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
  return RegenerateAccountApiKeyResponse.fromJson(await res.json());
}
