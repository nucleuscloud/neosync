import { TransformerHandler } from '@/components/jobs/SchemaTable/transformer-handler';
import { useQuery } from '@connectrpc/connect-query';
import { TransformerSource, TransformersService } from '@neosync/sdk';
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
  } = useQuery(TransformersService.method.getSystemTransformers);
  const {
    data: customTransformersData,
    isLoading: isLoadingCustomTransformers,
    isFetching: isCustomValidating,
  } = useQuery(
    TransformersService.method.getUserDefinedTransformers,
    { accountId: accountId },
    { enabled: !!accountId }
  );

  const systemTransformers = systemTransformersData?.transformers ?? [];
  const userDefinedTransformers = customTransformersData?.transformers ?? [];

  const isLoading = isLoadingSystemTransformers || isLoadingCustomTransformers;
  const isValidating = isSystemValidating || isCustomValidating;

  const handler = useMemo(
    (): TransformerHandler =>
      new TransformerHandler(
        systemTransformers.filter(
          (st) => st.source !== TransformerSource.TRANSFORM_PII_TEXT
        ),
        userDefinedTransformers
      ),
    [isValidating]
  );

  return {
    handler,
    isLoading,
    isValidating,
  };
}
