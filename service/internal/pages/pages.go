package pages

type MenuLink struct {
	Label    string
	Key      string
	Href     string
	ID       string
	IsActive bool
}

type Menu []MenuLink

func (m *Menu) Inflate() *Menu {
	*m = Menu{
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
			ID:    "login-link",
		},
	}

	return m
}

func (m *Menu) SetActiveLink(key string) {
	menu := *m
	for idx := range menu {
		link := &menu[idx]
		if link.Key == key {
			link.IsActive = true
			return
		}
	}
}

type Base struct {
	Title string
	Menu  *Menu
}

func NewBase(title string, activeLink string) Base {
	var b = Base{
		Title: title,
		Menu:  &Menu{},
	}

	b.Menu.Inflate().SetActiveLink(activeLink)

	return b
}
