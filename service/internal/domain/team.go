package domain

import (
	"fmt"
	footballdata "prediction-league/service/internal/clients/football-data-org"
	"prediction-league/service/internal/models"
)

// Teams returns a pre-determined data structure of all Teams that can be referenced within the system
func Teams() models.TeamCollection {
	return map[string]models.Team{
		"AFC": {
			ID:        "AFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 57},
			Name:      "Arsenal",
			ShortName: "Arsenal",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/53/Arsenal_FC.svg",
		},
		"AFCB": {
			ID:        "AFCB",
			ClientID:  footballdata.TeamIdentifier{TeamID: 1044},
			Name:      "Bournemouth",
			ShortName: "Cherries",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e5/AFC_Bournemouth_%282013%29.svg",
		},
		"AVFC": {
			ID:        "AVFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 58},
			Name:      "Aston Villa",
			ShortName: "Villa",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/de/9/9f/Aston_Villa_logo.svg",
		},
		"BFC": {
			ID:        "BFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 328},
			Name:      "Burnley",
			ShortName: "Burnley",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/6/62/Burnley_F.C._Logo.svg",
		},
		"BHAFC": {
			ID:        "BHAFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 397},
			Name:      "Brighton & Hove Albion",
			ShortName: "Brighton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/f/fd/Brighton_%26_Hove_Albion_logo.svg",
		},
		"CFC": {
			ID:        "CFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 61},
			Name:      "Chelsea",
			ShortName: "Chelsea",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/cc/Chelsea_FC.svg",
		},
		"CPFC": {
			ID:        "CPFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 354},
			Name:      "Crystal Palace",
			ShortName: "Palace",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Crystal_Palace_FC_logo.svg",
		},
		"EFC": {
			ID:        "EFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 62},
			Name:      "Everton",
			ShortName: "Everton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7c/Everton_FC_logo.svg",
		},
		"LFC": {
			ID:        "LFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 64},
			Name:      "Liverpool",
			ShortName: "Liverpool",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Liverpool_FC.svg",
		},
		"LCFC": {
			ID:        "LCFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 338},
			Name:      "Leicester City",
			ShortName: "Leicester",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/2/2d/Leicester_City_crest.svg",
		},
		"MCFC": {
			ID:        "MCFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 65},
			Name:      "Manchester City",
			ShortName: "Man City",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/eb/Manchester_City_FC_badge.svg",
		},
		"MUFC": {
			ID:        "MUFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 66},
			Name:      "Manchester United",
			ShortName: "Man Utd",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7a/Manchester_United_FC_crest.svg",
		},
		"NCFC": {
			ID:        "NCFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 68},
			Name:      "Norwich City",
			ShortName: "Norwich",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/8/8c/Norwich_City.svg",
		},
		"NUFC": {
			ID:        "NUFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 67},
			Name:      "Newcastle United",
			ShortName: "Newcastle",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/56/Newcastle_United_Logo.svg",
		},
		"SUFC": {
			ID:        "SUFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 356},
			Name:      "Sheffield United",
			ShortName: "Sheffield",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/9/9c/Sheffield_United_FC_logo.svg",
		},
		"SFC": {
			ID:        "SFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 340},
			Name:      "Southampton",
			ShortName: "Saints",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c9/FC_Southampton.svg",
		},
		"THFC": {
			ID:        "THFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 73},
			Name:      "Tottenham Hotspur",
			ShortName: "Spurs",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/b/b4/Tottenham_Hotspur.svg",
		},
		"WFC": {
			ID:        "WFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 346},
			Name:      "Watford",
			ShortName: "Watford",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e2/Watford.svg",
		},
		"WWFC": {
			ID:        "WWFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 76},
			Name:      "Wolverhampton Wanderers",
			ShortName: "Wolves",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/f/fc/Wolverhampton_Wanderers.svg",
		},
		"WHUFC": {
			ID:        "WHUFC",
			ClientID:  footballdata.TeamIdentifier{TeamID: 563},
			Name:      "West Ham United",
			ShortName: "West Ham",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c2/West_Ham_United_FC_logo.svg",
		},
	}
}

// ValidateTeam returns an error if validation rules are not satisfied for the provided Team
func ValidateTeam(t models.Team) error {
	var validationMsgs []string

	// validate values
	for k, v := range map[string]struct {
		actual  string
		invalid string
	}{
		"ID":        {actual: t.ID, invalid: ""},
		"Name":      {actual: t.Name, invalid: ""},
		"ShortName": {actual: t.ShortName, invalid: ""},
		"CrestURL":  {actual: t.CrestURL, invalid: ""},
		"ClientID":  {actual: t.ClientID.Value(), invalid: "0"},
	} {
		if v.actual == v.invalid {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}

	if len(validationMsgs) > 0 {
		return ValidationError{
			Reasons: validationMsgs,
		}
	}

	return nil
}
