import {
  SystemTransformer,
  UserDefinedTransformer,
} from '@/neosync-api-client/mgmt/v1alpha1/transformer_pb';

export type Transformer = SystemTransformer | UserDefinedTransformer;

/**
 * combines both transformer types into a single Transformer interface
 */
export function joinTransformers(
  system: SystemTransformer[],
  custom: UserDefinedTransformer[]
): Transformer[] {
  return [...system, ...custom];
}

export function isUserDefinedTransformer(
  transformer: Transformer
): transformer is UserDefinedTransformer {
  return !!(transformer as unknown as UserDefinedTransformer).id;
}

export function isSystemTransformer(
  transformer: Transformer
): transformer is SystemTransformer {
  return !isUserDefinedTransformer(transformer);
}
