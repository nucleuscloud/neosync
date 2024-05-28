'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import LearnMoreTag from '@/components/labels/LearnMoreTag';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  ValidateUserRegexCodeRequest,
  ValidateUserRegexCodeResponse,
} from '@neosync/sdk';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';

import { Button } from '@/components/ui/button';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';
interface Props {
  isDisabled?: boolean;
}

export type ValidRegex = 'valid' | 'invalid' | 'null';

export default function UserDefinedTransformCharacterScrambleForm(
  props: Props
): ReactElement {
  const { isDisabled } = props;

  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();

  const [userRegex, setRegex] = useState<string>(
    fc.getValues('config.value.userProvidedRegex')
  );

  const [isValidatingRegex, setIsValidatingRegex] = useState<boolean>(false);
  const [isRegexValid, setIsRegexValid] = useState<ValidRegex>('null');

  const account = useAccount();

  async function handleValidateCode(): Promise<void> {
    if (!account) {
      return;
    }
    setIsValidatingRegex(true);

    try {
      const res = await ValidateUserRegex(userRegex, account.account?.id ?? '');
      setIsValidatingRegex(false);
      if (res.valid === true) {
        setIsRegexValid('valid');
      } else {
        setIsRegexValid('invalid');
      }
    } catch (err) {
      setIsValidatingRegex(false);
      setIsRegexValid('invalid');
    }
  }

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <div className="flex flex-row gap-2 justify-end">
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
      </div>
      <FormField
        name={`config.value.userProvidedRegex`}
        render={({ field }) => (
          <FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
            <div className="space-y-0.5">
              <FormLabel>Regular Expression</FormLabel>
              <FormDescription className="w-[90%]">
                Provide a Go regular expression to match and transform a
                substring of the value. Leave this blank to transform the entire
                value. Note: the regex needs to compile in Go.
              </FormDescription>
              <LearnMoreTag href="https://docs.neosync.dev/transformers/user-defined#transform-character-scramble" />
            </div>
            <FormControl>
              <div className="w-[300px]">
                <Input
                  type="string"
                  value={field.value}
                  onChange={(e) => {
                    field.onChange(e);
                    setRegex(e.target.value ?? '');
                  }}
                  disabled={isDisabled}
                />
              </div>
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}

export async function ValidateUserRegex(
  regex: string,
  accountId: string
): Promise<ValidateUserRegexCodeResponse> {
  const body = new ValidateUserRegexCodeRequest({
    userProvidedRegex: regex,
    accountId: accountId,
  });
  const res = await fetch(
    `/api/accounts/${accountId}/transformers/user-defined/validate-regex`,
    {
      method: 'POST',
      headers: {
        'content-type': 'application/json',
      },
      body: JSON.stringify(body),
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return ValidateUserRegexCodeResponse.fromJson(await res.json());
}
