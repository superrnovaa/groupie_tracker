package main
import (
	//"groupie-tracker/controller"
	"groupie-tracker/models"
	"log"
	"net/http"
	//"strconv"
)
func main() {
	PORT := "7878"
	bandsData := &models.ApiData{}
	bandsData.FeedApi()
	//bandsData.CreateCaches()
	//coords := &models.ApiCoords{}
	staticFiles := http.FileServer(http.Dir("view/"))
	http.Handle("/view/", http.StripPrefix("/view/", staticFiles))
	http.HandleFunc("/", bandsData.RootHandler)
	//http.HandleFunc("/search", models.SearchHandler)

	log.Printf("[INFO] - Starting server on http://localhost:" + PORT + "/")
	go bandsData.WaitThenRefreshApi()
	err := http.ListenAndServe(":"+PORT, nil)
	if err != nil {
		log.Fatal("[ERROR] - Server not started properly.\n" + err.Error())
		
	}
}



