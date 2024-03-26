import { SystemTransformer, UserDefinedTransformer } from '@neosync/sdk';

export type Transformer = SystemTransformer | UserDefinedTransformer;

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
