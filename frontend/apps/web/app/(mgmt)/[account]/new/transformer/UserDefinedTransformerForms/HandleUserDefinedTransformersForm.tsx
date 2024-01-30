import { ReactElement } from 'react';
import UserDefinedGenerateCardNumberForm from './UserDefinedGenerateCardNumber';
import UserDefinedGenerateCategoricalForm from './UserDefinedGenerateCategoricalForm';
import UserDefinedGenerateE164NumberForm from './UserDefinedGenerateE164PhoneNumberForm';
import UserDefinedGenerateFloat64Form from './UserDefinedGenerateFloat64Form';
import UserDefinedGenerateGenderForm from './UserDefinedGenerateGenderForm';
import UserDefinedGenerateInt64Form from './UserDefinedGenerateInt64Form';
import UserDefinedGenerateStringForm from './UserDefinedGenerateStringForm';
import UserDefinedGenerateStringPhoneNumberNumberForm from './UserDefinedGenerateStringPhoneNumberForm';
import UserDefinedGenerateUuidForm from './UserDefinedGenerateUuidForm';
import UserDefinedTransformE164NumberForm from './UserDefinedTransformE164PhoneNumberForm';
import UserDefinedTransformEmailForm from './UserDefinedTransformEmailForm';
import UserDefinedTransformFirstNameForm from './UserDefinedTransformFirstNameForm';
import UserDefinedTransformFloat64Form from './UserDefinedTransformFloat64Form';
import UserDefinedTransformFullNameForm from './UserDefinedTransformFullNameForm';
import UserDefinedTransformInt64Form from './UserDefinedTransformInt64Form';
import UserDefinedTransformIntPhoneNumberForm from './UserDefinedTransformInt64PhoneForm';
import UserDefinedTransformJavascriptForm from './UserDefinedTransformJavascriptForm';
import UserDefinedTransformLastNameForm from './UserDefinedTransformLastNameForm';
import UserDefinedTransformPhoneNumberForm from './UserDefinedTransformPhoneNumberForm';
import UserDefinedTransformStringForm from './UserDefinedTransformStringForm';

// handles rendering custom transformer configs
export function handleUserDefinedTransformerForm(
  value: string | undefined,
  disabled?: boolean
): ReactElement {
  switch (value) {
    case 'generate_card_number':
      return <UserDefinedGenerateCardNumberForm isDisabled={disabled} />;
    case 'generate_categorical':
      return <UserDefinedGenerateCategoricalForm isDisabled={disabled} />;
    case 'generate_e164_phone_number':
      return <UserDefinedGenerateE164NumberForm isDisabled={disabled} />;
    case 'generate_float64':
      return <UserDefinedGenerateFloat64Form isDisabled={disabled} />;
    case 'generate_gender':
      return <UserDefinedGenerateGenderForm isDisabled={disabled} />;
    case 'generate_int64':
      return <UserDefinedGenerateInt64Form isDisabled={disabled} />;
    case 'generate_string':
      return <UserDefinedGenerateStringForm isDisabled={disabled} />;
    case 'generate_string_phone_number':
      return (
        <UserDefinedGenerateStringPhoneNumberNumberForm isDisabled={disabled} />
      );
    case 'generate_uuid':
      return <UserDefinedGenerateUuidForm isDisabled={disabled} />;
    case 'transform_e164_phone_number':
      return <UserDefinedTransformE164NumberForm isDisabled={disabled} />;
    case 'transform_email':
      return <UserDefinedTransformEmailForm isDisabled={disabled} />;
    case 'transform_first_name':
      return <UserDefinedTransformFirstNameForm isDisabled={disabled} />;
    case 'transform_float64':
      return <UserDefinedTransformFloat64Form isDisabled={disabled} />;
    case 'transform_full_name':
      return <UserDefinedTransformFullNameForm isDisabled={disabled} />;
    case 'transform_int64':
      return <UserDefinedTransformInt64Form isDisabled={disabled} />;
    case 'transform_int64_phone_number':
      return <UserDefinedTransformIntPhoneNumberForm isDisabled={disabled} />;
    case 'transform_last_name':
      return <UserDefinedTransformLastNameForm isDisabled={disabled} />;
    case 'transform_phone_number':
      return <UserDefinedTransformPhoneNumberForm isDisabled={disabled} />;
    case 'transform_string':
      return <UserDefinedTransformStringForm isDisabled={disabled} />;
    case 'transform_javascript':
      return <UserDefinedTransformJavascriptForm isDisabled={disabled} />;
    default:
      <div>No transformer found</div>;
  }
  return <div></div>;
}
