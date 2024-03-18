import {
  filterInputFreeSystemTransformers,
  filterInputFreeUdfTransformers,
} from '@/app/(mgmt)/[account]/transformers/EditTransformerOptions';
import { Transformer, joinTransformers } from '@/shared/transformers';
import { SystemTransformer, UserDefinedTransformer } from '@neosync/sdk';
import { useMemo } from 'react';
import { useGetSystemTransformers } from './useGetSystemTransformers';
import { useGetUserDefinedTransformers } from './useGetUserDefinedTransformers';

interface GetMergedTransformersResponse {
  transformers: Transformer[];
  systemTransformers: SystemTransformer[];
  systemMap: Map<string, SystemTransformer>;
  userDefinedTransformers: UserDefinedTransformer[];
  userDefinedMap: Map<string, UserDefinedTransformer>;
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

  const { filteredSystem, filteredCustom, merged, userMap, sysMap } =
    useMemo(() => {
      if (!systemTransformers || !customTransformers) {
        return {
          merged: [],
          filteredCustom: [],
          filteredSystem: [],
          sysMap: new Map(),
          userMap: new Map(),
        };
      }

      const filteredSystemTransformers = excludeInputReqTransformers
        ? filterInputFreeSystemTransformers(
            systemTransformers.transformers ?? []
          )
        : systemTransformers.transformers ?? [];

      const filteredCustomTransformers = excludeInputReqTransformers
        ? filterInputFreeUdfTransformers(
            customTransformers.transformers ?? [],
            filteredSystemTransformers
          )
        : customTransformers.transformers ?? [];

      return {
        merged: joinTransformers(
          filteredSystemTransformers,
          filteredCustomTransformers
        ),
        filteredSystem: filteredSystemTransformers,
        filteredCustom: filteredCustomTransformers,
        sysMap: new Map(filteredSystemTransformers.map((t) => [t.source, t])),
        userMap: new Map(filteredCustomTransformers.map((t) => [t.id, t])),
      };
    }, [isValidating, excludeInputReqTransformers]);

  return {
    transformers: merged,
    systemTransformers: filteredSystem,
    userDefinedTransformers: filteredCustom,
    systemMap: sysMap,
    userDefinedMap: userMap,

    isLoading: isLoading,
    isValidating: isValidating,
  };
}
