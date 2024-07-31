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
import { useMutation } from '@connectrpc/connect-query';
import { yupResolver } from '@hookform/resolvers/yup';
import { Editor, useMonaco } from '@monaco-editor/react';
import { TransformJavascript } from '@neosync/sdk';
import { validateUserJavascriptCode } from '@neosync/sdk/connectquery';
import { CheckCircledIcon, CrossCircledIcon } from '@radix-ui/react-icons';
import { useTheme } from 'next-themes';
import { ReactElement, useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { ValidCode } from '../../new/transformer/UserDefinedTransformerForms/UserDefinedTransformJavascriptForm';
import { TRANSFORMER_SCHEMA_CONFIGS } from '../../new/transformer/schema';
import { TransformerFormProps } from './util';
interface Props extends TransformerFormProps<TransformJavascript> {}

export default function TransformJavascriptForm(props: Props): ReactElement {
  const { existingConfig, onSubmit, isReadonly } = props;
  const { mutateAsync: validateUserJsCodeAsync } = useMutation(
    validateUserJavascriptCode
  );

  const { account } = useAccount();
  const form = useForm({
    mode: 'onChange',
    resolver: yupResolver(TRANSFORMER_SCHEMA_CONFIGS.transformJavascriptConfig),
    defaultValues: {
      code: existingConfig?.code ?? '',
    },
    context: {
      accountId: account?.id,
      isUserJavascriptCodeValid: validateUserJavascriptCode,
    },
  });

  const [isValidatingCode, setIsValidatingCode] = useState<boolean>(false);
  const [codeStatus, setCodeStatus] = useState<ValidCode>('null');
  const { resolvedTheme } = useTheme();

  const monaco = useMonaco();

  useEffect(() => {
    if (monaco) {
      const provider = monaco.languages.registerCompletionItemProvider('sql', {
        triggerCharacters: [' ', '.'], // Trigger autocomplete on space and dot

        provideCompletionItems: (model, position) => {
          // const textUntilPosition = model.getValueInRange({
          //   startLineNumber: 1,
          //   startColumn: 1,
          //   endLineNumber: position.lineNumber,
          //   endColumn: position.column,
          // });

          const columnSet = new Set<string>('neosync');

          // Check if the last character or word should trigger the auto-complete
          // if (!shouldTriggerAutocomplete(textUntilPosition)) {
          //   return { suggestions: [] };
          // }

          const word = model.getWordUntilPosition(position);

          const range = {
            startLineNumber: position.lineNumber,
            startColumn: word.startColumn,
            endLineNumber: position.lineNumber,
            endColumn: word.endColumn,
          };

          const suggestions = Array.from(columnSet).map((name) => ({
            label: name, // would be nice if we could add the type here as well?
            kind: monaco.languages.CompletionItemKind.Field,
            insertText: name,
            range: range,
          }));

          return { suggestions: suggestions };
        },
      });
      /* disposes of the instance if the component re-renders, otherwise the auto-compelte list just keeps appending the column names to the auto-complete, so you get liek 20 'address' entries for ex. then it re-renders and then it goes to 30 'address' entries
       */
      return () => {
        provider.dispose();
      };
    }
  }, [monaco]);

  async function handleValidateCode(): Promise<void> {
    if (!account) {
      return;
    }
    setIsValidatingCode(true);

    try {
      const res = await validateUserJsCodeAsync({
        accountId: account.id,
        code: form.getValues('code') ?? '',
      });
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
              <div className="flex flex-row justify-between gap-6 items-center">
                <div className="space-y-0.5 w-[90%]">
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
                <div className="flex flex-row gap-2 w-[80px]">
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
                  new TransformJavascript({
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

// function shouldTriggerAutocomplete(text: string): boolean {
//   const trimmedText = text.trim();
//   const textSplit = trimmedText.split(/\s+/);
//   const lastSignificantWord = trimmedText.split(/\s+/).pop()?.toUpperCase();
//   const triggerKeywords = ['neo'];

//   if (textSplit.length == 2 && textSplit[0].toUpperCase() == 'WHERE') {
//     /* since we pre-pend the 'WHERE', we want the autocomplete to show up for the first letter typed
//      which would come through as 'WHERE a' if the user just typed the letter 'a'
//      so the when we split that text, we check if the length is 2 (as a way of checking if the user has only typed one letter or is still on the first word) and if it is and the first word is 'WHERE' which it should be since we pre-pend it, then show the auto-complete */
//     return true;
//   } else {
//     return (
//       triggerKeywords.includes(lastSignificantWord || '') ||
//       triggerKeywords.some((keyword) => trimmedText.endsWith(keyword + ' '))
//     );
//   }
// }
