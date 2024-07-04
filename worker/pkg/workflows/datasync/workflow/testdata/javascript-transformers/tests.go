package testdata_javascripttransformers

import (
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	workflow_testdata "github.com/nucleuscloud/neosync/worker/pkg/workflows/datasync/workflow/testdata"
)

func GetSyncTests() []*workflow_testdata.IntegrationTest {
	return []*workflow_testdata.IntegrationTest{
		{
			Name:            "Javascript transformer sync",
			Folder:          "javascript-transformers",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create.sql"},
			JobMappings:     getJsTransformerJobmappings(),
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"javascript.transformers": &workflow_testdata.ExpectedOutput{RowCount: 20},
			},
		},
		{
			Name:            "Javascript generator sync",
			Folder:          "javascript-transformers",
			SourceFilePaths: []string{"create.sql", "insert.sql"},
			TargetFilePaths: []string{"create.sql"},
			JobMappings:     getJsGeneratorJobmappings(),
			Expected: map[string]*workflow_testdata.ExpectedOutput{
				"javascript.transformers": &workflow_testdata.ExpectedOutput{RowCount: 20},
			},
		},
	}
}

func getJsGeneratorJobmappings() []*mgmtv1alpha1.JobMapping {
	colTransformerMap := map[string]*mgmtv1alpha1.JobMappingTransformer{
		"e164_phone_number":   getJavascriptTransformerConfig("return neosync.generateInternationalPhoneNumber({ min: 9, max: 15});"),
		"email":               getJavascriptTransformerConfig("return neosync.generateEmail({ maxLength: 255});"),
		"str":                 getJavascriptTransformerConfig("return neosync.generateRandomString({ min: 1, max: 50});"),
		"measurement":         getJavascriptTransformerConfig("return neosync.generateFloat64({ min: 3.14, max: 300.10});"),
		"int64":               getJavascriptTransformerConfig("return neosync.generateInt64({ min: 1, max: 50});"),
		"int64_phone_number":  getJavascriptTransformerConfig("return neosync.generateInt64PhoneNumber({});"),
		"string_phone_number": getJavascriptTransformerConfig("return neosync.generateStringPhoneNumber({ min: 1, max: 15});"),
		"first_name":          getJavascriptTransformerConfig("return neosync.generateFirstName({ maxLength: 25});"),
		"last_name":           getJavascriptTransformerConfig("return neosync.generateLastName({ maxLength: 25});"),
		"full_name":           getJavascriptTransformerConfig("return neosync.generateFullName({ maxLength: 25});"),
		"character_scramble":  getJavascriptTransformerConfig("return neosync.generateCity({ maxLength: 100});"),
	}
	jobmappings := GetDefaultSyncJobMappings()
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		if _, ok := colTransformerMap[jm.Column]; !ok {
			updatedJobmappings = append(updatedJobmappings, jm)
		} else {
			updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
				Schema:      jm.Schema,
				Table:       jm.Table,
				Column:      jm.Column,
				Transformer: colTransformerMap[jm.Column],
			})
		}
	}
	return updatedJobmappings
}

func getJsTransformerJobmappings() []*mgmtv1alpha1.JobMapping {
	colTransformerMap := map[string]*mgmtv1alpha1.JobMappingTransformer{
		"e164_phone_number":   getJavascriptTransformerConfig("return neosync.transformE164PhoneNumber(value, { preserveLength: true, maxLength: 20});"),
		"email":               getJavascriptTransformerConfig("return neosync.transformEmail(value, { preserveLength: true, maxLength: 255});"),
		"str":                 getJavascriptTransformerConfig("return neosync.transformString(value, { preserveLength: true, maxLength: 30});"),
		"measurement":         getJavascriptTransformerConfig("return neosync.transformFloat64(value, { randomizationRangeMin: 3.14, randomizationRangeMax: 300.10});"),
		"int64":               getJavascriptTransformerConfig("return neosync.transformInt64(value, { randomizationRangeMin: 1, randomizationRangeMax: 300});"),
		"int64_phone_number":  getJavascriptTransformerConfig("return neosync.transformInt64PhoneNumber(value, { preserveLength: true});"),
		"string_phone_number": getJavascriptTransformerConfig("return neosync.transformStringPhoneNumber(value, { preserveLength: true, maxLength: 200});"),
		"first_name":          getJavascriptTransformerConfig("return neosync.transformFirstName(value, { preserveLength: true, maxLength: 25});"),
		"last_name":           getJavascriptTransformerConfig("return neosync.transformLastName(value, { preserveLength: true, maxLength: 25});"),
		"full_name":           getJavascriptTransformerConfig("return neosync.transformFullName(value, { preserveLength: true, maxLength: 25});"),
		"character_scramble":  getJavascriptTransformerConfig("return neosync.transformCharacterScramble(value, { preserveLength: false, maxLength: 100});"),
	}
	jobmappings := GetDefaultSyncJobMappings()
	updatedJobmappings := []*mgmtv1alpha1.JobMapping{}
	for _, jm := range jobmappings {
		updatedJobmappings = append(updatedJobmappings, &mgmtv1alpha1.JobMapping{
			Schema:      jm.Schema,
			Table:       jm.Table,
			Column:      jm.Column,
			Transformer: colTransformerMap[jm.Column],
		})
	}
	return updatedJobmappings
}

func getJavascriptTransformerConfig(code string) *mgmtv1alpha1.JobMappingTransformer {
	return &mgmtv1alpha1.JobMappingTransformer{
		Source: mgmtv1alpha1.TransformerSource_TRANSFORMER_SOURCE_TRANSFORM_JAVASCRIPT,
		Config: &mgmtv1alpha1.TransformerConfig{
			Config: &mgmtv1alpha1.TransformerConfig_TransformJavascriptConfig{
				TransformJavascriptConfig: &mgmtv1alpha1.TransformJavascript{Code: code},
			},
		},
	}
}
