'use client';
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';

import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import LearnMoreTag from '@/components/labels/LearnMoreTag';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import Editor from '@monaco-editor/react';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  CreateUserDefinedTransformerSchema,
  UpdateUserDefinedTransformer,
} from '../schema';
import {
  IsUserJavascriptCodeValid,
  ValidCode,
} from './UserDefinedTransformJavascriptForm';

interface Props {
  isDisabled?: boolean;
}

export default function UserDefinedGenerateJavascriptForm(
  props: Props
): ReactElement {
  const { isDisabled } = props;

  const fc = useFormContext<
    UpdateUserDefinedTransformer | CreateUserDefinedTransformerSchema
  >();

  const options = {
    minimap: { enabled: false },
    readOnly: isDisabled,
  };

  const { resolvedTheme } = useTheme();

  const [userCode, setUserCode] = useState<string>(
    fc.getValues('config.value.code')
  );

  const [isValidatingCode, setIsValidatingCode] = useState<boolean>(false);
  const [isCodeValid, setIsCodeValid] = useState<ValidCode>('null');

  const account = useAccount();

  async function handleValidateCode(): Promise<void> {
    if (!account) {
      return;
    }
    setIsValidatingCode(true);

    try {
      const res = await IsUserJavascriptCodeValid(
        userCode,
        account.account?.id ?? ''
      );
      setIsValidatingCode(false);
      if (res.valid === true) {
        setIsCodeValid('valid');
      } else {
        setIsCodeValid('invalid');
      }
    } catch (err) {
      console.error(err);
      setIsValidatingCode(false);
      setIsCodeValid('invalid');
    }
  }

  return (
    <div className="pt-4">
      <FormField
        name={`config.value.code`}
        control={fc.control}
        render={({ field }) => (
          <FormItem>
            <div className="flex flex-row justify-between">
              <div className="space-y-0.5">
                <FormLabel>Transformer Code</FormLabel>
                <div className="text-[0.8rem] text-muted-foreground">
                  Define your own Transformation below using Javascript.
                  <LearnMoreTag href="https://docs.neosync.dev/transformers/user-defined#custom-code-transformers" />
                </div>
              </div>
              <div className="flex flex-row gap-2">
                {isCodeValid !== 'null' && (
                  <Badge
                    variant={
                      isCodeValid === 'valid' ? 'success' : 'destructive'
                    }
                    className="h-9 px-4 py-2"
                  >
                    <ButtonText
                      leftIcon={
                        isCodeValid === 'valid' ? (
                          <CheckCircledIcon />
                        ) : isCodeValid === 'invalid' ? (
                          <CrossCircledIcon />
                        ) : null
                      }
                      text={isCodeValid === 'invalid' ? 'invalid' : 'valid'}
                    />
                  </Badge>
                )}
                <Button type="button" onClick={handleValidateCode}>
                  <ButtonText
                    leftIcon={isValidatingCode ? <Spinner /> : null}
                    text={'Validate'}
                  />
                </Button>
              </div>
            </div>
            <FormControl>
              <div className="flex flex-col items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm">
                <Editor
                  height="50vh"
                  width="100%"
                  language="javascript"
                  value={field.value}
                  theme={resolvedTheme === 'dark' ? 'vs-dark' : 'cobalt'}
                  defaultValue={field.value}
                  onChange={(e) => {
                    field.onChange(e);
                    setUserCode(e ?? '');
                  }}
                  options={options}
                />
              </div>
            </FormControl>
          </FormItem>
        )}
      />
    </div>
  );
}
