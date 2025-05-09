package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// ------------General Responses------------ //
func (app *application) redirectResponse(c echo.Context, path string, jsonStatus int, message any) error {
	var err error
	if strings.Contains(c.Request().Header.Get("Accept"), "application/json") {
		env := envelope{"message": message, "redirect": path}
		err = c.JSON(jsonStatus, env)
	} else {
		err = c.Redirect(http.StatusSeeOther, path)
	}
	if err != nil {
		app.logError(c, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	return err
}

// TODO: Incorporate config values from DB into responses, eg. so that service name reflects the one configured by the client
// ------------KAMAR Responses------------ //
func (app *application) kamarResponse(c echo.Context, status int, j map[string]any) error {
	c.Response().Header().Set(echo.HeaderServer, "WHS KAMAR Refresh")
	c.Response().Header().Set(echo.HeaderConnection, "close")

	env := envelope{"SMSDirectoryData": j}

	return c.JSON(status, env)

	// err := c.JSON(status, env)
	// if err != nil {
	// 	app.logError(c, err)
	// 	return c.NoContent(http.StatusInternalServerError)
	// }
	// return nil
}

// The responses below meet the requirements of KAMAR by adding expected headers and the expected JSON body - only these responses should ever be sent to KAMAR.
func (app *application) kamarSuccessResponse(c echo.Context) error {
	j := map[string]any{
		"error":   0,
		"result":  "OK",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	return app.kamarResponse(c, http.StatusOK, j)
}

// NOTE: The expected failed response here: https://directoryservices.kamar.nz/?listening-service/standard-response - includes a Content-Length: 123 header, whereas Content-Length is only 82 with this response.
func (app *application) kamarAuthFailedResponse(c echo.Context) error {
	j := map[string]any{
		"error":   403,
		"result":  "Authentication Failed",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	return app.kamarResponse(c, http.StatusForbidden, j)
}

func (app *application) kamarNoCredentialsResponse(c echo.Context) error {
	j := map[string]any{
		"error":   401,
		"result":  "No Credentials Provided",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	return app.kamarResponse(c, http.StatusUnauthorized, j)
}

// TODO: Receive specific error message indicating which aspect of the data was malformed (auth or body), and reflect in "result" message
func (app *application) kamarUnprocessableEntityResponse(c echo.Context) error {
	j := map[string]any{
		"error":   422,
		"result":  "Request From KAMAR Was Malformed",
		"service": "WHS KAMAR Refresh",
		"version": "1.0",
	}

	return app.kamarResponse(c, http.StatusUnprocessableEntity, j)
}

// TODO: Get check options etc. from DB
func (app *application) kamarCheckResponse(c echo.Context) error {
	j := map[string]any{
		"error":             0,
		"result":            "OK",
		"service":           "WHS KAMAR Refresh",
		"version":           "1.1",
		"status":            "Ready",
		"infourl":           "https://wakatipu.school.nz/",
		"privacystatement":  "This service only collects results data, and stores it locally on a secure device. Only staff members of the school have access to the data.",
		"countryDataStored": "New Zealand",
		"options": map[string]any{
			"ics": true,
			"students": map[string]any{
				"details":         true,
				"passwords":       true,
				"photos":          false,
				"groups":          true,
				"awards":          true,
				"timetables":      true,
				"attendance":      true,
				"assessments":     true,
				"pastoral":        true,
				"recognitions":    true,
				"classefforts":    true,
				"learningsupport": true,
				"fields": map[string]string{
					"required": "firstname;lastname;gender;gendercode;nsn;uniqueid",
					"optional": "schoolindex;firstnamelegal;lastnamelegal;forenames;forenameslegal;genderpreferred;username;mobile;email;house;whanau;boarder;byodinfo;ece;esol;ors;languagespoken;datebirth;startingdate;startschooldate;created;leavingdate;leavingreason;leavingschool;leavingactivity;res;resa;resb;res.title;res.salutation;res.email;res.numFlatUnit;res.numStreet;res.ruralDelivery;res.suburb;res.town;res.postcode;caregivers;caregivers1;caregivers2;caregivers3;caregivers4;caregiver.name;caregiver.relationship;caregiver.status;caregiver.address,caregiver.mobile;caregiver.email;emergency;emergency1;emergency2;emergency.name;emergency.relationship;emergency.mobile;moetype;ethnicityL1;ethnicityL2;ethnicity;iwi;yearlevel;fundinglevel;tutor;timetablebottom1;timetablebottom2;timetablebottom3;timetablebottom4;timetabletop1;timetabletop2;timetabletop3;timetabletop4;maorilevel;pacificlanguage;pacificlevel;flags;flag.alert;flag.conditions;flag.dietary;flag.general;flag.ibuprofen;flag.medical;flag.notes;flag.paracetamol;flag.pastoral;flag.reactions;flag.specialneeds;flag.vaccinations;flag.eotcconsent;flag.eotcform;custom;custom.custom1;custom.custom2;custom.custom3;custom.custom4;custom.custom5;siblinglink;photocopierid;signedagreement;accountdisabled;networkaccess;altdescription;althomedrive",
				},
			},
			"staff": map[string]any{
				"details":    true,
				"photos":     false,
				"timetables": true,
				"fields": map[string]string{
					"required": "uniqueid;firstname;lastname;username;gender;email",
					"optional": "schoolindex;title;mobile;extension;classification;position;house;tutor;groups;groups.departments;datebirth;created;leavingdate;startingdate;eslguid;moenumber;photocopierid;registrationnumber;custom;custom.custom1;custom.custom2;custom.custom3;custom.custom4;custom.custom5",
				},
			},
			"common": map[string]bool{
				"subjects": true,
				"notices":  false,
				"calendar": false,
				"bookings": false,
			},
		},
	}

	app.appMetrics.SetLastCheckTime()

	return app.kamarResponse(c, http.StatusOK, j)
}
