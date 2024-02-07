'use client';
import { handleUserDefinedTransformerForm } from '@/app/(mgmt)/[account]/new/transformer/UserDefinedTransformerForms/HandleUserDefinedTransformersForm';
import {
  SYSTEM_TRANSFORMER_SCHEMA,
  SystemTransformersSchema,
} from '@/app/(mgmt)/[account]/new/transformer/schema';
import ButtonText from '@/components/ButtonText';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
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
import { Skeleton } from '@/components/ui/skeleton';
import { useGetSystemTransformerBySource } from '@/libs/hooks/useGetSystemTransformerBySource';
import { convertTransformerConfigToForm } from '@/yup-validations/jobs';
import { yupResolver } from '@hookform/resolvers/yup';
import Error from 'next/error';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';

export default function ViewSystemTransformers({
  params,
}: PageProps): ReactElement {
  const sourceParam = params?.name ?? '';
  const { data: systemTransformerData, isLoading } =
    useGetSystemTransformerBySource(sourceParam);
  const { account } = useAccount();
  const systemTransformer = systemTransformerData?.transformer;

  const form = useForm<SystemTransformersSchema>({
    resolver: yupResolver(SYSTEM_TRANSFORMER_SCHEMA),
    values: {
      name: systemTransformer?.name ?? '',
      description: systemTransformer?.description ?? '',
      type: systemTransformer?.dataType ?? '',
      source: systemTransformer?.source ?? '',
      config: convertTransformerConfigToForm(systemTransformer?.config),
    },
  });

  if (isLoading) {
    return <Skeleton />;
  }
  if (!systemTransformer) {
    return <Error statusCode={404} />;
  }

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header={systemTransformer?.name ?? 'System Transformer'}
          extraHeading={
            <CloneTransformerButton source={systemTransformer?.source ?? ''} />
          }
        />
      }
      containerClassName="px-12 md:px-24 lg:px-32"
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
                  <FormDescription>The Transformer name</FormDescription>
                  <FormControl>
                    <Input placeholder="Transformer Name" {...field} />
                  </FormControl>
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
                    <FormDescription>
                      The Transformer decription.
                    </FormDescription>
                    <FormControl>
                      <Input placeholder="Transformer Name" {...field} />
                    </FormControl>
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
                    <FormLabel>Data Type</FormLabel>
                    <FormDescription>The Transformer type.</FormDescription>
                    <FormControl>
                      <Input placeholder="Transformer type" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <div className="pt-10">
              <FormField
                control={form.control}
                disabled={true}
                name="source"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Source</FormLabel>
                    <FormDescription>
                      The unique key associated with the transformer.
                    </FormDescription>
                    <FormControl>
                      <Input placeholder="Transformer Source" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>
          <div>
            {handleUserDefinedTransformerForm(systemTransformer?.source, true)}
          </div>
          <div className="flex flex-row justify-start">
            <NextLink href={`/${account?.name}/transformers?tab=system`}>
              <Button type="button">Back</Button>
            </NextLink>
          </div>
        </form>
      </Form>
    </OverviewContainer>
  );
}

interface CloneTransformerProps {
  source: string;
}

function CloneTransformerButton(props: CloneTransformerProps): ReactElement {
  const { source } = props;
  const { account } = useAccount();
  return (
    <NextLink href={`/${account?.name}/new/transformer?transformer=${source}`}>
      <Button>
        <ButtonText text="Clone Transformer" />
      </Button>
    </NextLink>
  );
}
