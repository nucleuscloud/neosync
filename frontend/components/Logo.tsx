'use client';
import { useTheme } from 'next-themes';
import Image from 'next/image';
import { useEffect, useState } from 'react';

interface LogoProps {
  className?: string;
}

export default function Logo({ className }: LogoProps) {
  const { resolvedTheme } = useTheme();

  const [src, setSrc] = useState(
    'https://assets.nucleuscloud.com/neosync/neosync_black.svg'
  );

  useEffect(() => {
    if (resolvedTheme === 'dark') {
      setSrc('https://assets.nucleuscloud.com/neosync/neosync_white.svg');
    } else {
      setSrc('https://assets.nucleuscloud.com/neosync/neosync_black.svg');
    }
  }, [resolvedTheme]);

  return (
    <Image
      src={src}
      alt="NeosyncLogo"
      className={className}
      width={64}
      height={20}
    />
  );
}
