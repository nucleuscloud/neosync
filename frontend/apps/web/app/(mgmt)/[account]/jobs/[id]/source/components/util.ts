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
  if (js?.options?.config.case === 'aiGenerate') {
    return js.options.config.value.fkSourceConnectionId;
  }
  return undefined;
}

export function getSetDelta(
  newSet: Set<string>,
  oldSet: Set<string>
): [Set<string>, Set<string>] {
  const added = new Set<string>();
  const removed = new Set<string>();

  oldSet.forEach((val) => {
    if (!newSet.has(val)) {
      removed.add(val);
    }
  });
  newSet.forEach((val) => {
    if (!oldSet.has(val)) {
      added.add(val);
    }
  });

  return [added, removed];
}
