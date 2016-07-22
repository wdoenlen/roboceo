package main

type YandexResponse struct {
	Code int      `json:"code"`
	Lang string   `json:"lang"`
	Text []string `json:"text"`
}

// func (s *Scraper) yandexTranslate(texts ...string) ([]string, error) {
// 	apiURL := "https://translate.yandex.net/api/v1.5/tr.json/translate?"

// 	params := make(url.Values)
// 	params.Set("key", s.YandexKey)
// 	for _, t := range texts {
// 		params.Add("text", t)
// 	}
// 	params.Set("lang", "en")
// 	apiURL += params.Encode()

// 	fmt.Println(apiURL)

// 	resp, err := http.Get(apiURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var yandex YandexResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&yandex); err != nil {
// 		return nil, err
// 	}

// 	if yandex.Code != 200 {
// 		return nil, fmt.Errorf("error %d", yandex.Code)
// 	}

// 	return yandex.Text, nil
// }

// func (s *Scraper) getTranslations(event Event) error {
// 	var eventJS struct {
// 		Name        string `json:"name"`
// 		Description string `json:"description"`
// 	}
// 	if err := json.Unmarshal(json.RawMessage(*event), &eventJS); err != nil {
// 		return err
// 	}

// 	// TODO(maxhawkins): DRY with corresponding in getMessages
// 	nameSummary := truncate(eventJS.Name, 50)
// 	descriptionSummary := truncate(eventJS.Description, 100)

// 	trans, err := s.yandexTranslate(nameSummary, descriptionSummary)
// 	if err != nil {
// 		return err
// 	}

// 	if err := s.db.SaveTranslation(nameSummary, trans[0]); err != nil {
// 		return err
// 	}
// 	if err := s.db.SaveTranslation(descriptionSummary, trans[1]); err != nil {
// 		return err
// 	}

// 	return nil
// }
