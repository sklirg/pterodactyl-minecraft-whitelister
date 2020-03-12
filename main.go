package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

// APIClient contains all the things we need to do API requests against pterodactyl
type APIClient struct {
	apiKey     string
	apiURL     string
	serverID   string
	httpClient http.Client
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	pterodactylURL := os.Getenv("PT_API")
	pterodactylServerID := os.Getenv("PT_SERVER_ID")
	pterodactylAPIKey := os.Getenv("PT_API_KEY")
	listenHost := os.Getenv("HOST")
	listenPort := os.Getenv("PORT")

	if listenPort == "" {
		listenPort = "8008"
	}

	c := APIClient{
		apiKey:   pterodactylAPIKey,
		apiURL:   pterodactylURL,
		serverID: pterodactylServerID,
	}
	c.httpClient = http.Client{}

	router := mux.NewRouter()
	router.HandleFunc("/whitelist", c.handleAddToWhitelist).Methods("POST")
	router.HandleFunc("/whitelist", c.handleRemoveFromWhitelist).Methods("DELETE")

	logrus.WithFields(logrus.Fields{
		"host": listenHost,
		"port": listenPort,
	}).Infof("Listening on %s:%s", listenHost, listenPort)

	http.ListenAndServe(fmt.Sprintf("%s:%s", listenHost, listenPort), router)
}

// APIResponse is a struct for building API response messages
type APIResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (c *APIClient) handleAddToWhitelist(w http.ResponseWriter, r *http.Request) {
	queryparams := r.URL.Query()

	// Illegal query
	if queryparams == nil || queryparams["username"] == nil || len(queryparams["username"]) != 1 {
		w.WriteHeader(400)

		e := APIResponse{
			Message: "Provide ONE username to add to the whitelist",
			Error:   "Provide ONE username to add to the whitelist",
		}

		r, err := json.Marshal(e)

		if err != nil {
			logrus.Error("Failed to convert error message to JSON")
			return
		}

		w.Write([]byte(r))
		return
	}

	// Update whitelist
	err := c.updateWhitelist(queryparams["username"][0], "add")

	// Handle whitelist update failures, if any
	if err != nil {
		w.WriteHeader(400)
		d, err := json.Marshal(APIResponse{
			Message: "Failed to whitelist user",
			Error:   err.Error(),
		})

		if err != nil {
			logrus.Error("Failed to marshal api response")
			return
		}

		w.Write(d)
		return
	}

	d, err := json.Marshal(APIResponse{
		Message: "success",
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to marshal api response")
	}

	w.WriteHeader(201)
	w.Write([]byte(d))
	return
}

func (c *APIClient) handleRemoveFromWhitelist(w http.ResponseWriter, r *http.Request) {
	queryparams := r.URL.Query()

	if queryparams == nil || queryparams["username"] == nil || len(queryparams["username"]) != 1 {
		w.WriteHeader(400)

		e := APIResponse{
			Message: "Provide ONE username to remove from the whitelist",
			Error:   "Provide ONE username to remove from the whitelist",
		}

		r, err := json.Marshal(e)

		if err != nil {
			logrus.Error("Failed to convert error message to JSON")
			return
		}

		w.Write([]byte(r))
		return
	}

	err := c.updateWhitelist(queryparams["username"][0], "remove")

	if err != nil {
		w.WriteHeader(400)

		d, err := json.Marshal(APIResponse{
			Message: "Failed to remove whitelisted user",
			Error:   err.Error(),
		})

		if err != nil {
			logrus.Error("Failed to marshal api response")
			return
		}

		w.Write(d)
	}

	w.WriteHeader(204)
}

func (c *APIClient) updateWhitelist(username, action string) error {
	endpoint := fmt.Sprintf("%s/api/client/servers/%s/command", c.apiURL, c.serverID)

	if action != "add" && action != "remove" {
		return fmt.Errorf("Illegal whitelist operation '%s'", action)
	}

	logrus.WithFields(logrus.Fields{
		"url":      endpoint,
		"action":   action,
		"username": username,
	}).Info("Whitelisting user")

	b := []byte(fmt.Sprintf("command=whitelist %s %s", action, username))
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(b))

	if err != nil {
		logrus.WithError(err).Error("Failed to create request")
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Add("Accept", "Application/vnd.pterodactyl.v1+json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"url":   endpoint,
		}).Error("Failed ")
		return err
	}

	if resp.StatusCode == 412 {
		logrus.Warning("Server might not be online. Status code is 412.")
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		logrus.WithError(err).Error("Failed to read response body")
		return err
	}

	logrus.Debugf("Response: %s", body)

	return nil
}
