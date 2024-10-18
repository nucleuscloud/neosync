import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "reference/mgmt-v-1-alpha-1",
    },
    {
      type: "category",
      label: "mgmt.v1alpha1.AnonymizationService",
      items: [
        {
          type: "doc",
          id: "reference/mgmt-v-1-alpha-1-anonymization-service-anonymize-many",
          label: "AnonymizeMany",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "reference/mgmt-v-1-alpha-1-anonymization-service-anonymize-single",
          label: "AnonymizeSingle",
          className: "api-method post",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
