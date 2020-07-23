package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

func fixRequestSpaces(request string) string {
	re := strings.NewReplacer(" ", "+")
	formatted := strings.ToLower(re.Replace(request))
	return formatted
}

func main() {
	// Build URL - Need current epoch time in nanoseconds
	// requestTime := time.Now().UnixNano()
	url := strings.Join([]string{"https://themetaldetektor.com/detektor.php?time=", strconv.FormatInt(time.Now().UnixNano(), 10)}, "")

	// Build Request Body
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	var artist = flag.String("artist", "", "The artist to search for")
	var album = flag.String("album", "", "The album to search for")
	flag.Parse()
	// TODO - Parameterize this
	if *artist == "" {
		fmt.Println("Artist not specified. Aborting...")
		os.Exit(1)
	}
	writer.WriteField("artist", fixRequestSpaces(*artist))
	writer.WriteField("album-title", fixRequestSpaces(*album))
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	// Build Request
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		fmt.Println(err)
	}

	// Send Request to site
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	// Unmarshal Results into Struct
	body, err := ioutil.ReadAll(res.Body)
	var results Results
	json.Unmarshal(body, &results)

	// Build the Table. If no matches, return message
	if results.MatchCount < 1 {
		fmt.Println("No matches found...")
	} else {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Shop", "Listing", "Price"})
		for r := range results.Matches {
			table.Append([]string{results.Matches[r].Shop, results.Matches[r].Listing, results.Matches[r].Price})
		}

		table.Render()
	}
}
