import { TransformerHandler } from '@/components/jobs/SchemaTable/transformer-handler';
import { useMemo } from 'react';
import { useGetSystemTransformers } from './useGetSystemTransformers';
import { useGetUserDefinedTransformers } from './useGetUserDefinedTransformers';

export function useGetTransformersHandler(accountId: string): {
  handler: TransformerHandler;
  isLoading: boolean;
  isValidating: boolean;
} {
  const {
    data: systemTransformersData,
    isLoading: isLoadingSystemTransformers,
    isValidating: isSystemValidating,
  } = useGetSystemTransformers();
  const {
    data: customTransformersData,
    isLoading: isLoadingCustomTransformers,
    isValidating: isCustomValidating,
  } = useGetUserDefinedTransformers(accountId);

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
