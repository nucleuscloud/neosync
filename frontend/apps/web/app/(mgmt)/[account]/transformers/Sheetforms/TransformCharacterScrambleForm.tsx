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
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';

import { TransformCharacterScramble } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useForm } from 'react-hook-form';
import { TransformerFormProps } from './util';
interface Props extends TransformerFormProps<TransformCharacterScramble> {}

export default function TransformCharacterScrambleForm(
  props: Props
): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;

  const form = useForm({
    mode: 'onChange',
    defaultValues: {
      ...existingConfig,
    },
  });

  // async function handleValidateCode(): Promise<void> {
  //   if (!account) {
  //     return;
  //   }
  //   setIsValidatingRegex(true);

  //   try {
  //     const res = await ValidateUserRegex(userRegex, account.account?.id ?? '');
  //     setIsValidatingRegex(false);
  //     if (res.valid === true) {
  //       setIsRegexValid('valid');
  //     } else {
  //       setIsRegexValid('invalid');
  //     }
  //   } catch (err) {
  //     setIsValidatingRegex(false);
  //     setIsRegexValid('invalid');
  //   }
  // }

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      {/* <div className="flex flex-row gap-2 justify-end">
        {isRegexValid !== 'null' && (
          <Badge
            variant={isRegexValid === 'valid' ? 'success' : 'destructive'}
            className="h-9 px-4 py-2"
          >
            <ButtonText
              leftIcon={
                isRegexValid === 'valid' ? (
                  <CheckCircledIcon />
                ) : isRegexValid === 'invalid' ? (
                  <CrossCircledIcon />
                ) : null
              }
              text={isRegexValid === 'invalid' ? 'invalid' : 'valid'}
            />
          </Badge>
        )}
        <Button type="button" onClick={handleValidateCode}>
          <ButtonText
            leftIcon={isValidatingRegex ? <Spinner /> : null}
            text={'Validate'}
          />
        </Button>
      </div> */}
      <Form {...form}>
        <FormField
          control={form.control}
          name={`userProvidedRegex`}
          render={({ field }) => (
            <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
              <div className="space-y-0.5">
                <FormLabel>Regular Expression</FormLabel>
                <FormDescription className="w-[90%]">
                  Provide a Go regular expression to match and transform a
                  substring of the value. Leave this blank to transform the
                  entire value. Note: the regex needs to compile in Go.
                </FormDescription>
                <LearnMoreTag href="https://docs.neosync.dev/transformers/user-defined#transform-character-scramble" />
              </div>
              <FormControl>
                <Input
                  {...field}
                  type="string"
                  className="min-w-[300px]"
                  disabled={isReadonly}
                />
              </FormControl>
            </FormItem>
          )}
        />
        <div className="flex justify-end">
          <Button
            type="button"
            disabled={isReadonly}
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
