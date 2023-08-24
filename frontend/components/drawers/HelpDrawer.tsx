// import {
//   AskAiResponse,
//   GetLinkFromPageIdRequest,
//   GetLinkFromPageIdResponse,
// } from '@/api_types';
// import { isNotNil } from '@/libs/utils';
// import { titleCase } from '@/util/util';
// import {
//   ChevronDownIcon,
//   ChevronUpIcon,
//   ExternalLinkIcon,
// } from '@chakra-ui/icons';
// import {
//   Badge,
//   Box,
//   Button,
//   Collapse,
//   Drawer,
//   DrawerBody,
//   DrawerCloseButton,
//   DrawerContent,
//   DrawerHeader,
//   DrawerOverlay,
//   Flex,
//   Icon,
//   Input,
//   Progress,
//   Stack,
//   Text,
//   useColorModeValue,
//   useDisclosure,
//   useToast,
// } from '@chakra-ui/react';
// import NextLink from 'next/link';
// import { ReactElement, RefObject, useRef, useState } from 'react';
// import { BsMagic } from 'react-icons/bs';
// import { IoDocumentOutline } from 'react-icons/io5';
// import { Toast } from '../overlays/alerts/toasts';

// interface Props {
//   isOpen: boolean;
//   onClose: () => void;
//   finalFocusRef?: RefObject<HTMLButtonElement>;
// }

// export default function HelpDrawer(props: Props): ReactElement {
//   const { onClose, isOpen, finalFocusRef } = props;
//   const toast = useToast();
//   const [query, setQuery] = useState<string>('');
//   const [queryAnswer, setQueryAnswer] = useState<AskAiResponse>({
//     answer: { text: '', pages: [], followupQuestions: [] },
//   });
//   const [isGettingAnswer, setIsGettingAnswer] = useState<boolean>(false);
//   const [queryAnswerLinks, setQueryAnswerLinks] = useState<
//     GetLinkFromPageIdResponse[]
//   >([]);
//   const [error, setError] = useState<boolean>(false);
//   const initialFocusRef = useRef<HTMLInputElement | null>(null);

//   async function askAi(query?: string) {
//     setIsGettingAnswer(true);
//     try {
//       const data = await getAiResults(query);

//       setQueryAnswer(data);

//       const results = (
//         await Promise.all(
//           data.answer.pages.map(async (page) => {
//             return getLinkFromPage(page.space, page.page);
//           })
//         )
//       ).filter(isNotNil);
//       setQueryAnswerLinks(results);
//     } catch (err) {
//       setError(true);
//       Toast({
//         id: 'ask-ai-err',
//         status: 'error',
//         title:
//           'The Nucleus Help AI is unable to get an answer for your question. ',
//         description:
//           'This may be caused due to grammar issues or the AI not being able to detect the question. Try another question',
//         toast,
//       });
//     } finally {
//       setIsGettingAnswer(false);
//     }
//   }

//   return (
//     <Drawer
//       isOpen={isOpen}
//       placement="right"
//       onClose={onClose}
//       size="lg"
//       finalFocusRef={finalFocusRef}
//       initialFocusRef={initialFocusRef}
//     >
//       <DrawerOverlay />
//       <DrawerContent borderLeftRadius="8px">
//         <DrawerCloseButton />
//         <DrawerHeader borderBottomWidth="1px" borderBottomColor="gray.600">
//           <Stack direction="row" alignItems="center" spacing={5}>
//             <Icon as={BsMagic} />
//             <Text>Nucleus Help - Powered by AI</Text>
//             <Badge colorScheme="twitter" borderRadius="8px">
//               Beta
//             </Badge>
//           </Stack>
//         </DrawerHeader>
//         <DrawerBody pb="10">
//           <Stack direction="column">
//             <Box pt="6">
//               <Text>
//                 Ask a question to the Nucleus Help AI using natural language to
//                 get started.
//               </Text>
//               <Stack direction="row" alignItems="center" pt="6">
//                 <Input
//                   placeholder="Ask the AI a question ..."
//                   onChange={(e) => {
//                     setQuery(e.target.value);
//                   }}
//                   onKeyDown={(e) => {
//                     if (e.key === 'Enter' && query.length > 0) {
//                       askAi(query);
//                     }
//                   }}
//                   value={query}
//                   ref={initialFocusRef}
//                 />
//                 <Button onClick={() => askAi(query)}> Submit</Button>
//               </Stack>
//             </Box>
//             {isGettingAnswer && !error && (
//               <Flex pt="10" justifyContent="center">
//                 <Progress
//                   size="xs"
//                   isIndeterminate
//                   borderRadius="10px"
//                   w="80%"
//                   colorScheme="purple"
//                 />
//               </Flex>
//             )}
//             {queryAnswer.answer?.text &&
//             queryAnswerLinks.length > 0 &&
//             !isGettingAnswer ? (
//               <Box>
//                 <AnswerBox answer={queryAnswer.answer.text} />
//                 <PageSources links={queryAnswerLinks} />
//                 {queryAnswer.answer.followupQuestions.length > 0 ? (
//                   <FollowupQuestions
//                     questions={queryAnswer.answer.followupQuestions}
//                     askAi={askAi}
//                     setQuery={setQuery}
//                   />
//                 ) : null}
//               </Box>
//             ) : queryAnswer.answer?.followupQuestions.length > 0 ? (
//               <FollowupQuestions
//                 questions={queryAnswer.answer.followupQuestions}
//                 askAi={askAi}
//                 setQuery={setQuery}
//               />
//             ) : (
//               <FAQs askAi={askAi} setQuery={setQuery} />
//             )}
//           </Stack>
//         </DrawerBody>
//       </DrawerContent>
//     </Drawer>
//   );
// }

