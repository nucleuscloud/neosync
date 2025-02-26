import * as React from 'react';

import { cn } from '@/libs/utils';

export function Table({
  className,
  ...props
}: React.ComponentPropsWithRef<'table'>) {
  return (
    <div className="w-full overflow-auto">
      <table
        className={cn('w-full caption-bottom text-sm', className)}
        {...props}
      />
    </div>
  );
}

export function StickyHeaderTable({
  className,
  ...props
}: React.ComponentPropsWithRef<'table'>) {
  return (
    <table
      className={cn('w-full caption-bottom text-sm grid', className)}
      {...props}
    />
  );
}

export function TableHeader({
  className,
  ...props
}: React.ComponentPropsWithRef<'thead'>) {
  return <thead className={cn('[&_tr]:border-b', className)} {...props} />;
}

export function TableBody({
  className,
  ...props
}: React.ComponentPropsWithRef<'tbody'>) {
  return (
    <tbody className={cn('[&_tr:last-child]:border-0', className)} {...props} />
  );
}

export function TableFooter({
  className,
  ...props
}: React.ComponentPropsWithRef<'tfoot'>) {
  return (
    <tfoot
      className={cn(
        'bg-primary font-medium text-primary-foreground',
        className
      )}
      {...props}
    />
  );
}

export function TableRow({
  className,
  ...props
}: React.ComponentPropsWithRef<'tr'>) {
  return (
    <tr
      className={cn(
        'border-b dark:border-gray-700 transition-colors hover:bg-muted/50 data-[state=selected]:bg-muted',
        className
      )}
      {...props}
    />
  );
}

export function TableHead({
  className,
  ...props
}: React.ComponentPropsWithRef<'th'>) {
  return (
    <th
      className={cn(
        'h-10 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]',
        className
      )}
      {...props}
    />
  );
}

export function TableCell({
  className,
  ...props
}: React.ComponentPropsWithRef<'td'>) {
  return (
    <td
      className={cn(
        'p-2 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]',
        className
      )}
      {...props}
    />
  );
}

export function TableCaption({
  className,
  ...props
}: React.ComponentPropsWithRef<'caption'>) {
  return (
    <caption
      className={cn('mt-4 text-sm text-muted-foreground', className)}
      {...props}
    />
  );
}
