package pages

type MenuLink struct {
	Label    string
	Key      string
	Href     string
	IsActive bool
}

type Menu []MenuLink

func (m *Menu) SetActiveLink(key string) {
	menu := *m
	for idx := range menu {
		if menu[idx].Key == key {
			menu[idx].IsActive = true
			return
		}
	}
}

type Main struct {
	Title string
	Menu  *Menu
}

func InflateWithMenu(p Main) Main {
	p.Menu = &Menu{
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
		},
	}

	return p
}
