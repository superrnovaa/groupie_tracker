package models

import (
	"html/template"
	"log"
	"net/http"
	"time"
)

func ServeFile(webpage http.ResponseWriter, pageName string, object interface{}) {
	content, err := template.ParseFiles("./view/html/" + pageName)
	p500, _ := template.ParseFiles("./view/html/500.html")
	if err != nil {
		log.Printf("[ERROR] - File \"%v\" does not exist or is not accessible.\n%v",
			pageName, err.Error())
		webpage.WriteHeader(http.StatusInternalServerError)
		p500.ExecuteTemplate(webpage, "500.html.html", nil)
	}
	err = content.ExecuteTemplate(webpage, pageName, object)
	if err != nil {
		log.Printf("[ERROR] - Template execution.\n" + err.Error() + "\n\n")
		webpage.WriteHeader(http.StatusInternalServerError)
		p500.ExecuteTemplate(webpage, "500.html.html", nil)
	}
}

func (self *ApiData) WaitThenRefreshApi() {
	/*
		Method of ApiData
		Loop forever, waits 24 hours then refresh the api and empty the caches
	*/
	for true {
		time.Sleep(24 * time.Hour)
		log.Printf("[INFO] - 30s since last cache update, refreshing the API.")
		self.FeedApi()
	}
}

func KeepOnlyDuplicates(longarray []BandInfo, shorterarray []BandInfo) []BandInfo {
	result := []BandInfo{}
	allKeys := make(map[int]bool)

	for _, element := range longarray {
		allKeys[element.Id] = true
	}

	for _, element := range shorterarray {
		if allKeys[element.Id] {
			result = append(result, element)
		}
	}

	return result
}
