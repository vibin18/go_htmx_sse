package main

import (
	"fmt"
	tmpl "html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Person struct {
	Name string
	Age  int
}

var FinalList *[]Person
var ch1 = make(chan Person)
var mu = sync.Mutex{}

func randName() string {
	wg := sync.WaitGroup{}
	mata := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m"}
	rand.Seed(time.Now().UnixNano())
	time.Sleep(1 * time.Microsecond)
	var myname []string
	ch := make(chan string)
	wg.Add(1)
	go func() {
		for i := 0; i < 6; i++ {
			rin := rand.Intn(12) + 1
			alpa := mata[rin]
			if i == 0 {
				alpa = strings.ToUpper(alpa)
			}

			ch <- alpa
		}
		wg.Done()
	}()
	wg.Add(1)
	go func() {
		for i := 0; i < 6; i++ {
			item := <-ch
			myname = append(myname, item)
		}
		wg.Done()
	}()
	wg.Wait()
	return strings.Join(myname, "")
}

func PersonGenerator() Person {
	rand.Seed(time.Now().UnixNano())
	nme := randName()
	age := rand.Intn(90) + 1
	return Person{
		nme,
		age,
	}
}

func RandomGen(ch1 chan Person) {
	for {
		ch1 <- PersonGenerator()
		time.Sleep(1 * time.Second)
	}
}

func Mytime(w http.ResponseWriter, r *http.Request) {

	log.Println("Received SSE message request")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("Failed to flush")
	}
	//htm := fmt.Sprintf("<h1>Time: %s </h1>", time.Now().Format("Jan 2 15:04:05 2006"))

loop:
	for {
		select {
		case data, ok := <-ch1:
			if !ok {
				break loop
			}

			mu.Lock()
			*FinalList = append(*FinalList, data)
			mu.Unlock()
			//data, err := json.Marshal(event)
			//if err != nil {
			//	fmt.Println("Error marshalling event:", err)
			//	return
			//}
			_, err := fmt.Fprintf(w, "event: sse1\ndata: %s\n\n", data)
			if err != nil {
				log.Println("Writing to response failed")
			}
			flusher.Flush()
		case <-r.Context().Done():
			// Close client connection on context cancellation
			// close(ch1)
			return
		}
	}

}

func RenderTable(w http.ResponseWriter, r *http.Request) {
	tmpl, err := tmpl.ParseGlob("html/*.html")
	if err != nil {
		log.Println("Failed to pasre template" + err.Error())
	}
	mu.Lock()
	mylist := &FinalList
	mu.Unlock()

	tmpl.ExecuteTemplate(w, "time.html", &mylist)
}

func main() {
	newslice := make([]Person, 0)
	FinalList = &newslice
	go RandomGen(ch1)

	http.HandleFunc("/time", Mytime)
	http.HandleFunc("/", index)
	http.HandleFunc("/table", RenderTable)
	log.Println("Starting SSE server...")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	tmpl, err := tmpl.ParseGlob("html/*.html")
	if err != nil {
		log.Println("Failed to parse template" + err.Error())
	}
	tmpl.ExecuteTemplate(w, "index.html", nil)
}
