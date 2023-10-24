package v1alpha1_transformersservice

import (
	"context"

	"connectrpc.com/connect"
	mgmtv1alpha1 "github.com/nucleuscloud/neosync/backend/gen/go/protos/mgmt/v1alpha1"
	"github.com/nucleuscloud/neosync/backend/internal/dtomaps"
)

type Transformation string

const (
	Invalid        Transformation = "invalid"
	Passthrough    Transformation = "passthrough"
	Uuid           Transformation = "uuid"
	FirstName      Transformation = "first_name"
	LastName       Transformation = "last_name"
	FullName       Transformation = "full_name"
	PhoneNumber    Transformation = "phone_number"
	IntPhoneNumber Transformation = "int_phone_number"
	Email          Transformation = "email"
	Null           Transformation = "null"
	RandomString   Transformation = "random_string"
	RandomBool     Transformation = "random_bool"
	RandomInt      Transformation = "random_int"
	RandomFloat    Transformation = "random_float"
	Gender         Transformation = "gender"
	UTCTimestamp   Transformation = "utc_timestamp"
	UnixTimestamp  Transformation = "unix_timestamp"
	StreetAddress  Transformation = "street_address"
	City           Transformation = "city"
	Zipcode        Transformation = "zipcode"
	State          Transformation = "state"
	FullAddress    Transformation = "full_address"
)

func (s *Service) GetSystemTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetSystemTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetSystemTransformersResponse], error) {
	return connect.NewResponse(&mgmtv1alpha1.GetSystemTransformersResponse{
		Transformers: []*mgmtv1alpha1.Transformer{
			{Value: string(Passthrough), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(Uuid), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_UuidConfig{
					UuidConfig: &mgmtv1alpha1.Uuid{
						IncludeHyphen: true,
					},
				},
			},
			},
			{Value: string(FirstName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FirstNameConfig{
					FirstNameConfig: &mgmtv1alpha1.FirstName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(LastName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_LastNameConfig{
					LastNameConfig: &mgmtv1alpha1.LastName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(FullName), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_FullNameConfig{
					FullNameConfig: &mgmtv1alpha1.FullName{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(PhoneNumber), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_PhoneNumberConfig{
					PhoneNumberConfig: &mgmtv1alpha1.PhoneNumber{
						PreserveLength: false,
						E164Format:     false,
						IncludeHyphens: false,
					},
				},
			}},
			{Value: string(IntPhoneNumber), Config: &mgmtv1alpha1.TransformerConfig{
				Config: &mgmtv1alpha1.TransformerConfig_IntPhoneNumberConfig{
					IntPhoneNumberConfig: &mgmtv1alpha1.IntPhoneNumber{
						PreserveLength: false,
					},
				},
			}},
			{Value: string(Null), Config: &mgmtv1alpha1.TransformerConfig{}},
			{
				Value: string(Email),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_EmailConfig{
						EmailConfig: &mgmtv1alpha1.EmailConfig{
							PreserveDomain: false,
							PreserveLength: false,
						},
					},
				}},
			{
				Value: string(RandomString),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_RandomStringConfig{
						RandomStringConfig: &mgmtv1alpha1.RandomString{
							PreserveLength: false,
							StrLength:      0,
							StrCase:        mgmtv1alpha1.RandomString_STRING_CASE_LOWER,
						},
					},
				}},
			{Value: string(RandomBool), Config: &mgmtv1alpha1.TransformerConfig{}},
			{
				Value: string(RandomInt),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_RandomIntConfig{
						RandomIntConfig: &mgmtv1alpha1.RandomInt{
							PreserveLength: false,
							IntLength:      0,
						},
					},
				}},
			{
				Value: string(RandomFloat),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_RandomFloatConfig{
						RandomFloatConfig: &mgmtv1alpha1.RandomFloat{
							PreserveLength:      false,
							DigitsBeforeDecimal: 3,
							DigitsAfterDecimal:  3,
						},
					},
				}},
			{
				Value: string(Gender),
				Config: &mgmtv1alpha1.TransformerConfig{
					Config: &mgmtv1alpha1.TransformerConfig_GenderConfig{
						GenderConfig: &mgmtv1alpha1.Gender{
							Abbreviate: false,
						},
					},
				}},
			{Value: string(UTCTimestamp), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(UnixTimestamp), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(StreetAddress), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(City), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(Zipcode), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(State), Config: &mgmtv1alpha1.TransformerConfig{}},
			{Value: string(FullAddress), Config: &mgmtv1alpha1.TransformerConfig{}},
		},
	}), nil
}

func (s *Service) GetCustomTransformers(
	ctx context.Context,
	req *connect.Request[mgmtv1alpha1.GetCustomTransformersRequest],
) (*connect.Response[mgmtv1alpha1.GetCustomTransformersResponse], error) {

	accountUuid, err := s.verifyUserInAccount(ctx, req.Msg.AccountId)
	if err != nil {
		return nil, err
	}

	transformers, err := s.db.Q.GetCustomTransformersByAccount(ctx, *accountUuid)
	if err != nil {
		return nil, err
	}

	dtoTransformers := []*mgmtv1alpha1.CustomTransformer{}
	for idx := range transformers {
		transformer := transformers[idx]
		dtoTransformers = append(dtoTransformers, dtomaps.ToCustomTransformerDto(&transformer))
	}

	return connect.NewResponse(&mgmtv1alpha1.GetCustomTransformersResponse{
		Transformers: dtoTransformers,
	}), nil
}
