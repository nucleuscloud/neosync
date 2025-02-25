import useMonacoOnMount from '@/libs/hooks/monaco/useMonacoOnMount';
import useMonacoResizer from '@/libs/hooks/monaco/useMonacoResizer';
import useMonacoTheme from '@/libs/hooks/monaco/useMonacoTheme';
import { Editor, useMonaco } from '@monaco-editor/react';
import { editor, IRange, languages } from 'monaco-editor';
import { ReactElement, useEffect } from 'react';

interface Props {
  whereClause: string;
  onWhereChange(whereClause: string): void;
  columns: string[];
}

// options for the sql editor
const BASE_EDITOR_OPTS: editor.IStandaloneEditorConstructionOptions = {
  minimap: { enabled: false },
  roundedSelection: false,
  scrollBeyondLastLine: false,
  renderLineHighlight: 'none' as const,
  overviewRulerBorder: false,
  overviewRulerLanes: 0,
  lineNumbers: 'on',
};

export default function WhereEditor(props: Props): ReactElement {
  const { whereClause, onWhereChange, columns } = props;

  const theme = useMonacoTheme();
  const { ref, width: editorWidth } = useMonacoResizer();
  const { onMount } = useMonacoOnMount();
  useAutocomplete(columns);

  return (
    <div
      className="flex flex-col items-center justify-between rounded-lg border dark:border-gray-700 p-3 shadow-xs"
      ref={ref}
    >
      <Editor
        height="10vh"
        width={editorWidth}
        language="sql"
        value={constructWhere(whereClause)}
        theme={theme}
        onChange={(e) => onWhereChange(e?.replace('WHERE ', '') ?? '')}
        options={BASE_EDITOR_OPTS}
        onMount={onMount}
      />
    </div>
  );
}

// enables auto complete of the columns in the where clause
function useAutocomplete(columns: string[]): void {
  const monaco = useMonaco();
  useEffect(() => {
    if (!monaco) {
      return;
    }

    const columnSet = new Set<string>(columns);

    const provider = monaco.languages.registerCompletionItemProvider('sql', {
      triggerCharacters: [' ', '.'], // Trigger autocomplete on space and dot

      provideCompletionItems(model, position) {
        const textUntilPosition = model.getValueInRange({
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: position.lineNumber,
          endColumn: position.column,
        });

        // Check if the last character or word should trigger the auto-complete
        if (!shouldTriggerAutocomplete(textUntilPosition)) {
          return { suggestions: [] };
        }

        const word = model.getWordUntilPosition(position);

        const range: IRange = {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: word.startColumn,
          endColumn: word.endColumn,
        };

        const suggestions: languages.CompletionItem[] = Array.from(
          columnSet
        ).map(
          (name): languages.CompletionItem => ({
            label: name, // would be nice if we could add the type here as well?
            kind: monaco.languages.CompletionItemKind.Field,
            insertText: name,
            range: range,
          })
        );

        return { suggestions: suggestions };
      },
    });
    /* disposes of the instance if the component re-renders, otherwise the auto-compelte list just keeps appending the column names to the auto-complete, so you get liek 20 'address' entries for ex. then it re-renders and then it goes to 30 'address' entries
     */
    return () => provider.dispose();
  }, [monaco, columns]);
}

function constructWhere(whereValue: string): string {
  if (!whereValue) return '';
  return whereValue.startsWith('WHERE ') ? whereValue : `WHERE ${whereValue}`;
}

function shouldTriggerAutocomplete(text: string): boolean {
  const trimmedText = text.trim();
  const textSplit = trimmedText.split(/\s+/);
  const lastSignificantWord = trimmedText.split(/\s+/).pop()?.toUpperCase();
  const triggerKeywords = ['SELECT', 'WHERE', 'AND', 'OR', 'FROM'];

  if (textSplit.length == 2 && textSplit[0].toUpperCase() == 'WHERE') {
    /* since we pre-pend the 'WHERE', we want the autocomplete to show up for the first letter typed
     which would come through as 'WHERE a' if the user just typed the letter 'a'
     so the when we split that text, we check if the length is 2 (as a way of checking if the user has only typed one letter or is still on the first word) and if it is and the first word is 'WHERE' which it should be since we pre-pend it, then show the auto-complete */
    return true;
  } else {
    return (
      triggerKeywords.includes(lastSignificantWord || '') ||
      triggerKeywords.some((keyword) => trimmedText.endsWith(keyword + ' '))
    );
  }
}
