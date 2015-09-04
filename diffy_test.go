package diffy

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

const (
	TestGerritInstanceURL = "https://go-review.googlesource.com/"
)

func TestNewClient_NoGerritInstance(t *testing.T) {
	mockData := []string{"", "://not-existing"}
	for _, data := range mockData {
		c, err := NewClient(data, nil)
		if c != nil {
			t.Errorf("NewClient return is not nil. Expected no client. Go %+v", c)
		}
		if err == nil {
			t.Error("No error occured by empty Gerrit Instance. Expected one.")
		}
	}
}

func TestNewClient_HttpClient(t *testing.T) {
	customHTTPClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	mockData := []struct {
		HTTPClient         *http.Client
		ExpectedHTTPClient *http.Client
	}{
		{nil, http.DefaultClient},
		{customHTTPClient, customHTTPClient},
	}

	for _, mock := range mockData {
		c, err := NewClient("https://gerrit-review.googlesource.com/", mock.HTTPClient)
		if err != nil {
			t.Errorf("An error occured. Expected nil. Got %+v.", err)
		}
		if reflect.DeepEqual(c.client, mock.ExpectedHTTPClient) == false {
			t.Errorf("Wrong HTTP Client. Expected %+v. Got %+v", mock.ExpectedHTTPClient, c.client)
		}
	}
}

func TestNewClient_Services(t *testing.T) {
	c, err := NewClient("https://gerrit-review.googlesource.com/", nil)
	if err != nil {
		t.Errorf("An error occured. Expected nil. Got %+v.", err)
	}

	if c.Access == nil {
		t.Error("No AccessService found.")
	}
	if c.Accounts == nil {
		t.Error("No AccountsService found.")
	}
	if c.Changes == nil {
		t.Error("No ChangesService found.")
	}
	if c.Config == nil {
		t.Error("No ConfigService found.")
	}
	if c.Groups == nil {
		t.Error("No GroupsService found.")
	}
	if c.Plugins == nil {
		t.Error("No PluginsService found.")
	}
	if c.Projects == nil {
		t.Error("No ProjectsService found.")
	}
}

func TestNewRequest(t *testing.T) {
	c, err := NewClient(TestGerritInstanceURL, nil)
	if err != nil {
		t.Errorf("An error occured. Expected nil. Got %+v.", err)
	}

	inURL, outURL := "/foo", TestGerritInstanceURL+"foo"
	inBody, outBody := &PermissionRuleInfo{Action: "ALLOW", Force: true, Min: 0, Max: 0}, `{"action":"ALLOW","force":true,"min":0,"max":0}`+"\n"
	req, _ := c.NewRequest("GET", inURL, inBody)

	// Test that relative URL was expanded
	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("NewRequest(%q) URL is %v, want %v", inURL, got, want)
	}

	// Test that body was JSON encoded
	body, _ := ioutil.ReadAll(req.Body)
	if got, want := string(body), outBody; got != want {
		t.Errorf("NewRequest(%q) Body is %v, want %v", inBody, got, want)
	}
}

func TestNewRequest_InvalidJSON(t *testing.T) {
	c, err := NewClient(TestGerritInstanceURL, nil)
	if err != nil {
		t.Errorf("An error occured. Expected nil. Got %+v.", err)
	}

	type T struct {
		A map[int]interface{}
	}
	_, err = c.NewRequest("GET", "/", &T{})

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("Expected a JSON error; got %#v.", err)
	}
}

func testURLParseError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

func TestNewRequest_BadURL(t *testing.T) {
	c, err := NewClient(TestGerritInstanceURL, nil)
	if err != nil {
		t.Errorf("An error occured. Expected nil. Got %+v.", err)
	}
	_, err = c.NewRequest("GET", ":", nil)
	testURLParseError(t, err)
}

// If a nil body is passed to diffy.NewRequest, make sure that nil is also passed to http.NewRequest.
// In most cases, passing an io.Reader that returns no content is fine,
// since there is no difference between an HTTP request body that is an empty string versus one that is not set at all.
// However in certain cases, intermediate systems may treat these differently resulting in subtle errors.
func TestNewRequest_EmptyBody(t *testing.T) {
	c, err := NewClient(TestGerritInstanceURL, nil)
	if err != nil {
		t.Errorf("An error occured. Expected nil. Got %+v.", err)
	}
	req, err := c.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("NewRequest returned unexpected error: %v", err)
	}
	if req.Body != nil {
		t.Fatalf("constructed request contains a non-nil Body")
	}
}
