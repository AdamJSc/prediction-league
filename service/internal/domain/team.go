package domain

import (
	"fmt"
	"prediction-league/service/internal/models"
)

// Teams returns a pre-determined data structure of all Teams that can be referenced within the system
func Teams() models.TeamCollection {
	return map[string]models.Team{
		"AFC": {
			ID:        "AFC",
			ClientID:  57,
			Name:      "Arsenal",
			ShortName: "Arsenal",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/53/Arsenal_FC.svg",
		},
		"AFCB": {
			ID:        "AFCB",
			ClientID:  1044,
			Name:      "Bournemouth",
			ShortName: "Cherries",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e5/AFC_Bournemouth_%282013%29.svg",
		},
		"AVFC": {
			ID:        "AVFC",
			ClientID:  58,
			Name:      "Aston Villa",
			ShortName: "Villa",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/de/9/9f/Aston_Villa_logo.svg",
		},
		"BHAFC": {
			ID:        "BHAFC",
			ClientID:  397,
			Name:      "Brighton & Hove Albion",
			ShortName: "Brighton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/f/fd/Brighton_%26_Hove_Albion_logo.svg",
		},
		"CFC": {
			ID:        "CFC",
			ClientID:  61,
			Name:      "Chelsea",
			ShortName: "Chelsea",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/cc/Chelsea_FC.svg",
		},
		"CPFC": {
			ID:        "CPFC",
			ClientID:  354,
			Name:      "Crystal Palace",
			ShortName: "Palace",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Crystal_Palace_FC_logo.svg",
		},
		"EFC": {
			ID:        "EFC",
			ClientID:  62,
			Name:      "Everton",
			ShortName: "Everton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7c/Everton_FC_logo.svg",
		},
		"LFC": {
			ID:        "LFC",
			ClientID:  64,
			Name:      "Liverpool",
			ShortName: "Liverpool",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Liverpool_FC.svg",
		},
		"LCFC": {
			ID:        "LCFC",
			ClientID:  338,
			Name:      "Leicester City",
			ShortName: "Leicester",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/2/2d/Leicester_City_crest.svg",
		},
		"MCFC": {
			ID:        "MCFC",
			ClientID:  65,
			Name:      "Manchester City",
			ShortName: "Man City",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/eb/Manchester_City_FC_badge.svg",
		},
		"MUFC": {
			ID:        "MUFC",
			ClientID:  66,
			Name:      "Manchester United",
			ShortName: "Man Utd",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7a/Manchester_United_FC_crest.svg",
		},
		"NUFC": {
			ID:        "NUFC",
			ClientID:  67,
			Name:      "Newcastle United",
			ShortName: "Newcastle",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/56/Newcastle_United_Logo.svg",
		},
		"SUFC": {
			ID:        "SUFC",
			ClientID:  356,
			Name:      "Sheffield United",
			ShortName: "Sheffield",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/9/9c/Sheffield_United_FC_logo.svg",
		},
		"SFC": {
			ID:        "SFC",
			ClientID:  340,
			Name:      "Southampton",
			ShortName: "Saints",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c9/FC_Southampton.svg",
		},
		"WFC": {
			ID:        "WFC",
			ClientID:  346,
			Name:      "Watford",
			ShortName: "Watford",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e2/Watford.svg",
		},
		"WHUFC": {
			ID:        "WHUFC",
			ClientID:  563,
			Name:      "West Ham United",
			ShortName: "West Ham",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c2/West_Ham_United_FC_logo.svg",
		},
	}
}

// ValidateTeam returns an error if validation rules are not satisfied for the provided Team
func ValidateTeam(t models.Team) error {
	var validationMsgs []string

	// validate strings
	for k, v := range map[string]string{
		"ID":        t.ID,
		"Name":      t.Name,
		"ShortName": t.ShortName,
		"CrestURL":  t.CrestURL,
	} {
		if v == "" {
			validationMsgs = append(validationMsgs, fmt.Sprintf("%s must not be empty", k))
		}
	}

	// validate ints
	for k, v := range map[string]int{
		"ClientID": t.ClientID,
	} {
		if v == 0 {
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
