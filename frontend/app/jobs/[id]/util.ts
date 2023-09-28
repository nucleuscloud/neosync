import { GetConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { GetJobResponse } from '@/neosync-api-client/mgmt/v1alpha1/job_pb';

export async function getJob(
  jobId?: string
): Promise<GetJobResponse | undefined> {
  if (!jobId) {
    return;
  }
  const res = await fetch(`/api/jobs/${jobId}`, {
    method: 'GET',
    headers: {
      'content-type': 'application/json',
    },
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetJobResponse.fromJson(await res.json());
}

export async function getConnection(
  connectionId?: string
): Promise<GetConnectionResponse | undefined> {
  if (!connectionId) {
    return;
  }
  const res = await fetch(`/api/connections/${connectionId}`, {
    method: 'GET',
    headers: {
      'content-type': 'application/json',
    },
  });
  if (!res.ok) {
    const body = await res.json();
    throw new Error(body.message);
  }
  return GetConnectionResponse.fromJson(await res.json());
}
