import { useColorMode } from '@docusaurus/theme-common';
import { GitHubLogoIcon } from '@radix-ui/react-icons';
import React, { ReactElement } from 'react';

export default function Gitlink(): ReactElement {
  const { colorMode } = useColorMode();
  console.log('color', colorMode);
  return (
    <a
      href="https://github.com/nucleuscloud/neosync"
      className="hover:no-underline"
    >
      {colorMode == 'light' ? (
        <div className="flex flex-row items-center gap-2 mr-10 p-2 rounded-full hover:bg-gray-100  no-underline text-black">
          <GitHubLogoIcon />
        </div>
      ) : (
        <div className="flex flex-row items-center gap-2 mr-10 p-2 rounded-full text-gray-100  no-underline hover:bg-gray-700">
          <GitHubLogoIcon />
        </div>
      )}
    </a>
  );
}
