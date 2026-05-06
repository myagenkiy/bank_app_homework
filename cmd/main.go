package main

import (
	"bank-api/internal/handlers"
	"bank-api/internal/middleware"
	"bank-api/internal/repositories"
	"bank-api/internal/scheduler"
	"bank-api/internal/services"
	"bank-api/pkg/database"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Подключение к БД
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Инициализация репозиториев
	userRepo := repositories.NewUserRepository(db)
	accountRepo := repositories.NewAccountRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	cardRepo := repositories.NewCardRepository(db)
	creditRepo := repositories.NewCreditRepository(db)
	paymentScheduleRepo := repositories.NewPaymentScheduleRepository(db)

	// Инициализация сервисов
	transferService := services.NewTransferService(db, accountRepo, transactionRepo)
	cardService := services.NewCardService(cardRepo, accountRepo)
	creditService := services.NewCreditService(creditRepo, paymentScheduleRepo, accountRepo, transferService)
	analyticsService := services.NewAnalyticsService(accountRepo, transactionRepo, creditRepo, cardRepo)

	// Интеграционные сервисы
	cbrService := services.NewCBRService()
	emailService := services.NewEmailService()

	// Инициализация обработчиков
	authHandler := handlers.NewAuthHandler(userRepo)
	accountHandler := handlers.NewAccountHandler(accountRepo)
	transferHandler := handlers.NewTransferHandler(transferService)
	cardHandler := handlers.NewCardHandler(cardService)
	creditHandler := handlers.NewCreditHandler(creditService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	integrationHandler := handlers.NewIntegrationHandler(cbrService, emailService)

	// Инициализация шедулера
	sched := scheduler.NewScheduler(paymentScheduleRepo, creditRepo, accountRepo, transferService)

	// Запуск шедулера в отдельной горутине
	ctx, cancel := context.WithCancel(context.Background())
	go sched.Start(ctx)

	// Настройка роутера
	r := mux.NewRouter()

	// Публичные маршруты (без авторизации)
	r.HandleFunc("/health", healthCheck).Methods("GET")
	r.HandleFunc("/api/register", authHandler.Register).Methods("POST")
	r.HandleFunc("/api/login", authHandler.Login).Methods("POST")

	// Защищенные маршруты (с JWT)
	// Счета
	r.HandleFunc("/api/accounts", middleware.AuthMiddleware(accountHandler.CreateAccount)).Methods("POST")
	r.HandleFunc("/api/accounts", middleware.AuthMiddleware(accountHandler.GetAccounts)).Methods("GET")

	// Переводы и операции
	r.HandleFunc("/api/transfer", middleware.AuthMiddleware(transferHandler.Transfer)).Methods("POST")
	r.HandleFunc("/api/deposit", middleware.AuthMiddleware(transferHandler.Deposit)).Methods("POST")
	r.HandleFunc("/api/withdraw", middleware.AuthMiddleware(transferHandler.Withdraw)).Methods("POST")
	r.HandleFunc("/api/transactions/{accountId}", middleware.AuthMiddleware(transferHandler.GetTransactions)).Methods("GET")

	// Карты
	r.HandleFunc("/api/cards", middleware.AuthMiddleware(cardHandler.CreateCard)).Methods("POST")
	r.HandleFunc("/api/cards", middleware.AuthMiddleware(cardHandler.GetCards)).Methods("GET")

	// Кредиты
	r.HandleFunc("/api/credits/apply", middleware.AuthMiddleware(creditHandler.ApplyForCredit)).Methods("POST")
	r.HandleFunc("/api/credits/{creditId}/schedule", middleware.AuthMiddleware(creditHandler.GetCreditSchedule)).Methods("GET")

	// Аналитика
	r.HandleFunc("/api/analytics/dashboard", middleware.AuthMiddleware(analyticsHandler.GetDashboard)).Methods("GET")
	r.HandleFunc("/api/analytics/credit-burden", middleware.AuthMiddleware(analyticsHandler.GetCreditBurden)).Methods("GET")
	r.HandleFunc("/api/accounts/{accountId}/predict", middleware.AuthMiddleware(analyticsHandler.PredictBalance)).Methods("GET")

	// Интеграции (ЦБ РФ и Email)
	r.HandleFunc("/api/central-bank/rate", middleware.AuthMiddleware(integrationHandler.GetCentralBankRate)).Methods("GET")
	r.HandleFunc("/api/email/test", middleware.AuthMiddleware(integrationHandler.TestEmail)).Methods("POST")

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Available routes:")
	log.Printf("  GET  /health")
	log.Printf("  POST /api/register")
	log.Printf("  POST /api/login")
	log.Printf("  POST /api/accounts (JWT required)")
	log.Printf("  GET  /api/accounts (JWT required)")
	log.Printf("  POST /api/transfer (JWT required)")
	log.Printf("  POST /api/deposit (JWT required)")
	log.Printf("  POST /api/withdraw (JWT required)")
	log.Printf("  GET  /api/transactions/{accountId} (JWT required)")
	log.Printf("  POST /api/cards (JWT required)")
	log.Printf("  GET  /api/cards (JWT required)")
	log.Printf("  POST /api/credits/apply (JWT required)")
	log.Printf("  GET  /api/credits/{creditId}/schedule (JWT required)")
	log.Printf("  GET  /api/analytics/dashboard (JWT required)")
	log.Printf("  GET  /api/analytics/credit-burden (JWT required)")
	log.Printf("  GET  /api/accounts/{accountId}/predict (JWT required)")
	log.Printf("  GET  /api/central-bank/rate (JWT required)")
	log.Printf("  POST /api/email/test (JWT required)")

	// Graceful shutdown
	go func() {
		log.Printf("🚀 Server started on port %s", port)
		if err := http.ListenAndServe(":"+port, r); err != nil {
			log.Fatal(err)
		}
	}()

	// Ожидаем сигнал прерывания
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	cancel()
	log.Println("Server stopped")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","message":"Bank API is running"}`))
}
