import { DescFile } from '@bufbuild/protobuf';
import type { ImportSymbol, Schema } from '@bufbuild/protoplugin/ecmascript';

/**
 * Temporary function to retrieve the import symbol for the proto2 or proto3
 * runtime.
 *
 * For syntax "editions", this function raises and error.
 *
 * Once support for "editions" is implemented in the runtime, this function can
 * be removed.
 */
export function getNonEditionRuntime(
  schema: Schema,
  file: DescFile
): ImportSymbol {
  if (file.syntax === 'editions') {
    // TODO support editions
    throw new Error(
      `${file.proto.name ?? ''}: syntax "editions" is not supported yet`
    );
  }
  return schema.runtime[file.syntax];
}
