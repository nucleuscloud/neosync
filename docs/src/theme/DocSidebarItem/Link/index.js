import Link from '@docusaurus/Link';
import isInternalUrl from '@docusaurus/isInternalUrl';
import { ThemeClassNames } from '@docusaurus/theme-common';
import { isActiveSidebarItem } from '@docusaurus/theme-common/internal';
import {
  HomeIcon,
  LayersIcon,
  LinkBreak1Icon,
  Share1Icon,
  TokensIcon,
} from '@radix-ui/react-icons';
import IconExternalLink from '@theme/Icon/ExternalLink';
import clsx from 'clsx';
import React from 'react';
import {
  AiOutlineCreditCard,
  AiOutlineMail,
  AiOutlinePhone,
} from 'react-icons/ai';
import { BiLogoPostgresql, BiSolidCity, BiTimeFive } from 'react-icons/bi';
import {
  BsCalendarDate,
  BsFillKeyFill,
  BsFunnel,
  BsGenderAmbiguous,
  BsPinMap,
  BsShieldCheck,
} from 'react-icons/bs';
import {
  FaAws,
  FaDocker,
  FaRegAddressBook,
  FaStreetView,
} from 'react-icons/fa';
import { GiTexas } from 'react-icons/gi';
import { GoCode } from 'react-icons/go';
import { GrMysql } from 'react-icons/gr';
import { IoBuildOutline } from 'react-icons/io5';
import { MdPassword } from 'react-icons/md';
import { PiArrowsSplitLight, PiFlaskLight } from 'react-icons/pi';
import {
  RxAvatar,
  RxComponentBoolean,
  RxLetterCaseCapitalize,
} from 'react-icons/rx';
import { SiKubernetes } from 'react-icons/si';
import { TbDecimal } from 'react-icons/tb';
import { TiSortNumerically } from 'react-icons/ti';
import styles from './styles.module.css';

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
        'menu__list-item',
        className
      )}
      key={label}
    >
      <Link
        className={clsx(
          'menu__link',
          !isInternalLink && styles.menuExternalLink,
          {
            'menu__link--active': isActive,
          }
        )}
        autoAddBaseUrl={autoAddBaseUrl}
        aria-current={isActive ? 'page' : undefined}
        to={href}
        {...(isInternalLink && {
          onClick: onItemClick ? () => onItemClick(item) : undefined,
        })}
        {...props}
      >
        <div className="gap-4 flex flex-row items-center font-normal text-gray-800">
          {RenderIcon(item.label)}
          {label}
          {!isInternalLink && <IconExternalLink />}
        </div>
      </Link>
    </li>
  );
}

//when adding new side links, add an icon to the switch here

export const RenderIcon = (name) => {
  switch (name) {
    case 'Platform':
      return <TokensIcon />;
    case 'Introduction':
      return <HomeIcon />;
    case 'Architecture':
      return <Share1Icon />;
    case 'Kubernetes':
      return <SiKubernetes />;
    case 'Docker Compose':
      return <FaDocker />;
    case 'Postgres':
      return <BiLogoPostgresql />;
    case 'Mysql':
      return <GrMysql />;
    case 'S3':
      return <FaAws />;
    case 'Email':
      return <AiOutlineMail />;
    case 'Phone (integer)':
      return <AiOutlinePhone />;
    case 'Phone (string)':
      return <AiOutlinePhone />;
    case 'SSN':
      return <MdPassword />;
    case 'Custom':
      return <GoCode />;
    case 'System':
      return <IoBuildOutline />;
    case 'Use cases':
      return <BsShieldCheck />;
    case 'Anonymize Data':
      return <LinkBreak1Icon />;
    case 'Replicate Data':
      return <PiArrowsSplitLight />;
    case 'Synthetic Data':
      return <PiFlaskLight />;
    case 'Subset Data':
      return <BsFunnel />;
    case 'City':
      return <BiSolidCity />;
    case 'Card Number':
      return <AiOutlineCreditCard />;
    case 'First Name':
      return <RxAvatar />;
    case 'Full Name':
      return <RxAvatar />;
    case 'Last Name':
      return <RxAvatar />;
    case 'Full Address':
      return <FaRegAddressBook />;
    case 'Gender':
      return <BsGenderAmbiguous />;
    case 'Random Boolean':
      return <RxComponentBoolean />;
    case 'Random Float':
      return <TbDecimal />;
    case 'Random Integer':
      return <TiSortNumerically />;
    case 'Random String':
      return <RxLetterCaseCapitalize />;
    case 'State':
      return <GiTexas />;
    case 'Street Address':
      return <FaStreetView />;
    case 'Unix Timestamp':
      return <BiTimeFive />;
    case 'UTC Timestamp':
      return <BsCalendarDate />;
    case 'UUID':
      return <BsFillKeyFill />;
    case 'Zipcode':
      return <BsPinMap />;
    default:
      return <LayersIcon />;
  }
};
