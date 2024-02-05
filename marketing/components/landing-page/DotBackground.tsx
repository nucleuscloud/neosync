import { ReactNode } from 'react';

interface Props {
  children: ReactNode;
}

export function DotBackground(props: Props) {
  const { children } = props;
  return (
    <div className="h-auto w-full dark:bg-black bg-white  dark:bg-dot-white/[0.2] bg-dot-black/[0.2] relative flex items-center justify-center">
      <div className="absolute pointer-events-none inset-0 flex py-10 items-center justify-center dark:bg-black bg-white [mask-image:radial-gradient(ellipse_at_center_700px,transparent_30%,black)]"></div>
      {children}
    </div>
  );
}
