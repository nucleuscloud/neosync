import { TransformerHandler } from '@/components/jobs/SchemaTable/transformer-handler';
import { useQuery } from '@connectrpc/connect-query';
import {
  getSystemTransformers,
  getUserDefinedTransformers,
} from '@neosync/sdk/connectquery';
import { useMemo } from 'react';

export function useGetTransformersHandler(accountId: string): {
  handler: TransformerHandler;
  isLoading: boolean;
  isValidating: boolean;
} {
  const {
    data: systemTransformersData,
    isLoading: isLoadingSystemTransformers,
    isFetching: isSystemValidating,
  } = useQuery(getSystemTransformers);
  const {
    data: customTransformersData,
    isLoading: isLoadingCustomTransformers,
    isFetching: isCustomValidating,
  } = useQuery(
    getUserDefinedTransformers,
    { accountId: accountId },
    { enabled: !!accountId }
  );

  const systemTransformers = systemTransformersData?.transformers ?? [];
  const userDefinedTransformers = customTransformersData?.transformers ?? [];

  const isLoading = isLoadingSystemTransformers || isLoadingCustomTransformers;
  const isValidating = isSystemValidating || isCustomValidating;

  const handler = useMemo(
    (): TransformerHandler =>
      new TransformerHandler(systemTransformers, userDefinedTransformers),
    [isValidating]
  );

  return {
    handler,
    isLoading,
    isValidating,
  };
}
