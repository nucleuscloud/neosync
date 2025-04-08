import { TanstackQueryProviderIgnore404Errors } from '@/components/providers/query-provider';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useQuery } from '@connectrpc/connect-query';
import { Code, Connection, ConnectionService, JobService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { IoAlertCircleOutline } from 'react-icons/io5';

interface JobRunActivityErrorsProps {
  jobRunId: string;
  jobId: string;
  accountId: string;
}

interface InitSchemaReport {
  ConnectionId: string;
  Errors: { Statement: string; Error: string }[];
}

function parseUint8ArrayToInitSchemaReport(
  data: Uint8Array
): InitSchemaReport[] | null {
  try {
    const jsonString = new TextDecoder().decode(data);
    const parsedData: InitSchemaReport[] = JSON.parse(jsonString);
    return parsedData;
  } catch (error) {
    console.error('Error parsing JSON:', error);
    return null;
  }
}

function parseUint8ArrayToReconcileSchemaReport(
  data: Uint8Array
): InitSchemaReport[] | null {
  try {
    const jsonString = new TextDecoder().decode(data);
    const parsedData: InitSchemaReport = JSON.parse(jsonString);
    return [parsedData];
  } catch (error) {
    console.error('Error parsing JSON:', error);
    return null;
  }
}

export default function JobRunActivityErrors(
  props: JobRunActivityErrorsProps
): ReactElement {
  const { data: jobData } = useQuery(
    JobService.method.getJob,
    { id: props.jobId },
    { enabled: !!props.jobId }
  );

  const job = jobData?.job;
  return (
    <TanstackQueryProviderIgnore404Errors>
      <JobRunInitSchemaErrorViewer
        jobRunId={props.jobRunId}
        accountId={props.accountId}
        externalId="init-schema-report"
        parseUint8ArrayToInitSchemaReport={parseUint8ArrayToInitSchemaReport}
      />
      {job?.destinations?.map((destination) => (
        <JobRunInitSchemaErrorViewer
          key={destination.id}
          jobRunId={props.jobRunId}
          accountId={props.accountId}
          externalId={`init-schema-report-${destination.id}`}
          parseUint8ArrayToInitSchemaReport={parseUint8ArrayToInitSchemaReport}
        />
      ))}
      {job?.destinations?.map((destination) => (
        <JobRunInitSchemaErrorViewer
          key={destination.id}
          jobRunId={props.jobRunId}
          accountId={props.accountId}
          externalId={`reconcile-schema-report-${destination.id}`}
          parseUint8ArrayToInitSchemaReport={
            parseUint8ArrayToReconcileSchemaReport
          }
        />
      ))}
    </TanstackQueryProviderIgnore404Errors>
  );
}

interface JobRunInitSchemaErrorViewerProps {
  jobRunId: string;
  accountId: string;
  externalId: string;
  parseUint8ArrayToInitSchemaReport: (
    data: Uint8Array
  ) => InitSchemaReport[] | null;
}

function JobRunInitSchemaErrorViewer(
  props: JobRunInitSchemaErrorViewerProps
): ReactElement {
  const { jobRunId, accountId, externalId, parseUint8ArrayToInitSchemaReport } =
    props;

  const { data: runContextData, error: runContextError } = useQuery(
    JobService.method.getRunContext,
    {
      id: {
        jobRunId,
        externalId,
        accountId: accountId,
      },
    },
    { enabled: !!jobRunId && !!accountId, retry: false }
  );

  const { data: connectionsData } = useQuery(
    ConnectionService.method.getConnections,
    { accountId },
    { enabled: !!accountId }
  );

  if (runContextError && runContextError.code === Code.NotFound) {
    return <div></div>;
  }

  const runContext = runContextData?.value
    ? parseUint8ArrayToInitSchemaReport(runContextData.value)
    : null;

  const filteredRunContext = runContext?.filter(
    (item) => item.Errors && item.Errors.length > 0
  );

  const connectionsMap =
    connectionsData?.connections?.reduce(
      (acc, connection) => {
        acc[connection.id] = connection;
        return acc;
      },
      {} as Record<string, Connection>
    ) ?? {};

  return (
    <div>
      {filteredRunContext && filteredRunContext.length > 0 && (
        <ScrollArea className="h-[400px] w-full rounded-md border">
          {filteredRunContext.map((connectionError) => (
            <div key={connectionError.ConnectionId} className="p-4">
              <Alert className="border-red-500">
                <div className="flex flex-row items-center gap-2">
                  <IoAlertCircleOutline className="h-6 w-6" />
                  <AlertTitle className="text-lg">
                    Connection:{' '}
                    {connectionsMap[connectionError.ConnectionId]?.name}
                    <span className="text-muted-foreground text-sm pl-4">
                      {connectionError.ConnectionId}
                    </span>
                  </AlertTitle>
                </div>
                <AlertDescription>
                  <h2 className="text-md font-semibold text-muted-foreground mt-2">
                    Schema Initialization Errors
                  </h2>
                  <Accordion type="single" collapsible className="w-full">
                    {connectionError.Errors.map((error, errorIdx) => (
                      <AccordionItem value={`item-${errorIdx}`} key={errorIdx}>
                        <AccordionTrigger className="text-left">
                          <div className="font-medium">{error.Error}</div>
                        </AccordionTrigger>
                        <AccordionContent>
                          <div className="space-y-2">
                            <div className="rounded-md bg-muted p-3">
                              <p className="font-medium">Statement:</p>
                              <pre className="mt-2 whitespace-pre-wrap text-sm">
                                {error.Statement}
                              </pre>
                            </div>
                          </div>
                        </AccordionContent>
                      </AccordionItem>
                    ))}
                  </Accordion>
                </AlertDescription>
              </Alert>
            </div>
          ))}
        </ScrollArea>
      )}
    </div>
  );
}
