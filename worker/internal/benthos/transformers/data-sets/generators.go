package transformers_dataset

//go:generate go run data_generator.go datasets/emaildomains.txt EmailDomain $GOPACKAGE
//go:generate go run data_generator.go datasets/first_names.txt FirstName $GOPACKAGE
//go:generate go run data_generator.go datasets/last_names.txt LastName $GOPACKAGE

// type Corpus struct {
// 	Values     []string
// 	LengthMap  map[int64][2]int
// 	MapIndices []int64
// }

// func (c *Corpus) GetValuesInRange(leftKey, rightKey int64) []string {
// 	range1, ok := c.LengthMap[leftKey]
// 	if !ok {
// 		return []string{}
// 	}
// 	range2, ok := c.LengthMap[rightKey]
// 	if !ok {
// 		return []string{}
// 	}
// 	minIdx := range1[0]
// 	maxIdx := range2[1]
// }
