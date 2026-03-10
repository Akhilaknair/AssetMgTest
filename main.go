package main

import (
	"AssetManagement/db"
	"AssetManagement/handlers"
	"AssetManagement/middleware"
	"AssetManagement/models"
	"AssetManagement/utils"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {

	if err := db.CreateAndMigrate(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),

		db.SSLModeDisabled); err != nil {
		log.Fatal(err)
	}

	log.Println("Database connected")

	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		utils.RespondJSON(w, http.StatusOK, map[string]string{
			"message": "server is running"})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", handlers.RegisterUser)
		r.Post("/login", handlers.LoginUser)
		r.With(middleware.AuthMiddleware).Group(func(r chi.Router) {
			r.Post("/logout", handlers.LogoutUser)
			r.Get("/profile", handlers.GetUser)
		})
	})

	r.Route("/users", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.With(middleware.RoleMiddleware(
			string(models.Admin),
			string(models.AssetManager),
		)).Group(func(r chi.Router) {
			r.Get("/", handlers.GetAllUsers)
			r.Put("/{user_id}/role", handlers.UpdateUserRole)
			r.Delete("/{user_id}", handlers.DeleteUser)
		})
	})

	r.Route("/assets", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Get("/", handlers.GetAssetsByEmpID)
		r.Get("/{asset_id}", handlers.GetAssetByID)

		r.With(middleware.RoleMiddleware(
			string(models.Admin),
			string(models.AssetManager),
		)).Group(func(r chi.Router) {
			r.Post("/", handlers.CreateAsset)
			r.Get("/dashboard", handlers.GetAssetsDashboard)
			r.Put("/{asset_id}", handlers.UpdateAsset)
			r.Delete("/{asset_id}", handlers.DeleteAsset)
			r.Post("/{asset_id}/assign", handlers.AssignAsset)
			r.Post("/{asset_id}/return", handlers.ReturnAsset)
			r.Put("/{asset_id}/need-service", handlers.SentToService)
		})
	})
	fmt.Println("listening on port 8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Println(err)
	}
}

// change empid-> user_id
// specs of device to show i
// include - exclude toggle button
