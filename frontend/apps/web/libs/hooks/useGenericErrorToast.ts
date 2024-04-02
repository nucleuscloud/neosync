'use client';
import { useToast } from '@/components/ui/use-toast';
import { getErrorMessage } from '@/util/util';
import { useEffect } from 'react';

export function useGenericErrorToast(error?: Error): void {
  const { toast } = useToast();
  useEffect(() => {
    if (error?.message || error?.name) {
      console.error(error);
      toast({
        title:
          'There was a problem making your request. Please try again later',
        description: getErrorMessage(error),
        variant: 'destructive',
      });
    }
  }, [error?.message, error?.name]);
}
