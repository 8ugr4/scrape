package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	inputFilepath = "C:/Users/Bugra/Desktop/listOfUrl.txt"
)

type Scrapper struct {
}

type response struct {
	Status string // Fields must start with capital letters to be exported.
	Url    string // Only exported fields will be encoded/decoded in JSON.
	Length int64
}

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	workersCnt = runtime.NumCPU()
)

func New() *Scrapper {
	return &Scrapper{}
}

// inputFilepath contains the path of the given document
// for the URL file.

// readFile reads the file from filePath, and while reading sends the url input to the urlFlowSender channel.
func readFile(filePath string, urlFlowSender chan<- string) {
	content, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer content.Close()

	s := bufio.NewScanner(content)
	for s.Scan() {
		urlStr := s.Text()
		if urlStr == "" {
			continue
		}

		urlFlowSender <- urlStr
	}

	if err := s.Err(); err != nil {
		log.Fatalf("could not read a line from the file: %v", err)
	}

	close(urlFlowSender)
	//important. always close the channel if other functions also use it, but wait for it to finish.
}

func httpWorker(wg *sync.WaitGroup, urlStrCh <-chan string, parsedFlowSender chan<- *response) {
	defer wg.Done()
	defer func(start time.Time) {
		log.Printf("it took %v to finish for the worker", time.Since(start))
	}(time.Now())

	// what to do if 2024/06/20 11:36:47 Get "https://chat.insomnia.rest": dial tcp 34.133.30.248:443: connectex: No connection could be made because the target machine actively refused it.
	// make an error control if this happens.
	for urlStr := range urlStrCh {
		request, err := http.NewRequest(http.MethodGet, urlStr, nil) //  GETS A RESPONSE
		if err != nil {                                              //a non 2-xx httpResp doesnt cause error. when err is nil, resp always contains a non-nil resp.Body.
			log.Fatal(err)
		}

		request.URL, err = url.Parse(urlStr)
		if err != nil {

		}
		resp := response{
			Url: urlStr,
		} //new response instance

		httpResp, err := httpClient.Do(request)
		if err != nil {
			log.Fatalf("error when sending request to the server: %v", err)
		}
		defer httpResp.Body.Close()

		resp.Status = httpResp.Status
		if httpResp.ContentLength == -1 {
			body, err := io.ReadAll(httpResp.Body)
			resp.Length = int64(len(body))
			if err != nil {
				log.Fatalln("could not read the size of the content.", err)
			}
		} else {
			resp.Length = httpResp.ContentLength
		}

		parsedFlowSender <- &resp
	}
}

// writeIntoFile writes the parsed URL queries to a file.
// reads from parsedFlowReceiver(actually parsedFlowSender), puts the inputs into a .csv file.
func writeIntoFile(parsedFlowReceiver <-chan *response) {
	//but in here we read from it (parsedFlowSender = parsedFlowReceiver).

	fp, err := os.Create("scrapedFile.csv")
	if err != nil {
		log.Fatalln("Failed to create the fp:", err)
	}
	defer func() {
		if err := fp.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	for resp := range parsedFlowReceiver {
		fmt.Fprintf(fp, "%v: %v: %v\n", resp.Status, resp.Url, resp.Length) // todo check for the error
	}
}

// readFileWg := &sync.WaitGroup{}, use : readFileWg
// var readFileWg sync.WaitGroup, use :  &readFileWg

func main() {
	//myScrapper := New("my input file", "my output file")
	//myScrapper.Run()

	log.Printf("using %d workers", workersCnt)

	urlFlowSender := make(chan string, workersCnt) // urls channel
	parsedFlowSender := make(chan *response, 3)    // parsed channel

	var workerWg sync.WaitGroup

	// reads file from filePath and sends the inputs
	go readFile(inputFilepath, urlFlowSender)

	//into urlFlowSender channel as it reads.
	for i := 0; i < workersCnt; i++ { //while readFile is reading, parseUrlQuery starts to read from
		workerWg.Add(1) //readFile with(urlFlowReceiver = urlFlowSender)
		// and parse the URL's (atm it does not)
		go httpWorker(&workerWg, urlFlowSender, parsedFlowSender) //and sends to parsedFlowSender channel
	}

	go writeIntoFile(parsedFlowSender) //reads from urlFlowSender and writes into .csv data.

	workerWg.Wait()

	close(parsedFlowSender)
}

// next goal: read <meta> tags with http and html, return using smth similar to the following struct.
// maybe additional stuff too, it depends.
/*
type definedUrl struct {
	urlname string `json:"url"`
	charset string `json:"<charset>"`
}

func newUrl(urlname string, charset string) *definedUrl {
	u := definedUrl{urlname: urlname, charset: charset}
	return &u
}
*/
