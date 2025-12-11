package database

import (
	"fmt"
	"log"

	"github.com/mm-api/mm-api/models"
	"gorm.io/gorm"
)

// createDefaultSubscriptionPlans создает планы подписки по умолчанию
func createDefaultSubscriptionPlans() error {
	plans := []models.SubscriptionPlan{
		{
			Name:             "Месячная подписка",
			Description:      "Доступ к приложению на 1 месяц",
			SubscriptionType: models.SubscriptionTypeMonthly,
			Price:            29.99,
			Currency:         "USD",
			DurationMonths:   1,
			IsActive:         true,
			Features:         `{"products": true, "orders": true, "analytics": true}`,
			SortOrder:        1,
		},
		{
			Name:             "Годовая подписка",
			Description:      "Доступ к приложению на 1 год (экономия 20%)",
			SubscriptionType: models.SubscriptionTypeYearly,
			Price:            299.99,
			Currency:         "USD",
			DurationMonths:   12,
			IsActive:         true,
			Features:         `{"products": true, "orders": true, "analytics": true, "priority_support": true}`,
			SortOrder:        2,
		},
		{
			Name:             "Пожизненная подписка",
			Description:      "Пожизненный доступ к приложению",
			SubscriptionType: models.SubscriptionTypeLifetime,
			Price:            999.99,
			Currency:         "USD",
			DurationMonths:   0,
			IsActive:         true,
			Features:         `{"products": true, "orders": true, "analytics": true, "priority_support": true, "lifetime_updates": true}`,
			SortOrder:        3,
		},
	}

	for _, plan := range plans {
		var existingPlan models.SubscriptionPlan
		if err := DB.Where("name = ? AND subscription_type = ?", plan.Name, plan.SubscriptionType).First(&existingPlan).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := DB.Create(&plan).Error; err != nil {
					return fmt.Errorf("failed to create subscription plan %s: %w", plan.Name, err)
				}
				log.Printf("✅ Subscription plan created: %s", plan.Name)
			} else {
				return fmt.Errorf("failed to check subscription plan %s: %w", plan.Name, err)
			}
		} else {
			log.Printf("✅ Subscription plan already exists: %s", plan.Name)
		}
	}

	return nil
}

