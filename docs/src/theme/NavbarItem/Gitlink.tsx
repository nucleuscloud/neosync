import Link from "@docusaurus/Link";
import { ExternalLinkIcon, GitHubLogoIcon } from "@radix-ui/react-icons";
import { ReactElement } from "react";
import React from "react";

export default function Gitlink(): ReactElement {
  return (
    <a href="https://github.com/nucleuscloud/neosync" className="no-underline">
      <div className="flex flex-row items-center gap-2 pr-10">
        <div>
          <GitHubLogoIcon />
        </div>
        <div>Github</div>
      </div>
    </a>
  );
}
