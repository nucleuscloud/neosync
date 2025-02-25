import { create } from '@bufbuild/protobuf';
import { useQuery } from '@connectrpc/connect-query';
import {
  DayResult,
  MetricsService,
  Date as NeosyncDate,
  DateSchema as NeosyncDateSchema,
  RangedMetricName,
} from '@neosync/sdk';
import { format } from 'date-fns';
import { useTheme } from 'next-themes';
import { ReactElement } from 'react';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { useAccount } from '../providers/account-provider';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../ui/card';
import { Skeleton } from '../ui/skeleton';
import {
  dateToNeoDate,
  MetricIdentifierType,
  periodToDateRange,
  shortNumberFormatter,
  UsagePeriod,
} from './util';

interface Props {
  period: UsagePeriod;
  metric: RangedMetricName;
  idtype: MetricIdentifierType;
  identifier: string;
}

export default function DailyMetricCount(props: Props): ReactElement<any> {
  const { period, metric, idtype, identifier } = props;
  const { account } = useAccount();
  const [start, end] = periodToDateRange(period);
  const { data: metricCountData, isLoading } = useQuery(
    MetricsService.method.getDailyMetricCount,
    {
      metric,
      start: dateToNeoDate(start),
      end: dateToNeoDate(end),
      identifier:
        idtype === 'accountId'
          ? { case: 'accountId', value: identifier }
          : idtype === 'jobId'
            ? { case: 'jobId', value: identifier }
            : idtype === 'runId'
              ? { case: 'runId', value: identifier }
              : undefined,
    },
    {
      enabled:
        !!metric && !!account?.id && !!identifier && !!idtype && !!period,
    }
  );
  const { resolvedTheme } = useTheme();
  const tickColor = resolvedTheme === 'dark' ? 'white' : 'black';
  const labelBg = resolvedTheme === 'dark' ? 'white' : 'black';
  const tooltipLabelColor = resolvedTheme === 'dark' ? 'black' : 'white';

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
        <CardDescription>
          There will be a delay before the count for the current day appears
        </CardDescription>
      </CardHeader>
      <CardContent>
        <>
          <style jsx global>{`
            .recharts-legend-item {
              cursor: pointer;
            }
          `}</style>
          <ResponsiveContainer width="100%" height={400}>
            <AreaChart data={toDayResultPlotPoints(results)} margin={{}}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis
                dataKey={(obj: DayResult) => {
                  const date = obj.date ?? create(NeosyncDateSchema);
                  return format(
                    new Date(date.year, date.month - 1, date.day),
                    'MMM d'
                  );
                }}
                tick={{ fill: tickColor }}
              />
              <YAxis
                tickFormatter={(value) =>
                  shortNumberFormatter(numformatter, value)
                }
                tick={{ fill: tickColor }}
              />
              <Area dataKey="count" name="ingested" />
              <Tooltip
                labelStyle={{
                  color: tooltipLabelColor,
                  borderBottomColor: tickColor,
                  borderBottomWidth: '1px',
                }}
                contentStyle={{ background: labelBg, borderRadius: '8px' }}
                formatter={(value) =>
                  shortNumberFormatter(numformatter, value as number)
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
    date: result.date ?? create(NeosyncDateSchema),
  }));
}
