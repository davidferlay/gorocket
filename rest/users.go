package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/skilld-labs/gorocket/api"
)

type logoutResponse struct {
	statusResponse
	data struct {
		message string `json:"message"`
	} `json:"data"`
}

type logonResponse struct {
	statusResponse
	Data struct {
		Token  string `json:"authToken"`
		UserId string `json:"userId"`
	} `json:"data"`
}

type usersResponse struct {
	Users   []api.User `json:"users"`
	Success bool       `json:"success"`
}

type userResponse struct {
	User    api.User `json:"user,omitempty"`
	Success bool     `json:"success"`
}

type preferencesResponse struct {
	Preferences api.UserPreferences `json:"preferences"`
	Success     bool                `json:"success"`
}

// Login a user. The Email and the Password are mandatory. The auth token of the user is stored in the Client instance.
//
// https://rocket.chat/docs/developer-guides/rest-api/authentication/login
func (c *Client) Login(credentials api.UserCredentials) error {
	body, err := json.Marshal(struct {
		User     string `json:"user"`
		Password string `json:"password"`
	}{
		User:     credentials.Email,
		Password: credentials.Password,
	})
	if err != nil {
		return err
	}
	request, _ := http.NewRequest(http.MethodPost, c.getUrl()+"/api/v1/login", bytes.NewReader(body))
	response := new(logonResponse)

	if err := c.doRequest(request, response); err != nil {
		return err
	}

	if response.Status == "success" {
		c.auth = &authInfo{id: response.Data.UserId, token: response.Data.Token}
		return nil
	} else {
		return errors.New("Response status: " + response.Status)
	}
}

// Logout a user. The function returns the response message of the server.
//
// https://rocket.chat/docs/developer-guides/rest-api/authentication/logout
func (c *Client) Logout() (string, error) {

	if c.auth == nil {
		return "Was not logged in", nil
	}

	request, _ := http.NewRequest(http.MethodPost, c.getUrl()+"/api/v1/logout", nil)

	response := new(logoutResponse)

	if err := c.doRequest(request, response); err != nil {
		return "", err
	}

	if response.Status == "success" {
		return response.data.message, nil
	} else {
		return "", errors.New("Response status: " + response.Status)
	}
}

func (c *Client) Create(user *api.User) (*userResponse, error) {
	u := new(userResponse)
	body, err := json.Marshal(user)
	if err != nil {
		return &userResponse{}, err
	}
	request, _ := http.NewRequest(http.MethodPost, c.getUrl()+"/api/v1/users.create", bytes.NewReader(body))

	err = c.doRequest(request, u)
	return u, err
}

func (c *Client) Delete(id string) error {
	body, err := json.Marshal(struct {
		UserID string `json:"userId"`
	}{
		UserID: id,
	})
	if err != nil {
		return err
	}
	request, _ := http.NewRequest(http.MethodPost, c.getUrl()+"/api/v1/users.delete", bytes.NewReader(body))

	return c.doRequest(request, &userResponse{})
}

func (c *Client) GetUsers(query map[string]string) (*usersResponse, error) {
	users := new(usersResponse)
	queryJson, err := json.Marshal(query)
	q := url.QueryEscape(string(queryJson))

	request, _ := http.NewRequest("GET", c.getUrl()+"/api/v1/users.list?query="+q, nil)

	err = c.doRequest(request, users)
	return users, err
}

func (c *Client) GetUserInfo(user *api.User) (*api.User, error) {
	url := c.getUrl() + "/api/v1/users.info?"
	if user.Id != "" {
		url += "userId=" + user.Id
	} else {
		url += "username=" + user.UserName
	}
	request, _ := http.NewRequest("GET", url, nil)
	response := new(userResponse)

	if err := c.doRequest(request, response); err != nil {
		return nil, err
	}
	return &response.User, nil
}

func (c *Client) SetPreferences(id string, preferences *api.UserPreferences) error {
	body, err := json.Marshal(struct {
		UserID string               `json:"userId"`
		Data   *api.UserPreferences `json:"data"`
	}{
		UserID: id,
		Data:   preferences,
	})
	if err != nil {
		return err
	}
	request, _ := http.NewRequest(http.MethodPost, c.getUrl()+"/api/v1/users.setPreferences", bytes.NewReader(body))

	return c.doRequest(request, &preferencesResponse{})
}
