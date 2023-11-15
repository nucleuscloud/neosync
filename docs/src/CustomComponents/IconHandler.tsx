import {
  HomeIcon,
  LayersIcon,
  LinkBreak1Icon,
  Share1Icon,
  TokensIcon,
} from '@radix-ui/react-icons';
import React, { ReactElement } from 'react';
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
import { GrMysql, GrSecure } from 'react-icons/gr';
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

export function IconHandler(name: string): ReactElement {
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
    case 'Hash':
      return <GrSecure />;
    default:
      return <LayersIcon />;
  }
}
