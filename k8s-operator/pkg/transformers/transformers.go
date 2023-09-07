package transformers

type Transformation string

const (
	UuidV4      Transformation = "uuid_v4"
	Latitude    Transformation = "latitude"
	Longitude   Transformation = "longitude"
	FirstName   Transformation = "first_name"
	PhoneNumber Transformation = "phone_number"
)
