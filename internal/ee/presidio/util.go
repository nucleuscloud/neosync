package presidioapi

func ToAnonymizeRecognizerResults(
	input []RecognizerResultWithAnaysisExplanation,
) []RecognizerResult {
	output := make([]RecognizerResult, 0, len(input))
	for _, rr := range input {
		output = append(output, ToAnonymizeRecognizerResult(rr))
	}
	return output
}

func ToAnonymizeRecognizerResult(input RecognizerResultWithAnaysisExplanation) RecognizerResult {
	return RecognizerResult{
		End:                 input.End,
		EntityType:          input.EntityType,
		RecognitionMetadata: input.RecognitionMetadata,
		Score:               input.Score,
		Start:               input.Start,
	}
}
