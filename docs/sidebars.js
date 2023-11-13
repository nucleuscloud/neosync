/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */

const sidebars = {
  // By default, Docusaurus generates a sidebar from the docs folder structure

  mainSideBar: [
    {
      type: 'html',
      value: '<div>Overview</div>',
      className: 'sidebarcategory',
    },
    {
      type: 'doc',
      id: 'overview/intro', // document ID
      label: 'Introduction', // sidebar label
    },
    {
      type: 'doc',
      id: 'overview/platform',
      label: 'Platform',
    },
    {
      type: 'category',
      label: 'Use cases',
      collapsible: true,
      collapsed: true,
      items: [
        {
          type: 'doc',
          id: 'overview/use-cases/anonymization',
          label: 'Anonymize Data',
        },
        {
          type: 'doc',
          id: 'overview/use-cases/synthetic-data',
          label: 'Synthetic Data',
        },
        {
          type: 'doc',
          id: 'overview/use-cases/subsetting',
          label: 'Subset Data',
        },
        {
          type: 'doc',
          id: 'overview/use-cases/replication',
          label: 'Replicate Data',
        },
      ],
    },
    {
      type: 'html',
      value: '<div>Deploy Neosync</div>',
      className: 'sidebarcategory',
    },
    {
      type: 'doc',
      id: 'deploy/intro',
      label: 'Introduction',
    },
    {
      type: 'doc',
      id: 'deploy/kubernetes',
      label: 'Kubernetes',
    },
    {
      type: 'doc',
      id: 'deploy/docker-compose',
      label: 'Docker Compose',
    },
    {
      type: 'html',
      value: '<div>Connections</div>',
      className: 'sidebarcategory',
    },
    {
      type: 'doc',
      id: 'connections/postgres',
      label: 'Postgres',
    },
    {
      type: 'doc',
      id: 'connections/mysql',
      label: 'Mysql',
    },
    {
      type: 'doc',
      id: 'connections/s3',
      label: 'S3',
    },
    {
      type: 'html',
      value: '<div>Transformers</div>',
      className: 'sidebarcategory',
    },
    {
      type: 'doc',
      id: 'transformers/introduction',
      label: 'Introduction',
    },
    {
      type: 'category',
      label: 'System',
      collapsible: true,
      collapsed: true,
      items: [
        {
          type: 'doc',
          id: 'transformers/system/email',
          label: 'Email',
        },
        {
          type: 'doc',
          id: 'transformers/system/physical-address',
          label: 'Physical Address',
        },
        {
          type: 'doc',
          id: 'transformers/system/int-phone',
          label: 'Phone (integer)',
        },
        {
          type: 'doc',
          id: 'transformers/system/string-phone',
          label: 'Phone (string)',
        },
        {
          type: 'doc',
          id: 'transformers/system/city',
          label: 'City',
        },
        {
          type: 'doc',
          id: 'transformers/system/full-address',
          label: 'Full Address',
        },
        {
          type: 'doc',
          id: 'transformers/system/card-number',
          label: 'Card Number',
        },
        {
          type: 'doc',
          id: 'transformers/system/first-name',
          label: 'First Name',
        },
        {
          type: 'doc',
          id: 'transformers/system/full-name',
          label: 'Full Name',
        },
        {
          type: 'doc',
          id: 'transformers/system/ssn',
          label: 'SSN',
        },
      ],
    },
    {
      type: 'doc',
      id: 'transformers/custom',
      label: 'Custom',
    },
  ],
};

module.exports = sidebars;
