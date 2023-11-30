import { ReactElement } from 'react';
import CustomGenerateCardNumberForm from './CustomGenerateCardNumber';
import CustomGenerateE164NumberForm from './CustomGenerateE164NumberForm';
import CustomGenerateFloatForm from './CustomGenerateFloatForm';
import CustomGenerateGenderForm from './CustomGenerateGenderForm';
import CustomGenerateIntForm from './CustomGenerateIntForm';
import CustomGenerateStringForm from './CustomGenerateStringForm';
import CustomGenerateStringPhoneNumberForm from './CustomGenerateStringPhoneForm';
import CustomGenerateUuidForm from './CustomGenerateUuidForm';
import CustomTransformE164NumberForm from './CustomTransformE164PhoneForm';
import CustomTransformEmailForm from './CustomTransformEmailForm';
import CustomTransformFirstNameForm from './CustomTransformFirstNameForm';
import CustomTransformFloatForm from './CustomTransformFloatForm';
import CustomTransformFullNameForm from './CustomTransformFullNameForm';
import CustomTransformIntForm from './CustomTransformIntForm';
import CustomTransformIntPhoneNumberForm from './CustomTransformIntPhoneForm';
import CustomTransformLastNameForm from './CustomTransformLastNameForm';
import CustomTransformPhoneForm from './CustomTransformPhoneForm';
import CustomTransformStringForm from './CustomTransformStringForm';

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
