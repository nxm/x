package kuchniaviking

import (
	"git.jakub.app/jakub/X/internal/env"
	"net/http"
	"time"
)

var (
	BASE_URL        = "https://panel.kuchniavikinga.pl"
	VIKING_LOGIN    = env.GetEnv("VIKING_LOGIN", "")
	VIKING_PASSWORD = env.GetEnv("VIKING_PASSWORD", "")
)

type KuchniaVikinga interface {
	GetActiveIds() ([]int, error)
	GetOrderData(orderId int) (*GetOrderDataResponse, error)
	GetDeliveryInfo(deliveryId int) (*DeliveryMenuResponse, error)
	GetNearestDeliveries(deliveries []Delivery, limit int) ([]Delivery, error)
}

type kuchniaViking struct {
	httpClient *http.Client
	baseUrl    string
	login      string
	password   string
}

type authTransport struct {
	cookies    []*http.Cookie
	underlying http.RoundTripper
}

func (t *authTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	for _, c := range t.cookies {
		if c != nil {
			r.AddCookie(c)
		}
	}
	return t.underlying.RoundTrip(r)
}

func New() (KuchniaVikinga, error) {
	t := &authTransport{}
	err := t.authLogin(BASE_URL, VIKING_LOGIN, VIKING_PASSWORD)
	if err != nil {
		return nil, err
	}
	t.underlying = http.DefaultTransport

	kv := &kuchniaViking{
		httpClient: &http.Client{
			Transport: t,
			Timeout:   30 * time.Second,
		},
		baseUrl:  BASE_URL,
		login:    VIKING_LOGIN,
		password: VIKING_PASSWORD,
	}

	return kv, nil
}
