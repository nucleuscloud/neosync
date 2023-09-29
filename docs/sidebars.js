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

  tutorialSidebar: [
    {
      type: "html",
      value: "<div>Overview</div>",
      className: "sidebarcategory",
    },
    {
      type: "doc",
      id: "overview/intro", // document ID
      label: "Introduction", // sidebar label
    },
    {
      type: "doc",
      id: "overview/platform",
      label: "Platform",
    },
    {
      type: "doc",
      id: "overview/architecture",
      label: "Architecture",
    },
    {
      type: "html",
      value: "<div>Concepts</div>",
      className: "sidebarcategory",
    },
    {
      type: "doc",
      id: "concepts/jobs",
      label: "Jobs",
    },
    {
      type: "doc",
      id: "concepts/connections",
      label: "Connections",
    },
    {
      type: "doc",
      id: "concepts/transformers",
      label: "Transformers",
    },
    {
      type: "html",
      value: "<div>Self-host Neosync</div>",
      className: "sidebarcategory",
    },
    {
      type: "doc",
      id: "self-host/intro",
      label: "Introduction",
    },
    {
      type: "doc",
      id: "self-host/kubernetes",
      label: "Kubernetes",
    },
    {
      type: "doc",
      id: "self-host/docker-compose",
      label: "Docker Compose",
    },
    {
      type: "html",
      value: "<div>Connections</div>",
      className: "sidebarcategory",
    },
    {
      type: "doc",
      id: "connections/postgres",
      label: "Postgres",
    },
    {
      type: "doc",
      id: "connections/mysql",
      label: "Mysql",
    },
    {
      type: "doc",
      id: "connections/s3",
      label: "S3",
    },
    {
      type: "html",
      value: "<div>Transformers</div>",
      className: "sidebarcategory",
    },
    {
      type: "category",
      label: "Pre-built",
      collapsible: true,
      collapsed: true,
      items: [
        {
          type: "doc",
          id: "transformers/pre-built/email",
          label: "Email",
        },
        {
          type: "doc",
          id: "transformers/pre-built/physical-address",
          label: "Physical Address",
        },
        {
          type: "doc",
          id: "transformers/pre-built/phone",
          label: "Phone",
        },
        {
          type: "doc",
          id: "transformers/pre-built/ssn",
          label: "SSN",
        },
      ],
    },
    {
      type: "doc",
      id: "transformers/custom",
      label: "Custom",
    },
  ],
};

module.exports = sidebars;
