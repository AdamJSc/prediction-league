package datastore

import (
	"log"
	"prediction-league/service/internal/models"
	"time"
)

const FakeSeasonID = "FakeSeason"

// Teams provides a pre-determined data structure of all Teams that can be referenced within the system
var Teams = models.TeamCollection{
	"AFC": {
		ID:        "AFC",
		ClientID:  models.TeamIdentifier{TeamID: 57},
		Name:      "Arsenal",
		ShortName: "Arsenal",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/53/Arsenal_FC.svg",
	},
	"AFCB": {
		ID:        "AFCB",
		ClientID:  models.TeamIdentifier{TeamID: 1044},
		Name:      "Bournemouth",
		ShortName: "Bournemouth",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e5/AFC_Bournemouth_%282013%29.svg",
	},
	"AVFC": {
		ID:        "AVFC",
		ClientID:  models.TeamIdentifier{TeamID: 58},
		Name:      "Aston Villa",
		ShortName: "Villa",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/de/9/9f/Aston_Villa_logo.svg",
	},
	"BFC": {
		ID:        "BFC",
		ClientID:  models.TeamIdentifier{TeamID: 328},
		Name:      "Burnley",
		ShortName: "Burnley",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/6/62/Burnley_F.C._Logo.svg",
	},
	"BHAFC": {
		ID:        "BHAFC",
		ClientID:  models.TeamIdentifier{TeamID: 397},
		Name:      "Brighton & Hove Albion",
		ShortName: "Brighton",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/f/fd/Brighton_%26_Hove_Albion_logo.svg",
	},
	"CFC": {
		ID:        "CFC",
		ClientID:  models.TeamIdentifier{TeamID: 61},
		Name:      "Chelsea",
		ShortName: "Chelsea",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/cc/Chelsea_FC.svg",
	},
	"CPFC": {
		ID:        "CPFC",
		ClientID:  models.TeamIdentifier{TeamID: 354},
		Name:      "Crystal Palace",
		ShortName: "Palace",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Crystal_Palace_FC_logo.svg",
	},
	"EFC": {
		ID:        "EFC",
		ClientID:  models.TeamIdentifier{TeamID: 62},
		Name:      "Everton",
		ShortName: "Everton",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7c/Everton_FC_logo.svg",
	},
	"LFC": {
		ID:        "LFC",
		ClientID:  models.TeamIdentifier{TeamID: 64},
		Name:      "Liverpool",
		ShortName: "Liverpool",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Liverpool_FC.svg",
	},
	"LCFC": {
		ID:        "LCFC",
		ClientID:  models.TeamIdentifier{TeamID: 338},
		Name:      "Leicester City",
		ShortName: "Leicester",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/2/2d/Leicester_City_crest.svg",
	},
	"MCFC": {
		ID:        "MCFC",
		ClientID:  models.TeamIdentifier{TeamID: 65},
		Name:      "Manchester City",
		ShortName: "Man City",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/eb/Manchester_City_FC_badge.svg",
	},
	"MUFC": {
		ID:        "MUFC",
		ClientID:  models.TeamIdentifier{TeamID: 66},
		Name:      "Manchester United",
		ShortName: "Man Utd",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7a/Manchester_United_FC_crest.svg",
	},
	"NCFC": {
		ID:        "NCFC",
		ClientID:  models.TeamIdentifier{TeamID: 68},
		Name:      "Norwich City",
		ShortName: "Norwich",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/8/8c/Norwich_City.svg",
	},
	"NUFC": {
		ID:        "NUFC",
		ClientID:  models.TeamIdentifier{TeamID: 67},
		Name:      "Newcastle United",
		ShortName: "Newcastle",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/56/Newcastle_United_Logo.svg",
	},
	"SUFC": {
		ID:        "SUFC",
		ClientID:  models.TeamIdentifier{TeamID: 356},
		Name:      "Sheffield United",
		ShortName: "Sheff Utd",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/9/9c/Sheffield_United_FC_logo.svg",
	},
	"SFC": {
		ID:        "SFC",
		ClientID:  models.TeamIdentifier{TeamID: 340},
		Name:      "Southampton",
		ShortName: "Saints",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c9/FC_Southampton.svg",
	},
	"THFC": {
		ID:        "THFC",
		ClientID:  models.TeamIdentifier{TeamID: 73},
		Name:      "Tottenham Hotspur",
		ShortName: "Spurs",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/b/b4/Tottenham_Hotspur.svg",
	},
	"WFC": {
		ID:        "WFC",
		ClientID:  models.TeamIdentifier{TeamID: 346},
		Name:      "Watford",
		ShortName: "Watford",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e2/Watford.svg",
	},
	"WWFC": {
		ID:        "WWFC",
		ClientID:  models.TeamIdentifier{TeamID: 76},
		Name:      "Wolverhampton Wanderers",
		ShortName: "Wolves",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/f/fc/Wolverhampton_Wanderers.svg",
	},
	"WHUFC": {
		ID:        "WHUFC",
		ClientID:  models.TeamIdentifier{TeamID: 563},
		Name:      "West Ham United",
		ShortName: "West Ham",
		CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c2/West_Ham_United_FC_logo.svg",
	},
}

