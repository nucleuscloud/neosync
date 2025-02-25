'use client';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Alert } from '@/components/ui/alert';
import { Skeleton } from '@/components/ui/skeleton';
import { useQuery } from '@connectrpc/connect-query';
import { JobMapping, JobService } from '@neosync/sdk';
import { ReactElement } from 'react';
import { IoAlertCircleOutline } from 'react-icons/io5';

interface Props {
  connectionId: string;
  jobmappings?: JobMapping[];
  initTableSchemaEnabled: boolean;
}

export default function DestinationAlerts({
  connectionId,
  jobmappings,
  initTableSchemaEnabled,
}: Props): ReactElement {
  const { data: validateSchemaResponse, isLoading: isValidatingSchema } =
    useQuery(
      JobService.method.validateSchema,
      {
        connectionId: connectionId,
        mappings: jobmappings,
      },
      {
        enabled: !!jobmappings && !!connectionId,
      }
    );

  const columnErrors =
    !!validateSchemaResponse?.missingColumns?.length ||
    !!validateSchemaResponse?.extraColumns?.length;

  const tableErrors =
    !!validateSchemaResponse?.missingTables?.length ||
    !!validateSchemaResponse?.missingSchemas?.length;

  return (
    <div className="flex flex-col gap-2">
      {isValidatingSchema && <Skeleton className="w-full h-24 rounded-lg" />}
      {!isValidatingSchema && columnErrors && (
        <div>
          <Alert className="border-red-400">
            <Accordion type="single" collapsible className="w-full">
              <AccordionItem value={`table-error`} key={1}>
                <AccordionTrigger className="text-left">
                  <div className="font-medium flex flex-row items-center gap-2">
                    <IoAlertCircleOutline className="h-6 w-6" />
                    Found issues with columns in schema. Please resolve before
                    next job run.
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-2">
                    <div className="rounded-md bg-muted p-3">
                      <pre className="mt-2 whitespace-pre-wrap text-sm">
                        {(() => {
                          const { missingColumns, extraColumns } =
                            validateSchemaResponse || {};
                          let output = '';
                          if (missingColumns && missingColumns.length > 0) {
                            output += 'Columns Missing in Destination:\n';
                            missingColumns.forEach((col) => {
                              output += ` - ${col.schema}.${col.table}.${col.column}\n`;
                            });
                            output += '\n';
                          }
                          if (extraColumns && extraColumns.length > 0) {
                            output += 'Columns Not Found in Job Mappings:\n';
                            extraColumns.forEach((col) => {
                              output += ` - ${col.schema}.${col.table}.${col.column}\n`;
                            });
                          }
                          return output || 'No differences found.';
                        })()}
                      </pre>
                    </div>
                  </div>
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          </Alert>
        </div>
      )}
      {!initTableSchemaEnabled && !isValidatingSchema && tableErrors && (
        <div>
          <Alert className="border-red-400">
            <Accordion type="single" collapsible className="w-full">
              <AccordionItem value={`table-error`} key={1}>
                <AccordionTrigger className="text-left">
                  <div className="font-medium flex flex-row items-center gap-2">
                    <IoAlertCircleOutline className="h-6 w-6" />
                    This destination is missing tables found in Job Mappings.
                    Either enable Init Table Schema or create the tables
                    manually.
                  </div>
                </AccordionTrigger>
                <AccordionContent>
                  <div className="space-y-2">
                    <div className="rounded-md bg-muted p-3">
                      <pre className="mt-2 whitespace-pre-wrap text-sm">
                        {(() => {
                          const { missingSchemas, missingTables } =
                            validateSchemaResponse || {};
                          let output = '';
                          if (missingSchemas && missingSchemas.length > 0) {
                            output += 'Schemas Missing in Destination:\n';
                            missingSchemas.forEach((schema) => {
                              output += ` - ${schema}\n`;
                            });
                            output += '\n';
                          }
                          if (missingTables && missingTables.length > 0) {
                            output += 'Tables Missing in Destination:\n';
                            missingTables.forEach((table) => {
                              output += ` - ${table.schema}.${table.table}\n`;
                            });
                            output += '\n';
                          }
                          return output || 'No differences found.';
                        })()}
                      </pre>
                    </div>
                  </div>
                </AccordionContent>
              </AccordionItem>
            </Accordion>
          </Alert>
        </div>
      )}
    </div>
  );
}
