'use client';
import { useEffect } from 'react';

export function useGenericErrorToast(error?: Error): void {
  // const toast = useToast();
  useEffect(() => {
    if (error?.message || error?.name) {
      console.error(error);
      // Toast({
      //   id: 'generic-err',
      //   title:
      //     'There was a problem making your request. Please try again later',
      //   description: error.message,
      //   status: 'error',
      //   toast,
      // });
    }
  }, [error?.message, error?.name]);
}
