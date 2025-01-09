package transformers_dataset

//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/emaildomains.txt EmailDomain $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/first_names.txt FirstName $GOPACKAGE
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/last_names.txt LastName $GOPACKAGE
//go:generate go run ../../../../../tools/generators/business_names/main.go datasets/business_names.txt
//go:generate go run ../../../../../tools/generators/dataset_generator/main.go datasets/business_names.txt BusinessName $GOPACKAGE
