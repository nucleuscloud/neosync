import { useQuery, UseQueryResult } from '@tanstack/react-query';
import { fetcher } from '../fetcher';

export function useReadNeosyncTransformerDeclarationFile(): UseQueryResult<string> {
  return useQuery({
    queryKey: [`/api/files/neosync-transformer-declarations`],
    queryFn: (ctx) => fetcher(ctx.queryKey.join('/')),
  });
}
