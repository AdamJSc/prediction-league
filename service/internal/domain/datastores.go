package domain

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	// FakeSeasonID defines the ID of the fake Season for running on localhost
	FakeSeasonID = "FakeSeason"

	// fakeSeasonKey defines the key of the real-world Season to replicate as fake season
	fakeSeasonKey = "202223_1"
)

// GetTeamCollection returns the required TeamCollection
func GetTeamCollection() TeamCollection {
	return TeamCollection{
		"AFC": {
			ID:        "AFC",
			ClientID:  TeamIdentifier{TeamID: 57},
			Name:      "Arsenal",
			ShortName: "Arsenal",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/53/Arsenal_FC.svg",
		},
		"AFCB": {
			ID:        "AFCB",
			ClientID:  TeamIdentifier{TeamID: 1044},
			Name:      "AFC Bournemouth",
			ShortName: "Bournemouth",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e5/AFC_Bournemouth_%282013%29.svg",
		},
		"AVFC": {
			ID:        "AVFC",
			ClientID:  TeamIdentifier{TeamID: 58},
			Name:      "Aston Villa",
			ShortName: "Villa",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/de/9/9f/Aston_Villa_logo.svg",
		},
		"BFC": {
			ID:        "BFC",
			ClientID:  TeamIdentifier{TeamID: 328},
			Name:      "Burnley",
			ShortName: "Burnley",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/6/62/Burnley_F.C._Logo.svg",
		},
		"BFC2": {
			ID:        "BFC2",
			ClientID:  TeamIdentifier{TeamID: 357},
			Name:      "Barnsley",
			ShortName: "Barnsley",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/c/c9/Barnsley_FC.svg/237px-Barnsley_FC.svg.png",
		},
		"BFC3": {
			ID:        "BFC3",
			ClientID:  TeamIdentifier{TeamID: 402},
			Name:      "Brentford",
			ShortName: "Brentford",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/2/2a/Brentford_FC_crest.svg/240px-Brentford_FC_crest.svg.png",
		},
		"BCFC": {
			ID:        "BCFC",
			ClientID:  TeamIdentifier{TeamID: 332},
			Name:      "Birmingham City",
			ShortName: "Birmingham",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/6/68/Birmingham_City_FC_logo.svg/249px-Birmingham_City_FC_logo.svg.png",
		},
		"BCFC2": {
			ID:        "BCFC2",
			ClientID:  TeamIdentifier{TeamID: 387},
			Name:      "Bristol City",
			ShortName: "Bristol C",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/f/f5/Bristol_City_crest.svg/240px-Bristol_City_crest.svg.png",
		},
		"BHAFC": {
			ID:        "BHAFC",
			ClientID:  TeamIdentifier{TeamID: 397},
			Name:      "Brighton & Hove Albion",
			ShortName: "Brighton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/f/fd/Brighton_%26_Hove_Albion_logo.svg",
		},
		"BRFC": {
			ID:        "BRFC",
			ClientID:  TeamIdentifier{TeamID: 59},
			Name:      "Blackburn Rovers",
			ShortName: "Blackburn",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/0/0f/Blackburn_Rovers.svg/232px-Blackburn_Rovers.svg.png",
		},
		"CAFC": {
			ID:        "CAFC",
			ClientID:  TeamIdentifier{TeamID: 348},
			Name:      "Charlton Athletic",
			ShortName: "Charlton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/commons/thumb/6/6a/CharltonBadge_30Jan2020.png/240px-CharltonBadge_30Jan2020.png",
		},
		"CCFC": {
			ID:        "CCFC",
			ClientID:  TeamIdentifier{TeamID: 715},
			Name:      "Cardiff City",
			ShortName: "Cardiff",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/3/3c/Cardiff_City_crest.svg/230px-Cardiff_City_crest.svg.png",
		},
		"CCFC2": {
			ID:        "CCFC2",
			ClientID:  TeamIdentifier{TeamID: 1076},
			Name:      "Coventry City",
			ShortName: "Coventry",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/9/94/Coventry_City_FC_logo.svg/278px-Coventry_City_FC_logo.svg.png",
		},
		"CFC": {
			ID:        "CFC",
			ClientID:  TeamIdentifier{TeamID: 61},
			Name:      "Chelsea",
			ShortName: "Chelsea",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/cc/Chelsea_FC.svg",
		},
		"CPFC": {
			ID:        "CPFC",
			ClientID:  TeamIdentifier{TeamID: 354},
			Name:      "Crystal Palace",
			ShortName: "Palace",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Crystal_Palace_FC_logo.svg",
		},
		"DCFC": {
			ID:        "DCFC",
			ClientID:  TeamIdentifier{TeamID: 342},
			Name:      "Derby County",
			ShortName: "Derby",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/4/4a/Derby_County_crest.svg/320px-Derby_County_crest.svg.png",
		},
		"EFC": {
			ID:        "EFC",
			ClientID:  TeamIdentifier{TeamID: 62},
			Name:      "Everton",
			ShortName: "Everton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7c/Everton_FC_logo.svg",
		},
		"FFC": {
			ID:        "FFC",
			ClientID:  TeamIdentifier{TeamID: 63},
			Name:      "Fulham",
			ShortName: "Fulham",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/e/eb/Fulham_FC_%28shield%29.svg/180px-Fulham_FC_%28shield%29.svg.png",
		},
		"HCFC": {
			ID:        "HCFC",
			ClientID:  TeamIdentifier{TeamID: 322},
			Name:      "Hull City",
			ShortName: "Hull",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/5/54/Hull_City_A.F.C._logo.svg/189px-Hull_City_A.F.C._logo.svg.png",
		},
		"HTAFC": {
			ID:        "HTAFC",
			ClientID:  TeamIdentifier{TeamID: 394},
			Name:      "Huddersfield Town",
			ShortName: "Huddersfield",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7d/Huddersfield_Town_A.F.C._logo.png",
		},
		"LFC": {
			ID:        "LFC",
			ClientID:  TeamIdentifier{TeamID: 64},
			Name:      "Liverpool",
			ShortName: "Liverpool",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/0/0c/Liverpool_FC.svg",
		},
		"LCFC": {
			ID:        "LCFC",
			ClientID:  TeamIdentifier{TeamID: 338},
			Name:      "Leicester City",
			ShortName: "Leicester",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/2/2d/Leicester_City_crest.svg",
		},
		"LTFC": {
			ID:        "LTFC",
			ClientID:  TeamIdentifier{TeamID: 389},
			Name:      "Luton Town",
			ShortName: "Luton",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/8/8b/LutonTownFC2009.png",
		},
		"LUFC": {
			ID:        "LUFC",
			ClientID:  TeamIdentifier{TeamID: 341},
			Name:      "Leeds United",
			ShortName: "Leeds",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/5/54/Leeds_United_F.C._logo.svg/196px-Leeds_United_F.C._logo.svg.png",
		},
		"MFC": {
			ID:        "MFC",
			ClientID:  TeamIdentifier{TeamID: 343},
			Name:      "Middlesbrough",
			ShortName: "Middlesbrough",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/2/2c/Middlesbrough_FC_crest.svg/230px-Middlesbrough_FC_crest.svg.png",
		},
		"MFC2": {
			ID:        "MFC2",
			ClientID:  TeamIdentifier{TeamID: 384},
			Name:      "Millwall",
			ShortName: "Millwall",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/c/c9/Millwall_F.C._logo.svg/240px-Millwall_F.C._logo.svg.png",
		},
		"MCFC": {
			ID:        "MCFC",
			ClientID:  TeamIdentifier{TeamID: 65},
			Name:      "Manchester City",
			ShortName: "Man City",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/eb/Manchester_City_FC_badge.svg",
		},
		"MUFC": {
			ID:        "MUFC",
			ClientID:  TeamIdentifier{TeamID: 66},
			Name:      "Manchester United",
			ShortName: "Man Utd",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/7/7a/Manchester_United_FC_crest.svg",
		},
		"NCFC": {
			ID:        "NCFC",
			ClientID:  TeamIdentifier{TeamID: 68},
			Name:      "Norwich City",
			ShortName: "Norwich",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/8/8c/Norwich_City.svg",
		},
		"NFFC": {
			ID:        "NFFC",
			ClientID:  TeamIdentifier{TeamID: 351},
			Name:      "Nottingham Forest",
			ShortName: "Nottm Forest",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/e/e5/Nottingham_Forest_F.C._logo.svg/225px-Nottingham_Forest_F.C._logo.svg.png",
		},
		"NUFC": {
			ID:        "NUFC",
			ClientID:  TeamIdentifier{TeamID: 67},
			Name:      "Newcastle United",
			ShortName: "Newcastle",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/5/56/Newcastle_United_Logo.svg",
		},
		"PNEFC": {
			ID:        "PNEFC",
			ClientID:  TeamIdentifier{TeamID: 1081},
			Name:      "Preston North End",
			ShortName: "Preston",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/8/82/Preston_North_End_FC.svg/261px-Preston_North_End_FC.svg.png",
		},
		"QPRFC": {
			ID:        "QPRFC",
			ClientID:  TeamIdentifier{TeamID: 69},
			Name:      "Queens Park Rangers",
			ShortName: "QPR",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/3/31/Queens_Park_Rangers_crest.svg/240px-Queens_Park_Rangers_crest.svg.png",
		},
		"RFC": {
			ID:        "RFC",
			ClientID:  TeamIdentifier{TeamID: 355},
			Name:      "Reading",
			ShortName: "Reading",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/1/11/Reading_FC.svg/200px-Reading_FC.svg.png",
		},
		"RUFC": {
			ID:        "RUFC",
			ClientID:  TeamIdentifier{TeamID: 385},
			Name:      "Rotherham United",
			ShortName: "Rotherham",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/c/c0/Rotherham_United_FC.svg/250px-Rotherham_United_FC.svg.png",
		},
		"SCAFC": {
			ID:        "SCAFC",
			ClientID:  TeamIdentifier{TeamID: 72},
			Name:      "Swansea City",
			ShortName: "Swansea",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/f/f9/Swansea_City_AFC_logo.svg/220px-Swansea_City_AFC_logo.svg.png",
		},
		"SCFC": {
			ID:        "SCFC",
			ClientID:  TeamIdentifier{TeamID: 70},
			Name:      "Stoke City",
			ShortName: "Stoke",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/2/29/Stoke_City_FC.svg/207px-Stoke_City_FC.svg.png",
		},
		"SFC": {
			ID:        "SFC",
			ClientID:  TeamIdentifier{TeamID: 340},
			Name:      "Southampton",
			ShortName: "Saints",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c9/FC_Southampton.svg",
		},
		"SUFC": {
			ID:        "SUFC",
			ClientID:  TeamIdentifier{TeamID: 356},
			Name:      "Sheffield United",
			ShortName: "Sheff Utd",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/9/9c/Sheffield_United_FC_logo.svg",
		},
		"SWFC": {
			ID:        "SWFC",
			ClientID:  TeamIdentifier{TeamID: 345},
			Name:      "Sheffield Wednesday",
			ShortName: "Sheff Wed",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/8/88/Sheffield_Wednesday_badge.svg/263px-Sheffield_Wednesday_badge.svg.png",
		},
		"THFC": {
			ID:        "THFC",
			ClientID:  TeamIdentifier{TeamID: 73},
			Name:      "Tottenham Hotspur",
			ShortName: "Spurs",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/b/b4/Tottenham_Hotspur.svg",
		},
		"WAFC": {
			ID:        "WAFC",
			ClientID:  TeamIdentifier{TeamID: 75},
			Name:      "Wigan Athletic",
			ShortName: "Wigan",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/4/43/Wigan_Athletic.svg/240px-Wigan_Athletic.svg.png",
		},
		"WBAFC": {
			ID:        "WBAFC",
			ClientID:  TeamIdentifier{TeamID: 74},
			Name:      "West Bromwich Albion",
			ShortName: "West Brom",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/8/8b/West_Bromwich_Albion.svg/200px-West_Bromwich_Albion.svg.png",
		},
		"WFC": {
			ID:        "WFC",
			ClientID:  TeamIdentifier{TeamID: 346},
			Name:      "Watford",
			ShortName: "Watford",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/e/e2/Watford.svg",
		},
		"WHUFC": {
			ID:        "WHUFC",
			ClientID:  TeamIdentifier{TeamID: 563},
			Name:      "West Ham United",
			ShortName: "West Ham",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/c/c2/West_Ham_United_FC_logo.svg",
		},
		"WWFC": {
			ID:        "WWFC",
			ClientID:  TeamIdentifier{TeamID: 76},
			Name:      "Wolverhampton Wanderers",
			ShortName: "Wolves",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/f/fc/Wolverhampton_Wanderers.svg",
		},
		"WWFC2": {
			ID:        "WWFC2",
			ClientID:  TeamIdentifier{TeamID: 1146},
			Name:      "Wycombe Wanderers",
			ShortName: "Wycombe",
			CrestURL:  "https://upload.wikimedia.org/wikipedia/en/thumb/f/fb/Wycombe_Wanderers_FC_logo.svg/240px-Wycombe_Wanderers_FC_logo.svg.png",
		},
	}
}

