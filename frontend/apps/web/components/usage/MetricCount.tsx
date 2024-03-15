import {
  MetricIdentifierType,
  useGetMetricCount,
} from '@/libs/hooks/useGetMetricCount';
import { RangedMetricName } from '@neosync/sdk';
import { ReactElement } from 'react';
import { useAccount } from '../providers/account-provider';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../ui/card';
import { Skeleton } from '../ui/skeleton';
import { UsagePeriod, periodToDateRange } from './util';

interface Props {
  period: UsagePeriod;
  metric: RangedMetricName;
  idtype: MetricIdentifierType;
  identifier: string;
}

export default function MetricCount(props: Props): ReactElement {
  const { period, metric, identifier, idtype } = props;
  const { account } = useAccount();
  const [start, end] = periodToDateRange(period);
  const { data: metricCountData, isLoading } = useGetMetricCount(
    account?.id ?? '',
    start,
    end,
    metric,
    idtype,
    identifier
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
        <CardTitle>{getCardTitle(metric)}</CardTitle>
        <CardDescription className="max-w-72">
          {getCardDescription(metric, idtype)}
        </CardDescription>
      </CardHeader>
      <CardContent>{count}</CardContent>
    </Card>
  );
}

function getCardTitle(metric: RangedMetricName): string {
  switch (metric) {
    case RangedMetricName.INPUT_RECEIVED:
    default:
      return 'Total number of records ingested';
  }
}

function getCardDescription(
  metric: RangedMetricName,
  idtype: MetricIdentifierType
): string {
  switch (metric) {
    case RangedMetricName.INPUT_RECEIVED:
    default:
      return `This metric shows the total number of records ingested for this ${getLabelByIdType(idtype)} for the selected usage period. Note there will be a delay before the count appears for the current day`;
  }
}

function getLabelByIdType(idtype: MetricIdentifierType): string {
  switch (idtype) {
    case 'accountId': {
      return 'account';
    }
    case 'jobId': {
      return 'job';
    }
    case 'runId': {
      return 'run';
    }
  }
}
