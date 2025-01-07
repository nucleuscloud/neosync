import { editor } from 'monaco-editor';

interface UseMonacoOnMountReturn {
  onMount(editor: editor.IStandaloneCodeEditor): void;
}

// Standard on mount that auto focuses the editor and moves the cursor to the end of the document
// Only use this if you want the editor to be focused and the cursor to be at the end of the document when it comes into view
// Mostly useful in Dialogs if the editor is the main focus of the dialog
export default function useMonacoOnMount(): UseMonacoOnMountReturn {
  return {
    onMount(editor: editor.IStandaloneCodeEditor): void {
      editor.focus();
      const model = editor.getModel();
      if (model) {
        const lastLine = model.getLineCount();
        const lastColumn = model.getLineMaxColumn(lastLine);
        editor.setPosition({ lineNumber: lastLine, column: lastColumn });
      }
    },
  };
}
