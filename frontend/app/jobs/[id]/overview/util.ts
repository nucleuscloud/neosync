import { GetJobResponse } from "@/neosync-api-client/mgmt/v1alpha1/job_pb";

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