// async function getLinkFromPage(
//   spaceId: string,
//   pageId: string
// ): Promise<GetLinkFromPageIdResponse | undefined> {
//   const req: GetLinkFromPageIdRequest = {
//     spaceId,
//     pageId,
//   };
//   const response = await fetch(`/api/ai-help/get-link-from-pageId`, {
//     method: 'POST',
//     headers: {
//       'Content-Type': 'application/json',
//     },
//     body: JSON.stringify(req),
//   });
//   if (!response.ok) {
//     throw new Error('Request failed');
//   }
//   return await response.json();
// }

// async function getAiResults(query: string | undefined): Promise<AskAiResponse> {
//   const response = await fetch(`/api/ai-help/ask-ai`, {
//     method: 'POST',
//     headers: {
//       'Content-Type': 'application/json',
//     },
//     body: JSON.stringify({ query: query }),
//   });
//   if (!response.ok) {
//     throw new Error('Request failed');
//   }
//   return (await response.json()) as AskAiResponse;
// }

// interface AnswerBoxProps {
//   answer: string;
// }

// function AnswerBox(props: AnswerBoxProps): ReactElement {
//   const { answer } = props;
//   return (
//     <Box
//       overflowY="scroll"
//       h="70%"
//       fontSize="14px"
//       borderWidth="1px"
//       borderColor="gray.500"
//       borderRadius="8px"
//       p="2"
//       mt="10"
//     >
//       <pre style={{ whiteSpace: 'pre-wrap' }}>{answer}</pre>
//     </Box>
//   );
// }

// interface PageSourcesProps {
//   links: GetLinkFromPageIdResponse[];
// }

// function PageSources(props: PageSourcesProps): ReactElement {
//   const { links } = props;

//   const baseUrl = 'https://docs.nucleuscloud.com/';
//   const { isOpen, onToggle } = useDisclosure();
//   const [linksExpanded, setLinksExpanded] = useState<boolean>(false);
//   const borderColors = useColorModeValue('gray.200', 'gray.600');

//   return (
//     <Box pt="10">
//       <Button
//         onClick={() => {
//           onToggle();
//           setLinksExpanded(linksExpanded ? false : true);
//         }}
//         variant="unstyled"
//       >
//         <Stack direction="row" alignItems="center">
//           <Text fontSize="14px" textColor="blue.400">
//             Answer based on {links.length} sources
//           </Text>
//           {linksExpanded ? <ChevronDownIcon /> : <ChevronUpIcon />}
//         </Stack>
//       </Button>
//       <Collapse in={isOpen}>
//         <Stack direction="column" pt="5">
//           {links.map((link) => (
//             <NextLink
//               key={link?.id}
//               href={baseUrl + link?.path}
//               target="_blank"
//             >
//               <Box
//                 _hover={{ bg: 'whiteAlpha.200' }}
//                 p="2"
//                 borderRadius="8px"
//                 borderWidth="1px"
//                 borderColor={borderColors}
//               >
//                 <Stack direction="row" alignItems="center">
//                   <Stack direction="column">
//                     <Stack direction="row" alignItems="center">
//                       <Icon as={IoDocumentOutline} />
//                       <Text fontWeight="semibold">{link?.title}</Text>
//                       <ExternalLinkIcon />
//                     </Stack>
//                     <Text fontSize="12px" pl="6">
//                       {titleCase(link?.path.split('/')[0] ?? '')}
//                     </Text>
//                   </Stack>
//                 </Stack>
//               </Box>
//             </NextLink>
//           ))}
//         </Stack>
//       </Collapse>
//     </Box>
//   );
// }

// interface FollowupQuestionsProps {
//   questions: string[];
//   askAi: (val: string) => Promise<void>;
//   setQuery: (val: string) => void;
// }

// function FollowupQuestions(props: FollowupQuestionsProps): ReactElement {
//   const { questions, askAi, setQuery } = props;
//   return (
//     <Box>
//       <Stack direction="column" spacing={3}>
//         <Text pt="10" fontWeight="semibold">
//           Ask a follow up question ...{' '}
//         </Text>
//         {questions.map((question) => (
//           <Button
//             variant="outline"
//             key={question}
//             onClick={() => {
//               askAi(question);
//               setQuery(question);
//             }}
//             justifyContent="left"
//           >
//             {question}
//           </Button>
//         ))}
//       </Stack>
//     </Box>
//   );
// }

// interface FAQsProps {
//   askAi: (val: string) => Promise<void>;
//   setQuery: (val: string) => void;
// }

// function FAQs(props: FAQsProps): ReactElement {
//   const { askAi, setQuery } = props;
//   const faqs = [
//     'How do I create an environment?',
//     'How do I deploy a service?',
//     'How do I set an environment variable?',
//   ];
//   return (
//     <Box>
//       <Stack direction="column" spacing={3}>
//         <Text pt="10" fontWeight="semibold">
//           Ask a FAQ:
//         </Text>
//         {faqs.map((question) => (
//           <Button
//             variant="outline"
//             key={question}
//             onClick={() => {
//               askAi(question);
//               setQuery(question);
//             }}
//             justifyContent="left"
//           >
//             <Stack direction="row">
//               <Icon as={BsMagic} />
//               <Text>{question}</Text>
//             </Stack>
//           </Button>
//         ))}
//       </Stack>
//     </Box>
//   );
// }
