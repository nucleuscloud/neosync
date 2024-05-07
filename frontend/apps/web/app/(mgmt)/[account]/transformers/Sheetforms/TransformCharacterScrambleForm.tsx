'use client';
import LearnMoreTag from '@/components/labels/LearnMoreTag';
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

import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { yupResolver } from '@hookform/resolvers/yup';
import { TransformCharacterScramble } from '@neosync/sdk';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import {
  ValidRegex,
  ValidateUserRegex,
} from '../../new/transformer/UserDefinedTransformerForms/UserDefinedTransformCharacterScrambleForm';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps } from './util';
interface Props extends TransformerFormProps<TransformCharacterScramble> {}

export default function TransformCharacterScrambleForm(
  props: Props
): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(
      TRANSFORMER_SCHEMA_CONFIGS.transformCharacterScrambleConfig
    ),
    defaultValues: {
      userProvidedRegex: existingConfig?.userProvidedRegex ?? '',
    },
  });

  const [isValidating, setIsValidating] = useState(false);
  const [status, setStatus] = useState<ValidRegex>('null');
  const account = useAccount();

  async function handleValidateCode(): Promise<void> {
    if (!account) {
      return;
    }
    setIsValidating(true);

    try {
      const res = await ValidateUserRegex(
        form.getValues('userProvidedRegex') ?? '',
        account.account?.id ?? ''
      );
      if (res.valid === true) {
        setStatus('valid');
      } else {
        setStatus('invalid');
      }
    } catch (err) {
      setStatus('invalid');
    } finally {
      setIsValidating(false);
    }
  }

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row gap-2 justify-end">
        {status !== 'null' && (
          <Badge
            variant={status === 'valid' ? 'success' : 'destructive'}
            className="h-9 px-4 py-2"
          >
            <ButtonText
              leftIcon={
                status === 'valid' ? (
                  <CheckCircledIcon />
                ) : status === 'invalid' ? (
                  <CrossCircledIcon />
                ) : null
              }
              text={status}
            />
          </Badge>
        )}
      </div>
      <Form {...form}>
        <FormField
          control={form.control}
          name={`userProvidedRegex`}
          render={({ field }) => (
            <FormItem className="rounded-lg border p-3 shadow-sm">
              <div className="flex flex-row items-start justify-between">
                <div className="flex flex-col space-y-2">
                  <FormLabel>Regular Expression</FormLabel>
                  <FormDescription className="w-[90%]">
                    Provide a Go regular expression to match and transform a
                    substring of the value. Leave this blank to transform the
                    entire value. Note: the regex needs to compile in Go.
                  </FormDescription>
                  <LearnMoreTag href="https://docs.neosync.dev/transformers/user-defined#transform-character-scramble" />
                </div>
                <FormControl>
                  <div className="flex flex-col items-center">
                    <Input
                      {...field}
                      type="string"
                      className="min-w-[300px]"
                      disabled={isReadonly}
                    />
                    <FormMessage />
                  </div>
                </FormControl>
              </div>
            </FormItem>
          )}
        />
        <div className="flex justify-end gap-4">
          <Button type="button" onClick={() => handleValidateCode()}>
            <ButtonText
              leftIcon={isValidating ? <Spinner /> : null}
              text="Validate"
            />
          </Button>
          <Button
            type="button"
            disabled={isReadonly || status !== 'valid'}
            onClick={(e) => {
              form.handleSubmit((values) => {
                onSubmit(
                  new TransformCharacterScramble({
                    ...values,
                  })
                );
              })(e);
            }}
          >
            Save
          </Button>
        </div>
      </Form>
    </div>
  );
}
