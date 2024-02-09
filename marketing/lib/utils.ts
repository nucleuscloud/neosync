import { env } from '@/env';
import { clsx, type ClassValue } from 'clsx';
import { format } from 'date-fns';
import { utcToZonedTime } from 'date-fns-tz';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(input: string | number): string {
  const date = new Date(input);
  return format(utcToZonedTime(date, 'UTC'), 'MMMM do, yyyy');
}

export function absoluteUrl(path: string) {
  return `${env.NEXT_PUBLIC_APP_URL}${path}`;
}
