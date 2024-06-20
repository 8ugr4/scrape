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

// inputFilepath contains the path of the given document for the URL file.

const (
	inputFilepath = "C:/Users/Bugra/Desktop/listOfUrl.txt"
	newFilepath   = "C:/Users/Bugra/Desktop/reportOfUrl.txt"
)

type Scrapper struct {
	oldFilepath string // old filepath
	newFilepath string // new filepath
	workersCnt  int
}

type response struct {
	Status string
	Url    string
	Length int64
}

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	workersCnt = runtime.NumCPU()
)

func New(newFilepath string, workersCnt int) *Scrapper {
	return &Scrapper{oldFilepath: inputFilepath, newFilepath: newFilepath, workersCnt: workersCnt}
}

func (scrapper *Scrapper) Run() {
	defer func(start time.Time) {
		log.Printf("it took %v to finish the Run()", time.Since(start))
	}(time.Now())

	log.Printf("using %d workers", workersCnt)

	urlStrCh := make(chan string, workersCnt) // urls channel
	parseCh := make(chan *response, 3)        // parsed channel
	errCh := make(chan error, 1)              //error channel

	var workerWg sync.WaitGroup

	// reads file from filePath and sends the inputs
	go readFile(inputFilepath, urlStrCh, errCh)
	go func() {
		time.Sleep(2 * time.Millisecond)
		select {
		case msg1 := <-errCh:
			if msg1 != nil {
				log.Println(msg1)
			}
		}
	}()
	//into urlStrCh channel as it reads.
	for i := 0; i < workersCnt; i++ {
		workerWg.Add(1)
		go httpWorker(&workerWg, urlStrCh, parseCh, errCh)
	}

	//reads from urlStrCh and writes into .csv data.
	go writeIntoFile(parseCh, errCh)

	workerWg.Wait()
	close(errCh)
}

// readFile reads the file from filePath, and while reading sends the url input to the urlFlowSender channel.
func readFile(filePath string, urlStrCh chan<- string, errCh chan<- error) {

	content, err := os.Open(filePath)
	if err != nil {
		errCh <- err
	}
	defer func(content *os.File) {
		err := content.Close()
		if err != nil {
			errCh <- err
		}
	}(content)

	s := bufio.NewScanner(content)
	for s.Scan() {
		urlStr := s.Text()
		if urlStr == "" {
			errCh <- err
			continue
		}

		urlStrCh <- urlStr
	}

	if err := s.Err(); err != nil {
		errCh <- err
		log.Fatalf("could not read a line from the file: %v", err)
	}

	close(urlStrCh)
	//important. always close the channel if other functions also use it, but wait for it to finish.
}

func httpWorker(wg *sync.WaitGroup, urlStrCh <-chan string, ParseCh chan<- *response, errCh chan<- error) {
	defer wg.Done()
	defer func(start time.Time) {
		log.Printf("it took %v to finish for the worker", time.Since(start))
	}(time.Now())

	for urlStr := range urlStrCh {
		request, err := http.NewRequest(http.MethodGet, urlStr, nil) //  GETS A RESPONSE
		if err != nil {
			errCh <- err
		}

		request.URL, err = url.Parse(urlStr)
		if err != nil {
			errCh <- err
		}
		resp := response{
			Url: urlStr,
		} //new response instance

		httpResp, err := httpClient.Do(request)
		if err != nil {
			log.Fatalf("error when sending request to the server: %v", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(httpResp.Body)

		resp.Status = httpResp.Status
		if httpResp.ContentLength == -1 {
			body, err := io.ReadAll(httpResp.Body)
			resp.Length = int64(len(body))
			if err != nil {
				errCh <- err
			}
		} else {
			resp.Length = httpResp.ContentLength
		}

		ParseCh <- &resp
	}
}

// writeIntoFile writes the parsed Input(URL) to a file.
// reads from parsedFlowReceiver puts the inputs into a .csv file.

func writeIntoFile(parseCh <-chan *response, errCh chan<- error) {

	//"scrapedFile.csv" changed with newFilepath
	fp, err := os.Create(newFilepath)
	if err != nil {
		errCh <- err
	}
	defer func() {
		if err := fp.Close(); err != nil {
			errCh <- err
		}
	}()

	for resp := range parseCh {
		_, err := fmt.Fprintf(fp, "%v: %v: %v\n", resp.Status, resp.Url, resp.Length)
		if err != nil {
			errCh <- err
		} // todo check for the error
		if _, err := fp.Write([]byte(resp.Url)); err != nil {
			log.Fatalf("could not write into the file %v\n", err)
		}
	}
}

// readFileWg := &sync.WaitGroup{}, use : readFileWg
// var readFileWg sync.WaitGroup, use :  &readFileWg

func main() {
	myScrapper := New("newFilepath", 1)
	myScrapper.Run()

}

/*
	after writing run function go for error channel
	smth like this:

	val, ok := mymap[123]

	err, ok := <- errCh
	if err =
	// if ch is closed then ok.
	close(parsedFlowSender)

*/
