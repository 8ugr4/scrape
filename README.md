
// UPDATE 19.06.24/21.20/: net/url is not enough to scrape <meta> tags.

// use net/http and maybe html to parse the html response
// to extract charset from <meta>

// goal:
// input: list of URL's in a text file. (.txt)
// <meta charset="UTF-8"> --> UTF-8
// output: "url":"<charset>" in this format

// ::: WORKING STRUCTURE OF THE PROGRAM :::
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
