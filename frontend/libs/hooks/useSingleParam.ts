'use client';
import { getSingleOrUndefined } from '@/util/util';
import { useParams } from 'next/navigation';

export function useSingleParam(name: string): string | undefined {
  const params = useParams();
  return getSingleOrUndefined(params[name]);
}
