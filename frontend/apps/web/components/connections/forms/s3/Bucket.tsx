import FormErrorMessage from '@/components/FormErrorMessage';
import FormHeader from '@/components/forms/FormHeader';
import { Input } from '@/components/ui/input';
import { AwsFormValues } from '@/yup-validations/connections';
import { ReactElement } from 'react';

interface Props {
  value: AwsFormValues['s3'];
  onChange: (value: AwsFormValues['s3']) => void;
  errors: Record<string, string>;
}

export default function Bucket(props: Props): ReactElement<any> {
  const { value, onChange, errors } = props;

  return (
    <>
      <div className="space-y-2">
        <FormHeader
          htmlFor="bucket"
          title="Bucket"
          description="The S3 bucket"
          isErrored={!!errors['s3.bucket']}
          isRequired={true}
        />
        <Input
          id="bucket"
          value={value.bucket || ''}
          onChange={(e) => onChange({ ...value, bucket: e.target.value })}
        />
        <FormErrorMessage message={errors['s3.bucket']} />
      </div>
      <div className="space-y-2">
        <FormHeader
          htmlFor="pathPrefix"
          title="Path Prefix"
          description="The path prefix for the bucket"
          isErrored={!!errors['s3.pathPrefix']}
        />
        <Input
          id="pathPrefix"
          value={value.pathPrefix || ''}
          onChange={(e) => onChange({ ...value, pathPrefix: e.target.value })}
        />
        <FormErrorMessage message={errors['s3.pathPrefix']} />
      </div>
    </>
  );
}
