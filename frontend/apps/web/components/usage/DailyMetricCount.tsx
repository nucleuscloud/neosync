import { useGetDailyMetricCount } from '@/libs/hooks/useGetDailyMetricCount';
import { MetricIdentifierType } from '@/libs/hooks/useGetMetricCount';
import { DayResult, Date as NeosyncDate, RangedMetricName } from '@neosync/sdk';
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
  UsagePeriod,
  dateToNeoDate,
  periodToDateRange,
  shortNumberFormatter,
} from './util';

interface Props {
  period: UsagePeriod;
  metric: RangedMetricName;
  idtype: MetricIdentifierType;
  identifier: string;
}

export default function DailyMetricCount(props: Props): ReactElement {
  const { period, metric, idtype, identifier } = props;
  const { account } = useAccount();
  const [start, end] = periodToDateRange(period);
  const { data: metricCountData, isLoading } = useGetDailyMetricCount(
    account?.id ?? '',
    dateToNeoDate(start),
    dateToNeoDate(end),
    metric,
    idtype,
    identifier
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
                  const date = obj.date ?? new NeosyncDate();
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
    date: result.date ?? new NeosyncDate(),
  }));
}
