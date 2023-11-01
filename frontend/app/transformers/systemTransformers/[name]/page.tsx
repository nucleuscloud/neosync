'use client';
import { handleCustomTransformerForm } from '@/app/new/transformer/CustomTransformerForms/HandleCustomTransformersForm';
import {
  SYSTEM_TRANSFORMER_SCHEMA,
  SystemTransformersSchema,
} from '@/app/new/transformer/schema';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { PageProps } from '@/components/types';
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
import { useGetSystemTransformers } from '@/libs/hooks/useGetSystemTransformers';
import { yupResolver } from '@hookform/resolvers/yup';
import NextLink from 'next/link';
import { useRouter } from 'next/navigation';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { handleTransformerMetadata } from '../../EditTransformerOptions';

export default function ViewSystemTransformers({
  params,
}: PageProps): ReactElement {
  const { data: systemTransformers, isLoading } = useGetSystemTransformers();

  const tName = params?.name ?? '';

  const currentTransformer = systemTransformers?.transformers.find(
    (item) => item.value == tName
  );

  const router = useRouter();

  const form = useForm<SystemTransformersSchema>({
    resolver: yupResolver(SYSTEM_TRANSFORMER_SCHEMA),
    defaultValues: {
      name: '',
      description: '',
      type: '',
      config: { config: { case: '', value: {} } },
    },
    values: {
      name: currentTransformer?.value ?? '',
      description:
        handleTransformerMetadata(currentTransformer?.value).description ?? '',
      type: handleTransformerMetadata(currentTransformer?.value).type ?? '',
      config: {
        config: {
          case: currentTransformer?.config?.config.case,
          value: currentTransformer?.config?.config.value ?? {},
        },
      },
    },
  });

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header={currentTransformer?.value ?? 'System Transformer'}
          extraHeading={
            <NewTransformerButton
              transformer={currentTransformer?.value ?? ''}
            />
          }
        />
      }
    >
      <Form {...form}>
        <form className="space-y-8">
          <div>
            <FormField
              control={form.control}
              name="name"
              disabled={true}
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
                  <FormDescription>The Transformer name</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <div className="pt-10">
              <FormField
                control={form.control}
                name="description"
                disabled={true}
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
            <div className="pt-10">
              <FormField
                control={form.control}
                disabled={true}
                name="type"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Type</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="Transformer type"
                        {...field}
                        className="w-[1000px]"
                      />
                    </FormControl>
                    <FormDescription>The Transformer type.</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>
          <div className="w-[1000px]">
            {handleCustomTransformerForm(currentTransformer?.value, true)}
          </div>
          <div className="flex flex-row justify-start">
            <Button type="button" onClick={() => router.push('/transformers')}>
              Back
            </Button>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

interface NewTransformerButtonProps {
  transformer: string;
}

function NewTransformerButton(props: NewTransformerButtonProps): ReactElement {
  const { transformer } = props;

  return (
    <NextLink href={'/new/transformer'}>
      <Button> Clone Transformer</Button>
    </NextLink>
  );
}
