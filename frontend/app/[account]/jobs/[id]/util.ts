import { GetConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { Job } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';

export async function getConnection(
  accountId: string,
  connectionId?: string
): Promise<GetConnectionResponse | undefined> {
  if (!connectionId) {
    return;
  }
  const res = await fetch(
    `/api/accounts/${accountId}/connections/${connectionId}`,
    {
      method: 'GET',
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionResponse.fromJson(await res.json());
}

export function isDataGenJob(job?: Job): boolean {
  return job?.source?.options?.config.case === 'generate';
}
