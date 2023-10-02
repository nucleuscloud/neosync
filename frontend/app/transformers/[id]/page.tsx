'use client';
import { PageProps } from '@/components/types';

export default function TransformerPage({ params }: PageProps) {
  const id = params?.id ?? '';
  // const { data, isLoading, mutate } = useGetTransformers(); //udpate with tranformesr

  // const { toast } = useToast();

  console.log('id', id);
  // if (id) {
  //   return <div>Not Found ... yet</div>;
  // }
  // if (isLoading) {
  //   return (
  //     <div className="mt-10">
  //       <SkeletonForm />
  //     </div>
  //   );
  // }
  // const tranformerComponent = getTransformerComponentDetails({
  //   transformer: data?.transformers[0]!,
  //   onSaved: (resp) => {
  //     mutate(
  //       new GetConnectionResponse({
  //         //udpate this to transformer
  //         connection: resp.connection,
  //       })
  //     );
  //     toast({
  //       title: 'Successfully updated transformer!',
  //       variant: 'default',
  //     });
  //   },
  //   onSaveFailed: (err) =>
  //     toast({
  //       title: 'Unable to update transformer',
  //       description: getErrorMessage(err),
  //       variant: 'destructive',
  //     }),
  //   extraPageHeading: (
  //     <div>
  //       <RemoveTransformerButton transformerID={id} />
  //     </div>
  //   ),
  // });

  // console.log('trans', tranformerComponent);
  return (
    // <OverviewContainer Header={tranformerComponent.header}>
    //   <div className="transformer-details-container">
    //     <div>
    //       <div className="flex flex-col">
    //         {/* <div>{tranformerComponent.body}</div> */}
    //         <div>this is the transformer body</div>
    //       </div>
    //     </div>
    //   </div>
    // </OverviewContainer>
    <div>this is the overview</div>
  );
}
