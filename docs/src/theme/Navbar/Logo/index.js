import React from 'react';
import Logo from '@theme/Logo';
export default function NavbarLogo() {
  return (
    <Logo
      className="navbar__brand"
      imageClassName="navbar__logo w-10 h-6"
      titleClassName="navbar__title text--truncate"
    />
  );
}
