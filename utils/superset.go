package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// global http client dengan cookie jar
var client *http.Client

func init() {
	jar, _ := cookiejar.New(nil)
	client = &http.Client{Jar: jar}
}

// --- helper function ---
func CreateGuestToken(c *gin.Context, companyID string) (string, error) {
	// Ambil role dari context
	roleVal, _ := c.Get("role")
	role, _ := roleVal.(string)

	// login ke superset pakai service account
	accessToken, err := supersetLogin()
	if err != nil {
		return "", err
	}

	// ambil csrf token (akan otomatis simpan cookie session)
	csrfToken, err := getCSRFToken(accessToken)
	if err != nil {
		return "", err
	}

	url := os.Getenv("SUPERSET_URL") + "/api/v1/security/guest_token/"
	/*
		// siapkan payload RLS
		var rls []map[string]string
		if role != "super-admin" {
			rls = []map[string]string{
				{"clause": fmt.Sprintf("company_id = '%s'", companyID)},
			}
		}
	*/
	// Default rls: array kosong (bukan nil)
	rls := []map[string]interface{}{}

	if role != "super-admin" {
		//datasetID := os.Getenv("SUPERSET_DATASET_ID") // <- simpan dataset_id di env
		rls = append(rls, map[string]interface{}{
			"clause": fmt.Sprintf("company_id = '%s'", companyID),
			//"dataset": datasetID,
		})
	}

	dashboardIDs := os.Getenv("DASHBOARD_IDS") // contoh: "1,2,3"
	ids := strings.Split(dashboardIDs, ",")

	resources := []map[string]string{}
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			resources = append(resources, map[string]string{
				"type": "dashboard",
				"id":   id,
			})
		}
	}

	payload := map[string]interface{}{
		"user": map[string]string{
			"username":   "guest_" + companyID,
			"first_name": "Guest",
			"last_name":  "User",
		},
		"resources": resources,
		"rls":       rls,
	}

	return postSupersetGuestToken(url, accessToken, csrfToken, payload)
}

// --- login superset ---
func supersetLogin() (string, error) {
	url := os.Getenv("SUPERSET_URL") + "/api/v1/security/login"

	payload := map[string]interface{}{
		"username": os.Getenv("SUPERSET_USERNAME"),
		"password": os.Getenv("SUPERSET_PASSWORD"),
		"provider": "db",
		"refresh":  true,
	}

	jsonPayload, _ := json.Marshal(payload)
	//fmt.Println("[DEBUG] Login Payload:", string(jsonPayload)) // debug

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("[DEBUG] Login Response:", string(bodyBytes)) // debug

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("superset login failed: %s", string(bodyBytes))
	}

	var respData struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return "", err
	}

	return respData.AccessToken, nil
}

// --- post guest token ---
func postSupersetGuestToken(url, accessToken, csrfToken string, payload map[string]interface{}) (string, error) {
	jsonPayload, _ := json.Marshal(payload)
	//fmt.Println("[DEBUG] GuestToken Payload:", string(jsonPayload))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-CSRFToken", csrfToken) // penting: cocok dengan cookie session

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("[DEBUG] GuestToken Response:", string(bodyBytes))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("guest token failed: %s", string(bodyBytes))
	}

	var respData struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return "", err
	}

	return respData.Token, nil
}

// --- get csrf token ---
func getCSRFToken(accessToken string) (string, error) {
	url := os.Getenv("SUPERSET_URL") + "/api/v1/security/csrf_token/"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("[DEBUG] CSRF Response:", string(bodyBytes)) // debug

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("csrf token failed: %s", string(bodyBytes))
	}

	var respData struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(bodyBytes, &respData); err != nil {
		return "", err
	}

	return respData.Result, nil
}
