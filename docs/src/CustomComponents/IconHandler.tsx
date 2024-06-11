import {
  AvatarIcon,
  GitHubLogoIcon,
  HomeIcon,
  LayersIcon,
  LinkBreak1Icon,
  ListBulletIcon,
  RocketIcon,
  Share1Icon,
  SymbolIcon,
  TokensIcon,
} from '@radix-ui/react-icons';
import React, { ReactElement } from 'react';
import {
  AiOutlineExperiment,
  AiOutlineMail,
  AiOutlinePhone,
  AiOutlineSolution,
} from 'react-icons/ai';
import {
  BiDownload,
  BiLogInCircle,
  BiLogoPostgresql,
  BiSolidWrench,
  BiTerminal,
} from 'react-icons/bi';
import { BsFunnel } from 'react-icons/bs';
import { CiMicrochip } from 'react-icons/ci';
import { DiMongodb } from 'react-icons/di';
import { FaAws, FaDocker, FaFolder, FaKey, FaLaptop } from 'react-icons/fa';
import { GoLightBulb, GoSync } from 'react-icons/go';

import { GoCode, GoTable, GoVersions } from 'react-icons/go';
import { GrAnalytics, GrMysql } from 'react-icons/gr';
import { IoBuildOutline } from 'react-icons/io5';
import {
  MdCenterFocusStrong,
  MdOutlineSchema,
  MdPassword,
  MdStart,
} from 'react-icons/md';
import { PiArrowsSplitLight, PiFlaskLight } from 'react-icons/pi';
import { RiOpenaiFill } from 'react-icons/ri';
import { SiGo, SiKubernetes, SiTerraform, SiTypescript } from 'react-icons/si';
import { TbCloudLock, TbSdk, TbVariable } from 'react-icons/tb';

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
    case 'Mongo':
    case 'MongoDB':
      return <DiMongodb />;
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
    case 'Installing':
      return <BiDownload />;
    case 'whoami':
      return <AvatarIcon />;
    case 'jobs':
    case 'Creating a Sync Job':
      return <SymbolIcon />;
    case 'Creating a Data Generation Job':
      return <AiOutlineExperiment />;
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
    case 'Using Neosync in CI':
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
    case 'Authentication':
      return <FaKey />;
    case 'Configuring Analytics':
      return <GrAnalytics />;
    case 'Developing Neosync Locally':
      return <FaLaptop />;
    case 'Neosync Terraform Provider':
      return <SiTerraform />;
    case 'Quickstart':
      return <RocketIcon />;
    case 'Using a Custom LLM Transformer':
      return <RiOpenaiFill />;
    case 'AI Synthetic Data Generation':
      return <CiMicrochip />;
    case 'Troubleshooting':
      return <BiSolidWrench />;
    case 'Initializing your Schema':
      return <MdOutlineSchema />;
    case 'Cloud Security Overview':
      return <TbCloudLock />;
    case 'Use cases':
      return <AiOutlineSolution />;
    case 'Core Features':
      return <MdCenterFocusStrong />;
    case 'Connect Postgres via Bastion Host':
      return <BiLogoPostgresql />;
    default:
      return <LayersIcon />;
  }
}
