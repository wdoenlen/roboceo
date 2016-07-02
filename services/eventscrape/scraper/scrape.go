package scraper

import (
	"fmt"
	"time"

	"github.com/tebeka/selenium"
)

const MaxPages = 500

func EnsureLoggedIn(wd selenium.WebDriver, username, password string) error {
	if err := wd.Get("https://www.facebook.com"); err != nil {
		return err
	}

	if _, err := OnWelcomePage(wd); err == nil {
		return nil // already logged in
	}

	loginPage, err := OnLoginPage(wd)
	if err != nil {
		return fmt.Errorf("login: %v", err)
	}
	if err := loginPage.Login(username, password); err != nil {
		return fmt.Errorf("login: %v", err)
	}

	return nil
}

func GetAllEvents(wd selenium.WebDriver, searchURL string, ids chan string) error {
	if err := wd.Get(searchURL); err != nil {
		return err
	}

	eventsPage, err := OnEventsPage(wd)
	if err != nil {
		return err
	}

	if err := eventsPage.InjectScraper(); err != nil {
		return err
	}

	for i := 0; i < MaxPages; i++ {
		pageletIDs, err := eventsPage.GetEventPagelet()
		if err != nil {
			return err
		}
		if len(pageletIDs) == 0 {
			break
		}

		for _, id := range pageletIDs {
			ids <- id
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}
