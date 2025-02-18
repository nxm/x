package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"git.jakub.app/jakub/X/cmd/layla/modules/discord"
	"git.jakub.app/jakub/X/internal/env"
	"git.jakub.app/jakub/X/internal/kuchniaviking"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

var (
	DISCORD_WEBHOOK_URL = env.GetEnv("DISCORD_WEBHOOK", "")
)

type Server struct {
	router        *mux.Router
	discordModule *discord.Discord
	kvService     kuchniaviking.KuchniaVikinga
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type DeliveryResponse struct {
	Date         string                           `json:"date"`
	DeliveryID   int                              `json:"deliveryId"`
	Meals        []kuchniaviking.DeliveryMenuItem `json:"meals"`
	AllergyMeals []kuchniaviking.DeliveryMenuItem `json:"allergyMeals"`
}

type HTMLDeliveryData struct {
	Date       string
	DeliveryID int
	Meals      []MealData
}

type MealData struct {
	MealName     string
	MenuMealName string
	Nutrition    kuchniaviking.Nutrition
	Ingredients  []IngredientData
	Allergens    []string
}

type IngredientData struct {
	Name  string
	Major bool
}

func NewServer() (*Server, error) {
	kvService, err := kuchniaviking.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kuchnia viking: %w", err)
	}

	discordModule, err := discord.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize discord module: %w", err)
	}

	server := &Server{
		router:        mux.NewRouter(),
		discordModule: discordModule,
		kvService:     kvService,
	}

	server.setupRoutes()
	return server, nil
}

func (s *Server) setupRoutes() {
	s.router.HandleFunc("/api/deliveries", s.GetDeliveriesHandler).Methods("GET")
	s.router.HandleFunc("/api/deliveries/html", s.GetMenuHTMLHandler).Methods("GET")
}

func (s *Server) GetDeliveriesHandler(w http.ResponseWriter, r *http.Request) {
	ids, err := s.kvService.GetActiveIds()
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Failed to get active orders")
		return
	}

	if len(ids) == 0 {
		s.respondWithError(w, http.StatusNotFound, "No active orders found")
		return
	}

	orderDataResp, err := s.kvService.GetOrderData(ids[0])
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Failed to get order data")
		return
	}

	nearestDeliveries, err := s.kvService.GetNearestDeliveries(orderDataResp.Deliveries, 7)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Failed to get nearest deliveries")
		return
	}

	var response []DeliveryResponse
	for _, delivery := range nearestDeliveries {
		deliveryInfo, err := s.kvService.GetDeliveryInfo(delivery.DeliveryID)
		if err != nil {
			log.Error().Err(err).Int("deliveryId", delivery.DeliveryID).Msg("Failed to get delivery info")
			continue
		}

		response = append(response, DeliveryResponse{
			Date:         delivery.Date,
			DeliveryID:   delivery.DeliveryID,
			Meals:        deliveryInfo.DeliveryMenuMeal,
		})
	}

	s.respondWithJSON(w, http.StatusOK, response)
}

