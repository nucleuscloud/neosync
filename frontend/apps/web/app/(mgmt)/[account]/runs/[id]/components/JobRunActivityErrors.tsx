import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { ScrollArea } from '@/components/ui/scroll-area';
import { ReactElement } from 'react';
import { IoAlertCircle } from 'react-icons/io5';

interface JobRunActivityErrorsProps {}

// SELECT
// convert_from(value, 'UTF8') AS text_content
// --     encode(value, 'hex') AS hex_content,
// --     encode(value, 'base64') AS base64_content,
// --     encode(value, 'escape') AS escaped_content
//  FROM neosync_api.runcontexts where workflow_id = '5febf1b2-a984-45cd-b49b-a5b459b40d49-2025-01-18T05:34:32Z'
// and external_id = 'init-schema-report';

export default function JobRunActivityErrors(
  props: JobRunActivityErrorsProps
): ReactElement {
  const errors = [
    {
      ConnectionId: '8411a2e3-38e8-4711-b784-d331360b726e',
      Errors: [
        {
          Statement: 'CREATE EXTENSION "citext" VERSION "1.6";',
          Error: 'ERROR: extension "citext" already exists (SQLSTATE 42710)',
        },
        {
          Statement: 'CREATE EXTENSION "cube" VERSION "1.5" SCHEMA "test";',
          Error: 'ERROR: schema "test" does not exist',
        },
        {
          Statement: 'CREATE EXTENSION "uuid-ossp" VERSION "1.1";',
          Error:
            'ERROR: could not open extension control file "/usr/share/postgresql/15/extension/uuid-ossp.control": No such file or directory',
        },
      ],
    },
  ];
  return (
    // <Card className="w-full max-w-4xl">
    //   {/* <CardHeader>
    //     <CardTitle>Job Run Details</CardTitle>
    //   </CardHeader> */}
    //   <CardContent>
    <div className="w-full max-w-4xl">
      {errors.length > 0 && (
        <ScrollArea className="h-[400px] w-full rounded-md border">
          {errors.map((connectionError, idx) => (
            <div key={connectionError.ConnectionId} className="p-4">
              <Alert variant="info">
                <IoAlertCircle className="h-4 w-4" />
                <AlertTitle className="font-mono text-sm">
                  Connection ID: {connectionError.ConnectionId}
                </AlertTitle>
                <AlertDescription>
                  <Accordion type="single" collapsible className="w-full">
                    {connectionError.Errors.map((error, errorIdx) => (
                      <AccordionItem value={`item-${errorIdx}`} key={errorIdx}>
                        <AccordionTrigger className="text-left">
                          <div className="font-medium">
                            Error {errorIdx + 1}
                          </div>
                        </AccordionTrigger>
                        <AccordionContent>
                          <div className="space-y-2">
                            <div className="rounded-md bg-destructive/10 p-3">
                              <p className="font-medium">Error Message:</p>
                              <p className="font-mono text-sm">{error.Error}</p>
                            </div>
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
