import { cva, type VariantProps } from 'class-variance-authority';
import * as React from 'react';

import { cn } from '@/libs/utils';

const alertVariants = cva(
  'relative w-full rounded-lg border px-4 py-3 text-sm [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:text-foreground [&>svg~*]:pl-7',
  {
    variants: {
      variant: {
        default: 'bg-background text-foreground',
        success:
          'bg-background text-foreground bg-green-100 dark:bg-green-800 border border-green-300 dark:border-green-400',
        destructive:
          'border-destructive/50 dark:border-destructive bg-destructive text-destructive-foreground',
        warning: 'border border-yellow-300 dark:border-orange-500 ',
        info: 'border border-blue-300 shadow-xs bg-blue-50 dark:bg-gray-800 dark:border-gray-700',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  }
);

export function Alert({
  className,
  variant,
  ...props
}: React.ComponentPropsWithRef<'div'> & VariantProps<typeof alertVariants>) {
  return (
    <div
      role="alert"
      className={cn(alertVariants({ variant }), className)}
      {...props}
    />
  );
}

export function AlertTitle({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h5
      className={cn('mb-1 font-medium leading-none tracking-tight', className)}
      {...props}
    />
  );
}

export function AlertDescription({
  className,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <div
      className={cn('text-sm [&_p]:leading-relaxed', className)}
      {...props}
    />
  );
}
