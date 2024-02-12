import { cn } from '@/lib/utils';

export const BentoGrid = ({
  className,
  children,
}: {
  className?: string;
  children?: React.ReactNode;
}) => {
  return (
    <div
      className={cn(
        'grid md:auto-rows-[18rem] lg:auto-rows-[26rem] grid-cols-1 md:grid-cols-3 gap-4 max-w-7xl mx-auto',
        className
      )}
    >
      {children}
    </div>
  );
};

export const BentoGridItem = ({
  className,
  title,
  description,
  header,
  icon,
}: {
  className?: string;
  title?: string | React.ReactNode;
  description?: string | React.ReactNode;
  header?: React.ReactNode;
  icon?: React.ReactNode;
}) => {
  return (
    <div
      className={cn(
        'row-span-1 rounded-xl shadow-xl p-4 bg-gradient-to-tr from-[#1E1E1E] to-[#232222] border border-gray-600 justify-between flex flex-col space-y-4',
        className
      )}
    >
      {header}
      <div>
        {icon}
        <div className="font-sans font-bold text-gray-100 mb-2 mt-2">
          {title}
        </div>
        <div className="font-sans font-normal text-gray-100 text-sm">
          {description}
        </div>
      </div>
    </div>
  );
};
