'use client';
import ButtonText from '@/components/ButtonText';
import Spinner from '@/components/Spinner';
import LearnMoreTag from '@/components/labels/LearnMoreTag';
import { useAccount } from '@/components/providers/account-provider';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
} from '@/components/ui/form';
import { yupResolver } from '@hookform/resolvers/yup';
import { Editor } from '@monaco-editor/react';
import { GenerateJavascript } from '@neosync/sdk';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement, useState } from 'react';
import { useForm } from 'react-hook-form';
import {
  IsUserJavascriptCodeValid,
  ValidCode,
} from '../../new/transformer/UserDefinedTransformerForms/UserDefinedTransformJavascriptForm';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps } from './util';
interface Props extends TransformerFormProps<GenerateJavascript> {}

export default function GenerateJavascriptForm(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;
  const account = useAccount();

  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.transformJavascriptConfig),
    defaultValues: {
      code: existingConfig?.code ?? '',
    },
    context: { accountId: account.account?.id },
  });

  const [isValidatingCode, setIsValidatingCode] = useState<boolean>(false);
  const [codeStatus, setCodeStatus] = useState<ValidCode>('null');
  const { resolvedTheme } = useTheme();

  async function handleValidateCode(): Promise<void> {
    if (!account) {
      return;
    }
    setIsValidatingCode(true);

    try {
      const res = await IsUserJavascriptCodeValid(
        form.getValues('code') ?? '',
        account.account?.id ?? ''
      );
      if (res.valid === true) {
        setCodeStatus('valid');
      } else {
        setCodeStatus('invalid');
      }
    } catch (err) {
      console.error(err);
      setCodeStatus('invalid');
    } finally {
      setIsValidatingCode(false);
    }
  }

  return (
    <div className="flex flex-col w-full space-y-4 pt-4">
      <Form {...form}>
        <FormField
          control={form.control}
          name={`code`}
          render={({ field }) => (
            <FormItem>
              <div className="flex flex-row justify-between">
                <div className="space-y-0.5">
                  <FormLabel>Transformer Code</FormLabel>
                  <div className="text-[0.8rem] text-muted-foreground">
                    Define your own Transformation below using Javascript. .{' '}
                    <LearnMoreTag href="https://docs.neosync.dev/transformers/user-defined#custom-code-transformers" />
                  </div>
                </div>
                <div className="flex flex-row gap-2">
                  {codeStatus !== 'null' && (
                    <Badge
                      variant={
                        codeStatus === 'valid' ? 'success' : 'destructive'
                      }
                      className="h-9 px-4 py-2"
                    >
                      <ButtonText
                        leftIcon={
                          codeStatus === 'valid' ? (
                            <CheckCircledIcon />
                          ) : codeStatus === 'invalid' ? (
                            <CrossCircledIcon />
                          ) : null
                        }
                        text={codeStatus}
                      />
                    </Badge>
                  )}
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
                    onChange={field.onChange}
                    options={{
                      minimap: { enabled: false },
                      readOnly: isReadonly,
                    }}
                  />
                </div>
              </FormControl>
            </FormItem>
          )}
        />
        <div className="flex justify-end gap-4">
          <Button
            type="button"
            variant="secondary"
            onClick={() => handleValidateCode()}
          >
            <ButtonText
              leftIcon={isValidatingCode ? <Spinner /> : null}
              text="Validate"
            />
          </Button>
          <Button
            type="button"
            disabled={isReadonly || codeStatus !== 'valid'}
            onClick={(e) => {
              form.handleSubmit((values) => {
                onSubmit(
                  new GenerateJavascript({
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
