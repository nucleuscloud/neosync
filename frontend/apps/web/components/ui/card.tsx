import * as React from 'react';

import { cn } from '@/libs/utils';

export function Card({
  ref,
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & {
  ref: React.RefObject<HTMLDivElement>;
}) {
  return (
    <div
      ref={ref}
      className={cn(
        'rounded-xl border border-gray-200 dark:border-gray-700 bg-background text-card-foreground',
        className
      )}
      {...props}
    />
  );
}

export function CardHeader({
  ref,
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & {
  ref: React.RefObject<HTMLDivElement>;
}) {
  return (
    <div
      ref={ref}
      className={cn(
        'flex flex-col space-y-1.5 p-6',
        'dark:border-gray-900 ',
        className
      )}
      {...props}
    />
  );
}

export function CardTitle({
  className,
  ...props
}: React.HTMLAttributes<HTMLHeadingElement>) {
  return (
    <h3
      className={cn('font-semibold leading-none tracking-tight', className)}
      {...props}
    />
  );
}

export function CardDescription({
  className,
  ...props
}: React.HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p className={cn('text-sm text-muted-foreground', className)} {...props} />
  );
}

export function CardContent({
  ref,
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement> & {
  ref: React.RefObject<HTMLDivElement>;
}) {
  return <div ref={ref} className={cn('p-6 pt-0', className)} {...props} />;
}

export function CardFooter({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn('flex items-center p-6 pt-0', className)} {...props} />
  );
}
