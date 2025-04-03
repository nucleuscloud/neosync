import { Job } from '@neosync/sdk';

export function isDataGenJob(job?: Job): boolean {
  return job?.source?.options?.config?.case === 'generate';
}

export function isAiDataGenJob(job?: Job): boolean {
  return job?.source?.options?.config.case === 'aiGenerate';
}

export function isPiiDetectJob(job?: Job): boolean {
  return job?.jobType?.jobType.case === 'piiDetect';
}
