package core

import (
	"content/core/model"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rokwire/rokwire-building-block-sdk-go/services/core/auth"
)

// Adapter is the adapter for Core BB APIs
type Adapter struct {
	coreURL               string
	serviceAccountManager *auth.ServiceAccountManager
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(coreURL string, serviceAccountManager *auth.ServiceAccountManager) *Adapter {
	return &Adapter{coreURL: coreURL, serviceAccountManager: serviceAccountManager}
}

// LoadDeletedMemberships loads deleted memberships
func (a *Adapter) LoadDeletedMemberships() ([]model.DeletedUserData, error) {

	if a.serviceAccountManager == nil {
		log.Println("LoadDeletedMemberships: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/deleted-memberships?service_id=%s", a.coreURL, a.serviceAccountManager.AuthService.ServiceID)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("delete membership: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
	if err != nil {
		log.Printf("LoadDeletedMemberships: error sending request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("LoadDeletedMemberships: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("LoadDeletedMemberships: error with response code != 200")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to read json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	var deletedMemberships []model.DeletedUserData
	err = json.Unmarshal(data, &deletedMemberships)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to parse json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	return deletedMemberships, nil
}
