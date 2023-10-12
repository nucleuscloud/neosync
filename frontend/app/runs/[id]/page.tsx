'use client';
import { PageProps } from '@/components/types';

import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import SkeletonProgress from '@/components/skeleton/SkeletonProgress';
import { Alert, AlertTitle } from '@/components/ui/alert';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { useGetJobRun } from '@/libs/hooks/useGetJobRun';
import { useGetJobRunEvents } from '@/libs/hooks/useGetJobRunEvents';
import { formatDateTime, formatDateTimeMilliseconds } from '@/util/util';
import { ArrowRightIcon } from '@radix-ui/react-icons';
import { useRouter } from 'next/navigation';
import { ReactElement, useState } from 'react';
import { JOB_RUN_STATUS } from '../components/status';

export default function Page({ params }: PageProps): ReactElement {
  const id = params?.id ?? '';
  const { data, isLoading } = useGetJobRun(id);
  const { data: jobRunEvents, isLoading: jobRunEventsLoading } =
    useGetJobRunEvents(id);

  const jobRun = data?.jobRun;
  const events = jobRunEvents?.events || [];

  const [isOpen, setIsOpen] = useState(events.map((_) => 0));

  const status = JOB_RUN_STATUS.find(
    (status) => status.value === jobRun?.status
  );

  function onRowClick(index: number): void {
    const newOpen = [...isOpen];
    newOpen[index] = isOpen[index] == 1 ? 0 : 1;
    setIsOpen(newOpen);
  }

  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Job Run Details"
          description={jobRun?.id || ''}
          extraHeading={<ButtonLink jobId={jobRun?.jobId} />}
        />
      }
      containerClassName="runs-page"
    >
      {isLoading ? (
        <div className="space-y-24">
          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4`}
          >
            <Skeleton className="w-full h-24 rounded-lg" />
            <Skeleton className="w-full h-24 rounded-lg" />
            <Skeleton className="w-full h-24 rounded-lg" />
            <Skeleton className="w-full h-24 rounded-lg" />
          </div>

          <SkeletonProgress />
        </div>
      ) : (
        <div className="space-y-12">
          <div
            className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4`}
          >
            <StatCard header="Status" content={status?.badge} />
            <StatCard
              header="Start Time"
              content={formatDateTime(jobRun?.startedAt?.toDate())}
            />
            <StatCard
              header="Completion Time"
              content={formatDateTime(jobRun?.completedAt?.toDate())}
            />
            <StatCard
              header="Duration"
              content={getDuration(
                jobRun?.completedAt?.toDate(),
                jobRun?.startedAt?.toDate()
              )}
            />
          </div>
          <div className="space-y-4">
            {jobRun?.pendingActivities.map((a) => {
              if (a.lastFailure) {
                return (
                  <AlertDestructive
                    key={a.activityName}
                    title={a.activityName}
                    description={a.lastFailure?.message || ''}
                  />
                );
              }
            })}
          </div>
          {/*  */}
          <div className="space-y-2">
            <h1 className="text-2xl font-semibold tracking-tight">
              Activities
            </h1>
            <div>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Id</TableHead>
                    <TableHead>Time</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Schema</TableHead>
                    <TableHead>Table</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {events.map((e, index) => {
                    const isError = e.Tasks.some((t) => t.error);
                    return (
                      <>
                        <TableRow
                          key={e.id}
                          onClick={(_) => onRowClick(index)}
                          className={
                            isError
                              ? 'border-t-2 border-b-2 border-destructive'
                              : ''
                          }
                        >
                          <TableCell>{e.id.toString()}</TableCell>
                          <TableCell>
                            {e.startTime &&
                              formatDateTimeMilliseconds(e.startTime.toDate())}
                          </TableCell>
                          <TableCell>{e.type}</TableCell>
                          <TableCell>
                            {e.metadata?.metadata.case == 'syncMetadata' &&
                              e.metadata.metadata.value.schema}
                          </TableCell>
                          <TableCell>
                            {e.metadata?.metadata.case == 'syncMetadata' &&
                              e.metadata.metadata.value.table}
                          </TableCell>
                        </TableRow>
                        <TableRow
                          key={`${e.id}-collapse`}
                          className={`${
                            isOpen[index] == 1 ? 'visible' : 'collapse'
                          }`}
                        >
                          <TableCell>
                            <div className="p-6">
                              <Card className="p-5">
                                <Table>
                                  <TableHeader>
                                    <TableRow>
                                      <TableHead>Id</TableHead>
                                      <TableHead>Type</TableHead>
                                      <TableHead>Time</TableHead>
                                      {isError && (
                                        <>
                                          <TableHead>Error</TableHead>
                                          <TableHead>Retry State</TableHead>
                                        </>
                                      )}
                                    </TableRow>
                                  </TableHeader>
                                  <TableBody>
                                    {e.Tasks.map((t) => {
                                      const cn =
                                        t.error &&
                                        'border-t border-b border-destructive';
                                      return (
                                        <TableRow key={t.id}>
                                          <TableCell className={cn}>
                                            {t.id.toString()}
                                          </TableCell>
                                          <TableCell className={cn}>
                                            {t.type}
                                          </TableCell>
                                          <TableCell className={cn}>
                                            {t.eventTime &&
                                              formatDateTimeMilliseconds(
                                                t.eventTime.toDate()
                                              )}
                                          </TableCell>
                                          {isError && (
                                            <>
                                              <TableCell className={cn}>
                                                {t.error?.message}
                                              </TableCell>
                                              <TableCell className={cn}>
                                                {t.error?.retryState}
                                              </TableCell>
                                            </>
                                          )}
                                        </TableRow>
                                      );
                                    })}
                                  </TableBody>
                                </Table>
                              </Card>
                            </div>
                          </TableCell>
                        </TableRow>
                      </>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </div>
          {/*  */}
          {/*  */}
          {/* <div className="space-y-2">
            <h1 className="text-2xl font-semibold tracking-tight">
              Activities
            </h1>
            <div>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>
                      <div
                        className={`grid  grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 p-5 font-medium  text-left  text-muted-foreground`}
                      >
                        <div className="lg:pl-10">Id</div>
                        <div>Time</div>
                        <div>Type</div>
                        <div>Schema</div>
                        <div>Table</div>
                      </div>
                    </TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {events.map((e) => {
                    const isError = e.Tasks.some((t) => t.error);
                    return (
                      <TableRow key={e.id}>
                        <TableCell
                          className={` ${
                            isError && 'border border-destructive'
                          }`}
                        >
                          <Collapsible>
                            <CollapsibleTrigger className="w-full text-left">
                              <div
                                className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 p-5 `}
                              >
                                <div className="lg:pl-10">
                                  {e.id.toString()}
                                </div>
                                <div>
                                  {e.startTime &&
                                    formatDateTimeMilliseconds(
                                      e.startTime.toDate()
                                    )}
                                </div>
                                <div>{e.type}</div>
                                <div>
                                  {e.metadata?.metadata.case ==
                                    'syncMetadata' &&
                                    e.metadata.metadata.value.schema}
                                </div>
                                <div>
                                  {e.metadata?.metadata.case ==
                                    'syncMetadata' &&
                                    e.metadata.metadata.value.table}
                                </div>
                              </div>
                            </CollapsibleTrigger>
                            <CollapsibleContent>
                              <div>
                                <Separator />
                                <div className="p-6">
                                  <Card className="p-5">
                                    <Table>
                                      <TableHeader>
                                        <TableRow>
                                          <TableHead>Id</TableHead>
                                          <TableHead>Type</TableHead>
                                          <TableHead>Time</TableHead>
                                          {isError && (
                                            <>
                                              <TableHead>Error</TableHead>
                                              <TableHead>Retry State</TableHead>
                                            </>
                                          )}
                                        </TableRow>
                                      </TableHeader>
                                      <TableBody>
                                        {e.Tasks.map((t) => {
                                          const cn =
                                            t.error &&
                                            'border-t border-b border-destructive';
                                          return (
                                            <TableRow key={t.id}>
                                              <TableCell className={cn}>
                                                {t.id.toString()}
                                              </TableCell>
                                              <TableCell className={cn}>
                                                {t.type}
                                              </TableCell>
                                              <TableCell className={cn}>
                                                {t.eventTime &&
                                                  formatDateTimeMilliseconds(
                                                    t.eventTime.toDate()
                                                  )}
                                              </TableCell>
                                              {isError && (
                                                <>
                                                  <TableCell className={cn}>
                                                    {t.error?.message}
                                                  </TableCell>
                                                  <TableCell className={cn}>
                                                    {t.error?.retryState}
                                                  </TableCell>
                                                </>
                                              )}
                                            </TableRow>
                                          );
                                        })}
                                      </TableBody>
                                    </Table>
                                  </Card>
                                </div>
                              </div>
                            </CollapsibleContent>
                          </Collapsible>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </div>
          </div> */}
          {/*  */}
          {/* <div className="space-y-2">
            <h1 className="text-2xl font-semibold tracking-tight">
              Activities
            </h1>
            <div className="border">
              <div
                className={`grid  grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 p-5 font-medium border-b text-left  text-muted-foreground`}
              >
                <div className="lg:pl-10">Id</div>
                <div>Time</div>
                <div>Type</div>
                <div>Schema</div>
                <div>Table</div>
              </div>
              {events.map((e) => {
                const isError = e.Tasks.some((t) => t.error);
                return (
                  <div key={e.id} className="border-b">
                    <Collapsible>
                      <CollapsibleTrigger className="w-full text-left">
                        <div
                          className={`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4 p-5 ${
                            isError && 'border border-destructive'
                          }`}
                        >
                          <div className="lg:pl-10">{e.id.toString()}</div>
                          <div>
                            {e.startTime &&
                              formatDateTimeMilliseconds(e.startTime.toDate())}
                          </div>
                          <div>{e.type}</div>
                          <div>
                            {e.metadata?.metadata.case == 'syncMetadata' &&
                              e.metadata.metadata.value.schema}
                          </div>
                          <div>
                            {e.metadata?.metadata.case == 'syncMetadata' &&
                              e.metadata.metadata.value.table}
                          </div>
                        </div>
                      </CollapsibleTrigger>
                      <CollapsibleContent>
                        <div>
                          <Separator />
                          <div className="p-6">
                            <Card className="p-5">
                              <Table>
                                <TableHeader>
                                  <TableRow>
                                    <TableHead>Id</TableHead>
                                    <TableHead>Type</TableHead>
                                    <TableHead>Time</TableHead>
                                    {isError && (
                                      <>
                                        <TableHead>Error</TableHead>
                                        <TableHead>Retry State</TableHead>
                                      </>
                                    )}
                                  </TableRow>
                                </TableHeader>
                                <TableBody>
                                  {e.Tasks.map((t) => {
                                    const cn =
                                      t.error &&
                                      'border-t border-b border-destructive';
                                    return (
                                      <TableRow key={t.id}>
                                        <TableCell className={cn}>
                                          {t.id.toString()}
                                        </TableCell>
                                        <TableCell className={cn}>
                                          {t.type}
                                        </TableCell>
                                        <TableCell className={cn}>
                                          {t.eventTime &&
                                            formatDateTimeMilliseconds(
                                              t.eventTime.toDate()
                                            )}
                                        </TableCell>
                                        {isError && (
                                          <>
                                            <TableCell className={cn}>
                                              {t.error?.message}
                                            </TableCell>
                                            <TableCell className={cn}>
                                              {t.error?.retryState}
                                            </TableCell>
                                          </>
                                        )}
                                      </TableRow>
                                    );
                                  })}
                                </TableBody>
                              </Table>
                            </Card>
                          </div>
                        </div>
                      </CollapsibleContent>
                    </Collapsible>
                  </div>
                );
              })}
            </div>
          </div> */}
        </div>
      )}
    </OverviewContainer>
  );
}

interface StatCardProps {
  header: string;
  content?: JSX.Element | string;
}

function StatCard(props: StatCardProps): ReactElement {
  const { header, content } = props;
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{header}</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-lg font-bold">{content}</div>
      </CardContent>
    </Card>
  );
}

function getDuration(dateTimeValue2?: Date, dateTimeValue1?: Date): string {
  if (!dateTimeValue1 || !dateTimeValue2) {
    return '';
  }
  var differenceValue =
    (dateTimeValue2.getTime() - dateTimeValue1.getTime()) / 1000;
  const minutes = Math.abs(Math.round(differenceValue / 60));
  const seconds = differenceValue % 60;
  if (minutes == 0) {
    return `${seconds} seconds`;
  }
  return `${minutes} minutes ${seconds} seconds`;
}

interface AlertProps {
  title: string;
  description: string;
}

function AlertDestructive(props: AlertProps): ReactElement {
  return (
    <Alert variant="destructive">
      <AlertTitle>{`${props.title}: ${props.description}`}</AlertTitle>
    </Alert>
  );
}

interface ButtonProps {
  jobId?: string;
}

function ButtonLink(props: ButtonProps): ReactElement {
  const router = useRouter();
  if (!props.jobId) {
    return <></>;
  }
  return (
    <Button
      variant="outline"
      onClick={() => router.push(`/jobs/${props.jobId}`)}
    >
      View Job
      <ArrowRightIcon className="ml-2 h-4 w-4" />
    </Button>
  );
}
