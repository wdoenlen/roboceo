package scraper

import (
	"errors"

	"github.com/tebeka/selenium"
)

type WelcomePage struct {
	wd selenium.WebDriver
}

func (p *WelcomePage) Verify() error {
	_, err := p.wd.FindElement(selenium.ByCSSSelector, "#contentArea")
	if err != nil {
		return errors.New("expected to be on welcome page")
	}
	return nil
}

func OnWelcomePage(wd selenium.WebDriver) (*WelcomePage, error) {
	page := &WelcomePage{wd}
	if err := page.Verify(); err != nil {
		return nil, err
	}
	return page, nil
}
