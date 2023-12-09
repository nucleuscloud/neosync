import Logo from '@theme/Logo';
import React from 'react';
export default function NavbarLogo() {
  return (
    <Logo
      className="navbar__brand"
      imageClassName="navbar__logo w-[84px] h-[20px] flex"
      titleClassName="navbar__title text--truncate"
    />
  );
}
