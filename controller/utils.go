package controller

import (
	"log"
	//"strconv"
	"net/http"
	"html/template"
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

