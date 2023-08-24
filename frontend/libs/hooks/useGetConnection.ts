'use client';
import { GetConnectionResponse } from '@/neosync-api-client/mgmt/v1alpha1/connection_pb';
import { JsonValue } from '@bufbuild/protobuf';
import { HookReply } from './types';
import { useNucleusAuthenticatedFetch } from './useNucleusAuthenticatedFetch';

export function useGetConnection(id: string): HookReply<GetConnectionResponse> {
  return useNucleusAuthenticatedFetch<
    GetConnectionResponse,
    JsonValue | GetConnectionResponse
  >(`/api/connections/${id}`, !!id, undefined, (data) =>
    data instanceof GetConnectionResponse
      ? data
      : GetConnectionResponse.fromJson(data)
  );
}

// export async function getConnection(
//   id: string
// ): Promise<GetConnectionResponse> {
//   const headersIns = headers();
//   const cookiesIns = cookies();
//   const result = await fetcher2(`http://localhost:3000/api/connections/${id}`, {
//     headers: {
//       cookie: headersIns.get('cookie') ?? cookiesIns.toString(),
//     },
//   });
//   return GetConnectionResponse.fromJson(result);
// }
