import { ReactElement } from 'react';

import ConnectionSelectContent from '@/app/(mgmt)/[account]/new/job/connect/ConnectionSelectContent';
import FormErrorMessage from '@/components/FormErrorMessage';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import useMonacoResizer from '@/libs/hooks/monaco/useMonacoResizer';
import useMonacoTheme from '@/libs/hooks/monaco/useMonacoTheme';
import { splitConnections } from '@/libs/utils';
import { Editor } from '@monaco-editor/react';
import { Connection } from '@neosync/sdk';
import { editor } from 'monaco-editor';
import FormHeader from './FormHeader';
import { JobHookSqlFormValues, SqlTimingFormValue } from './validation';

interface Props {
  values: JobHookSqlFormValues;
  setValues(values: JobHookSqlFormValues): void;
  jobConnections: Connection[];
  errors: Record<string, string>;
}

export default function JobConfigSqlForm(props: Props): ReactElement<any> {
  const { values, setValues, jobConnections, errors } = props;
  return (
    <>
      <SelectConnections
        jobConnections={jobConnections}
        selectedConnectionId={values.connectionId}
        setSelectedConnectionId={(updatedId) => {
          setValues({ ...values, connectionId: updatedId });
        }}
        error={errors['config.sql.connectionId']}
      />
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Query"
          description="The SQL query that will be invoked"
          isErrored={!!errors['config.sql.query']}
        />
        <EditSqlQuery
          query={values.query}
          setQuery={(query) => setValues({ ...values, query })}
        />
        <FormErrorMessage message={errors['config.sql.query']} />
      </div>
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Timing"
          description="The lifecycle of when the hook will run"
          htmlFor="timing"
          isErrored={!!errors['config.sql.timing']}
        />
        <Select
          name="timing"
          value={values.timing}
          onValueChange={(newValue) => {
            if (newValue) {
              setValues({ ...values, timing: newValue as SqlTimingFormValue });
            }
          }}
        >
          <SelectTrigger>
            <SelectValue placeholder="Select a timing value" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="preSync">
              Pre Sync - Before first table sync, truncation, schema init
            </SelectItem>
            <SelectItem value="postSync">
              Post Sync - After the last table sync
            </SelectItem>
          </SelectContent>
        </Select>
        <FormErrorMessage message={errors['config.sql.timing']} />
      </div>
    </>
  );
}

interface EditSqlQueryProps {
  query: string;
  setQuery(query: string): void;
}

const editorOptions: editor.IStandaloneEditorConstructionOptions = {
  minimap: { enabled: false },
  lineNumbers: 'off',
};

function EditSqlQuery(props: EditSqlQueryProps): ReactElement<any> {
  const { query, setQuery } = props;

  const theme = useMonacoTheme();
  const { ref, width: editorWidth } = useMonacoResizer();

  return (
    <div
      className="monaco-editor-container border border-gray-700 p-2 rounded-lg"
      ref={ref}
    >
      <Editor
        height="10vh"
        width={editorWidth}
        language="sql"
        theme={theme}
        options={editorOptions}
        value={query}
        onChange={(value) => {
          if (value) {
            setQuery(value);
          }
        }}
      />
    </div>
  );
}

interface SelectConnectionsProps {
  jobConnections: Connection[];

  selectedConnectionId: string;
  setSelectedConnectionId(id: string): void;
  error?: string;
}
function SelectConnections(props: SelectConnectionsProps): ReactElement<any> {
  const {
    jobConnections,
    selectedConnectionId,
    setSelectedConnectionId,
    error,
  } = props;
  const { postgres, mysql, mssql } = splitConnections(jobConnections);
  return (
    <div className="flex flex-col gap-3">
      <FormHeader
        htmlFor="connectionId"
        title="Connection"
        description="The connection that this hook will be invoked against"
        isErrored={!!error}
      />
      <Select
        name="connectionId"
        value={selectedConnectionId}
        onValueChange={(newValue) => {
          if (newValue) {
            setSelectedConnectionId(newValue);
          }
        }}
      >
        <SelectTrigger>
          <SelectValue placeholder="Select a connection..." />
        </SelectTrigger>
        <SelectContent>
          <ConnectionSelectContent
            postgres={postgres}
            mysql={mysql}
            mssql={mssql}
          />
        </SelectContent>
      </Select>
      <FormErrorMessage message={error} />
    </div>
  );
}
