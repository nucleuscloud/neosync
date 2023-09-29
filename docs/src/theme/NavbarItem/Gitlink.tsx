import { GitHubLogoIcon } from "@radix-ui/react-icons";
import React, { ReactElement } from "react";

export default function Gitlink(): ReactElement {
  return (
    <a
      href="https://github.com/nucleuscloud/neosync"
      className="hover:no-underline"
    >
      <div className="flex flex-row items-center gap-2 mr-10 p-2 rounded-full hover:bg-gray-100 hover:no-underline">
        <GitHubLogoIcon />
        <div className="text-sm">Support us on Github</div>
      </div>
    </a>
  );
}
