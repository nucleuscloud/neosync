import { useColorMode } from '@docusaurus/theme-common';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import React, { ReactElement } from 'react';

export default function Gitlink(): ReactElement {
  const { colorMode } = useColorMode();
  return (
    <a
      href="https://github.com/nucleuscloud/neosync"
      className="hover:no-underline"
    >
      {colorMode == 'light' ? (
        <div className="flex flex-row items-center rounded-full hover:bg-gray-100  no-underline text-black">
          <GitHubLogoIcon width={20} height={20} />
        </div>
      ) : (
        <div className="flex flex-row items-center rounded-full text-gray-100  no-underline hover:bg-gray-700">
          <GitHubLogoIcon />
        </div>
      )}
    </a>
  );
}
