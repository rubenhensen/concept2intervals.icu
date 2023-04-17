package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Data struct {
	Api_key_concept2  string
	Api_key_intervals string
}

type Results struct {
	Data []struct {
		ID            int    `json:"id"`
		UserID        int    `json:"user_id"`
		Date          string `json:"date"`
		Timezone      any    `json:"timezone"`
		DateUtc       any    `json:"date_utc"`
		Distance      int    `json:"distance"`
		Type          string `json:"type"`
		Time          int    `json:"time"`
		TimeFormatted string `json:"time_formatted"`
		WorkoutType   string `json:"workout_type"`
		Source        string `json:"source"`
		WeightClass   string `json:"weight_class"`
		Verified      bool   `json:"verified"`
		Ranked        bool   `json:"ranked"`
		Comments      any    `json:"comments"`
		StrokeData    bool   `json:"stroke_data"`
		RealTime      any    `json:"real_time"`
	} `json:"data"`
	Meta struct {
		Pagination struct {
			Total       int   `json:"total"`
			Count       int   `json:"count"`
			PerPage     int   `json:"per_page"`
			CurrentPage int   `json:"current_page"`
			TotalPages  int   `json:"total_pages"`
			Links       []any `json:"links"`
		} `json:"pagination"`
	} `json:"meta"`
}

func main() {
	t := time.NewTicker(24 * time.Hour)
	for range t.C {
		SyncConceptIntervals()
	}

	fmt.Println("Ticker stopped")
}

func SyncConceptIntervals() {
	prevhour := time.Now().Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	from := "?from=" + prevhour
	url := "https://log.concept2.com/api/users/me/results" + from
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}

	c2key := os.Getenv("conceptkey")
	req.Header.Add("Authorization", "Bearer  "+c2key)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer res.Body.Close()
	body, readErr := io.ReadAll(res.Body)

	if readErr != nil {
		fmt.Print(err.Error())
	}
	var results Results
	err = json.Unmarshal(body, &results)
	if err != nil {
		print("Error during Unmarshal(): ", err)
	}
	for _, v := range results.Data {
		url := "https://log.concept2.com/api/users/me/results/" + strconv.Itoa(v.ID) + "/export/fit"
		println(url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Print(err.Error())
		}
		req.Header.Add("Authorization", "Bearer  "+c2key)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Print(err.Error())
		}
		defer res.Body.Close()
		body, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			fmt.Print(err.Error())
		}

		err = os.WriteFile("temp.fit", body, 0644)
		if err != nil {
			println("error writing to file")
		}

		values := map[string]io.Reader{
			"file": mustOpen("temp.fit"),
		}

		err = Upload("https://intervals.icu/api/v1/athlete/"+os.Getenv("intervalsid")+"/activities", values)
		if err != nil {
			println("error uploading to file")
			panic(err)
		}

		os.Remove("temp.fit")
	}
}

func Upload(url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				panic(err)
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				panic(err)
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			panic(err)
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		panic(err)
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth("API_KEY", os.Getenv("intervalskey"))

	// Submit the request
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}
