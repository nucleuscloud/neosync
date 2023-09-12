import { cn } from '@/libs/utils';
import { CheckCircledIcon } from '@radix-ui/react-icons';
interface ProgressNavProps extends React.HTMLAttributes<HTMLElement> {
  items: {
    href?: string;
    title: string;
    description: string;
  }[];
}

export default function ProgressNav({
  className,
  items,
  ...props
}: ProgressNavProps) {
  return (
    <nav aria-label="Progress" className={cn('', className)} {...props}>
      <ol role="list" className="overflow-hidden">
        {items.map((item, index) => {
          if (index == items.length - 1) {
            return (
              <li className="relative" key={`${item.title}-${index}`}>
                <a href="#" className="group relative flex items-start">
                  <span className="flex h-9 items-center" aria-hidden="true">
                    <span className="relative z-10 flex h-8 w-8 items-center justify-center  ">
                      <CheckCircledIcon className="bg-white h-8 w-8 bg:white dark:bg-black" />
                    </span>
                  </span>
                  <span className="ml-4 flex min-w-0 flex-col">
                    <span className="text-sm font-medium">{item.title}</span>
                    <span className="text-sm text-gray-500">
                      {item.description}
                    </span>
                  </span>
                </a>
              </li>
            );
          }
          return (
            <li className="relative pb-10" key={`${item.title}-${index}`}>
              <div
                className="absolute left-4 top-8 -ml-px mt-0.5 h-full w-0.5 bg-gray-300"
                aria-hidden="true"
              ></div>
              <a
                href="#"
                className="group relative flex items-start"
                aria-current="step"
              >
                <span className="flex h-9 items-center" aria-hidden="true">
                  <span className="relative z-10 flex h-8 w-8 items-center justify-center ">
                    <CheckCircledIcon className="bg-white h-8 w-8 dark:bg-black" />
                  </span>
                </span>
                <span className="ml-4 flex min-w-0 flex-col">
                  <span className="text-sm font-medium">{item.title}</span>
                  <span className="text-sm text-gray-500">
                    {item.description}
                  </span>
                </span>
              </a>
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
