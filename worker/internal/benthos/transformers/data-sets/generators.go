package transformers_dataset

//go:generate go run data_generator.go datasets/emaildomains.txt EmailDomain $GOPACKAGE
//go:generate go run data_generator.go datasets/first_names.txt FirstName $GOPACKAGE
//go:generate go run data_generator.go datasets/last_names.txt LastName $GOPACKAGE
