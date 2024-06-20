package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

// UPDATE 19.06.24/21.20/: net/url is not enough to scrape <meta> tags.
// use net/http and maybe html to parse the html response
// to extract charset from <meta>

// goal:
// input: list of URL's in a text file. (.txt)
// <meta charset="UTF-8"> --> UTF-8
// output: "url":"<charset>" in this format
// read the file, save the urls.
// http scraping
// use goroutines to scrape every URL
// use channels(buffer with the number of total goroutines)
// while reading with goroutines, use one goroutine to carry the input to the output file
// use another goroutine to convert the taken input from READER goroutines into expected format
// format is: "url":"<charset>"
// target address URL

// EXAMPLE OUTPUT AT THE MOMENT: (21.40)
/*
scrapedInput:= https://drstearns.github.io/tutorials/gojson/
scrapedInput:= https://leangaurav.medium.com/common-mistakes-when-using-golangs-sync-waitgroup-88188556ca54
scrapedInput:= https://stackoverflow.com/questions/48271388/for-loop-with-buffered-channel

Process finished with the exit code 0

*/

// InputFileAddress contains the path of the given document
// for the URL file.
const InputFileAddress string = "C:/Users/Bugra/Desktop/listOfUrl.txt"

// takeFile reads the file from filePath, and while reading sends the url input to the urlFlowSender channel.
func takeFile(wg *sync.WaitGroup, filePath string, urlFlowSender chan<- string) {
	defer wg.Done()

	content, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer content.Close()

	s := bufio.NewScanner(content)
	for s.Scan() {
		urlvar := s.Text()
		urlFlowSender <- urlvar
		// todo send to the channel
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
	close(urlFlowSender)
	//important. always close the channel if other functions also use it, but wait for it to finish.
}

// UPDATE 21.20: at the moment it does not parse anything, just sends the URLs around.
// parseUrlQuery: parses raw URLs (from urlFlowReceiver) into URL structures and sends
// the parsedURL's (still just same URL's given at the start until HTTP and HTML update)
// into parsedFlowSender channel.
/*
func parseUrlQuery(wg *sync.WaitGroup, urlFlowReceiver <-chan string, parsedFlowSender chan<- string) {
	defer wg.Done()

	for urlInput := range urlFlowReceiver {
		parsedURL, err := url.Parse(urlInput)
		if err != nil {
			log.Fatal(err)
			continue
		}
		parsedFlowSender <- parsedURL.String()
	}
}
*/

type Response struct {
	Status string `json:"status"`        // Fields must start with capital letters to be exported.
	Url    string `json:"url,omitempty"` // Only exported fields will be encoded/decoded in JSON.
	Length int    `json:"length"`
}

func httpWorker(wg *sync.WaitGroup, urlFlowReceiver <-chan string, parsedFlowSender chan<- *Response) {
	defer wg.Done()
	client := &http.Client{}
	// what to do if 2024/06/20 11:36:47 Get "https://chat.insomnia.rest": dial tcp 34.133.30.248:443: connectex: No connection could be made because the target machine actively refused it.
	// make an error control if this happens.
	for urlInput := range urlFlowReceiver {
		request, err := http.NewRequest(http.MethodGet, urlInput, nil) //  GETS A RESPONSE
		if err != nil {                                                //a non 2-xx response doesnt cause error. when err is nil, resp always contains a non-nil resp.Body.
			log.Fatal(err)
		}
		q := request.URL.Query()
		q.Add("<meta charset>", "UTF-8")  //adding <meta> to query Argument.
		request.URL.RawQuery = q.Encode() // assigning encoded query string to http request.
		response, err := client.Do(request)
		if err != nil {
			log.Fatalln("error when sending request to the server")
			return
		}
		defer response.Body.Close()
		responseBody, err := io.ReadAll(response.Body)
		size := len(responseBody)
		if err != nil {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}
		resp := Response{} //new Responce instance
		resp.Url = urlInput
		resp.Status = response.Status
		resp.Length = size
		parsedFlowSender <- &resp
	}
}

// writeInFile writes the parsed URL queries to a file.
// reads from parsedFlowReceiver(actually parsedFlowSender), puts the inputs into a .csv file.
func writeInFile(wg *sync.WaitGroup, parsedFlowReceiver <-chan *Response) {
	//but in here we read from it (parsedFlowSender = parsedFlowReceiver).
	defer wg.Done()

	file, err := os.Create("scrapedFile.csv")
	if err != nil {
		log.Fatalln("Failed to create the file:", err)
	}
	defer file.Close()
	for resp := range parsedFlowReceiver {
		fmt.Fprintf(file, "%v: %v: %v\n", resp.Status, resp.Url, resp.Length)
	}
	/*
		for resp := range parsedFlowReceiver {
			file.WriteString(resp.Status)
			file.WriteString(resp.Url)
			file.WriteString(string(resp.Length))
			log.Fatalln("Failed to write to file:", err)
		}
	*/

}

const workersCnt = 3 //number of Workers.

// readFileWg := &sync.WaitGroup{}, use : readFileWg
// var readFileWg sync.WaitGroup, use :  &readFileWg

func main() {
	urlFlowSender := make(chan string, workersCnt) // urls channel
	parsedFlowSender := make(chan *Response, 3)    // parsed channel
	var readFileWg sync.WaitGroup
	var workerWg sync.WaitGroup
	var writerWg sync.WaitGroup
	//var resps []*Response
	//var httpWg sync.WaitGroup

	//workerDone := make(chan struct{}, workersCnt) // helps synchronize the completion of worker goroutines
	// empty struct in go is a zero size type. (probably it's idiomatic way to use that so no extra memory usage.)
	// there is also no need to send any additional data. so kind of a way to send done signal. (not boolean but still)
	// especially with parsedFlowSender

	readFileWg.Add(1)
	go takeFile(&readFileWg, InputFileAddress, urlFlowSender) //reads file from filePath and sends the inputs
	//into urlFlowSender channel as it reads.

	for i := 0; i < workersCnt; i++ { //while takeFile is reading, parseUrlQuery starts to read from
		workerWg.Add(1) //takeFile with(urlFlowReceiver = urlFlowSender)
		// and parse the URL's (atm it does not)
		go httpWorker(&workerWg, urlFlowSender, parsedFlowSender) //and sends to parsedFlowSender channel
	}

	writerWg.Add(1)
	go writeInFile(&writerWg, parsedFlowSender) //reads from urlFlowSender and writes into .csv data.
	//httpWg.Wait()
	readFileWg.Wait()
	workerWg.Wait()
	// waiting for parseUrlQuery goroutines
	/*
		for i := 0; i < workersCnt; i++ {
			<-workerDone
		}
	*/
	close(parsedFlowSender)
	writerWg.Wait()
	// parse the url's with parseUrlQuery function
	// go parseUrlQuery(urlFlowSender, urlFlowReceiver)

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
