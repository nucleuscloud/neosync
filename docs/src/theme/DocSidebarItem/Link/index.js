import React from "react";
import clsx from "clsx";
import { ThemeClassNames } from "@docusaurus/theme-common";
import { isActiveSidebarItem } from "@docusaurus/theme-common/internal";
import Link from "@docusaurus/Link";
import isInternalUrl from "@docusaurus/isInternalUrl";
import IconExternalLink from "@theme/Icon/ExternalLink";
import styles from "./styles.module.css";
import { HomeIcon, LayersIcon, Share1Icon } from "@radix-ui/react-icons";
import { FaAws, FaDocker, FaRegAddressCard } from "react-icons/fa";
import { SiKubernetes } from "react-icons/si";
import { BiLogoPostgresql } from "react-icons/bi";
import { GrMysql, GrSecure } from "react-icons/gr";
import { AiOutlineMail, AiOutlinePhone } from "react-icons/ai";
import { GoCode } from "react-icons/go";
import { MdPassword } from "react-icons/md";

export default function DocSidebarItemLink({
  item,
  onItemClick,
  activePath,
  level,
  index,
  ...props
}) {
  const { href, label, className, autoAddBaseUrl } = item;
  const isActive = isActiveSidebarItem(item, activePath);
  const isInternalLink = isInternalUrl(href);

  return (
    <li
      className={clsx(
        ThemeClassNames.docs.docSidebarItemLink,
        ThemeClassNames.docs.docSidebarItemLinkLevel(level),
        "menu__list-item",
        className
      )}
      key={label}
    >
      <Link
        className={clsx(
          "menu__link",
          !isInternalLink && styles.menuExternalLink,
          {
            "menu__link--active": isActive,
          }
        )}
        autoAddBaseUrl={autoAddBaseUrl}
        aria-current={isActive ? "page" : undefined}
        to={href}
        {...(isInternalLink && {
          onClick: onItemClick ? () => onItemClick(item) : undefined,
        })}
        {...props}
      >
        <div className="gap-4 flex flex-row items-center font-normal">
          {RenderIcon(item.label)}
          {label}
          {!isInternalLink && <IconExternalLink />}
        </div>
      </Link>
    </li>
  );
}

//when adding new side links, add an icon to the switch here

const RenderIcon = (name) => {
  switch (name) {
    case "Platform":
      return <LayersIcon />;
    case "Introduction":
      return <HomeIcon />;
    case "Architecture":
      return <Share1Icon />;
    case "Kubernetes":
      return <SiKubernetes />;
    case "Docker Compose":
      return <FaDocker />;
    case "Postgres":
      return <BiLogoPostgresql />;
    case "Mysql":
      return <GrMysql />;
    case "S3":
      return <FaAws />;
    case "Email":
      return <AiOutlineMail />;
    case "Phone":
      return <AiOutlinePhone />;
    case "SSN":
      return <MdPassword />;
    case "Physical Address":
      return <FaRegAddressCard />;
    case "Custom":
      return <GoCode />;
    default:
      return <LayersIcon />;
  }
};
