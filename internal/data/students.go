package data

import (
	"database/sql"
	"encoding/json"
)

// Without clear examples or documentation from KAMAR, a number of fields' data types are unclear. These data types are Unmarshalled as "any" type to the struct, and then written as TEXT values to the database, to ensure they still come through
type Student struct {
	ID               int             `json:"id,omitempty"`
	UUID             string          `json:"uuid,omitempty"`
	Created          int64           `json:"created,omitempty"`
	Role             string          `json:"role,omitempty"`
	Uniqueid         int             `json:"uniqueid,omitempty"`
	SchoolIndex      int             `json:"schoolindex,omitempty"`
	Nsn              string          `json:"nsn,omitempty"`
	Username         string          `json:"username,omitempty"`
	Firstname        string          `json:"firstname,omitempty"`
	FirstnameLegal   string          `json:"firstnamelegal,omitempty"`
	Lastname         string          `json:"lastname,omitempty"`
	LastnameLegal    string          `json:"lastnamelegal,omitempty"`
	Forenames        string          `json:"forenames,omitempty"`
	ForenamesLegal   string          `json:"forenameslegal,omitempty"`
	Gender           any             `json:"gender,omitempty"`
	GenderPreferred  any             `json:"genderpreffered,omitempty"`
	Gendercode       int             `json:"gendercode,omitempty"`
	Mobile           string          `json:"mobile,omitempty"`
	Email            string          `json:"email,omitempty"`
	House            string          `json:"house,omitempty"`
	Whanau           string          `json:"whanau,omitempty"`
	Boarder          any             `json:"boarder,omitempty"`
	BYODInfo         any             `json:"byodinfo,omitempty"`
	ECE              any             `json:"ece,omitempty"`
	ESOL             any             `json:"esol,omitempty"`
	ORS              any             `json:"ors,omitempty"`
	LanguageSpoken   string          `json:"languagespoken,omitempty"`
	Datebirth        int             `json:"datebirth,omitempty"`
	Startingdate     int             `json:"startingdate,omitempty"`
	StartSchoolDate  any             `json:"startschooldate,omitempty"`
	Leavingdate      int             `json:"leavingdate,omitempty"`
	LeavingReason    string          `json:"leavingreason,omitempty"`
	LeavingSchool    string          `json:"leavingschool,omitempty"`
	LeavingActivity  any             `json:"leavingactivity,omitempty"`
	MOEType          string          `json:"moetype,omitempty"`
	EthnicityL1      any             `json:"ethnicityL1,omitempty"`
	EthnicityL2      any             `json:"ethnicityL2,omitempty"`
	Ethnicity        any             `json:"ethnicity,omitempty"`
	Iwi              string          `json:"iwi,omitempty"`
	YearLevel        any             `json:"yearlevel,omitempty"`
	FundingLevel     any             `json:"fundinglevel,omitempty"`
	Tutor            string          `json:"tutor,omitempty"`
	TimetableBottom1 any             `json:"timetablebottom1,omitempty"`
	TimetableBottom2 any             `json:"timetablebottom2,omitempty"`
	TimetableBottom3 any             `json:"timetablebottom3,omitempty"`
	TimetableBottom4 any             `json:"timetablebottom4,omitempty"`
	TimetableTop1    any             `json:"timetabletop1,omitempty"`
	TimetableTop2    any             `json:"timetabletop2,omitempty"`
	TimetableTop3    any             `json:"timetabletop3,omitempty"`
	TimetableTop4    any             `json:"timetabletop4,omitempty"`
	MaoriLevel       any             `json:"maorilevel,omitempty"`
	PacificLanguage  string          `json:"pacificlanguage,omitempty"`
	PacificLevel     any             `json:"pacificlevel,omitempty"`
	SiblingLink      any             `json:"siblinglink,omitempty"`
	PhotocopierID    any             `json:"photocopierid,omitempty"`
	SignedAgreement  any             `json:"signedagreement,omitempty"`
	AccountDisabled  any             `json:"accountdisabled,omitempty"`
	NetworkAccess    any             `json:"networkaccess,omitempty"`
	AltDescription   string          `json:"altdescription,omitempty"`
	AltHomeDrive     string          `json:"althomedrive,omitempty"`
	Flags            Flags           `json:"flags,omitempty"`
	Res              []Residence     `json:"res,omitempty"`
	Caregivers       []Caregiver     `json:"caregivers,omitempty"`
	Emergency        []Emergency     `json:"emergency,omitempty"`
	Groups           []Group         `json:"groups,omitempty"`
	Awards           []Award         `json:"awards,omitempty"`
	Datasharing      Datasharing     `json:"datasharing,omitempty"`
	Custom           json.RawMessage `json:"custom,omitempty"`
}

type Award struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Year int    `json:"year,omitempty"`
	Date string `json:"date,omitempty"`
}

type Caregiver struct {
	Ref          int    `json:"ref,omitempty"`
	Role         string `json:"role,omitempty"`
	Name         string `json:"name,omitempty"`
	Email        string `json:"email,omitempty"`
	Mobile       string `json:"mobile,omitempty"`
	Relationship string `json:"relationship,omitempty"`
	Status       any    `json:"status,omitempty"`
}

type Datasharing struct {
	Details int `json:"details,omitempty"`
	Photo   int `json:"photo,omitempty"`
	Other   int `json:"other,omitempty"`
}

type Emergency struct {
	Name         string `json:"name,omitempty"`
	Relationship string `json:"relationship,omitempty"`
	Mobile       string `json:"mobile,omitempty"`
}

type Flags struct {
	General string `json:"general,omitempty"`
	// TODO: Notes is a string, but its value (if it exists) is either 0 or 1 - better translated as a bool
	Notes        string `json:"notes,omitempty"`
	Alert        any    `json:"alert,omitempty"`
	Conditions   any    `json:"conditions,omitempty"`
	Dietary      any    `json:"dietary,omitempty"`
	Ibuprofen    any    `json:"ibuprofen,omitempty"`
	Medical      any    `json:"medical,omitempty"`
	Paracetamol  any    `json:"paracetamol,omitempty"`
	Pastoral     any    `json:"pastoral,omitempty"`
	Reactions    any    `json:"reactions,omitempty"`
	SpecialNeeds any    `json:"specialneeds,omitempty"`
	Vaccinations any    `json:"vaccinations,omitempty"`
	EOTCConsent  any    `json:"eotcconsent,omitempty"`
	EOTCForm     any    `json:"eotcform,omitempty"`
}

type Group struct {
	Type       string `json:"type,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Coreoption string `json:"coreoption,omitempty"`
}

type Residence struct {
	Title         string `json:"title,omitempty"`
	Salutation    string `json:"salutation,omitempty"`
	Email         string `json:"email,omitempty"`
	NumFlatUnit   any    `json:"numFlatUnit,omitempty"`
	NumStreet     any    `json:"numStreet,omitempty"`
	RuralDelivery any    `json:"ruralDelivery,omitempty"`
	Suburb        string `json:"suburb,omitempty"`
	Town          string `json:"town,omitempty"`
	Postcode      any    `json:"postcode,omitempty"`
}

type StudentModel struct {
	DB *sql.DB
}
