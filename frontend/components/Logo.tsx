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
    'https://assets.nucleuscloud.com/neosync/newbrand/logo_and_test_light_mode.svg'
  );

  useEffect(() => {
    if (resolvedTheme === 'dark') {
      setSrc(
        'https://assets.nucleuscloud.com/neosync/newbrand/logo_text_dark_mode.svg'
      );
    } else {
      setSrc(
        'https://assets.nucleuscloud.com/neosync/newbrand/logo_and_test_light_mode.svg'
      );
    }
  }, [resolvedTheme]);

  return (
    <Image
      src={src}
      alt="NeosyncLogo"
      className={className}
      width={84}
      height={40}
    />
  );
}
