package seed

import (
	"context"

	"github.com/DB-Vincent/personal-finance/services/finance/models"
	"github.com/google/uuid"
)

type defaultCategory struct {
	Group    string
	Name     string
	IsIncome bool
}

var defaults = []defaultCategory{
	{"Income", "Salary", true},
	{"Income", "Freelance", true},
	{"Income", "Interest", true},
	{"Income", "Other Income", true},
	{"Housing", "Rent/Mortgage", false},
	{"Housing", "Utilities", false},
	{"Housing", "Insurance", false},
	{"Housing", "Maintenance", false},
	{"Food", "Groceries", false},
	{"Food", "Restaurants", false},
	{"Food", "Coffee", false},
	{"Transportation", "Fuel", false},
	{"Transportation", "Public Transit", false},
	{"Transportation", "Parking", false},
	{"Transportation", "Car Maintenance", false},
	{"Entertainment", "Subscriptions", false},
	{"Entertainment", "Hobbies", false},
	{"Entertainment", "Events", false},
	{"Shopping", "Clothing", false},
	{"Shopping", "Electronics", false},
	{"Shopping", "Gifts", false},
	{"Health", "Medical", false},
	{"Health", "Pharmacy", false},
	{"Health", "Fitness", false},
	{"Financial", "Savings", false},
	{"Financial", "Investments", false},
	{"Financial", "Loan Payments", false},
	{"Financial", "Fees", false},
	{"Other", "Miscellaneous", false},
}

func DefaultCategories(userID uuid.UUID) []models.Category {
	cats := make([]models.Category, len(defaults))
	for i, d := range defaults {
		cats[i] = models.Category{
			UserID:    userID,
			GroupName: d.Group,
			Name:      d.Name,
			IsIncome:  d.IsIncome,
		}
	}
	return cats
}

type CategoryCreator interface {
	Create(ctx context.Context, cat *models.Category) error
}

func CategoriesForUser(ctx context.Context, repo CategoryCreator, userID uuid.UUID) error {
	cats := DefaultCategories(userID)
	for i := range cats {
		if err := repo.Create(ctx, &cats[i]); err != nil {
			return err
		}
	}
	return nil
}