func (s *Server) GetMenuHTMLHandler(w http.ResponseWriter, r *http.Request) {
	ids, err := s.kvService.GetActiveIds()
	if err != nil {
		http.Error(w, "Failed to get active orders", http.StatusInternalServerError)
		return
	}

	if len(ids) == 0 {
		http.Error(w, "No active orders found", http.StatusNotFound)
		return
	}

	orderDataResp, err := s.kvService.GetOrderData(ids[0])
	if err != nil {
		http.Error(w, "Failed to get order data", http.StatusInternalServerError)
		return
	}

	nearestDeliveries, err := s.kvService.GetNearestDeliveries(orderDataResp.Deliveries, 7)
	if err != nil {
		http.Error(w, "Failed to get nearest deliveries", http.StatusInternalServerError)
		return
	}

	var deliveriesData []HTMLDeliveryData
	for _, delivery := range nearestDeliveries {
		deliveryInfo, err := s.kvService.GetDeliveryInfo(delivery.DeliveryID)
		if err != nil {
			log.Error().Err(err).Int("deliveryId", delivery.DeliveryID).Msg("Failed to get delivery info")
			continue
		}

		meals := make([]MealData, len(deliveryInfo.DeliveryMenuMeal))
		for i, meal := range deliveryInfo.DeliveryMenuMeal {
			var majorIngredients []IngredientData
			for _, ing := range meal.Ingredients {
				if ing.Major {
					majorIngredients = append(majorIngredients, IngredientData{
						Name:  ing.Name,
						Major: true,
					})
				}
			}

			meals[i] = MealData{
				MealName:     meal.MealName,
				MenuMealName: meal.MenuMealName,
				Nutrition:    meal.Nutrition,
				Ingredients:  majorIngredients,
				Allergens:    meal.Allergens,
			}
		}

		deliveriesData = append(deliveriesData, HTMLDeliveryData{
			Date:       delivery.Date,
			DeliveryID: delivery.DeliveryID,
			Meals:      meals,
		})
	}

	tmpl := template.Must(template.New("menu").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Menu Overview</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        table {
            border-collapse: collapse;
            width: 100%;
            background-color: white;
            box-shadow: 0 1px 3px rgba(0,0,0,0.2);
        }
        th, td {
            border: 1px solid #ddd;
            padding: 12px 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
            font-weight: bold;
        }
        tr:nth-child(even) {
            background-color: #f9f9f9;
        }
        .ingredients {
            font-size: 0.9em;
            color: #444;
        }
        .ingredients::before {
            content: "Major: ";
            font-weight: bold;
            color: #666;
        }
        .date-cell {
            font-weight: bold;
            background-color: #e9ecef;
        }
        .nutrition {
            font-size: 0.9em;
            color: #666;
        }
        .allergens {
            color: #dc3545;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <table border="1" cellpadding="8" cellspacing="0">
        <thead>
            <tr>
                <th>Date</th>
                <th>Meal Type</th>
                <th>Menu Item</th>
                <th>Nutrition</th>
                <th>Ingredients</th>
                <th>Allergens</th>
            </tr>
        </thead>
        <tbody>
            {{range .}}
                {{$date := .Date}}
                {{$mealCount := len .Meals}}
                {{range $i, $meal := .Meals}}
                    <tr>
                        {{if eq $i 0}}
                            <td rowspan="{{$mealCount}}" class="date-cell">{{$date}}</td>
                        {{end}}
                        <td>{{$meal.MealName}}</td>
                        <td>{{$meal.MenuMealName}}</td>
                        <td class="nutrition">
                            Calories: {{.Nutrition.Calories}} kcal<br>
                            Protein: {{printf "%.2f" .Nutrition.Protein}}g<br>
                            Fat: {{printf "%.2f" .Nutrition.Fat}}g<br>
                            Carbs: {{printf "%.2f" .Nutrition.Carbohydrate}}g
                        </td>
                        <td class="ingredients">
                            {{range $j, $ing := .Ingredients}}
                                {{if $j}}, {{end}}{{$ing.Name}}
                            {{end}}
                        </td>
                        <td class="allergens">
                            {{range $j, $allergen := .Allergens}}
                                {{if $j}}, {{end}}{{$allergen}}
                            {{end}}
                        </td>
                    </tr>
                {{end}}
            {{end}}
        </tbody>
    </table>
</body>
</html>`))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, deliveriesData); err != nil {
		log.Error().Err(err).Msg("Failed to execute template")
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response := APIResponse{
		Success: code >= 200 && code < 300,
		Data:    payload,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func (s *Server) respondWithError(w http.ResponseWriter, code int, message string) {
	response := APIResponse{
		Success: false,
		Error:   message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create server")
	}

	port := env.GetEnv("PORT", "8080")
	log.Info().Msgf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, server.router); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

