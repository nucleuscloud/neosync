import { ReactElement, useMemo } from 'react';

import ConnectionSelectContent from '@/app/(mgmt)/[account]/new/job/connect/ConnectionSelectContent';
import { useAccount } from '@/components/providers/account-provider';
import { Label } from '@/components/ui/label';
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
import { JobHookSqlFormValues, SqlTimingFormValue } from './validation';

interface Props {
  values: JobHookSqlFormValues;
  setValues(values: JobHookSqlFormValues): void;
  jobConnectionIds: string[];
}

export default function JobConfigSqlForm(props: Props): ReactElement {
  const { values, setValues, jobConnectionIds } = props;
  return (
    <>
      <SelectConnections
        connectionIds={jobConnectionIds}
        selectedConnectionId={values.connectionId}
        setSelectedConnectionId={(updatedId) => {
          setValues({ ...values, connectionId: updatedId });
        }}
      />
      <div className="flex flex-col gap-3">
        <Label htmlFor="query">Query</Label>
        <EditSqlQuery
          query={values.query}
          setQuery={(query) => setValues({ ...values, query })}
        />
      </div>
      <div className="flex flex-col gap-3">
        <Label htmlFor="timing">Timing</Label>
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
}
function SelectConnections(props: SelectConnectionsProps): ReactElement {
  const { connectionIds, selectedConnectionId, setSelectedConnectionId } =
    props;
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
      <Label htmlFor="connectionId">Connection</Label>
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
    </div>
  );
}
