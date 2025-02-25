'use client';
import TransformerForm from '@/app/(mgmt)/[account]/new/transformer/TransformerForms/TransformerForm';
import LearnMoreLink from '@/components/labels/LearnMoreLink';
import { useAccount } from '@/components/providers/account-provider';
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
import { Select, SelectTrigger, SelectValue } from '@/components/ui/select';
import { getErrorMessage, getTransformerSourceString } from '@/util/util';
import {
  convertTransformerConfigSchemaToTransformerConfig,
  convertTransformerConfigToForm,
} from '@/yup-validations/jobs';
import {
  EditUserDefinedTransformerFormContext,
  UpdateUserDefinedTransformerFormValues,
} from '@/yup-validations/transformer-validations';
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { TransformersService, UserDefinedTransformer } from '@neosync/sdk';
import NextLink from 'next/link';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import { constructDocsLink } from '../../EditTransformerOptions';

interface Props {
  currentTransformer: UserDefinedTransformer;
  onUpdated(transformer: UserDefinedTransformer): void;
}

export default function UpdateTransformerForm(props: Props): ReactElement<any> {
  const { currentTransformer, onUpdated } = props;
  const { account } = useAccount();
  const { mutateAsync: isTransformerNameAvailableAsync } = useMutation(
    TransformersService.method.isTransformerNameAvailable
  );
  const { mutateAsync: isJavascriptCodeValid } = useMutation(
    TransformersService.method.validateUserJavascriptCode
  );

  const form = useForm<
    UpdateUserDefinedTransformerFormValues,
    EditUserDefinedTransformerFormContext
  >({
    mode: 'onChange',
    resolver: yupResolver(UpdateUserDefinedTransformerFormValues),
    values: {
      name: currentTransformer?.name ?? '',
      description: currentTransformer?.description ?? '',
      id: currentTransformer?.id ?? '',
      config: convertTransformerConfigToForm(currentTransformer.config),
    },
    context: {
      name: currentTransformer?.name,
      accountId: account?.id ?? '',
      isTransformerNameAvailable: isTransformerNameAvailableAsync,
      isUserJavascriptCodeValid: isJavascriptCodeValid,
    },
  });
  const { mutateAsync } = useMutation(
    TransformersService.method.updateUserDefinedTransformer
  );

  async function onSubmit(
    values: UpdateUserDefinedTransformerFormValues
  ): Promise<void> {
    if (!account || !currentTransformer) {
      return;
    }
    try {
      const transformer = await mutateAsync({
        transformerId: currentTransformer.id,
        description: values.description,
        name: values.name,
        transformerConfig: convertTransformerConfigSchemaToTransformerConfig(
          values.config
        ),
      });
      toast.success('Successfully updated transformer!');
      if (transformer.transformer) {
        onUpdated(transformer.transformer);
      }
    } catch (err) {
      console.error(err);
      toast.error('Unable to update transformer', {
        description: getErrorMessage(err),
      });
    }
  }

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)}>
        <FormField
          name="source"
          render={() => (
            <FormItem>
              <FormLabel>Source Transformer</FormLabel>
              <FormDescription>
                The system transformer source.{' '}
                <LearnMoreLink
                  href={constructDocsLink(currentTransformer.source)}
                />
              </FormDescription>
              <FormControl>
                <Select disabled={true}>
                  <SelectTrigger>
                    <SelectValue
                      placeholder={getTransformerSourceString(
                        currentTransformer.source
                      )}
                    />
                  </SelectTrigger>
                </Select>
              </FormControl>
            </FormItem>
          )}
        />
        <div className="pt-8">
          <FormField
            control={form.control}
            name="name"
            render={({ field: { onChange, ...field } }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormDescription>
                  The unique name of the Transformer.
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
                <FormMessage />
              </FormItem>
            )}
          />
          <div>
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormDescription>The Transformer decription.</FormDescription>
                  <FormControl>
                    <Input placeholder="Transformer Name" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </div>
        <div>
          <FormField
            control={form.control}
            name="config"
            render={({ field }) => (
              <FormItem>
                <FormControl>
                  <TransformerForm
                    value={convertTransformerConfigSchemaToTransformerConfig(
                      field.value
                    )}
                    setValue={(newValue) => {
                      field.onChange(convertTransformerConfigToForm(newValue));
                    }}
                    disabled={false}
                    errors={form.formState.errors}
                  />
                </FormControl>
              </FormItem>
            )}
          />
        </div>
        <div className="flex flex-row justify-between py-4">
          <NextLink href={`/${account?.name}/transformers?tab=ud`}>
            <Button type="button">Back</Button>
          </NextLink>
          <Button type="submit">Save</Button>
        </div>
      </form>
    </Form>
  );
}
