// "use client";
// import WaitlistForm from "@/components/buttons/WaitlistForm";
// import CTA from "@/components/cta/CTA";
// import { useRouter } from "next/navigation";
// import React, { ReactElement } from "react";
// import { IoCheckmarkSharp } from "react-icons/io5";

import { ReactElement } from 'react';

export default function Pricing(): ReactElement {
  return <h1>Pricing</h1>;
}

// const pricingPlans = [
//   {
//     name: "Starter",
//     description: "Best for small teams.",
//     features: [
//       "1 Job",
//       "Pre-built Transformers",
//       "Unlimited Integrations",
//       "3 users",
//       "Real time Run Logs",
//       "Community Slack Support",
//     ],
//     lockedFeatures: ["Audit logging", "Custom Transformers", "Private Support"],
//     pricing: "Free",
//   },
//   {
//     name: "Professional",
//     description: "Best for growing teams.",
//     features: [
//       "All Basic features",
//       "3 Jobs",
//       "Custom Transformers",
//       "10 users",
//       "Audit Logging",
//       "Private Slack Support",
//     ],
//     lockedFeatures: ["Unlimited Jobs, Dedicated infrastructure "],
//     pricing: "$0.024/record",
//   },
//   {
//     name: "Enterprise",
//     description: "Best for sophisticated teams.",
//     features: [
//       "All Professional features",
//       "Unlimited Jobs",
//       "Unlimited Users",
//       "SSO",
//       "Custom audit requirements",
//       "Data residency",
//       "Dedicated infrastructure",
//     ],
//     lockedFeatures: [],
//     pricing: "Custom",
//   },
// ];

// export default function Pricing() {
//   return (
//     <Box className="mainContainer">
//       <Stack direction="column" align="center" pt="40">
//         <MainGradient top="30%" left="40%" blur="200" height="30%" />
//         <MainGradient top="12%" left="-20%" blur="200" height="30%" />
//         <MainGradient top="50%" left="90%" blur="100" height="30%" />
//         <Box>
//           <Text fontFamily="Satoshi-Regular" textStyle="h1">
//             Simple, Transparent Pricing
//           </Text>
//         </Box>
//         <Box pt="10">
//           <Text textStyle="h3" fontFamily="Satoshi-Regular">
//             Easy and free to get started and scales with you as you grow.
//           </Text>
//         </Box>
//         <Box pt="20" zIndex="3">
//           <PricingPlanSection />
//         </Box>
//       </Stack>
//       <Box zIndex="3" w="100%">
//         <CTA />
//       </Box>
//     </Box>
//   );
// }

// function PricingPlanSection(): ReactElement {
//   return (
//     <Box>
//       <Stack
//         direction={{ base: "column", lg: "row" }}
//         justifyContent="center"
//         spacing="48px"
//       >
//         {pricingPlans.map((plan) => (
//           <Box
//             bgGradient="linear-gradient(111.31deg, #26262C -4.83%, #1D1A27 108.56%)"
//             borderRadius="8px"
//             borderWidth="1px"
//             borderColor="gray.600"
//             w="350px"
//             key={plan.name}
//           >
//             <Stack direction="column" align="left" px="24px" pt="10">
//               <Text textStyle="h2">{plan.name}</Text>
//               <Text textStyle="h3">{plan.description}</Text>
//               <Text textStyle="h1">{plan.pricing}</Text>
//               <Divider />
//               <Box pb="40px" pt="24px">
//                 <FeatureCheckmarkColumn data={plan.features} />
//               </Box>
//             </Stack>
//           </Box>
//         ))}
//       </Stack>
//     </Box>
//   );
// }

// interface FeatureProps {
//   data: string[];
// }

// function FeatureCheckmarkColumn(props: FeatureProps): ReactElement {
//   const { data } = props;
//   return (
//     <Stack direction="column">
//       <Box>
//         {data.map((item) => (
//           <Box key={item}>
//             <Stack
//               alignItems="center"
//               direction="row"
//               spacing={2}
//               py="12px"
//               borderBottomColor="rgba(41, 41, 41, 1)"
//               borderBottomWidth="1px"
//             >
//               <Icon as={IoCheckmarkSharp} color="rgba(227, 232, 239, 1)" />
//               <Text textStyle="h4" color="#E3E8EF" fontFamily="Satoshi-Regular">
//                 {item}
//               </Text>
//             </Stack>
//           </Box>
//         ))}
//       </Box>
//     </Stack>
//   );
// }
