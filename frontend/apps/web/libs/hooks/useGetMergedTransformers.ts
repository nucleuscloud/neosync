import {
  filterInputFreeSystemTransformers,
  filterInputFreeUdfTransformers,
} from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { Transformer, joinTransformers } from '@/shared/transformers';
import { useMemo } from 'react';
import { useGetSystemTransformers } from './useGetSystemTransformers';
import { useGetUserDefinedTransformers } from './useGetUserDefinedTransformers';

interface GetMergedTransformersResponse {
  transformers: Transformer[];
  isLoading: boolean;
  isValidating: boolean;
}

export function useGetMergedTransformers(
  excludeInputReqTransformers: boolean,
  accountId: string
): GetMergedTransformersResponse {
  const {
    data: systemTransformers,
    isLoading: isLoadingSystemTransformers,
    isValidating: isSystemValidating,
  } = useGetSystemTransformers();
  const {
    data: customTransformers,
    isLoading: isLoadingCustomTransformers,
    isValidating: isCustomValidating,
  } = useGetUserDefinedTransformers(accountId);

  const isLoading = isLoadingSystemTransformers || isLoadingCustomTransformers;
  const isValidating = isSystemValidating || isCustomValidating;

  const mergedTransformers = useMemo(() => {
    if (!systemTransformers || !customTransformers) return [];

    const filteredSystemTransformers = excludeInputReqTransformers
      ? filterInputFreeSystemTransformers(systemTransformers.transformers ?? [])
      : systemTransformers.transformers ?? [];

    const filteredCustomTransformers = excludeInputReqTransformers
      ? filterInputFreeUdfTransformers(
          customTransformers.transformers ?? [],
          filteredSystemTransformers
        )
      : customTransformers.transformers ?? [];

    return joinTransformers(
      filteredSystemTransformers,
      filteredCustomTransformers
    );
  }, [isValidating, excludeInputReqTransformers]);

  return {
    transformers: mergedTransformers,
    isLoading: isLoading,
    isValidating: isValidating,
  };
}
