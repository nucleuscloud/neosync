import { JobSource } from '@neosync/sdk';

export function getConnectionIdFromSource(
  js: JobSource | undefined
): string | undefined {
  if (
    js?.options?.config.case === 'postgres' ||
    js?.options?.config.case === 'mysql' ||
    js?.options?.config.case === 'awsS3'
  ) {
    return js.options.config.value.connectionId;
  }
  return undefined;
}

export function getFkIdFromGenerateSource(
  js: JobSource | undefined
): string | undefined {
  if (js?.options?.config.case === 'generate') {
    return js.options.config.value.fkSourceConnectionId;
  }
  return undefined;
}
