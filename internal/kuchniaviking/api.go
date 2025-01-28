package kuchniaviking

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Delivery struct {
	DeliveryID     int            `json:"deliveryId"`
	Date           string         `json:"date"`
	HourPreference string         `json:"hourPreference"`
	DietCaloriesID int            `json:"dietCaloriesId"`
	AddressID      int            `json:"addressId"`
	PickupPointID  *int           `json:"pickupPointId"`
	DeliverySpot   string         `json:"deliverySpot"`
	Deleted        bool           `json:"deleted"`
	DeliveryMeals  []DeliveryMeal `json:"deliveryMeals"`
	SideOrders     []any          `json:"sideOrders"`
}

type DeliveryMeal struct {
	DeliveryMealID     int  `json:"deliveryMealId"`
	Amount             int  `json:"amount"`
	DietCaloriesMealID int  `json:"dietCaloriesMealId"`
	AddedByUser        bool `json:"addedByUser"`
	Deleted            bool `json:"deleted"`
}

type GetOrderDataResponse struct {
	Deliveries []Delivery `json:"deliveries"`
}

// delivery menu

type DeliveryMenuResponse struct {
	MenuVisible      string             `json:"menuVisible"`
	ShowNutrition    bool               `json:"showNutrition"`
	ShowIngredients  bool               `json:"showIngredients"`
	DeliveryMenuMeal []DeliveryMenuItem `json:"deliveryMenuMeal"`
}

type DeliveryMenuItem struct {
	DeliveryMealID        int          `json:"deliveryMealId"`
	Amount                int          `json:"amount"`
	MealName              string       `json:"mealName"`
	MealPriority          int          `json:"mealPriority"`
	MenuMealID            int          `json:"menuMealId"`
	MenuMealName          string       `json:"menuMealName"`
	Thermo                string       `json:"thermo"`
	DietCaloriesMealID    int          `json:"dietCaloriesMealId"`
	DietCaloriesID        int          `json:"dietCaloriesId"`
	Nutrition             Nutrition    `json:"nutrition"`
	Allergens             []string     `json:"allergens"`
	AllergensWithExcluded []any        `json:"allergensWithExcluded"`
	Ingredients           []Ingredient `json:"ingredients"`
	Review                any          `json:"review"`
	AddedByUser           bool         `json:"addedByUser"`
	Switchable            bool         `json:"switchable"`
	MealAddingSource      bool         `json:"mealAddingSource"`
	DeliveryMealSeen      string       `json:"deliveryMealSeen"`
	ReviewSummary         any          `json:"reviewSummary"`
}

type Nutrition struct {
	Weight              float64 `json:"weight"`
	Calories            float64 `json:"calories"`
	Fat                 float64 `json:"fat"`
	Protein             float64 `json:"protein"`
	Carbohydrate        float64 `json:"carbohydrate"`
	DietaryFiber        float64 `json:"dietaryFiber"`
	Sugar               float64 `json:"sugar"`
	Salt                float64 `json:"salt"`
	SaturatedFattyAcids float64 `json:"saturatedFattyAcids"`
	CaloriesText        string  `json:"caloriesText"`
}

type Ingredient struct {
	Name      string      `json:"name"`
	Major     bool        `json:"major"`
	Exclusion []Exclusion `json:"exclusion"`
}

type Exclusion struct {
	DietaryExclusionID int    `json:"dietaryExclusionId"`
	Name               string `json:"name"`
	Chosen             bool   `json:"chosen"`
}

func (t *authTransport) authLogin(baseUrl, login, password string) error {

	formData := url.Values{}
	formData.Set("username", login)
	formData.Set("password", password)
	req, err := http.NewRequest(
		"POST",
		baseUrl+"/api/auth/login",
		strings.NewReader(formData.Encode()),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	t.cookies = resp.Cookies()
	return nil
}

func (kv *kuchniaViking) GetActiveIds() ([]int, error) {
	resp, err := kv.httpClient.Get(kv.baseUrl + "/api/company/customer/order/active-ids")
	if err != nil {
		log.Error().Err(err).Msg("can't send request to get active ids")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("can't read response body")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var ids []int
	if err := json.Unmarshal(body, &ids); err != nil {
		log.Error().Err(err).Str("body", string(body)).Msg("can't decode response")
		return nil, err
	}

	return ids, nil
}

func (kv *kuchniaViking) GetOrderData(orderId int) (*GetOrderDataResponse, error) {
	resp, err := kv.httpClient.Get(fmt.Sprintf("%s/api/company/customer/order/%d", kv.baseUrl, orderId))
	if err != nil {
		log.Error().Err(err).Int("orderId", orderId).Msg("can't send request to get order data")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("can't read response body")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result GetOrderDataResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error().Err(err).Str("body", string(body)).Msg("can't decode response")
		return nil, err
	}

	return &result, nil
}

func (kv *kuchniaViking) GetDeliveryInfo(deliveryId int) (*DeliveryMenuResponse, error) {
	resp, err := kv.httpClient.Get(fmt.Sprintf("%s/api/company/general/menus/delivery/%d/new", kv.baseUrl, deliveryId))
	if err != nil {
		log.Error().Err(err).Int("deliveryId", deliveryId).Msg("can't send request to get delivery info")
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("can't read response body")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result DeliveryMenuResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error().Err(err).Str("body", string(body)).Msg("can't decode response")
		return nil, err
	}

	return &result, nil
}
