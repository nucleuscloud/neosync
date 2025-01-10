package transformers_dataset

//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/emaildomains.txt EmailDomain $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/first_names.txt FirstName $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/last_names.txt LastName $GOPACKAGE
//go:generate go run ../../../../../tools/generators/business_names/main.go datasets/business_names.txt
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/business_names.txt BusinessName $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/us_states.txt UsState $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/us_state_codes.txt UsStateCode $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/us_areacodes.txt UsAreaCode $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/countries.txt Country $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/country_codes.txt CountryCode $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/addresses/address1.txt Address_Address1 $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/addresses/address2.txt Address_Address2 $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/addresses/city.txt Address_City $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/addresses/state.txt Address_State $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/addresses/zipcode.txt Address_ZipCode $GOPACKAGE
