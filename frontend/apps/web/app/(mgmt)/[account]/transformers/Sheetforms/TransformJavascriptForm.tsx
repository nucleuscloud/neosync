'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import LearnMoreTag from '@/components/labels/LearnMoreTag';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { Transformer, isUserDefinedTransformer } from '@/shared/transformers';
import { Editor } from '@monaco-editor/react';
import { TransformJavascript } from '@neosync/sdk';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement, useState } from 'react';
import { useFormContext } from 'react-hook-form';
import {
  IsUserJavascriptCodeValid,
  ValidCode,
} from '../../new/transformer/UserDefinedTransformerForms/UserDefinedTransformJavascriptForm';
interface Props {
  index?: number;
  transformer: Transformer;
  setIsSheetOpen?: (val: boolean) => void;
}

export default function TransformJavascriptForm(props: Props): ReactElement {
  const { index, setIsSheetOpen, transformer } = props;

  const fc = useFormContext();

  const codeValue = fc.getValues(
    `mappings.${index}.transformer.config.value.code`
  );
  const [userCode, setUserCode] = useState<string>(codeValue);
  const [isValidatingCode, setIsValidatingCode] = useState<boolean>(false);
  const [isCodeValid, setIsCodeValid] = useState<ValidCode>('null');
  const { resolvedTheme } = useTheme();

  const options = {
    minimap: { enabled: false },
    readOnly: isUserDefinedTransformer(transformer),
  };

  const handleSubmit = () => {
    fc.setValue(
      `mappings.${index}.transformer.config.value`,
      new TransformJavascript({ code: userCode }),
      {
        shouldValidate: false,
      }
    );
    setIsSheetOpen!(false);
  };

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
    <div className="flex flex-col w-full space-y-4 pt-4">
      <FormField
        name={`mappings.${index}.transformer.config.value.preserveLength`}
        render={() => (
          <FormItem>
            <div className="flex flex-row justify-between">
              <div className="space-y-0.5">
                <FormLabel>Transformer Code</FormLabel>
                <div className="text-[0.8rem] text-muted-foreground">
                  Define your own Transformation below using Javascript. The
                  source column value will be available at the{' '}
                  <code className="bg-gray-200 dark:bg-gray-800 text-gray-800 dark:text-gray-300 px-1 py-0.5 rounded">
                    value
                  </code>{' '}
                  keyword. While additional columns can be accessed at{' '}
                  <code className="bg-gray-200 dark:bg-gray-800 text-gray-800 dark:text-gray-300 px-1 py-0.5 rounded">
                    input.{'{'}column_name{'}'}
                  </code>
                  .{' '}
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
                  value={userCode}
                  theme={resolvedTheme === 'dark' ? 'vs-dark' : 'cobalt'}
                  defaultValue={userCode}
                  onChange={(e) => {
                    setUserCode(e ?? '');
                  }}
                  options={options}
                />
              </div>
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
