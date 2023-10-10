import { GetConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';

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
