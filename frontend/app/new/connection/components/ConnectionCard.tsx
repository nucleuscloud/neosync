'use client';
import ConnectionIcon from '@/components/connections/ConnectionIcon';
import { Avatar } from '@/components/ui/avatar';
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { useRouter } from 'next/navigation';

export interface ConnectionMeta {
  name: string;
  description: string;
  urlSlug: string;
}

interface Props {
  connection: ConnectionMeta;
}

export default function ConnectionCard(props: Props) {
  const { connection } = props;
  const router = useRouter();
  return (
    <Card
      onClick={() => router.push(`/new/connection/${connection.urlSlug}`)}
      className="cursor-pointer"
    >
      <CardHeader>
        <CardTitle>
          <div className="flex flex-row items-center space-x-2">
            <Avatar>
              {/* <AvatarImage /> */}
              <ConnectionIcon name={connection.name} />
              {/* <AvatarFallback>{connection.name}</AvatarFallback> */}
            </Avatar>
            <p>{connection.name}</p>
          </div>
        </CardTitle>
        <CardDescription>{connection.description}</CardDescription>
      </CardHeader>
    </Card>
  );
}

// function ConnectionCard(props: ConnectionCardProps): ReactElement {
//   const { connection } = props;
//   return (
//     <NextLink href={`/new/connection/${formatUrlParam(connection.urlSlug)}`}>
//       <div
//         borderWidth="1px"
//         borderColor={borderColor}
//         borderRadius="10"
//         h="250px"
//         key={connection.name}
//         _hover={{ borderColor: 'purple.400' }}
//       >
//         <Stack
//           direction="column"
//           spacing={5}
//           p="5"
//           alignItems="flex-start"
//           h="100%"
//         >
//           <Box>
//             <ConnectionIcon name={connection.name ?? ''} />
//           </Box>
//           <Box>
//             <Text textStyle="h3" align="left">
//               {connection.name}
//             </Text>
//           </Box>
//           <Container p="0" m="0" alignItems="flex-start">
//             <Text align="left" textStyle="body">
//               {connection.description}
//             </Text>
//           </Container>
//         </Stack>
//       </Box>
//     </NextLink>
//   );
// }
