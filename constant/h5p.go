package constant

const (
	ActivityEventVerbIDInitGame   = "http://verb.kidsloop.cn/verbs/initgame"
	ActivityEventVerbIDInteracted = "http://adlnet.gov/expapi/verbs/interacted"
	ActivityEventVerbIDCompleted  = "http://adlnet.gov/expapi/verbs/completed"
	ActivityEventVerbIDAnswered   = "http://adlnet.gov/expapi/verbs/answered"

	H5PGJSONPathLibrary                         = "library"
	H5PGJSONPathSequenceImagesCardsNumber       = "params.sequenceImages.#"
	H5PGJSONPathSequenceImagesCorrectCardsCount = "additionanProp1.correct_amount"
	H5PGJSONPathMemoryGamePairsNumber           = "params.cards.#"
	H5PGJSONPathImagePairPairsNumber            = "params.cards.#"
	H5PGJSONPathImagePairCorrectPairsCount      = "additionanProp1.correct_amount"
	H5PGJSONPathFlashCardsCardsNumber           = "params.cards.#"
	H5PGJSONPathFlashCardsCorrectCardsCount     = "additionanProp1.correct_amount"

	H5PServiceDefaultEndpoint = "https://api.alpha.kidsloop.net/assessment/graphql"
)
