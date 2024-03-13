'use client';
import OverviewContainer from '@/components/containers/OverviewContainer';
import PageHeader from '@/components/headers/PageHeader';
import { useAccount } from '@/components/providers/account-provider';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { useGetDailyMetricCount } from '@/libs/hooks/useGetDailyMetricCount';
import { useGetMetricCount } from '@/libs/hooks/useGetMetricCount';
import { useGetSystemAppConfig } from '@/libs/hooks/useGetSystemAppConfig';
import { DayResult, Date as NeosyncDate, RangedMetricName } from '@neosync/sdk';
import { endOfMonth, format, startOfMonth, subMonths } from 'date-fns';
import { ReactElement, useState } from 'react';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';

export default function UsagePage(): ReactElement {
  const [period, setPeriod] = useState<UsagePeriod>('current');
  const { data: configData, isLoading } = useGetSystemAppConfig();
  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }
  if (!configData?.isMetricsServiceEnabled) {
    return (
      <div>
        <Alert variant="warning">
          <AlertTitle>Metrics are not currently enabled</AlertTitle>
          <AlertDescription>
            To enable them, please update Neosync configuration or contact your
            system administrator.
          </AlertDescription>
        </Alert>
      </div>
    );
  }
  const [start, end] = periodToDateRange(period);
  return (
    <OverviewContainer
      Header={
        <PageHeader
          header="Usage"
          description={`${getPeriodLabel(period)}: ${getDateRangeLabel(start, end)}`}
          extraHeading={
            <UsagePeriodSelector period={period} setPeriod={setPeriod} />
          }
        />
      }
      containerClassName="usage-page"
    >
      <div className="flex">
        <DisplayMetricCount period={period} />
      </div>
      <div>
        <DailyMetricCount period={period} />
      </div>
    </OverviewContainer>
  );
}

interface DailyMetricCountProps {
  period: UsagePeriod;
}

function DailyMetricCount(props: DailyMetricCountProps): ReactElement {
  const { period } = props;
  const { account } = useAccount();
  const [start, end] = periodToDateRange(period);
  const { data: metricCountData, isLoading } = useGetDailyMetricCount(
    account?.id ?? '',
    dateToNeoDate(start),
    dateToNeoDate(end),
    RangedMetricName.INPUT_RECEIVED,
    'accountId',
    account?.id ?? ''
  );

  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }
  const browserLanugages = [...navigator.languages];
  const numformatter = new Intl.NumberFormat(browserLanugages, {
    style: 'decimal',
    minimumFractionDigits: 0,
  });
  const results = metricCountData?.results ?? [];

  if (results.length === 0) {
    return <div />;
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Total number of records ingested by day</CardTitle>
      </CardHeader>
      <CardContent>
        <>
          <style jsx global>{`
            .recharts-legend-item {
              cursor: pointer;
            }
          `}</style>
          <ResponsiveContainer width="100%" height={400}>
            <AreaChart data={toDayResultPlotPoints(results)}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey={(obj: DayResult) => {
                  const date = obj.date ?? new NeosyncDate();
                  return format(
                    new Date(date.year, date.month - 1, date.day),
                    'MMM d yy'
                  );
                }}
              />
              <YAxis tickFormatter={(value) => numformatter.format(value)} />
              {/* <Legend /> */}
              <Area dataKey="count" name="ingested" />
              <Tooltip
                formatter={(value, name) =>
                  numformatter.format(value as number)
                }
              />
            </AreaChart>
          </ResponsiveContainer>
        </>
      </CardContent>
    </Card>
  );
}

interface DayResultPlotPoint {
  count: number;
  date: NeosyncDate;
}

function toDayResultPlotPoints(results: DayResult[]): DayResultPlotPoint[] {
  return results.map((result) => ({
    count: Number(result.count),
    date: result.date ?? new NeosyncDate(),
  }));
}

interface DisplayMetricCountProps {
  period: UsagePeriod;
}

function DisplayMetricCount(props: DisplayMetricCountProps): ReactElement {
  const { period } = props;
  const { account } = useAccount();
  const [start, end] = periodToDateRange(period);
  const { data: metricCountData, isLoading } = useGetMetricCount(
    account?.id ?? '',
    start,
    end,
    RangedMetricName.INPUT_RECEIVED,
    'accountId',
    account?.id ?? ''
  );

  if (isLoading) {
    return <Skeleton className="w-full h-12" />;
  }
  const browserLanugages = [...navigator.languages];
  const numformatter = new Intl.NumberFormat(browserLanugages, {
    style: 'decimal',
    minimumFractionDigits: 0,
  });
  const count =
    metricCountData?.count !== undefined
      ? numformatter.format(metricCountData.count)
      : '0';
  return (
    <Card>
      <CardHeader>
        <CardTitle>Total number of records ingested</CardTitle>
        <CardDescription className="max-w-72">
          This metric shows the total number of records ingested from data
          sources for the given time period.
        </CardDescription>
      </CardHeader>
      <CardContent>{count}</CardContent>
    </Card>
  );
}

function getPeriodLabel(period: UsagePeriod): string {
  switch (period) {
    case 'current': {
      return 'Current Period';
    }
    case 'last-month': {
      return 'Last Month';
    }
  }
}

function getDateRangeLabel(start: Date, end: Date): string {
  return `${format(start, 'MM/dd/yy')} - ${format(end, 'MM/dd/yy')}`;
}

function periodToDateRange(period: UsagePeriod): [Date, Date] {
  const currentDate = new Date();
  switch (period) {
    case 'current': {
      const start = startOfMonth(currentDate);
      const end = endOfMonth(currentDate);
      return [start, end];
    }
    case 'last-month': {
      const prevMonthDate = subMonths(currentDate, 1);
      const start = startOfMonth(prevMonthDate);
      const end = endOfMonth(prevMonthDate);
      return [start, end];
    }
  }
}

function dateToNeoDate(date: Date): NeosyncDate {
  return new NeosyncDate({
    day: date.getDate(),
    month: date.getMonth() + 1,
    year: date.getFullYear(),
  });
}

type UsagePeriod = 'current' | 'last-month';

interface UsagePeriodSelectorProps {
  period: UsagePeriod;
  setPeriod(newVal: UsagePeriod): void;
}

function UsagePeriodSelector(props: UsagePeriodSelectorProps): ReactElement {
  const { period, setPeriod } = props;
  return (
    <Select
      onValueChange={(value: string) => {
        if (!value) {
          return;
        }
        const typedVal = value as UsagePeriod;
        setPeriod(typedVal);
      }}
      value={period}
    >
      <SelectTrigger>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem className="cursor-pointer" value="current">
          <p>Current Period</p>
        </SelectItem>
        <SelectItem className="cursor-pointer" value="last-month">
          <p>Last Month</p>
        </SelectItem>
      </SelectContent>
    </Select>
  );
}
