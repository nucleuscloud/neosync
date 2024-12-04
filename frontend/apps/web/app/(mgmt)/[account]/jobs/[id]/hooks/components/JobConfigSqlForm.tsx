import { ReactElement, useMemo } from 'react';

import ConnectionSelectContent from '@/app/(mgmt)/[account]/new/job/connect/ConnectionSelectContent';
import FormErrorMessage from '@/components/FormErrorMessage';
import { useAccount } from '@/components/providers/account-provider';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { splitConnections } from '@/libs/utils';
import { useQuery } from '@connectrpc/connect-query';
import { Editor } from '@monaco-editor/react';
import { getConnections } from '@neosync/sdk/connectquery';
import { editor } from 'monaco-editor';
import { useTheme } from 'next-themes';
import FormHeader from './FormHeader';
import { JobHookSqlFormValues, SqlTimingFormValue } from './validation';

interface Props {
  values: JobHookSqlFormValues;
  setValues(values: JobHookSqlFormValues): void;
  jobConnectionIds: string[];
  errors: Record<string, string>;
}

export default function JobConfigSqlForm(props: Props): ReactElement {
  const { values, setValues, jobConnectionIds, errors } = props;
  return (
    <>
      <SelectConnections
        connectionIds={jobConnectionIds}
        selectedConnectionId={values.connectionId}
        setSelectedConnectionId={(updatedId) => {
          setValues({ ...values, connectionId: updatedId });
        }}
        error={errors.connectionId}
      />
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Query"
          description="The SQL query that will be invoked"
          isErrored={!!errors.query}
        />
        <EditSqlQuery
          query={values.query}
          setQuery={(query) => setValues({ ...values, query })}
        />
        <FormErrorMessage message={errors.query} />
      </div>
      <div className="flex flex-col gap-3">
        <FormHeader
          title="Timing"
          description="The lifecycle of when the hook will run"
          htmlFor="timing"
          isErrored={!!errors.timing}
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
            <SelectItem value="preSync">Pre Sync</SelectItem>
            <SelectItem value="postSync">Post Sync</SelectItem>
          </SelectContent>
        </Select>
        <FormErrorMessage message={errors.timing} />
      </div>
    </>
  );
}

interface EditSqlQueryProps {
  query: string;
  setQuery(query: string): void;
}

const sqlEditorOptions: editor.IStandaloneEditorConstructionOptions = {
  minimap: { enabled: false },
  wordWrap: 'on',
  lineNumbers: 'off',
};

function EditSqlQuery(props: EditSqlQueryProps): ReactElement {
  const { query, setQuery } = props;

  const { resolvedTheme } = useTheme();

  return (
    <div>
      <Editor
        height="5vh"
        width="100%"
        language="sql"
        theme={resolvedTheme === 'dark' ? 'vs-dark' : 'cobalt'}
        options={sqlEditorOptions}
        value={query}
        onChange={(updatedValue) => setQuery(updatedValue ?? '')}
      />
    </div>
  );
}

interface SelectConnectionsProps {
  connectionIds: string[];

  selectedConnectionId: string;
  setSelectedConnectionId(id: string): void;
  error?: string;
}
function SelectConnections(props: SelectConnectionsProps): ReactElement {
  const {
    connectionIds,
    selectedConnectionId,
    setSelectedConnectionId,
    error,
  } = props;
  const { account } = useAccount();

  const { data: connectionsResp, isLoading } = useQuery(
    getConnections,
    { accountId: account?.id },
    { enabled: !!account?.id }
  );
  const { postgres, mysql, mssql } = useMemo(() => {
    const connections = connectionsResp?.connections ?? [];
    const uniqueConnectionIds = new Set(connectionIds);
    const filtered = connections.filter((c) => uniqueConnectionIds.has(c.id));
    return splitConnections(filtered);
  }, [connectionsResp?.connections, isLoading, connectionIds]);

  if (isLoading) {
    return <Skeleton />;
  }

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
