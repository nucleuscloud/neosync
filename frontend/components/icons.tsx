'use client';
import Image from 'next/image';

// type IconProps = React.HTMLAttributes<SVGElement>;

interface LogoProps {
  theme?: string;
  className?: string;
}

export const Icons = {
  logo: ({ theme, className }: LogoProps) => {
    const src =
      theme === 'light'
        ? 'https://assets.nucleuscloud.com/neosync/neosync_black.svg'
        : 'https://assets.nucleuscloud.com/neosync/neosync_white.svg';
    return (
      <Image
        src={src}
        alt="NeosyncLogo"
        className={className}
        width="64"
        height="20"
      />
    );
  },
};