// Seasons provides a pre-determined data structure of all Seasons that can be referenced within the system
var Seasons models.SeasonCollection

// MustInflate inflates our data stores
func MustInflate() {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		log.Fatal(err)
	}

	// inflate seasons using location
	// we can't do this directly when defining the struct because we need to load the location at runtime
	Seasons = models.SeasonCollection{
		FakeSeasonID: {
			ID:       FakeSeasonID,
			ClientID: nil,	// will not trigger requests to client when running in retrieve latest standings job
			Name:     "Localhost Season",
			EntriesAccepted: models.TimeFrame{
				From:  time.Now(),
				Until: time.Now().Add(20 * time.Minute),
			},
			SelectionsAccepted: []models.TimeFrame{
				{
					From:  time.Now(),
					Until: time.Now().Add(20 * time.Minute),
				},
				{
					From:  time.Now().Add(40 * time.Minute),
					Until: time.Now().Add(60 * time.Minute),
				},
			},
			TeamIDs: []string{
				"AFC",
				"AFCB",
				"AVFC",
				"BFC",
				"BHAFC",
				"CFC",
				"CPFC",
				"EFC",
				"LFC",
				"LCFC",
				"MCFC",
				"MUFC",
				"NCFC",
				"NUFC",
				"SUFC",
				"SFC",
				"THFC",
				"WFC",
				"WWFC",
				"WHUFC",
			},
			MaxRounds: 38,
		},
		"201920_1": {
			ID:       "201920_1",
			ClientID: models.SeasonIdentifier{SeasonID: "PL"},
			Name:     "Premier League 2019/20",
			Active: models.TimeFrame{
				From:  time.Date(2019, 8, 9, 19, 0, 0, 0, loc),
				Until: time.Date(2020, 5, 17, 23, 59, 59, 0, loc),
			},
			EntriesAccepted: models.TimeFrame{
				From:  time.Date(2019, 7, 1, 0, 0, 0, 0, loc),
				Until: time.Date(2019, 8, 9, 19, 0, 0, 0, loc),
			},
			SelectionsAccepted: []models.TimeFrame{
				{
					From:  time.Date(2019, 7, 1, 0, 0, 0, 0, loc),
					Until: time.Date(2019, 8, 9, 19, 0, 0, 0, loc),
				},
			},
			TeamIDs: []string{
				"AFC",
				"AFCB",
				"AVFC",
				"BFC",
				"BHAFC",
				"CFC",
				"CPFC",
				"EFC",
				"LFC",
				"LCFC",
				"MCFC",
				"MUFC",
				"NCFC",
				"NUFC",
				"SUFC",
				"SFC",
				"THFC",
				"WFC",
				"WWFC",
				"WHUFC",
			},
			MaxRounds: 38,
		},
	}
}
