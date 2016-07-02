package scraper

import (
	"errors"

	"github.com/tebeka/selenium"
)

type LoginPage struct {
	wd selenium.WebDriver
}

func (p *LoginPage) Verify() error {
	_, err := p.wd.FindElement(selenium.ByCSSSelector, "#login_form")
	if err != nil {
		return errors.New("expected to be on login page")
	}
	return nil
}

func (p *LoginPage) Login(username, password string) error {
	email, err := p.wd.FindElement(selenium.ByID, "email")
	if err != nil {
		return err
	}

	pass, err := p.wd.FindElement(selenium.ByID, "pass")
	if err != nil {
		return err
	}

	form, err := p.wd.FindElement(selenium.ByCSSSelector, "#login_form")
	if err != nil {
		return err
	}

	if err := email.SendKeys(username); err != nil {
		return err
	}

	if err := pass.SendKeys(password); err != nil {
		return err
	}

	if err := form.Submit(); err != nil {
		return err
	}

	if err := p.Verify(); err == nil { // still on login page
		return errors.New("login failed")
	}

	return nil
}

func OnLoginPage(wd selenium.WebDriver) (*LoginPage, error) {
	page := &LoginPage{wd}
	if err := page.Verify(); err != nil {
		return nil, err
	}
	return page, nil
}
