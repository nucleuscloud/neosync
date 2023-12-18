import {
  AvatarIcon,
  GitHubLogoIcon,
  HomeIcon,
  LayersIcon,
  LinkBreak1Icon,
  ListBulletIcon,
  Share1Icon,
  StackIcon,
  TokensIcon,
} from '@radix-ui/react-icons';
import React, { ReactElement } from 'react';
import { AiOutlineMail, AiOutlinePhone } from 'react-icons/ai';
import { BiLogInCircle, BiLogoPostgresql, BiTerminal } from 'react-icons/bi';
import { BsFunnel, BsShieldCheck } from 'react-icons/bs';
import { FaAws, FaDocker, FaFolder } from 'react-icons/fa';
import { GoLightBulb, GoSync } from 'react-icons/go';

import { GoCode, GoTable, GoVersions } from 'react-icons/go';
import { GrMysql } from 'react-icons/gr';
import { IoBuildOutline } from 'react-icons/io5';
import { MdPassword, MdStart } from 'react-icons/md';
import { PiArrowsSplitLight, PiFlaskLight } from 'react-icons/pi';
import { SiGo, SiKubernetes, SiTypescript } from 'react-icons/si';
import { TbSdk, TbVariable } from 'react-icons/tb';

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
    case 'User Defined':
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
    case 'Neosync CLI':
      return <BiTerminal />;
    case 'Environment Variables':
      return <TbVariable />;
    case 'Reference':
      return <GoTable />;
    case 'login':
      return <BiLogInCircle />;
    case 'whoami':
      return <AvatarIcon />;
    case 'jobs':
      return <StackIcon />;
    case 'list':
      return <ListBulletIcon />;
    case 'trigger':
      return <MdStart />;
    case 'version':
      return <GoVersions />;
    case 'sync':
      return <GoSync />;
    case 'Core Concepts':
      return <GoLightBulb />;
    case 'Github Actions':
      return <GitHubLogoIcon />;
    case 'SDK':
      return <TbSdk />;
    case 'Go':
    case 'Golang':
      return <SiGo />;
    case 'TypeScript':
    case 'Typescript':
    case 'ts':
    case 'TS':
      return <SiTypescript />;
    case 'Protos':
    case '/mgmt/v1alpha1':
      return <FaFolder />;
    default:
      return <LayersIcon />;
  }
}
