package zoom

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func httpGet(client *http.Client, url string, headers http.Header) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header = headers
	return client.Do(request)
}

func (session *ZoomSession) generateSignatureJWT() string {
	meetingNumber := session.MeetingNumber
	timestamp := strconv.FormatInt((time.Now().UTC().UnixNano()/1e6)-30000, 10)

	h := hmac.New(sha256.New, []byte(session.ZoomApiSecret))
	h.Write([]byte(base64.StdEncoding.EncodeToString([]byte(session.ZoomApiKey + meetingNumber + timestamp + ZOOM_ROLE))))
	return base64.StdEncoding.EncodeToString([]byte(session.ZoomApiKey + "." + meetingNumber + "." + timestamp + "." + ZOOM_ROLE + "." + base64.StdEncoding.EncodeToString(h.Sum(nil))))
}

func (session *ZoomSession) generateSignatureSDK() string {
	meetingNumber := session.MeetingNumber
	// Give ourselves 50 seconds leeway, this code is kinda brittle because it relies on
	// the clock of the system on which it's running, perhaps we can get this timestamp
	// from the server directly
	ts := time.Now().Unix() - 50

	header := []byte(`{"alg":"HS256","typ":"JWT"}`)
	payload := []byte(fmt.Sprintf(`{"sdkKey":"%s","iat":%d,"exp":%d,"mn":%s,"role":0}`, session.ZoomApiKey, ts, ts+1800, meetingNumber))
	message := base64.URLEncoding.EncodeToString(header) + "." + base64.URLEncoding.EncodeToString(payload)

	h := hmac.New(sha256.New, []byte(session.ZoomApiSecret))
	h.Write([]byte(message))
	return message + "." + base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (session *ZoomSession) GetMeetingInfoData() (*MeetingInfo, string, error) {
	var meetingInfo MeetingInfo

	// generate info url
	values := url.Values{}
	values.Set("meetingNumber", session.MeetingNumber)
	values.Set("userName", session.Username)
	values.Set("passWord", session.MeetingPassword)
	switch apiType := session.ZoomApiType; apiType {
	case ZOOM_JWT_API_TYPE:
		values.Set("signatureType", "api")
		values.Set("signature", session.generateSignatureJWT())
	case ZOOM_SDK_API_TYPE:
		values.Set("signatureType", "sdk")
		values.Set("signature", session.generateSignatureSDK())
	}
	values.Set("apiKey", session.ZoomApiKey)
	values.Set("lang", "en-US")
	values.Set("userEmail", "")
	values.Set("cv", "2.12.0")
	values.Set("proxy", "1")
	values.Set("sdkOrigin", "aHR0cDovL2xvY2FsaG9zdDo5OTk5")
	values.Set("tk", "")
	values.Set("ztk", "")
	values.Set("sdkUrl", "aHR0cDovL2xvY2FsaG9zdDo5OTk5L21lZXRpbmcuaHRtbA")
	values.Set("captcha", "")
	values.Set("captchaName", "")
	values.Set("suid", "")
	values.Set("callback", "axiosJsonpCallback1")
	infoUrl := (&url.URL{
		Scheme:   "https",
		Host:     "zoom.us",
		Path:     "/api/v1/wc/info",
		RawQuery: values.Encode(),
	}).String()

	response, err := httpGet(session.httpClient, infoUrl, httpHeaders())
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, "", err
	}

	getStringInBetweenTwoString := func(str []byte, startS []byte, endS []byte) []byte {
		s := bytes.Index(str, startS)
		if s == -1 {
			return nil
		}
		newS := str[s+len(startS):]

		e := bytes.Index(newS, endS)
		if e == -1 {
			return nil
		}
		return newS[:e]
	}

	log.Printf("%v", string(data))
	jsonData := getStringInBetweenTwoString(data, []byte("axiosJsonpCallback1("), []byte(")"))
	err = json.Unmarshal(jsonData, &meetingInfo)
	if err != nil {
		return nil, "", err
	}

	if meetingInfo.ErrorCode > 0 {
		return nil, "", fmt.Errorf(meetingInfo.ErrorMessage)
	}

	var cookieString string
	for _, cookieValue := range response.Cookies() {
		// skip "cred" and empty cookies
		if cookieValue.Name == "cred" || cookieValue.Value != "" {
			cookieString += fmt.Sprintf("%s=%s; ", cookieValue.Name, cookieValue.Value)
		}
	}
	// remove trailing ";"" if it exists
	cookieString = strings.TrimSuffix(cookieString, "; ")

	return &meetingInfo, cookieString, nil
}

func getRwgPingServer(meetingInfo *MeetingInfo) *RwgInfo {
	var rwgPingInfo RwgInfo

	for key, value := range meetingInfo.Result.EncryptedRWC {
		rwgPingInfo.Rwg = key
		rwgPingInfo.RwcAuth = value
		break
	}

	return &rwgPingInfo
}

// @TODO(bug): if meeting not joinable, returns all false.
func (session *ZoomSession) getRwgPingData(meetingInfo *MeetingInfo, pingRwcServer *RwgInfo) (*RwgInfo, error) {

	headers := httpHeaders()
	headers["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	pingUrl := (&url.URL{
		Scheme: "https",
		Host:   pingRwcServer.Rwg,
		Path:   fmt.Sprintf("/wc/ping/%s", meetingInfo.Result.MeetingNumber),
		// @TODO(security): THE ORDER OF PARAMTERS IS VERY IMPORTANT! IT DOES NOT WORK OTHERWISE
		RawQuery: fmt.Sprintf("ts=%d&auth=%s&rwcToken=%s&dmz=1", meetingInfo.Result.Ts, meetingInfo.Result.Auth, pingRwcServer.RwcAuth),
		// RawQuery: values.Encode(),
	}).String()

	response, err := httpGet(session.httpClient, pingUrl, headers)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var rwgPingInfo RwgInfo
	err = json.NewDecoder(response.Body).Decode(&rwgPingInfo)
	if err != nil {
		return nil, err
	}

	return &rwgPingInfo, nil
}
