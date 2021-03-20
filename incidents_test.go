package squadcast

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateIncident(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/services/5e8edb24668e003cb0b18ba1", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/services/payment-api-service.json")
	})
	mux.HandleFunc("/2f81ac8b2362990dd220f8bb4f7cd30ccc3dac43", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	client, err := NewTestClient(ts.URL)
	if err != nil {
		t.Error(err)
	}

	service, err := client.GetServiceByID(context.Background(), "5e8edb24668e003cb0b18ba1")
	if err != nil {
		t.Error(err)
	}

	t.Logf(service.APIKey)

	request := &PostIncidentRequest{
		Message:     "Foo",
		Description: "Bar",
		Status:      "trigger",
	}

	webhookClient := NewTestWebhookClient(ts.URL, service.APIKey)
	if err := webhookClient.CreateIncident(context.Background(), request); err != nil {
		t.Error(err)
	}
}
