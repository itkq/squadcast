package squadcast

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAllServices(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/services/services.json")
	})

	client, err := NewTestClient(ts.URL)
	if err != nil {
		t.Error(err)
	}

	services, err := client.GetAllServices(context.Background())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 2, len(services))
	assert.Equal(t, "Payment API Service", services[0].Name)
}

func TestGetServiceByName(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "Payment API Service" {
			http.ServeFile(w, r, "testdata/services/payment-api-service.json")
		}
	})

	client, err := NewTestClient(ts.URL)
	if err != nil {
		t.Error(err)
	}

	service, err := client.GetServiceByName(context.Background(), "Payment API Service")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "Payment API Service", service.Name)
}

func TestGetServiceByID(t *testing.T) {
	mux := http.NewServeMux()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	mux.HandleFunc("/services/5e8edb24668e003cb0b18ba1", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/services/payment-api-service.json")
	})

	client, err := NewTestClient(ts.URL)
	if err != nil {
		t.Error(err)
	}

	service, err := client.GetServiceByID(context.Background(), "5e8edb24668e003cb0b18ba1")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "5e8edb24668e003cb0b18ba1", service.ID)
}
