package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/repository"
	"tj-routes/internal/utils"

	"gorm.io/gorm"
)

// loadEnvFile reads .env file and sets environment variables
func loadEnvFile() {
	wd, _ := os.Getwd()
	envPaths := []string{
		filepath.Join(wd, ".env"),
		".env",
	}

	var envFile *os.File
	var err error
	for _, envPath := range envPaths {
		envFile, err = os.Open(envPath)
		if err == nil {
			break
		}
	}

	if envFile == nil {
		return
	}
	defer envFile.Close()

	scanner := bufio.NewScanner(envFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
}

func main() {
	// Load environment variables
	loadEnvFile()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := utils.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// Find admin user
	userRepo := repository.NewUserRepository(db)
	adminUser, err := findOrCreateAdmin(userRepo)
	if err != nil {
		log.Fatalf("Failed to find/create admin user: %v", err)
	}
	fmt.Printf("Using admin user: %s (ID: %d)\n", adminUser.Email, adminUser.ID)

	// Initialize repositories
	stopRepo := repository.NewStopRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	routeStopRepo := repository.NewRouteStopRepository(db)
	badgeRepo := repository.NewBadgeRepository(db)

	// Seed badges first
	if err := seedBadges(badgeRepo); err != nil {
		log.Fatalf("Failed to seed badges: %v", err)
	}

	// Seed data
	if err := seedData(db, stopRepo, routeRepo, routeStopRepo, adminUser.ID); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	fmt.Println("✅ Database seeding completed successfully!")
}

func findOrCreateAdmin(userRepo repository.UserRepository) (*models.User, error) {
	// Try to find admin by email
	admin, err := userRepo.FindByEmail("admin@transjakarta.go.id")
	if err == nil {
		if admin.Role == models.RoleAdmin {
			return admin, nil
		}
		// Update to admin if not already
		admin.Role = models.RoleAdmin
		if err := userRepo.Update(admin); err != nil {
			return nil, fmt.Errorf("failed to update user to admin: %w", err)
		}
		return admin, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to find admin: %w", err)
	}

	// Try to find any admin user
	users, _, err := userRepo.List(0, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range users {
		if user.Role == models.RoleAdmin {
			return &user, nil
		}
	}

	// Create admin user if none exists
	admin = &models.User{
		Email:    "admin@transjakarta.go.id",
		Username: "admin",
		Role:     models.RoleAdmin,
	}

	if err := userRepo.Create(admin); err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	return admin, nil
}

func seedBadges(badgeRepo repository.BadgeRepository) error {
	badges := []models.Badge{
		{
			Name:          "First Report",
			Description:   "Submitted your first report",
			Icon:          "🎯",
			CriteriaType:  models.BadgeCriteriaReportsAccepted,
			CriteriaValue: 1,
		},
		{
			Name:          "Helpful Contributor",
			Description:   "Submitted 5 accepted reports",
			Icon:          "⭐",
			CriteriaType:  models.BadgeCriteriaReportsAccepted,
			CriteriaValue: 5,
		},
		{
			Name:          "Dedicated Reporter",
			Description:   "Submitted 10 accepted reports",
			Icon:          "🏆",
			CriteriaType:  models.BadgeCriteriaReportsAccepted,
			CriteriaValue: 10,
		},
		{
			Name:          "Community Voice",
			Description:   "Made 10 comments",
			Icon:          "💬",
			CriteriaType:  models.BadgeCriteriaCommentsMade,
			CriteriaValue: 10,
		},
		{
			Name:          "Active Commenter",
			Description:   "Made 50 comments",
			Icon:          "🗣️",
			CriteriaType:  models.BadgeCriteriaCommentsMade,
			CriteriaValue: 50,
		},
		{
			Name:          "Popular Contributor",
			Description:   "Received 25 upvotes on your content",
			Icon:          "👍",
			CriteriaType:  models.BadgeCriteriaUpvotesReceived,
			CriteriaValue: 25,
		},
		{
			Name:          "Community Favorite",
			Description:   "Received 100 upvotes on your content",
			Icon:          "❤️",
			CriteriaType:  models.BadgeCriteriaUpvotesReceived,
			CriteriaValue: 100,
		},
		{
			Name:          "Rising Star",
			Description:   "Reached 50 reputation points",
			Icon:          "🌟",
			CriteriaType:  models.BadgeCriteriaReputationPoints,
			CriteriaValue: 50,
		},
		{
			Name:          "Trusted Member",
			Description:   "Reached 200 reputation points",
			Icon:          "✨",
			CriteriaType:  models.BadgeCriteriaReputationPoints,
			CriteriaValue: 200,
		},
		{
			Name:          "Expert Contributor",
			Description:   "Reached 500 reputation points",
			Icon:          "🎖️",
			CriteriaType:  models.BadgeCriteriaReputationPoints,
			CriteriaValue: 500,
		},
		{
			Name:          "Legend",
			Description:   "Reached 1000 reputation points",
			Icon:          "👑",
			CriteriaType:  models.BadgeCriteriaReputationPoints,
			CriteriaValue: 1000,
		},
	}

	fmt.Println("Seeding badges...")
	for _, badge := range badges {
		// Check if badge already exists
		existing, err := badgeRepo.FindByID(badge.ID)
		if err == nil && existing != nil {
			fmt.Printf("  ⊙ Badge already exists: %s\n", badge.Name)
			continue
		}

		// Try to find by name
		allBadges, _ := badgeRepo.FindAll()
		exists := false
		for _, b := range allBadges {
			if b.Name == badge.Name {
				exists = true
				break
			}
		}

		if exists {
			fmt.Printf("  ⊙ Badge already exists: %s\n", badge.Name)
			continue
		}

		if err := badgeRepo.Create(&badge); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to create badge %s: %v\n", badge.Name, err)
			continue
		}

		fmt.Printf("  ✓ Created badge: %s\n", badge.Name)
	}

	return nil
}

func seedData(
	db *gorm.DB,
	stopRepo repository.StopRepository,
	routeRepo repository.RouteRepository,
	routeStopRepo repository.RouteStopRepository,
	adminID uint,
) error {
	// Create stops map for easy lookup
	stopsMap := make(map[string]*models.Stop)

	// Define all stops with approximate Jakarta coordinates
	stops := []struct {
		name      string
		stopType  models.StopType
		latitude  float64
		longitude float64
		city      string
		district  string
	}{
		// Terminals
		{"Kota", models.StopTypeTerminal, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Blok M", models.StopTypeTerminal, -6.2442, 106.7996, "Jakarta Selatan", "Kebayoran Baru"},
		{"Pulo Gadung", models.StopTypeTerminal, -6.1878, 106.9078, "Jakarta Timur", "Pulo Gadung"},
		{"Monumen Nasional", models.StopTypeTerminal, -6.1751, 106.8270, "Jakarta Pusat", "Gambir"},
		{"Kalideres", models.StopTypeTerminal, -6.1586, 106.7156, "Jakarta Barat", "Kalideres"},
		{"Galunggung", models.StopTypeTerminal, -6.2000, 106.8500, "Jakarta Pusat", "Menteng"},
		{"Ancol", models.StopTypeTerminal, -6.1256, 106.8303, "Jakarta Utara", "Ancol"},
		{"Kampung Melayu", models.StopTypeTerminal, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Ragunan", models.StopTypeTerminal, -6.3011, 106.8203, "Jakarta Selatan", "Pasar Minggu"},
		{"Kampung Rambutan", models.StopTypeTerminal, -6.3200, 106.9000, "Jakarta Timur", "Ciracas"},
		{"Lebak Bulus", models.StopTypeTerminal, -6.2894, 106.7831, "Jakarta Selatan", "Cilandak"},
		{"Pasar Baru", models.StopTypeTerminal, -6.1697, 106.8311, "Jakarta Pusat", "Sawah Besar"},
		{"Pinang Ranti", models.StopTypeTerminal, -6.3300, 106.8800, "Jakarta Timur", "Makasar"},
		{"Pluit", models.StopTypeTerminal, -6.1156, 106.7881, "Jakarta Utara", "Penjaringan"},
		{"PGC", models.StopTypeTerminal, -6.2800, 106.8900, "Jakarta Timur", "Cakung"},
		{"Tanjung Priok", models.StopTypeTerminal, -6.1381, 106.8806, "Jakarta Utara", "Tanjung Priok"},
		{"Pulo Gebang", models.StopTypeTerminal, -6.2000, 106.9500, "Jakarta Timur", "Cakung"},
		{"Ciledug", models.StopTypeTerminal, -6.2400, 106.7500, "Jakarta Selatan", "Ciledug"},
		{"Tegal Mampang", models.StopTypeTerminal, -6.2500, 106.8000, "Jakarta Selatan", "Mampang"},
		{"Pasar Senen", models.StopTypeTerminal, -6.1756, 106.8411, "Jakarta Pusat", "Senen"},
		{"Jakarta Int'l Stadium", models.StopTypeTerminal, -6.2000, 106.8800, "Jakarta Utara", "Papanggo"},
		{"Dukuh Atas", models.StopTypeStop, -6.2081, 106.8203, "Jakarta Selatan", "Setiabudi"},

		// Additional major stops (simplified - you can add more based on the map)
		{"Harmoni", models.StopTypeStop, -6.1697, 106.8203, "Jakarta Pusat", "Gambir"},
		{"Bundaran HI", models.StopTypeStop, -6.1944, 106.8228, "Jakarta Pusat", "Menteng"},
		{"Senayan", models.StopTypeStop, -6.2278, 106.8003, "Jakarta Selatan", "Kebayoran Baru"},
		{"Cawang", models.StopTypeStop, -6.2500, 106.8700, "Jakarta Timur", "Cakung"},
		{"Cikoko", models.StopTypeStop, -6.2600, 106.8600, "Jakarta Timur", "Cakung"},
		{"Cawang UKI", models.StopTypeStop, -6.2400, 106.8700, "Jakarta Timur", "Cakung"},
		{"BKN", models.StopTypeStop, -6.2700, 106.8500, "Jakarta Timur", "Cakung"},
		{"Tugu Tani", models.StopTypeStop, -6.1811, 106.8361, "Jakarta Pusat", "Menteng"},
		{"Gambir", models.StopTypeStop, -6.1756, 106.8303, "Jakarta Pusat", "Gambir"},
		{"Sawah Besar", models.StopTypeStop, -6.1697, 106.8311, "Jakarta Pusat", "Sawah Besar"},
		{"Mangga Dua", models.StopTypeStop, -6.1400, 106.8300, "Jakarta Utara", "Sawah Besar"},
		{"Gunung Sahari", models.StopTypeStop, -6.1500, 106.8400, "Jakarta Pusat", "Sawah Besar"},
		{"Jembatan Merah", models.StopTypeStop, -6.1300, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Pecenongan", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Kwitang", models.StopTypeStop, -6.1800, 106.8400, "Jakarta Pusat", "Senen"},
		{"Kramat", models.StopTypeStop, -6.1900, 106.8500, "Jakarta Pusat", "Senen"},
		{"Pasar Rebo", models.StopTypeStop, -6.3100, 106.8700, "Jakarta Timur", "Ciracas"},
		{"Cililitan", models.StopTypeStop, -6.2600, 106.8700, "Jakarta Timur", "Kramat Jati"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Cikini", models.StopTypeStop, -6.1900, 106.8400, "Jakarta Pusat", "Menteng"},
		{"Menteng", models.StopTypeStop, -6.1950, 106.8350, "Jakarta Pusat", "Menteng"},
		{"Cikini", models.StopTypeStop, -6.1900, 106.8400, "Jakarta Pusat", "Menteng"},
		{"Kebon Jeruk", models.StopTypeStop, -6.2000, 106.7800, "Jakarta Barat", "Kebon Jeruk"},
		{"Tomang", models.StopTypeStop, -6.1800, 106.8000, "Jakarta Barat", "Grogol Petamburan"},
		{"Grogol", models.StopTypeStop, -6.1700, 106.7900, "Jakarta Barat", "Grogol Petamburan"},
		{"Slipi", models.StopTypeStop, -6.2000, 106.8000, "Jakarta Barat", "Palmerah"},
		{"Palmerah", models.StopTypeStop, -6.2100, 106.8000, "Jakarta Selatan", "Palmerah"},
		{"Tosari", models.StopTypeStop, -6.1950, 106.8100, "Jakarta Pusat", "Menteng"},
		{"Dukuh Atas 2", models.StopTypeStop, -6.2081, 106.8203, "Jakarta Selatan", "Setiabudi"},
		{"Setiabudi", models.StopTypeStop, -6.2100, 106.8200, "Jakarta Selatan", "Setiabudi"},
		{"Kuningan", models.StopTypeStop, -6.2200, 106.8300, "Jakarta Selatan", "Setiabudi"},
		{"Fatmawati", models.StopTypeStop, -6.2800, 106.8000, "Jakarta Selatan", "Cilandak"},
		{"Pondok Indah", models.StopTypeStop, -6.2700, 106.7800, "Jakarta Selatan", "Kebayoran Lama"},
		{"Blok A", models.StopTypeStop, -6.2450, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Blok M BCA", models.StopTypeStop, -6.2442, 106.7996, "Jakarta Selatan", "Kebayoran Baru"},
		{"Pasaraya", models.StopTypeStop, -6.2430, 106.7990, "Jakarta Selatan", "Kebayoran Baru"},
		{"ASEAN", models.StopTypeStop, -6.2300, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Senopati", models.StopTypeStop, -6.2250, 106.8050, "Jakarta Selatan", "Kebayoran Baru"},
		{"Sudirman", models.StopTypeStop, -6.2081, 106.8203, "Jakarta Selatan", "Setiabudi"},
		{"Thamrin", models.StopTypeStop, -6.1950, 106.8250, "Jakarta Pusat", "Menteng"},
		{"Sarinah", models.StopTypeStop, -6.1900, 106.8300, "Jakarta Pusat", "Menteng"},
		{"Bank Indonesia", models.StopTypeStop, -6.1800, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Stasiun Jakarta Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Glodok", models.StopTypeStop, -6.1400, 106.8150, "Jakarta Utara", "Jakarta Barat"},
		{"Mangga Dua Selatan", models.StopTypeStop, -6.1400, 106.8300, "Jakarta Utara", "Sawah Besar"},
		{"Ancol Barat", models.StopTypeStop, -6.1256, 106.8303, "Jakarta Utara", "Ancol"},
		{"Pademangan", models.StopTypeStop, -6.1300, 106.8400, "Jakarta Utara", "Pademangan"},
		{"Sunter", models.StopTypeStop, -6.1500, 106.8700, "Jakarta Utara", "Tanjung Priok"},
		{"Yos Sudarso", models.StopTypeStop, -6.1400, 106.8800, "Jakarta Utara", "Tanjung Priok"},
		{"Pluit Timur", models.StopTypeStop, -6.1156, 106.7881, "Jakarta Utara", "Penjaringan"},
		{"Muara Karang", models.StopTypeStop, -6.1100, 106.7900, "Jakarta Utara", "Penjaringan"},
		{"Penjaringan", models.StopTypeStop, -6.1200, 106.7850, "Jakarta Utara", "Penjaringan"},
		{"Kemayoran", models.StopTypeStop, -6.1600, 106.8500, "Jakarta Pusat", "Kemayoran"},
		{"Rawamangun", models.StopTypeStop, -6.1900, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pulomas", models.StopTypeStop, -6.1850, 106.9050, "Jakarta Timur", "Pulo Gadung"},
		{"Kayu Putih", models.StopTypeStop, -6.1800, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pramuka BPKP", models.StopTypeStop, -6.1750, 106.8700, "Jakarta Timur", "Matraman"},
		{"Matraman", models.StopTypeStop, -6.1700, 106.8600, "Jakarta Timur", "Matraman"},
		{"Pegangsaan", models.StopTypeStop, -6.1900, 106.8400, "Jakarta Pusat", "Menteng"},
		{"Cikini", models.StopTypeStop, -6.1900, 106.8400, "Jakarta Pusat", "Menteng"},
		{"Gondangdia", models.StopTypeStop, -6.1850, 106.8350, "Jakarta Pusat", "Menteng"},
		{"Juanda", models.StopTypeStop, -6.1800, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Istiqlal", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Lapangan Banteng", models.StopTypeStop, -6.1750, 106.8350, "Jakarta Pusat", "Gambir"},
		{"Pecenongan", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Harmoni", models.StopTypeStop, -6.1697, 106.8203, "Jakarta Pusat", "Gambir"},
		{"Gajah Mada", models.StopTypeStop, -6.1600, 106.8200, "Jakarta Pusat", "Gambir"},
		{"Hayam Wuruk", models.StopTypeStop, -6.1550, 106.8150, "Jakarta Pusat", "Gambir"},
		{"Mangga Besar", models.StopTypeStop, -6.1500, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Olahraga", models.StopTypeStop, -6.1450, 106.8180, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Kalideres", models.StopTypeStop, -6.1586, 106.7156, "Jakarta Barat", "Kalideres"},
		{"Rawa Buaya", models.StopTypeStop, -6.1600, 106.7300, "Jakarta Barat", "Kalideres"},
		{"Cengkareng", models.StopTypeStop, -6.1650, 106.7400, "Jakarta Barat", "Cengkareng"},
		{"Bojong Indah", models.StopTypeStop, -6.1700, 106.7500, "Jakarta Barat", "Cengkareng"},
		{"Puri Kembangan", models.StopTypeStop, -6.1900, 106.7600, "Jakarta Barat", "Kembangan"},
		{"Kembangan", models.StopTypeStop, -6.2000, 106.7700, "Jakarta Barat", "Kembangan"},
		{"Meruya", models.StopTypeStop, -6.2100, 106.7800, "Jakarta Barat", "Kembangan"},
		{"Kebon Jeruk", models.StopTypeStop, -6.2000, 106.7800, "Jakarta Barat", "Kebon Jeruk"},
		{"Tomang", models.StopTypeStop, -6.1800, 106.8000, "Jakarta Barat", "Grogol Petamburan"},
		{"Grogol", models.StopTypeStop, -6.1700, 106.7900, "Jakarta Barat", "Grogol Petamburan"},
		{"Slipi", models.StopTypeStop, -6.2000, 106.8000, "Jakarta Barat", "Palmerah"},
		{"Palmerah", models.StopTypeStop, -6.2100, 106.8000, "Jakarta Selatan", "Palmerah"},
		{"Tosari", models.StopTypeStop, -6.1950, 106.8100, "Jakarta Pusat", "Menteng"},
		{"Dukuh Atas", models.StopTypeStop, -6.2081, 106.8203, "Jakarta Selatan", "Setiabudi"},
		{"Setiabudi", models.StopTypeStop, -6.2100, 106.8200, "Jakarta Selatan", "Setiabudi"},
		{"Kuningan", models.StopTypeStop, -6.2200, 106.8300, "Jakarta Selatan", "Setiabudi"},
		{"Ragunan", models.StopTypeStop, -6.3011, 106.8203, "Jakarta Selatan", "Pasar Minggu"},
		{"Pasar Minggu", models.StopTypeStop, -6.2900, 106.8300, "Jakarta Selatan", "Pasar Minggu"},
		{"Tanjung Barat", models.StopTypeStop, -6.2800, 106.8400, "Jakarta Selatan", "Pasar Minggu"},
		{"Lenteng Agung", models.StopTypeStop, -6.2700, 106.8500, "Jakarta Selatan", "Jagakarsa"},
		{"Universitas Indonesia", models.StopTypeStop, -6.3600, 106.8300, "Depok", "Beji"},
		{"Pondok Cabe", models.StopTypeStop, -6.3400, 106.7700, "Tangerang Selatan", "Pondok Aren"},
		{"Lebak Bulus", models.StopTypeStop, -6.2894, 106.7831, "Jakarta Selatan", "Cilandak"},
		{"Fatmawati", models.StopTypeStop, -6.2800, 106.8000, "Jakarta Selatan", "Cilandak"},
		{"Cipete Raya", models.StopTypeStop, -6.2700, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Haji Nawi", models.StopTypeStop, -6.2600, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Blok A", models.StopTypeStop, -6.2450, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Blok M", models.StopTypeStop, -6.2442, 106.7996, "Jakarta Selatan", "Kebayoran Baru"},
		{"Pasaraya", models.StopTypeStop, -6.2430, 106.7990, "Jakarta Selatan", "Kebayoran Baru"},
		{"ASEAN", models.StopTypeStop, -6.2300, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Senopati", models.StopTypeStop, -6.2250, 106.8050, "Jakarta Selatan", "Kebayoran Baru"},
		{"Sudirman", models.StopTypeStop, -6.2081, 106.8203, "Jakarta Selatan", "Setiabudi"},
		{"Thamrin", models.StopTypeStop, -6.1950, 106.8250, "Jakarta Pusat", "Menteng"},
		{"Sarinah", models.StopTypeStop, -6.1900, 106.8300, "Jakarta Pusat", "Menteng"},
		{"Bank Indonesia", models.StopTypeStop, -6.1800, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Harmoni", models.StopTypeStop, -6.1697, 106.8203, "Jakarta Pusat", "Gambir"},
		{"Gajah Mada", models.StopTypeStop, -6.1600, 106.8200, "Jakarta Pusat", "Gambir"},
		{"Hayam Wuruk", models.StopTypeStop, -6.1550, 106.8150, "Jakarta Pusat", "Gambir"},
		{"Mangga Besar", models.StopTypeStop, -6.1500, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Olahraga", models.StopTypeStop, -6.1450, 106.8180, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Kampung Rambutan", models.StopTypeStop, -6.3200, 106.9000, "Jakarta Timur", "Ciracas"},
		{"Pasar Rebo", models.StopTypeStop, -6.3100, 106.8700, "Jakarta Timur", "Ciracas"},
		{"Cililitan", models.StopTypeStop, -6.2600, 106.8700, "Jakarta Timur", "Kramat Jati"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Cikini", models.StopTypeStop, -6.1900, 106.8400, "Jakarta Pusat", "Menteng"},
		{"Menteng", models.StopTypeStop, -6.1950, 106.8350, "Jakarta Pusat", "Menteng"},
		{"Gondangdia", models.StopTypeStop, -6.1850, 106.8350, "Jakarta Pusat", "Menteng"},
		{"Juanda", models.StopTypeStop, -6.1800, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Istiqlal", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Lapangan Banteng", models.StopTypeStop, -6.1750, 106.8350, "Jakarta Pusat", "Gambir"},
		{"Pecenongan", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Harmoni", models.StopTypeStop, -6.1697, 106.8203, "Jakarta Pusat", "Gambir"},
		{"Gajah Mada", models.StopTypeStop, -6.1600, 106.8200, "Jakarta Pusat", "Gambir"},
		{"Hayam Wuruk", models.StopTypeStop, -6.1550, 106.8150, "Jakarta Pusat", "Gambir"},
		{"Mangga Besar", models.StopTypeStop, -6.1500, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Olahraga", models.StopTypeStop, -6.1450, 106.8180, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Pulo Gadung", models.StopTypeStop, -6.1878, 106.9078, "Jakarta Timur", "Pulo Gadung"},
		{"Rawamangun", models.StopTypeStop, -6.1900, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pulomas", models.StopTypeStop, -6.1850, 106.9050, "Jakarta Timur", "Pulo Gadung"},
		{"Kayu Putih", models.StopTypeStop, -6.1800, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pramuka BPKP", models.StopTypeStop, -6.1750, 106.8700, "Jakarta Timur", "Matraman"},
		{"Matraman", models.StopTypeStop, -6.1700, 106.8600, "Jakarta Timur", "Matraman"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Galunggung", models.StopTypeStop, -6.2000, 106.8500, "Jakarta Pusat", "Menteng"},
		{"Cikini", models.StopTypeStop, -6.1900, 106.8400, "Jakarta Pusat", "Menteng"},
		{"Menteng", models.StopTypeStop, -6.1950, 106.8350, "Jakarta Pusat", "Menteng"},
		{"Gondangdia", models.StopTypeStop, -6.1850, 106.8350, "Jakarta Pusat", "Menteng"},
		{"Juanda", models.StopTypeStop, -6.1800, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Istiqlal", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Lapangan Banteng", models.StopTypeStop, -6.1750, 106.8350, "Jakarta Pusat", "Gambir"},
		{"Pecenongan", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Harmoni", models.StopTypeStop, -6.1697, 106.8203, "Jakarta Pusat", "Gambir"},
		{"Gajah Mada", models.StopTypeStop, -6.1600, 106.8200, "Jakarta Pusat", "Gambir"},
		{"Hayam Wuruk", models.StopTypeStop, -6.1550, 106.8150, "Jakarta Pusat", "Gambir"},
		{"Mangga Besar", models.StopTypeStop, -6.1500, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Olahraga", models.StopTypeStop, -6.1450, 106.8180, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Ancol", models.StopTypeStop, -6.1256, 106.8303, "Jakarta Utara", "Ancol"},
		{"Pademangan", models.StopTypeStop, -6.1300, 106.8400, "Jakarta Utara", "Pademangan"},
		{"Sunter", models.StopTypeStop, -6.1500, 106.8700, "Jakarta Utara", "Tanjung Priok"},
		{"Yos Sudarso", models.StopTypeStop, -6.1400, 106.8800, "Jakarta Utara", "Tanjung Priok"},
		{"Tanjung Priok", models.StopTypeStop, -6.1381, 106.8806, "Jakarta Utara", "Tanjung Priok"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Cawang UKI", models.StopTypeStop, -6.2400, 106.8700, "Jakarta Timur", "Cakung"},
		{"Cikoko", models.StopTypeStop, -6.2600, 106.8600, "Jakarta Timur", "Cakung"},
		{"Cawang", models.StopTypeStop, -6.2500, 106.8700, "Jakarta Timur", "Cakung"},
		{"BKN", models.StopTypeStop, -6.2700, 106.8500, "Jakarta Timur", "Cakung"},
		{"Tugu Tani", models.StopTypeStop, -6.1811, 106.8361, "Jakarta Pusat", "Menteng"},
		{"Gambir", models.StopTypeStop, -6.1756, 106.8303, "Jakarta Pusat", "Gambir"},
		{"Sawah Besar", models.StopTypeStop, -6.1697, 106.8311, "Jakarta Pusat", "Sawah Besar"},
		{"Mangga Dua", models.StopTypeStop, -6.1400, 106.8300, "Jakarta Utara", "Sawah Besar"},
		{"Gunung Sahari", models.StopTypeStop, -6.1500, 106.8400, "Jakarta Pusat", "Sawah Besar"},
		{"Jembatan Merah", models.StopTypeStop, -6.1300, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Pecenongan", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Kwitang", models.StopTypeStop, -6.1800, 106.8400, "Jakarta Pusat", "Senen"},
		{"Kramat", models.StopTypeStop, -6.1900, 106.8500, "Jakarta Pusat", "Senen"},
		{"Pasar Rebo", models.StopTypeStop, -6.3100, 106.8700, "Jakarta Timur", "Ciracas"},
		{"Cililitan", models.StopTypeStop, -6.2600, 106.8700, "Jakarta Timur", "Kramat Jati"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Ragunan", models.StopTypeStop, -6.3011, 106.8203, "Jakarta Selatan", "Pasar Minggu"},
		{"Pasar Minggu", models.StopTypeStop, -6.2900, 106.8300, "Jakarta Selatan", "Pasar Minggu"},
		{"Tanjung Barat", models.StopTypeStop, -6.2800, 106.8400, "Jakarta Selatan", "Pasar Minggu"},
		{"Lenteng Agung", models.StopTypeStop, -6.2700, 106.8500, "Jakarta Selatan", "Jagakarsa"},
		{"Universitas Indonesia", models.StopTypeStop, -6.3600, 106.8300, "Depok", "Beji"},
		{"Pondok Cabe", models.StopTypeStop, -6.3400, 106.7700, "Tangerang Selatan", "Pondok Aren"},
		{"Lebak Bulus", models.StopTypeStop, -6.2894, 106.7831, "Jakarta Selatan", "Cilandak"},
		{"Fatmawati", models.StopTypeStop, -6.2800, 106.8000, "Jakarta Selatan", "Cilandak"},
		{"Cipete Raya", models.StopTypeStop, -6.2700, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Haji Nawi", models.StopTypeStop, -6.2600, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Blok A", models.StopTypeStop, -6.2450, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Blok M", models.StopTypeStop, -6.2442, 106.7996, "Jakarta Selatan", "Kebayoran Baru"},
		{"Pasaraya", models.StopTypeStop, -6.2430, 106.7990, "Jakarta Selatan", "Kebayoran Baru"},
		{"ASEAN", models.StopTypeStop, -6.2300, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Senopati", models.StopTypeStop, -6.2250, 106.8050, "Jakarta Selatan", "Kebayoran Baru"},
		{"Sudirman", models.StopTypeStop, -6.2081, 106.8203, "Jakarta Selatan", "Setiabudi"},
		{"Thamrin", models.StopTypeStop, -6.1950, 106.8250, "Jakarta Pusat", "Menteng"},
		{"Sarinah", models.StopTypeStop, -6.1900, 106.8300, "Jakarta Pusat", "Menteng"},
		{"Bank Indonesia", models.StopTypeStop, -6.1800, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Harmoni", models.StopTypeStop, -6.1697, 106.8203, "Jakarta Pusat", "Gambir"},
		{"Gajah Mada", models.StopTypeStop, -6.1600, 106.8200, "Jakarta Pusat", "Gambir"},
		{"Hayam Wuruk", models.StopTypeStop, -6.1550, 106.8150, "Jakarta Pusat", "Gambir"},
		{"Mangga Besar", models.StopTypeStop, -6.1500, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Olahraga", models.StopTypeStop, -6.1450, 106.8180, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Galunggung", models.StopTypeStop, -6.2000, 106.8500, "Jakarta Pusat", "Menteng"},
		{"Kampung Rambutan", models.StopTypeStop, -6.3200, 106.9000, "Jakarta Timur", "Ciracas"},
		{"Pasar Rebo", models.StopTypeStop, -6.3100, 106.8700, "Jakarta Timur", "Ciracas"},
		{"Cililitan", models.StopTypeStop, -6.2600, 106.8700, "Jakarta Timur", "Kramat Jati"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Lebak Bulus", models.StopTypeStop, -6.2894, 106.7831, "Jakarta Selatan", "Cilandak"},
		{"Fatmawati", models.StopTypeStop, -6.2800, 106.8000, "Jakarta Selatan", "Cilandak"},
		{"Cipete Raya", models.StopTypeStop, -6.2700, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Haji Nawi", models.StopTypeStop, -6.2600, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Blok A", models.StopTypeStop, -6.2450, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Blok M", models.StopTypeStop, -6.2442, 106.7996, "Jakarta Selatan", "Kebayoran Baru"},
		{"Pasaraya", models.StopTypeStop, -6.2430, 106.7990, "Jakarta Selatan", "Kebayoran Baru"},
		{"ASEAN", models.StopTypeStop, -6.2300, 106.8000, "Jakarta Selatan", "Kebayoran Baru"},
		{"Senopati", models.StopTypeStop, -6.2250, 106.8050, "Jakarta Selatan", "Kebayoran Baru"},
		{"Sudirman", models.StopTypeStop, -6.2081, 106.8203, "Jakarta Selatan", "Setiabudi"},
		{"Thamrin", models.StopTypeStop, -6.1950, 106.8250, "Jakarta Pusat", "Menteng"},
		{"Sarinah", models.StopTypeStop, -6.1900, 106.8300, "Jakarta Pusat", "Menteng"},
		{"Bank Indonesia", models.StopTypeStop, -6.1800, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Harmoni", models.StopTypeStop, -6.1697, 106.8203, "Jakarta Pusat", "Gambir"},
		{"Gajah Mada", models.StopTypeStop, -6.1600, 106.8200, "Jakarta Pusat", "Gambir"},
		{"Hayam Wuruk", models.StopTypeStop, -6.1550, 106.8150, "Jakarta Pusat", "Gambir"},
		{"Mangga Besar", models.StopTypeStop, -6.1500, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Olahraga", models.StopTypeStop, -6.1450, 106.8180, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Pasar Baru", models.StopTypeStop, -6.1697, 106.8311, "Jakarta Pusat", "Sawah Besar"},
		{"Pinang Ranti", models.StopTypeStop, -6.3300, 106.8800, "Jakarta Timur", "Makasar"},
		{"Cililitan", models.StopTypeStop, -6.2600, 106.8700, "Jakarta Timur", "Kramat Jati"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Cawang UKI", models.StopTypeStop, -6.2400, 106.8700, "Jakarta Timur", "Cakung"},
		{"Cikoko", models.StopTypeStop, -6.2600, 106.8600, "Jakarta Timur", "Cakung"},
		{"Cawang", models.StopTypeStop, -6.2500, 106.8700, "Jakarta Timur", "Cakung"},
		{"BKN", models.StopTypeStop, -6.2700, 106.8500, "Jakarta Timur", "Cakung"},
		{"Tugu Tani", models.StopTypeStop, -6.1811, 106.8361, "Jakarta Pusat", "Menteng"},
		{"Gambir", models.StopTypeStop, -6.1756, 106.8303, "Jakarta Pusat", "Gambir"},
		{"Sawah Besar", models.StopTypeStop, -6.1697, 106.8311, "Jakarta Pusat", "Sawah Besar"},
		{"Mangga Dua", models.StopTypeStop, -6.1400, 106.8300, "Jakarta Utara", "Sawah Besar"},
		{"Gunung Sahari", models.StopTypeStop, -6.1500, 106.8400, "Jakarta Pusat", "Sawah Besar"},
		{"Jembatan Merah", models.StopTypeStop, -6.1300, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Pecenongan", models.StopTypeStop, -6.1700, 106.8300, "Jakarta Pusat", "Gambir"},
		{"Kwitang", models.StopTypeStop, -6.1800, 106.8400, "Jakarta Pusat", "Senen"},
		{"Kramat", models.StopTypeStop, -6.1900, 106.8500, "Jakarta Pusat", "Senen"},
		{"Pasar Rebo", models.StopTypeStop, -6.3100, 106.8700, "Jakarta Timur", "Ciracas"},
		{"Cililitan", models.StopTypeStop, -6.2600, 106.8700, "Jakarta Timur", "Kramat Jati"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Pluit", models.StopTypeStop, -6.1156, 106.7881, "Jakarta Utara", "Penjaringan"},
		{"Muara Karang", models.StopTypeStop, -6.1100, 106.7900, "Jakarta Utara", "Penjaringan"},
		{"Penjaringan", models.StopTypeStop, -6.1200, 106.7850, "Jakarta Utara", "Penjaringan"},
		{"Kemayoran", models.StopTypeStop, -6.1600, 106.8500, "Jakarta Pusat", "Kemayoran"},
		{"PGC", models.StopTypeStop, -6.2800, 106.8900, "Jakarta Timur", "Cakung"},
		{"Cakung", models.StopTypeStop, -6.2700, 106.8900, "Jakarta Timur", "Cakung"},
		{"Rawamangun", models.StopTypeStop, -6.1900, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pulomas", models.StopTypeStop, -6.1850, 106.9050, "Jakarta Timur", "Pulo Gadung"},
		{"Kayu Putih", models.StopTypeStop, -6.1800, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pramuka BPKP", models.StopTypeStop, -6.1750, 106.8700, "Jakarta Timur", "Matraman"},
		{"Matraman", models.StopTypeStop, -6.1700, 106.8600, "Jakarta Timur", "Matraman"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Tanjung Priok", models.StopTypeStop, -6.1381, 106.8806, "Jakarta Utara", "Tanjung Priok"},
		{"Sunter", models.StopTypeStop, -6.1500, 106.8700, "Jakarta Utara", "Tanjung Priok"},
		{"Yos Sudarso", models.StopTypeStop, -6.1400, 106.8800, "Jakarta Utara", "Tanjung Priok"},
		{"Ancol", models.StopTypeStop, -6.1256, 106.8303, "Jakarta Utara", "Ancol"},
		{"Pademangan", models.StopTypeStop, -6.1300, 106.8400, "Jakarta Utara", "Pademangan"},
		{"Kemayoran", models.StopTypeStop, -6.1600, 106.8500, "Jakarta Pusat", "Kemayoran"},
		{"Gunung Sahari", models.StopTypeStop, -6.1500, 106.8400, "Jakarta Pusat", "Sawah Besar"},
		{"Mangga Dua", models.StopTypeStop, -6.1400, 106.8300, "Jakarta Utara", "Sawah Besar"},
		{"Jembatan Merah", models.StopTypeStop, -6.1300, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Pluit", models.StopTypeStop, -6.1156, 106.7881, "Jakarta Utara", "Penjaringan"},
		{"Pulo Gebang", models.StopTypeStop, -6.2000, 106.9500, "Jakarta Timur", "Cakung"},
		{"Cakung", models.StopTypeStop, -6.2700, 106.8900, "Jakarta Timur", "Cakung"},
		{"Rawamangun", models.StopTypeStop, -6.1900, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pulomas", models.StopTypeStop, -6.1850, 106.9050, "Jakarta Timur", "Pulo Gadung"},
		{"Kayu Putih", models.StopTypeStop, -6.1800, 106.9000, "Jakarta Timur", "Pulo Gadung"},
		{"Pramuka BPKP", models.StopTypeStop, -6.1750, 106.8700, "Jakarta Timur", "Matraman"},
		{"Matraman", models.StopTypeStop, -6.1700, 106.8600, "Jakarta Timur", "Matraman"},
		{"Bidara Cina", models.StopTypeStop, -6.2200, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Stasiun Jatinegara", models.StopTypeStop, -6.2300, 106.8600, "Jakarta Timur", "Jatinegara"},
		{"Kampung Melayu", models.StopTypeStop, -6.2306, 106.8611, "Jakarta Timur", "Jatinegara"},
		{"Tanjung Priok", models.StopTypeStop, -6.1381, 106.8806, "Jakarta Utara", "Tanjung Priok"},
		{"Sunter", models.StopTypeStop, -6.1500, 106.8700, "Jakarta Utara", "Tanjung Priok"},
		{"Yos Sudarso", models.StopTypeStop, -6.1400, 106.8800, "Jakarta Utara", "Tanjung Priok"},
		{"Ancol", models.StopTypeStop, -6.1256, 106.8303, "Jakarta Utara", "Ancol"},
		{"Pademangan", models.StopTypeStop, -6.1300, 106.8400, "Jakarta Utara", "Pademangan"},
		{"Kemayoran", models.StopTypeStop, -6.1600, 106.8500, "Jakarta Pusat", "Kemayoran"},
		{"Gunung Sahari", models.StopTypeStop, -6.1500, 106.8400, "Jakarta Pusat", "Sawah Besar"},
		{"Mangga Dua", models.StopTypeStop, -6.1400, 106.8300, "Jakarta Utara", "Sawah Besar"},
		{"Jembatan Merah", models.StopTypeStop, -6.1300, 106.8200, "Jakarta Utara", "Jakarta Barat"},
		{"Stasiun Kota", models.StopTypeStop, -6.1352, 106.8133, "Jakarta Utara", "Jakarta Barat"},
		{"Pluit", models.StopTypeStop, -6.1156, 106.7881, "Jakarta Utara", "Penjaringan"},
		{"Ciledug", models.StopTypeStop, -6.2400, 106.7500, "Jakarta Selatan", "Ciledug"},
		{"Tegal Mampang", models.StopTypeStop, -6.2500, 106.8000, "Jakarta Selatan", "Mampang"},
		{"Pasar Senen", models.StopTypeStop, -6.1756, 106.8411, "Jakarta Pusat", "Senen"},
		{"Jakarta Int'l Stadium", models.StopTypeStop, -6.2000, 106.8800, "Jakarta Utara", "Papanggo"},
	}

	// Create stops (avoid duplicates)
	fmt.Println("Creating stops...")
	uniqueStops := make(map[string]bool)
	for _, stopData := range stops {
		if uniqueStops[stopData.name] {
			continue
		}
		uniqueStops[stopData.name] = true

		// Check if stop already exists by name
		var existingStop models.Stop
		if err := db.Where("name = ?", stopData.name).First(&existingStop).Error; err == nil {
			// Stop already exists, use it
			stopsMap[stopData.name] = &existingStop
			fmt.Printf("  ⊙ Stop already exists: %s\n", stopData.name)
			continue
		}

		stop := &models.Stop{
			Name:      stopData.name,
			Type:      stopData.stopType,
			Latitude:  stopData.latitude,
			Longitude: stopData.longitude,
			City:      stopData.city,
			District:  stopData.district,
			Status:    models.StatusActive,
		}

		if err := stopRepo.Create(stop); err != nil {
			// If it's a duplicate key error, try to fetch existing
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
				if err := db.Where("name = ?", stopData.name).First(stop).Error; err == nil {
					stopsMap[stopData.name] = stop
					fmt.Printf("  ⊙ Stop already exists: %s\n", stopData.name)
					continue
				}
			}
			fmt.Printf("  ⚠ Warning: Failed to create stop %s: %v\n", stopData.name, err)
			continue
		}

		stopsMap[stopData.name] = stop
		fmt.Printf("  ✓ Created stop: %s\n", stopData.name)
	}

	// Fetch all stops from DB to ensure we have all IDs
	var allStops []models.Stop
	if err := db.Find(&allStops).Error; err != nil {
		return fmt.Errorf("failed to fetch stops: %w", err)
	}
	for i := range allStops {
		stopsMap[allStops[i].Name] = &allStops[i]
	}

	// Define routes with their stops in sequence
	routes := []struct {
		routeNumber string
		name        string
		description string
		stopNames   []string
	}{
		{
			routeNumber: "1",
			name:        "Blok M - Kota",
			description: "Koridor 1: Blok M ke Kota",
			stopNames:   []string{"Blok M", "ASEAN", "Senopati", "Sudirman", "Thamrin", "Sarinah", "Bank Indonesia", "Harmoni", "Gajah Mada", "Hayam Wuruk", "Mangga Besar", "Olahraga", "Stasiun Kota", "Kota"},
		},
		{
			routeNumber: "2",
			name:        "Pulo Gadung - Monumen Nasional",
			description: "Koridor 2: Pulo Gadung ke Monumen Nasional",
			stopNames:   []string{"Pulo Gadung", "Rawamangun", "Pulomas", "Kayu Putih", "Pramuka BPKP", "Matraman", "Bidara Cina", "Stasiun Jatinegara", "Kampung Melayu", "Cikini", "Menteng", "Gondangdia", "Juanda", "Istiqlal", "Lapangan Banteng", "Pecenongan", "Harmoni", "Monumen Nasional"},
		},
		{
			routeNumber: "3",
			name:        "Kalideres - Monumen Nasional",
			description: "Koridor 3: Kalideres ke Monumen Nasional",
			stopNames:   []string{"Kalideres", "Rawa Buaya", "Cengkareng", "Bojong Indah", "Puri Kembangan", "Kembangan", "Meruya", "Kebon Jeruk", "Tomang", "Grogol", "Slipi", "Palmerah", "Tosari", "Dukuh Atas", "Setiabudi", "Kuningan", "Monumen Nasional"},
		},
		{
			routeNumber: "4",
			name:        "Pulo Gadung - Galunggung",
			description: "Koridor 4: Pulo Gadung ke Galunggung",
			stopNames:   []string{"Pulo Gadung", "Rawamangun", "Pulomas", "Kayu Putih", "Pramuka BPKP", "Matraman", "Bidara Cina", "Stasiun Jatinegara", "Kampung Melayu", "Galunggung"},
		},
		{
			routeNumber: "5",
			name:        "Ancol - Kampung Melayu",
			description: "Koridor 5: Ancol ke Kampung Melayu",
			stopNames:   []string{"Ancol", "Pademangan", "Sunter", "Yos Sudarso", "Tanjung Priok", "Kemayoran", "Gunung Sahari", "Mangga Dua", "Jembatan Merah", "Stasiun Kota", "Kota", "Harmoni", "Gajah Mada", "Hayam Wuruk", "Mangga Besar", "Olahraga", "Stasiun Kota", "Kampung Melayu"},
		},
		{
			routeNumber: "6",
			name:        "Ragunan - Galunggung",
			description: "Koridor 6: Ragunan ke Galunggung",
			stopNames:   []string{"Ragunan", "Pasar Minggu", "Tanjung Barat", "Lenteng Agung", "Universitas Indonesia", "Pondok Cabe", "Lebak Bulus", "Fatmawati", "Cipete Raya", "Haji Nawi", "Blok A", "Blok M", "Pasaraya", "ASEAN", "Senopati", "Sudirman", "Thamrin", "Sarinah", "Bank Indonesia", "Harmoni", "Gajah Mada", "Hayam Wuruk", "Mangga Besar", "Olahraga", "Stasiun Kota", "Galunggung"},
		},
		{
			routeNumber: "7",
			name:        "Kampung Rambutan - Kampung Melayu",
			description: "Koridor 7: Kampung Rambutan ke Kampung Melayu",
			stopNames:   []string{"Kampung Rambutan", "Pasar Rebo", "Cililitan", "Bidara Cina", "Stasiun Jatinegara", "Kampung Melayu"},
		},
		{
			routeNumber: "8",
			name:        "Lebak Bulus - Pasar Baru",
			description: "Koridor 8: Lebak Bulus ke Pasar Baru",
			stopNames:   []string{"Lebak Bulus", "Fatmawati", "Cipete Raya", "Haji Nawi", "Blok A", "Blok M", "Pasaraya", "ASEAN", "Senopati", "Sudirman", "Thamrin", "Sarinah", "Bank Indonesia", "Harmoni", "Gajah Mada", "Hayam Wuruk", "Mangga Besar", "Olahraga", "Stasiun Kota", "Pasar Baru"},
		},
		{
			routeNumber: "9",
			name:        "Pinang Ranti - Pluit",
			description: "Koridor 9: Pinang Ranti ke Pluit",
			stopNames:   []string{"Pinang Ranti", "Cililitan", "Bidara Cina", "Stasiun Jatinegara", "Cawang UKI", "Cikoko", "Cawang", "BKN", "Tugu Tani", "Gambir", "Sawah Besar", "Mangga Dua", "Gunung Sahari", "Jembatan Merah", "Pecenongan", "Kwitang", "Kramat", "Pasar Rebo", "Cililitan", "Bidara Cina", "Stasiun Jatinegara", "Kampung Melayu", "Pluit"},
		},
		{
			routeNumber: "10",
			name:        "PGC - Tanjung Priok",
			description: "Koridor 10: PGC ke Tanjung Priok",
			stopNames:   []string{"PGC", "Cakung", "Rawamangun", "Pulomas", "Kayu Putih", "Pramuka BPKP", "Matraman", "Bidara Cina", "Stasiun Jatinegara", "Kampung Melayu", "Tanjung Priok"},
		},
		{
			routeNumber: "11",
			name:        "Pulo Gebang - Kampung Melayu",
			description: "Koridor 11: Pulo Gebang ke Kampung Melayu",
			stopNames:   []string{"Pulo Gebang", "Cakung", "Rawamangun", "Pulomas", "Kayu Putih", "Pramuka BPKP", "Matraman", "Bidara Cina", "Stasiun Jatinegara", "Kampung Melayu"},
		},
		{
			routeNumber: "12",
			name:        "Tanjung Priok - Pluit",
			description: "Koridor 12: Tanjung Priok ke Pluit",
			stopNames:   []string{"Tanjung Priok", "Sunter", "Yos Sudarso", "Ancol", "Pademangan", "Kemayoran", "Gunung Sahari", "Mangga Dua", "Jembatan Merah", "Stasiun Kota", "Pluit"},
		},
		{
			routeNumber: "13",
			name:        "Ciledug - Tegal Mampang",
			description: "Koridor 13: Ciledug ke Tegal Mampang",
			stopNames:   []string{"Ciledug", "Tegal Mampang"},
		},
		{
			routeNumber: "14",
			name:        "Pasar Senen - Jakarta Int'l Stadium",
			description: "Koridor 14: Pasar Senen ke Jakarta International Stadium",
			stopNames:   []string{"Pasar Senen", "Jakarta Int'l Stadium"},
		},
	}

	// Create routes and route stops
	fmt.Println("\nCreating routes...")
	for _, routeData := range routes {
		// Check if route already exists
		var existingRoute models.Route
		if err := db.Where("route_number = ?", routeData.routeNumber).First(&existingRoute).Error; err == nil {
			fmt.Printf("  ⚠ Route %s already exists, skipping...\n", routeData.routeNumber)
			continue
		}

		route := &models.Route{
			RouteNumber: routeData.routeNumber,
			Name:        routeData.name,
			Description: routeData.description,
			Status:      models.StatusActive,
		}

		if err := routeRepo.Create(route); err != nil {
			return fmt.Errorf("failed to create route %s: %w", routeData.routeNumber, err)
		}

		fmt.Printf("  ✓ Created route: %s - %s\n", routeData.routeNumber, routeData.name)

		// Create route stops
		for seq, stopName := range routeData.stopNames {
			stop, ok := stopsMap[stopName]
			if !ok {
				fmt.Printf("    ⚠ Warning: Stop '%s' not found, skipping...\n", stopName)
				continue
			}

			routeStop := &models.RouteStop{
				RouteID:       route.ID,
				StopID:        stop.ID,
				SequenceOrder: seq + 1,
				IsOrigin:      seq == 0,
				IsDestination: seq == len(routeData.stopNames)-1,
			}

			if err := routeStopRepo.Create(routeStop); err != nil {
				fmt.Printf("    ⚠ Warning: Failed to create route stop for %s: %v\n", stopName, err)
				continue
			}
		}
		fmt.Printf("    ✓ Added %d stops to route %s\n", len(routeData.stopNames), routeData.routeNumber)
	}

	return nil
}
