'use client';
import { FormLabel } from '@/components/ui/form';

import ButtonText from '@/components/ButtonText';
import FormErrorMessage from '@/components/FormErrorMessage';
import Spinner from '@/components/Spinner';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { useReadNeosyncTransformerDeclarationFile } from '@/libs/hooks/useReadNeosyncTransfomerDeclarationFile';
import { create } from '@bufbuild/protobuf';
import { useMutation } from '@connectrpc/connect-query';
import Editor, { useMonaco } from '@monaco-editor/react';
import {
  GenerateJavascript,
  GenerateJavascriptSchema,
  TransformersService,
} from '@neosync/sdk';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement, useEffect, useState } from 'react';
import { ValidCode } from './TransformJavascriptForm';
import { TransformerConfigProps } from './util';

interface Props extends TransformerConfigProps<GenerateJavascript> {}

export default function GenerateJavascriptForm(props: Props): ReactElement {
  const { value, setValue, isDisabled, errors } = props;

  const options = {
    minimap: { enabled: false },
    readOnly: isDisabled,
  };

  const { resolvedTheme } = useTheme();
  const monaco = useMonaco();
  const { data: fileContent } = useReadNeosyncTransformerDeclarationFile();
  useEffect(() => {
    if (monaco && fileContent) {
      monaco.languages.typescript.javascriptDefaults.addExtraLib(
        fileContent,
        'neosync-transformer.d.ts'
      );
    }
  }, [monaco, fileContent]);

  const [isValidatingCode, setIsValidatingCode] = useState<boolean>(false);
  const [isCodeValid, setIsCodeValid] = useState<ValidCode>('null');

  const { account } = useAccount();
  const { mutateAsync: validateUserJsCodeAsync } = useMutation(
    TransformersService.method.validateUserJavascriptCode
  );

  async function handleValidateCode(): Promise<void> {
    if (!account) {
      return;
    }
    setIsValidatingCode(true);

    try {
      const res = await validateUserJsCodeAsync({
        code: value.code,
      });
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
    <div>
      <div className="flex flex-row items-center justify-between">
        <div className="space-y-0.5">
          <FormLabel>Transformer Code</FormLabel>
          <div className="text-sm text-muted-foreground">
            Define your own Transformation below using Javascript
          </div>
        </div>
        <div className="flex flex-row gap-2">
          {isCodeValid !== 'null' && (
            <Badge
              variant={isCodeValid === 'valid' ? 'success' : 'destructive'}
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
          <Button
            type="button"
            variant="secondary"
            onClick={handleValidateCode}
          >
            <ButtonText
              leftIcon={isValidatingCode ? <Spinner /> : null}
              text={'Validate'}
            />
          </Button>
        </div>
      </div>
      <div className="flex flex-col items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-sm mt-4">
        <Editor
          height="50vh"
          width="100%"
          language="javascript"
          value={value.code}
          theme={resolvedTheme === 'dark' ? 'vs-dark' : 'cobalt'}
          onChange={(e) => {
            setValue(
              create(GenerateJavascriptSchema, { ...value, code: e ?? '' })
            );
          }}
          options={options}
        />
      </div>
      <FormErrorMessage message={errors?.code?.message} />
    </div>
  );
}
