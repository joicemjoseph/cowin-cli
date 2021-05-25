package cowin

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

func reqAddHeaders(req *http.Request) {
	req.Header.Add("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.212 Safari/537.36")
	req.Header.Add("origin", "https://selfregistration.cowin.gov.in/")
	req.Header.Add("referer", "https://selfregistration.cowin.gov.in/")
	req.Header.Add("sec-fetch-site", "cross-site")
	req.Header.Add("sec-fetch-mode", "cors")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua", `" Not A;Brand";v="99", "Chromium";v="90", "Google Chrome";v="90"`)

}

func getReqAuth(URL, bearerToken string, auth bool) ([]byte, int) {
	var body []byte
	var respCode int
	client := http.Client{}

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		log.Println(err)
	}

	reqAddHeaders(req)
	if auth {
		req.Header.Add("authorization", fmt.Sprintf("Bearer %s", bearerToken))
	}
	resp, err := client.Do(req)

	if err == nil {
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		respCode = resp.StatusCode
	} else {
		log.Println(err)
	}

	return body, respCode

}

func postReq(URL string, postData []byte, bearerToken string) ([]byte, int) {
	var body []byte
	var respCode int

	client := http.Client{}

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(postData))
	if err != nil {
		log.Println(err)
	}

	reqAddHeaders(req)

	if bearerToken != "" {
		req.Header.Add("authorization", fmt.Sprintf("Bearer %s", bearerToken))
	}
	resp, err := client.Do(req)

	if err == nil {
		defer resp.Body.Close()
		body, _ = io.ReadAll(resp.Body)
		respCode = resp.StatusCode
	} else {
		log.Println(err)
	}

	return body, respCode
}
