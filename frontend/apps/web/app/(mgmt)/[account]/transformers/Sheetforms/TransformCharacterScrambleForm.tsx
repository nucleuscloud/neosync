'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import LearnMoreTag from '@/components/labels/LearnMoreTag';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';

import { useAccount } from '@/components/providers/account-provider';
import { Transformer, isUserDefinedTransformer } from '@/shared/transformers';
import { TransformCharacterScramble } from '@neosync/sdk';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  ValidRegex,
  ValidateUserRegex,
} from '../../new/transformer/UserDefinedTransformerForms/UserDefinedTransformCharacterScrambleForm';
interface Props {
  index?: number;
  transformer: Transformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function TransformCharacterScrambleForm(
  props: Props
): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const regexValue = fc.getValues(
    `mappings.${index}.transformer.config.value.regex`
  );

  const [regex, setRegex] = useState<string>(regexValue);

  const [userRegex, setUserRegex] = useState<string>(
    fc.getValues('config.value.regex')
  );

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.value`,
      new TransformCharacterScramble({ userProvidedRegex: regex }),
      {
        shouldValidate: false,
      }
    );

    setIsSheetOpen!(false);
  };
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
        name={`mappings.${index}.transformer.config.value.regex`}
        render={() => (
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
              <Input
                type="string"
                className="min-w-[300px]"
                value={regex}
                disabled={isUserDefinedTransformer(transformer)}
                onChange={(event) => {
                  setRegex(event.target.value);
                  setUserRegex(regex);
                }}
              />
            </FormControl>
          </FormItem>
        )}
      />
      <div className="flex justify-end">
        <Button type="button" onClick={handleSubmit}>
          Save
        </Button>
      </div>
    </div>
  );
}
