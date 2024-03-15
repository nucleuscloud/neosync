'use client';
import { JsonValue } from '@bufbuild/protobuf';
import { GetMetricCountResponse, RangedMetricName } from '@neosync/sdk';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export type MetricIdentifierType = 'accountId' | 'jobId' | 'runId';

export function useGetMetricCount(
  accountId: string,
  start: Date,
  end: Date,
  metric: RangedMetricName,
  idtype: MetricIdentifierType,
  identifier: string
): HookReply<GetMetricCountResponse> {
  const startSec = start.getTime() / 1000;
  const endSec = end.getTime() / 1000;
  const urlparams = new URLSearchParams({
    start: startSec.toString(),
    end: endSec.toString(),
    metric: metric.toString(),
    idtype,
    identifier,
  });
  return useNucleusAuthenticatedFetch<
    GetMetricCountResponse,
    JsonValue | GetMetricCountResponse
  >(
    `/api/accounts/${accountId}/metrics/count?${urlparams.toString()}`,
    !!accountId && !!start && !!end && !!metric && !!idtype && !!identifier,
    undefined,
    (data) =>
      data instanceof GetMetricCountResponse
        ? data
        : GetMetricCountResponse.fromJson(data)
  );
}
