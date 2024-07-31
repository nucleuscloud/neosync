import { Editor, useMonaco } from '@monaco-editor/react';
import { ReactElement, useEffect } from 'react';

interface Props {}
export default function TestEditor(_: Props): ReactElement {
  const monaco = useMonaco();

  useEffect(() => {
    if (monaco) {
      const provider = monaco.languages.registerCompletionItemProvider(
        'javascript',
        {
          provideCompletionItems: (model, position) => {
            // const textUntilPosition = model.getValueInRange({
            //   startLineNumber: 1,
            //   startColumn: 1,
            //   endLineNumber: position.lineNumber,
            //   endColumn: position.column,
            // });

            // const columnSet = new Set<string>('neosync');

            const word = model.getWordUntilPosition(position);

            const range = {
              startLineNumber: position.lineNumber,
              startColumn: word.startColumn,
              endLineNumber: position.lineNumber,
              endColumn: word.endColumn,
            };

            // const suggestions = Array.from(columnSet).map((name) => ({
            //   label: name, // would be nice if we could add the type here as well?
            //   kind: monaco.languages.CompletionItemKind.Field,
            //   insertText: name,
            //   range: range,
            // }));
            const suggestions = [
              {
                label: 'neosync.transformEmail',
                kind: monaco.languages.CompletionItemKind.Function,
                insertText: 'neosync.transformEmail({${1:email}})',
                insertTextRules:
                  monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                documentation: 'Transforms an email using neosync',
                range: range,
              },
              {
                label: 'neosync.transformPhoneNumber',
                kind: monaco.languages.CompletionItemKind.Function,
                insertText: 'neosync.transformPhoneNumber(${1:phoneNumber})',
                insertTextRules:
                  monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                documentation: 'Transforms a phone number using neosync',
                range: range,
              },
            ];
            return { suggestions: suggestions };
          },
        }
      );
      return () => {
        provider.dispose();
      };
    }
  }, [monaco]);

  return (
    <Editor
      height="600px"
      width="100%"
      language="javascript"
      theme="vs-light"
    />
  );
}
