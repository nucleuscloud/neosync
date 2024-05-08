import { TransformerSource } from '@neosync/sdk';
import { ReactElement } from 'react';
import UserDefinedGenerateCardNumberForm from './UserDefinedGenerateCardNumber';
import UserDefinedGenerateCategoricalForm from './UserDefinedGenerateCategoricalForm';
import UserDefinedGenerateEmailForm from './UserDefinedGenerateEmailForm';
import UserDefinedGenerateFloat64Form from './UserDefinedGenerateFloat64Form';
import UserDefinedGenerateGenderForm from './UserDefinedGenerateGenderForm';
import UserDefinedGenerateInt64Form from './UserDefinedGenerateInt64Form';
import UserDefinedGenerateInternationalPhoneNumberForm from './UserDefinedGenerateInternationalPhoneNumberForm';
import UserDefinedGenerateJavascriptForm from './UserDefinedGenerateJavascriptForm';
import UserDefinedGenerateStringForm from './UserDefinedGenerateStringForm';
import UserDefinedGenerateStringPhoneNumberNumberForm from './UserDefinedGenerateStringPhoneNumberForm';
import UserDefinedGenerateUuidForm from './UserDefinedGenerateUuidForm';
import UserDefinedTransformCharacterScrambleForm from './UserDefinedTransformCharacterScrambleForm';
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

interface Props {
  value: TransformerSource;
  disabled?: boolean;
}
// handles rendering custom transformer configs
export function UserDefinedTransformerForm(props: Props): ReactElement {
  const { value, disabled } = props;
  switch (value) {
    case TransformerSource.GENERATE_CARD_NUMBER:
      return <UserDefinedGenerateCardNumberForm isDisabled={disabled} />;
    case TransformerSource.GENERATE_CATEGORICAL:
      return <UserDefinedGenerateCategoricalForm isDisabled={disabled} />;
    case TransformerSource.GENERATE_E164_PHONE_NUMBER:
      return (
        <UserDefinedGenerateInternationalPhoneNumberForm
          isDisabled={disabled}
        />
      );
    case TransformerSource.GENERATE_FLOAT64:
      return <UserDefinedGenerateFloat64Form isDisabled={disabled} />;
    case TransformerSource.GENERATE_GENDER:
      return <UserDefinedGenerateGenderForm isDisabled={disabled} />;
    case TransformerSource.GENERATE_INT64:
      return <UserDefinedGenerateInt64Form isDisabled={disabled} />;
    case TransformerSource.GENERATE_RANDOM_STRING:
      return <UserDefinedGenerateStringForm isDisabled={disabled} />;
    case TransformerSource.GENERATE_STRING_PHONE_NUMBER:
      return (
        <UserDefinedGenerateStringPhoneNumberNumberForm isDisabled={disabled} />
      );
    case TransformerSource.GENERATE_UUID:
      return <UserDefinedGenerateUuidForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_E164_PHONE_NUMBER:
      return <UserDefinedTransformE164NumberForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_EMAIL:
      return <UserDefinedTransformEmailForm isDisabled={disabled} />;
    case TransformerSource.GENERATE_EMAIL:
      return <UserDefinedGenerateEmailForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_FIRST_NAME:
      return <UserDefinedTransformFirstNameForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_FLOAT64:
      return <UserDefinedTransformFloat64Form isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_FULL_NAME:
      return <UserDefinedTransformFullNameForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_INT64:
      return <UserDefinedTransformInt64Form isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_INT64_PHONE_NUMBER:
      return <UserDefinedTransformIntPhoneNumberForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_LAST_NAME:
      return <UserDefinedTransformLastNameForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_PHONE_NUMBER:
      return <UserDefinedTransformPhoneNumberForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_STRING:
      return <UserDefinedTransformStringForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_JAVASCRIPT:
      return <UserDefinedTransformJavascriptForm isDisabled={disabled} />;
    case TransformerSource.TRANSFORM_CHARACTER_SCRAMBLE:
      return (
        <UserDefinedTransformCharacterScrambleForm isDisabled={disabled} />
      );
    case TransformerSource.GENERATE_JAVASCRIPT:
      return <UserDefinedGenerateJavascriptForm isDisabled={disabled} />;
    default:
      <div>No transformer found</div>;
  }
  return <div></div>;
}
