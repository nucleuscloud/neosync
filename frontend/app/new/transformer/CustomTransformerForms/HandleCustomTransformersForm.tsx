import { ReactElement } from 'react';
import CustomCardNumberTransformerForm from './CustomCardNumberTransformerForm';
import CustomEmailTransformerForm from './CustomEmailTransformerForm';
import CustomFirstNameTransformerForm from './CustomFirstnameTransformerForm';
import CustomFullNameTransformerForm from './CustomFullnameTransformerForm';
import CustomGenderTransformerForm from './CustomGenderTransformerForm';
import CustomIntPhoneNumberTransformerForm from './CustomIntPhoneNumberTransformerForm';
import CustomLastNameTransformerForm from './CustomLastnameTransformerForm';
import CustomPhoneNumberTransformerForm from './CustomPhoneNumberTransformerForm';
import CustomRandomFloatTransformerForm from './CustomRandomFloatTransformerForm';
import CustomRandomIntTransformerForm from './CustomRandomIntTransformerForm';
import CustomRandomStringTransformerForm from './CustomRandomStringTransformerForm';
import CustomUuidTransformerForm from './CustomUuidTransformerForm';

// handles rendering custom tranformer configs
export function handleCustomTransformerForm(
  value: string | undefined,
  disabled?: boolean
): ReactElement {
  switch (value) {
    case 'email':
      return <CustomEmailTransformerForm isDisabled={disabled} />;
    case 'uuid':
      return <CustomUuidTransformerForm isDisabled={disabled} />;
    case 'first_name':
      return <CustomFirstNameTransformerForm isDisabled={disabled} />;
    case 'last_name':
      return <CustomLastNameTransformerForm isDisabled={disabled} />;
    case 'full_name':
      return <CustomFullNameTransformerForm isDisabled={disabled} />;
    case 'phone_number':
      return <CustomPhoneNumberTransformerForm isDisabled={disabled} />;
    case 'int_phone_number':
      return <CustomIntPhoneNumberTransformerForm isDisabled={disabled} />;
    case 'random_string':
      return <CustomRandomStringTransformerForm isDisabled={disabled} />;
    case 'random_int':
      return <CustomRandomIntTransformerForm isDisabled={disabled} />;
    case 'random_float':
      return <CustomRandomFloatTransformerForm isDisabled={disabled} />;
    case 'gender':
      return <CustomGenderTransformerForm isDisabled={disabled} />;
    case 'card_number':
      return <CustomCardNumberTransformerForm isDisabled={disabled} />;
    default:
      <div>No transformer component found</div>;
  }
  return <div></div>;
}
