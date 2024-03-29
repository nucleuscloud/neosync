import { cn } from '@/libs/utils';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { buttonVariants } from './ui/button';

interface SubNavProps extends React.HTMLAttributes<HTMLElement> {
  items: {
    href: string;
    title: string;
  }[];
  buttonClassName?: string;
}

export function SubNav({
  className,
  items,
  buttonClassName,
  ...props
}: SubNavProps) {
  const pathname = usePathname();
  return (
    <nav
      className={cn(
        'flex flex-col lg:flex-row gap-2 lg:gap-y-0 lg:gap-x-2',
        className
      )}
      {...props}
    >
      {items.map((item) => {
        return (
          <Link
            key={item.href}
            href={item.href}
            className={cn(
              buttonVariants({ variant: 'outline' }),
              pathname === item.href
                ? 'bg-muted hover:bg-muted'
                : 'hover:bg-gray-100 dark:hover:bg-gray-800',
              'justify-start',
              buttonClassName,
              'px-6'
            )}
          >
            {item.title}
          </Link>
        );
      })}
    </nav>
  );
}
