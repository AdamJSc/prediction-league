package pages

import "prediction-league/service/internal/models"

type MenuLink struct {
	Label    string
	Key      string
	Href     string
	ID       string
	IsActive bool
}

type Menu []MenuLink

func getInflatedMenu() Menu {
	return Menu{
		{
			Label: "Home",
			Key:   "home",
			Href:  "/",
		},
		{
			Label: "Results",
			Key:   "results",
			Href:  "/results",
		},
		{
			Label: "Enter",
			Key:   "enter",
			Href:  "/enter",
		},
		{
			Label: "FAQ",
			Key:   "faq",
			Href:  "/faq",
		},
		{
			Label: "Login",
			Key:   "login",
			Href:  "/login",
			ID:    "login-link",
		},
	}
}

func setActiveMenuLink(m Menu, key string) Menu {
	for idx := range m {
		link := &m[idx]
		if link.Key == key {
			link.IsActive = true
			return m
		}
	}

	return m
}

func InflatedMenu(activeLink string) Menu {
	return setActiveMenuLink(getInflatedMenu(), activeLink)
}

type Base struct {
	Title string
	Menu  Menu
	Data  interface{}
}

type SelectionPageData struct {
	Err                    error
	IsAcceptingSelections  bool
	SelectionsNextAccepted *models.TimeFrame
}