// GetSeasonCollection returns the required SeasonCollection
func GetSeasonCollection() (SeasonCollection, error) {
	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return nil, fmt.Errorf("cannot load time location Europe/London: %w", err)
	}

	// define seasons using location
	sc := SeasonCollection{

		// Premier League 2022/23
		"202223_1": {
			ID:        "202223_1",
			ClientID:  SeasonIdentifier{SeasonID: "PL"},
			Name:      "Premier League 2022/23",
			ShortName: "Prem 22/23",
			Live: TimeFrame{
				From:  time.Date(2022, 8, 5, 20, 0, 0, 0, loc),    // opening day kick-off
				Until: time.Date(2023, 5, 28, 23, 59, 59, 0, loc), // end of final day (regular season)
			},
			EntriesAccepted: TimeFrame{
				From:  time.Date(2022, 7, 22, 12, 0, 0, 0, loc), // competition opens
				Until: time.Date(2023, 5, 28, 16, 0, 0, 0, loc), // final day kick-off
			},
			PredictionsAccepted: TimeFrame{
				From:  time.Date(2022, 7, 22, 12, 0, 0, 0, loc),   // competition opens
				Until: time.Date(2023, 5, 28, 23, 59, 59, 0, loc), // end of final day (regular season)
			},
			TeamIDs: []string{
				"AFC",   // arsenal
				"AVFC",  // aston villa
				"AFCB",  // bournemouth
				"BFC3",  // brentford
				"BHAFC", // brighton
				"CFC",   // chelsea
				"CPFC",  // crystal palace
				"EFC",   // everton
				"FFC",   // fulham
				"LUFC",  // leeds united
				"LCFC",  // leicester city
				"LFC",   // liverpool
				"MCFC",  // man city
				"MUFC",  // man united
				"NFFC",  // nottm forest
				"NUFC",  // newcastle
				"SFC",   // southampton
				"THFC",  // tottenham hotspur
				"WHUFC", // west ham
				"WWFC",  // wolves
			},
			MaxRounds: 38,
		},

		// Premier League 2021/22
		"202122_1": {
			ID:        "202122_1",
			ClientID:  SeasonIdentifier{SeasonID: "PL"},
			Name:      "Premier League 2021/22",
			ShortName: "Prem 21/22",
			Live: TimeFrame{
				From:  time.Date(2021, 8, 13, 20, 0, 0, 0, loc),   // opening day kick-off
				Until: time.Date(2022, 5, 22, 23, 59, 59, 0, loc), // end of final day (regular season)
			},
			EntriesAccepted: TimeFrame{
				From:  time.Date(2021, 7, 30, 12, 0, 0, 0, loc), // competition opens
				Until: time.Date(2021, 8, 13, 20, 0, 0, 0, loc), // opening day kick-off
			},
			PredictionsAccepted: TimeFrame{
				From:  time.Date(2021, 7, 30, 12, 0, 0, 0, loc),   // competition opens
				Until: time.Date(2022, 5, 22, 23, 59, 59, 0, loc), // midnight on final day (regular season)
			},
			TeamIDs: []string{
				"AFC",
				"AVFC",
				"BFC3",
				"BHAFC",
				"BFC",
				"CFC",
				"CPFC",
				"EFC",
				"LUFC",
				"LCFC",
				"LFC",
				"MCFC",
				"MUFC",
				"NUFC",
				"NCFC",
				"SFC",
				"THFC",
				"WFC",
				"WHUFC",
				"WWFC",
			},
			MaxRounds: 38,
		},

		// Premier League 2020/21
		"202021_1": {
			ID:        "202021_1",
			ClientID:  SeasonIdentifier{SeasonID: "PL"},
			Name:      "Premier League 2020/21",
			ShortName: "Prem 20/21",
			Live: TimeFrame{
				From:  time.Date(2020, 9, 12, 15, 0, 0, 0, loc),   // opening day kick-off
				Until: time.Date(2021, 5, 23, 23, 59, 59, 0, loc), // end of final day (regular season)
			},
			EntriesAccepted: TimeFrame{
				From:  time.Date(2020, 8, 29, 9, 0, 0, 0, loc),  // competition opens
				Until: time.Date(2020, 9, 12, 15, 0, 0, 0, loc), // opening day kick-off
			},
			PredictionsAccepted: TimeFrame{
				From:  time.Date(2020, 8, 29, 9, 0, 0, 0, loc), // competition opens
				Until: time.Date(2021, 5, 23, 0, 0, 0, 0, loc), // midnight on final day (regular season)
			},
			TeamIDs: []string{
				"AFC",
				"AVFC",
				"BFC",
				"BHAFC",
				"CFC",
				"CPFC",
				"EFC",
				"FFC",
				"LFC",
				"LCFC",
				"LUFC",
				"MCFC",
				"MUFC",
				"NUFC",
				"SUFC",
				"SFC",
				"THFC",
				"WBAFC",
				"WHUFC",
				"WWFC",
			},
			MaxRounds: 38,
		},

		// Championship 2020/21
		"202021_2": {
			ID:        "202021_2",
			ClientID:  SeasonIdentifier{SeasonID: "ELC"},
			Name:      "Championship 2020/21",
			ShortName: "Champ 20/21",
			Live: TimeFrame{
				From:  time.Date(2020, 9, 12, 15, 0, 0, 0, loc),  // opening day kick-off
				Until: time.Date(2021, 5, 9, 23, 59, 59, 0, loc), // end of final day (regular season)
			},
			EntriesAccepted: TimeFrame{
				From:  time.Date(2020, 8, 29, 9, 0, 0, 0, loc),  // competition opens
				Until: time.Date(2020, 9, 12, 15, 0, 0, 0, loc), // opening day kick-off
			},
			PredictionsAccepted: TimeFrame{
				From:  time.Date(2020, 8, 29, 9, 0, 0, 0, loc), // competition opens
				Until: time.Date(2021, 5, 9, 0, 0, 0, 0, loc),  // midnight on final day (regular season)
			},
			TeamIDs: []string{
				"AFCB",
				"BFC2",
				"BFC3",
				"BCFC",
				"BCFC2",
				"BRFC",
				"CCFC",
				"CCFC2",
				"DCFC",
				"HTAFC",
				"LTFC",
				"MFC",
				"MFC2",
				"NCFC",
				"NFFC",
				"PNEFC",
				"QPRFC",
				"RFC",
				"RUFC",
				"SCAFC",
				"SCFC",
				"SWFC",
				"WFC",
				"WWFC2",
			},
			MaxRounds: 46,
		},

		// Premier League 2019/20
		"201920_1": {
			ID:       "201920_1",
			ClientID: SeasonIdentifier{SeasonID: "PL"},
			Name:     "Premier League 2019/20",
			Live: TimeFrame{
				From:  time.Date(2019, 8, 9, 19, 0, 0, 0, loc),    // opening day kick-off
				Until: time.Date(2020, 7, 26, 23, 59, 59, 0, loc), // end of final day (regular season)
			},
			EntriesAccepted: TimeFrame{
				From:  time.Date(2019, 7, 1, 0, 0, 0, 0, loc),  // competition opens
				Until: time.Date(2019, 8, 9, 19, 0, 0, 0, loc), // opening day kick-off
			},
			PredictionsAccepted: TimeFrame{
				From:  time.Date(2019, 7, 1, 0, 0, 0, 0, loc),  // competition opens
				Until: time.Date(2020, 7, 26, 0, 0, 0, 0, loc), // midnight on final day (regular season)
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
				"WHUFC",
				"WWFC",
			},
			MaxRounds: 38,
		},
	}

	// define fake season
	sc[FakeSeasonID] = Season{
		ID:       FakeSeasonID,
		ClientID: nil, // will not invoke requests to client when running in retrieve latest standings job
		Name:     "Localhost Season",
		Live: TimeFrame{
			From:  time.Now(),
			Until: time.Now().Add(60 * time.Minute),
		},
		EntriesAccepted: TimeFrame{
			From:  time.Now(),
			Until: time.Now().Add(20 * time.Minute),
		},
		PredictionsAccepted: TimeFrame{
			From:  time.Now(),
			Until: time.Now().Add(60 * time.Minute),
		},
		TeamIDs:   sc[fakeSeasonKey].TeamIDs,
		MaxRounds: sc[fakeSeasonKey].MaxRounds,
	}

	return sc, nil
}

// GetRealmCollection returns the required RealmCollection
func GetRealmCollection() (RealmCollection, error) {
	return parseRealmsFromPath(fmt.Sprintf("./data/realms.yml"))
}

// parseRealmsFromPath parses the realms from the contents of the YAML file at the provided path
func parseRealmsFromPath(yamlPath string) (RealmCollection, error) {
	contents, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file '%s': %w", yamlPath, err)
	}

	var payload struct {
		Realms RealmCollection `yaml:"realms"`
	}

	// parse file contents
	if err := yaml.Unmarshal(contents, &payload); err != nil {
		return nil, fmt.Errorf("cannot unmarshal yaml at path '%s': %w", yamlPath, err)
	}

	// populate names of realms with map key
	for key := range payload.Realms {
		r := payload.Realms[key]
		r.Name = key
		payload.Realms[key] = r
	}

	return payload.Realms, nil
}
