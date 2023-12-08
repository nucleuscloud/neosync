import { ReactElement } from 'react';
import CustomGenerateCardNumberForm from './UserDefinedGenerateCardNumber';
import CustomGenerateE164NumberForm from './UserDefinedGenerateE164NumberForm';
import CustomGenerateFloatForm from './UserDefinedGenerateFloatForm';
import CustomGenerateGenderForm from './UserDefinedGenerateGenderForm';
import CustomGenerateIntForm from './UserDefinedGenerateIntForm';
import CustomGenerateStringForm from './UserDefinedGenerateStringForm';
import CustomGenerateStringPhoneNumberForm from './UserDefinedGenerateStringPhoneForm';
import CustomGenerateUuidForm from './UserDefinedGenerateUuidForm';
import CustomTransformE164NumberForm from './UserDefinedTransformE164PhoneForm';
import CustomTransformEmailForm from './UserDefinedTransformEmailForm';
import CustomTransformFirstNameForm from './UserDefinedTransformFirstNameForm';
import CustomTransformFloatForm from './UserDefinedTransformFloatForm';
import CustomTransformFullNameForm from './UserDefinedTransformFullNameForm';
import CustomTransformIntForm from './UserDefinedTransformIntForm';
import CustomTransformIntPhoneNumberForm from './UserDefinedTransformIntPhoneForm';
import CustomTransformLastNameForm from './UserDefinedTransformLastNameForm';
import CustomTransformPhoneForm from './UserDefinedTransformPhoneForm';
import CustomTransformStringForm from './UserDefinedTransformStringForm';

// handles rendering custom tranformer configs
export function handleCustomTransformerForm(
  value: string | undefined,
  disabled?: boolean
): ReactElement {
  switch (value) {
    case 'generate_card_number':
      return <CustomGenerateCardNumberForm isDisabled={disabled} />;
    case 'generate_e164_number':
      return <CustomGenerateE164NumberForm isDisabled={disabled} />;
    case 'generate_float':
      return <CustomGenerateFloatForm isDisabled={disabled} />;
    case 'generate_gender':
      return <CustomGenerateGenderForm isDisabled={disabled} />;
    case 'generate_int':
      return <CustomGenerateIntForm isDisabled={disabled} />;
    case 'generate_string':
      return <CustomGenerateStringForm isDisabled={disabled} />;
    case 'generate_string_phone':
      return <CustomGenerateStringPhoneNumberForm isDisabled={disabled} />;
    case 'generate_uuid':
      return <CustomGenerateUuidForm isDisabled={disabled} />;
    case 'transform_e164_phone':
      return <CustomTransformE164NumberForm isDisabled={disabled} />;
    case 'transform_email':
      return <CustomTransformEmailForm isDisabled={disabled} />;
    case 'transform_first_name':
      return <CustomTransformFirstNameForm isDisabled={disabled} />;
    case 'transform_float':
      return <CustomTransformFloatForm isDisabled={disabled} />;
    case 'transform_full_name':
      return <CustomTransformFullNameForm isDisabled={disabled} />;
    case 'transform_int':
      return <CustomTransformIntForm isDisabled={disabled} />;
    case 'transform_int_phone':
      return <CustomTransformIntPhoneNumberForm isDisabled={disabled} />;
    case 'transform_last_name':
      return <CustomTransformLastNameForm isDisabled={disabled} />;
    case 'transform_phone':
      return <CustomTransformPhoneForm isDisabled={disabled} />;
    case 'transform_string':
      return <CustomTransformStringForm isDisabled={disabled} />;
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}
