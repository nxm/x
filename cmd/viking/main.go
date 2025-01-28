package main

import (
	"fmt"
	"git.jakub.app/jakub/X/cmd/layla/modules/discord"
	"git.jakub.app/jakub/X/internal/env"
	"git.jakub.app/jakub/X/internal/kuchniaviking"
	"github.com/rs/zerolog/log"
	"strings"
)

var (
	DISCORD_WEBHOOK_URL = env.GetEnv("DISCORD_WEBHOOK", "")
)

type svc struct {
	discordModule *discord.Discord
}

func run() error {
	var err error
	svc := svc{}
	svc.discordModule, err = discord.New()
	if err != nil {
		log.Error().Err(err).Msg("can't load discord module")
		return err
	}

	return nil
}

func main() {
	kv, err := kuchniaviking.New()
	if err != nil {
		log.Fatal().Err(err).Msg("can't initialize kuchnia vikinga")
	}

	ids, err := kv.GetActiveIds()
	if err != nil {
		log.Fatal().Err(err).Msg("can't get active ids")
	}

	if len(ids) == 0 {
		log.Fatal().Msg("you don't have active order!")
	}

	orderDataResp, err := kv.GetOrderData(ids[0])
	if err != nil {
		log.Fatal().Err(err).Int("orderId", ids[0]).Msg("can't get orderData")
	}
	nearestDeliveries, err := kv.GetNearestDeliveries(orderDataResp.Deliveries, 3)

	for _, nearestDelivery := range nearestDeliveries {
		deliveryInfo, err := kv.GetDeliveryInfo(nearestDelivery.DeliveryID)
		if err != nil {
			log.Error().Err(err).Int("deliveryId", nearestDelivery.DeliveryID).Msg("can't get delivery info")
			break
		}

		var allergyMeals []kuchniaviking.DeliveryMenuItem
		allergens := map[string]bool{
			"ryba":       true,
			"skorupiaki": true,
		}

		for _, meal := range deliveryInfo.DeliveryMenuMeal {
			fmt.Printf("- - - %s - - -\n", nearestDelivery.Date)
			fmt.Printf("mealName: %s\n", meal.MealName)
			fmt.Printf("menuMealName: %s\n", meal.MenuMealName)
			fmt.Printf("ingredients: %v\n", meal.Ingredients)

			for _, ingredient := range meal.Ingredients {
				lowerName := strings.ToLower(ingredient.Name)
				for allergen := range allergens {
					if strings.Contains(lowerName, allergen) {
						allergyMeals = append(allergyMeals, meal)
						break
					}
				}
			}
		}

		fmt.Printf("allergyMeals: %v\n", allergyMeals)
		fmt.Printf("- - - - - - - - - - - - - - - - - -\n")

		if len(allergyMeals) > 0 {
			var fields []discord.EmbedField
			for _, meal := range allergyMeals {
				fields = append(fields, discord.EmbedField{
					Name:   "Date",
					Value:  nearestDelivery.Date,
					Inline: false,
				})

				fields = append(fields, discord.EmbedField{
					Name:   "Meal",
					Value:  meal.MenuMealName,
					Inline: false,
				})

				var allergenIngredients []string
				for _, ingredient := range meal.Ingredients {
					lowerName := strings.ToLower(ingredient.Name)
					for allergen := range allergens {
						if strings.Contains(lowerName, allergen) {
							allergenIngredients = append(allergenIngredients, ingredient.Name)
						}
					}
				}

				fields = append(fields, discord.EmbedField{
					Name:   "Allergen Ingredients",
					Value:  strings.Join(allergenIngredients, "\n"),
					Inline: false,
				})
			}

			embed := discord.Embed{
				Title:       "⚠️ Allergen Alert",
				Description: fmt.Sprintf("Found %d meals containing allergens!", len(allergyMeals)),
				Color:       0xFF0000,
				Fields:      fields,
			}

			err := discord.SendMessageWithEmbed(DISCORD_WEBHOOK_URL, "", embed)
			if err != nil {
				log.Error().Err(err).Msg("failed to send Discord webhook")
			}
		}
	}

	if err := run(); err != nil {
		panic(err)
	}
}
