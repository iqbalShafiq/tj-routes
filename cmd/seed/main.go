package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

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

	// Run migrations first
	fmt.Println("Running database migrations...")
	if err := utils.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("✓ Database migrations completed")

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
	userBadgeRepo := repository.NewUserBadgeRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	reportRepo := repository.NewReportRepository(db)
	reportCategoryRepo := repository.NewReportCategoryRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	routeChangeRepo := repository.NewRouteChangeRepository(db)
	hashtagRepo := repository.NewHashtagRepository(db)
	userFollowRepo := repository.NewUserFollowRepository(db)
	bulkUploadLogRepo := repository.NewBulkUploadLogRepository(db)
	forumRepo := repository.NewForumRepository(db)
	forumPostRepo := repository.NewForumPostRepository(db)
	forumMemberRepo := repository.NewForumMemberRepository(db)

	// Seed badges first
	if err := seedBadges(badgeRepo); err != nil {
		log.Fatalf("Failed to seed badges: %v", err)
	}

	// Seed report categories
	if err := seedReportCategories(reportCategoryRepo); err != nil {
		log.Fatalf("Failed to seed report categories: %v", err)
	}

	// Seed users
	if err := seedUsers(userRepo); err != nil {
		log.Fatalf("Failed to seed users: %v", err)
	}

	// Seed data (stops and routes)
	if err := seedData(db, stopRepo, routeRepo, routeStopRepo, adminUser.ID); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	// Seed vehicles
	if err := seedVehicles(db, vehicleRepo, routeRepo); err != nil {
		log.Fatalf("Failed to seed vehicles: %v", err)
	}

	// Seed reports
	if err := seedReports(db, reportRepo, userRepo, routeRepo, stopRepo); err != nil {
		log.Fatalf("Failed to seed reports: %v", err)
	}

	// Seed comments
	if err := seedComments(db, commentRepo, reportRepo, userRepo); err != nil {
		log.Fatalf("Failed to seed comments: %v", err)
	}

	// Seed reactions
	if err := seedReactions(db, reactionRepo, reportRepo, commentRepo, userRepo); err != nil {
		log.Fatalf("Failed to seed reactions: %v", err)
	}

	// Seed route changes
	if err := seedRouteChanges(db, routeChangeRepo, routeRepo, userRepo, stopRepo); err != nil {
		log.Fatalf("Failed to seed route changes: %v", err)
	}

	// Seed forums
	if err := seedForums(db, forumRepo, routeRepo, forumPostRepo, forumMemberRepo, userRepo); err != nil {
		log.Fatalf("Failed to seed forums: %v", err)
	}

	// Seed user badges
	if err := seedUserBadges(db, userBadgeRepo, badgeRepo, userRepo); err != nil {
		log.Fatalf("Failed to seed user badges: %v", err)
	}

	// Seed hashtags
	if err := seedHashtags(hashtagRepo, reportRepo); err != nil {
		log.Fatalf("Failed to seed hashtags: %v", err)
	}

	// Seed user follows
	if err := seedUserFollows(userFollowRepo, userRepo); err != nil {
		log.Fatalf("Failed to seed user follows: %v", err)
	}

	// Seed bulk upload logs
	if err := seedBulkUploadLogs(bulkUploadLogRepo, userRepo); err != nil {
		log.Fatalf("Failed to seed bulk upload logs: %v", err)
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
		// Journey Check-in Badges
		{
			Name:          "First Journey",
			Description:   "Completed your first check-in",
			Icon:          "🚌",
			CriteriaType:  models.BadgeCriteriaCheckInsCount,
			CriteriaValue: 1,
		},
		{
			Name:          "Commuter",
			Description:   "Completed 10 check-ins",
			Icon:          "🚇",
			CriteriaType:  models.BadgeCriteriaCheckInsCount,
			CriteriaValue: 10,
		},
		{
			Name:          "Regular Rider",
			Description:   "Completed 50 check-ins",
			Icon:          "🎫",
			CriteriaType:  models.BadgeCriteriaCheckInsCount,
			CriteriaValue: 50,
		},
		{
			Name:          "Route Master",
			Description:   "Checked in on 10 different routes",
			Icon:          "🗺️",
			CriteriaType:  models.BadgeCriteriaUniqueRoutes,
			CriteriaValue: 10,
		},
		{
			Name:          "Consistent Commuter",
			Description:   "7 consecutive days of check-ins",
			Icon:          "📅",
			CriteriaType:  models.BadgeCriteriaConsecutiveDays,
			CriteriaValue: 7,
		},
		{
			Name:          "Dedicated Rider",
			Description:   "30 consecutive days of check-ins",
			Icon:          "🏅",
			CriteriaType:  models.BadgeCriteriaConsecutiveDays,
			CriteriaValue: 30,
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

func seedReportCategories(categoryRepo repository.ReportCategoryRepository) error {
	desc1 := "Vehicle accidents or crashes"
	desc2 := "Litter or garbage issues"
	desc3 := "Malfunctioning or broken buses"
	desc4 := "Potholes, cracks, or other road damage"
	desc5 := "Traffic congestion or flow problems"
	desc6 := "Safety-related issues or hazards"
	desc7 := "Other types of issues not covered above"

	categories := []models.ReportCategory{
		{
			Name:        "Crash",
			Description: &desc1,
		},
		{
			Name:        "Trash",
			Description: &desc2,
		},
		{
			Name:        "Broken Bus",
			Description: &desc3,
		},
		{
			Name:        "Road Damage",
			Description: &desc4,
		},
		{
			Name:        "Traffic Issue",
			Description: &desc5,
		},
		{
			Name:        "Safety Concern",
			Description: &desc6,
		},
		{
			Name:        "Other",
			Description: &desc7,
		},
	}

	for _, category := range categories {
		// Check if category already exists
		existing, err := categoryRepo.FindByName(category.Name)
		if err == nil && existing != nil {
			continue // Category already exists, skip
		}

		if err := categoryRepo.Create(&category); err != nil {
			return fmt.Errorf("failed to create category %s: %w", category.Name, err)
		}
	}

	return nil
}

func seedUsers(userRepo repository.UserRepository) error {
	// Initialize random generator
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Indonesian first names
	firstNames := []string{
		"Ahmad", "Budi", "Cahya", "Dedi", "Eko", "Fajar", "Gunawan", "Hadi", "Indra", "Joko",
		"Kurniawan", "Lukman", "Mulyadi", "Nugroho", "Oki", "Prasetyo", "Rahmat", "Sari", "Taufik", "Udin",
		"Wahyu", "Yanto", "Zainal", "Ayu", "Bunga", "Citra", "Dewi", "Eka", "Fitri", "Gita",
		"Hani", "Indah", "Jihan", "Kartika", "Lestari", "Maya", "Nina", "Oktavia", "Putri", "Rina",
		"Sinta", "Tika", "Umi", "Vina", "Wati", "Yuli", "Zahra", "Ade", "Bayu", "Candra",
	}

	// Indonesian last names
	lastNames := []string{
		"Santoso", "Wijaya", "Kurniawan", "Prasetyo", "Sari", "Lestari", "Nugroho", "Rahman", "Hidayat", "Saputra",
		"Setiawan", "Gunawan", "Siregar", "Siregar", "Hutapea", "Simanjuntak", "Nainggolan", "Sihombing", "Manurung", "Situmorang",
		"Purba", "Saragih", "Sinaga", "Lumban", "Tambunan", "Pangaribuan", "Samosir", "Pardede", "Sianturi", "Lubis",
		"Nasution", "Harahap", "Ritonga", "Dalimunthe", "Hasibuan", "Tanjung", "Pohan", "Siregar", "Lubis", "Nasution",
		"Wijaya", "Kusuma", "Dewi", "Sari", "Lestari", "Nugroho", "Rahman", "Hidayat", "Saputra", "Setiawan",
	}

	// Domains for emails
	domains := []string{
		"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "mail.com",
		"ymail.com", "rocketmail.com", "live.com", "msn.com", "icloud.com",
	}

	fmt.Println("Seeding users...")

	// Create 50 users (middle of 20-100 range)
	numUsers := 50
	createdCount := 0
	skippedCount := 0

	for i := 0; i < numUsers; i++ {
		firstName := firstNames[rng.Intn(len(firstNames))]
		lastName := lastNames[rng.Intn(len(lastNames))]
		username := strings.ToLower(firstName + lastName + fmt.Sprintf("%d", rng.Intn(9999)))
		email := strings.ToLower(firstName + "." + lastName + fmt.Sprintf("%d", rng.Intn(9999)) + "@" + domains[rng.Intn(len(domains))])

		// Check if user already exists
		_, err := userRepo.FindByEmail(email)
		if err == nil {
			skippedCount++
			continue
		}

		// Hash password
		hashedPassword, err := utils.HashPassword("password123") // Default password for all seeded users
		if err != nil {
			fmt.Printf("  ⚠ Warning: Failed to hash password for user %s: %v\n", username, err)
			continue
		}

		// Random reputation points (0-500)
		reputationPoints := rng.Intn(501)

		// Determine level based on reputation
		var level string
		switch {
		case reputationPoints >= 500:
			level = "expert"
		case reputationPoints >= 200:
			level = "trusted"
		case reputationPoints >= 50:
			level = "rising_star"
		default:
			level = "newcomer"
		}

		user := &models.User{
			Email:            email,
			Username:         username,
			Password:         &hashedPassword,
			Role:             models.RoleCommonUser,
			ReputationPoints: reputationPoints,
			Level:            level,
		}

		if err := userRepo.Create(user); err != nil {
			// Check if it's a duplicate error
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
				skippedCount++
				continue
			}
			fmt.Printf("  ⚠ Warning: Failed to create user %s: %v\n", username, err)
			continue
		}

		createdCount++
		if createdCount%10 == 0 {
			fmt.Printf("  ✓ Created %d users...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d users (skipped %d duplicates)\n", createdCount, skippedCount)
	return nil
}

func seedVehicles(db *gorm.DB, vehicleRepo repository.VehicleRepository, routeRepo repository.RouteRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all routes
	routes, _, err := routeRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch routes: %w", err)
	}
	if len(routes) == 0 {
		fmt.Println("  ⚠ No routes found, skipping vehicle seeding")
		return nil
	}

	vehicleTypes := []string{"Bus", "Bus Rapid Transit", "Microbus", "Minibus"}
	statuses := []models.Status{models.StatusActive, models.StatusActive, models.StatusActive, models.StatusInactive} // Mostly active

	fmt.Println("Seeding vehicles...")
	numVehicles := 50
	createdCount := 0

	for i := 0; i < numVehicles; i++ {
		route := routes[rng.Intn(len(routes))]
		vehicleType := vehicleTypes[rng.Intn(len(vehicleTypes))]
		status := statuses[rng.Intn(len(statuses))]

		// Generate plate number (Jakarta format: B 1234 XYZ)
		plateNumber := fmt.Sprintf("B %d%02d%02d %c%c%c",
			rng.Intn(9)+1,          // First digit 1-9
			rng.Intn(100),          // Two digits
			rng.Intn(100),          // Two digits
			'A'+rune(rng.Intn(26)), // Letter
			'A'+rune(rng.Intn(26)), // Letter
			'A'+rune(rng.Intn(26)), // Letter
		)

		// Check if vehicle already exists
		_, err := vehicleRepo.FindByVehiclePlate(plateNumber)
		if err == nil {
			continue
		}

		capacity := 0
		switch vehicleType {
		case "Bus", "Bus Rapid Transit":
			capacity = 40 + rng.Intn(20) // 40-60
		case "Microbus":
			capacity = 12 + rng.Intn(4) // 12-16
		case "Minibus":
			capacity = 20 + rng.Intn(10) // 20-30
		}

		vehicle := &models.Vehicle{
			VehiclePlate: plateNumber,
			RouteID:      route.ID,
			VehicleType:  vehicleType,
			Capacity:     capacity,
			Status:       status,
		}

		if err := vehicleRepo.Create(vehicle); err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
				continue
			}
			fmt.Printf("  ⚠ Warning: Failed to create vehicle %s: %v\n", plateNumber, err)
			continue
		}

		createdCount++
		if createdCount%10 == 0 {
			fmt.Printf("  ✓ Created %d vehicles...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d vehicles\n", createdCount)
	return nil
}

func seedReports(db *gorm.DB, reportRepo repository.ReportRepository, userRepo repository.UserRepository, routeRepo repository.RouteRepository, stopRepo repository.StopRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all users, routes, and stops
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	routes, _, err := routeRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch routes: %w", err)
	}
	stops, _, err := stopRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch stops: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("  ⚠ No users found, skipping report seeding")
		return nil
	}

	reportTypes := []models.ReportType{
		models.ReportTypeRouteIssue,
		models.ReportTypeStopIssue,
		models.ReportTypeTemporaryEvent,
		models.ReportTypePolicyChange,
	}
	statuses := []models.ReportStatus{
		models.ReportStatusPending,
		models.ReportStatusReviewed,
		models.ReportStatusResolved,
	}

	reportTitles := []string{
		"Bus terlambat di halte",
		"Halte rusak dan perlu perbaikan",
		"Rute baru dibuka",
		"Perubahan jadwal operasional",
		"Fasilitas halte tidak berfungsi",
		"Kemacetan di jalur busway",
		"Penumpang terlalu banyak",
		"AC bus tidak berfungsi",
		"Pintu bus rusak",
		"Kursi bus perlu diganti",
		"Lampu halte mati",
		"Papan informasi tidak update",
		"Kebersihan halte kurang",
		"Keamanan di halte perlu ditingkatkan",
		"Jadwal tidak sesuai",
	}

	descriptions := []string{
		"Bus sering terlambat di halte ini, mohon perbaikan jadwal",
		"Halte ini mengalami kerusakan pada atap dan perlu segera diperbaiki",
		"Rute baru telah dibuka untuk melayani area ini",
		"Ada perubahan jadwal operasional mulai bulan depan",
		"Fasilitas seperti kursi dan tempat sampah di halte tidak berfungsi dengan baik",
		"Kemacetan sering terjadi di jalur busway ini pada jam sibuk",
		"Penumpang terlalu banyak sehingga tidak semua bisa naik",
		"AC di beberapa bus tidak berfungsi dengan baik",
		"Pintu bus sering macet dan perlu perbaikan",
		"Beberapa kursi bus sudah rusak dan perlu diganti",
		"Lampu di halte mati pada malam hari",
		"Papan informasi tidak ter-update dengan jadwal terbaru",
		"Kebersihan halte perlu ditingkatkan, banyak sampah",
		"Keamanan di halte perlu ditingkatkan, terutama pada malam hari",
		"Jadwal yang tertera tidak sesuai dengan kenyataan",
	}

	fmt.Println("Seeding reports...")
	numReports := 50
	createdCount := 0

	for i := 0; i < numReports; i++ {
		user := users[rng.Intn(len(users))]
		reportType := reportTypes[rng.Intn(len(reportTypes))]
		status := statuses[rng.Intn(len(statuses))]

		var relatedRouteID *uint
		var relatedStopID *uint

		if reportType == models.ReportTypeRouteIssue && len(routes) > 0 {
			routeID := routes[rng.Intn(len(routes))].ID
			relatedRouteID = &routeID
		}
		if reportType == models.ReportTypeStopIssue && len(stops) > 0 {
			stopID := stops[rng.Intn(len(stops))].ID
			relatedStopID = &stopID
		}

		title := reportTitles[rng.Intn(len(reportTitles))]
		description := descriptions[rng.Intn(len(descriptions))]

		upvotes := rng.Intn(50)
		downvotes := rng.Intn(10)

		report := &models.Report{
			UserID:         user.ID,
			Type:           reportType,
			Title:          title,
			Description:    description,
			RelatedRouteID: relatedRouteID,
			RelatedStopID:  relatedStopID,
			Status:         status,
			Upvotes:        upvotes,
			Downvotes:      downvotes,
			CommentCount:   0, // Will be updated when comments are created
		}

		if err := reportRepo.Create(report); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to create report: %v\n", err)
			continue
		}

		createdCount++
		if createdCount%10 == 0 {
			fmt.Printf("  ✓ Created %d reports...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d reports\n", createdCount)
	return nil
}

func seedComments(db *gorm.DB, commentRepo repository.CommentRepository, reportRepo repository.ReportRepository, userRepo repository.UserRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all users and reports
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	reports, _, err := reportRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch reports: %w", err)
	}

	if len(reports) == 0 {
		fmt.Println("  ⚠ No reports found, skipping comment seeding")
		return nil
	}

	commentTexts := []string{
		"Setuju, halte ini memang perlu diperbaiki",
		"Terima kasih atas laporannya",
		"Saya juga mengalami hal yang sama",
		"Semoga segera ditindaklanjuti",
		"Informasi yang sangat membantu",
		"Bagus, halte ini memang perlu perhatian",
		"Saya setuju dengan laporan ini",
		"Terima kasih sudah melaporkan",
		"Semoga bisa segera diperbaiki",
		"Laporan yang sangat detail",
		"Saya juga melihat masalah yang sama",
		"Terima kasih atas informasinya",
		"Semoga pihak terkait segera menindaklanjuti",
		"Laporan yang sangat membantu",
		"Setuju dengan pendapat ini",
	}

	fmt.Println("Seeding comments...")
	numComments := 50
	createdCount := 0

	// Track comments for replies
	var topLevelComments []*models.Comment

	for i := 0; i < numComments; i++ {
		report := reports[rng.Intn(len(reports))]
		user := users[rng.Intn(len(users))]
		content := commentTexts[rng.Intn(len(commentTexts))]

		var parentID *uint
		// 30% chance of being a reply
		if len(topLevelComments) > 0 && rng.Float32() < 0.3 {
			parent := topLevelComments[rng.Intn(len(topLevelComments))]
			parentID = &parent.ID
		}

		upvotes := rng.Intn(20)
		downvotes := rng.Intn(5)

		reportID := report.ID
		comment := &models.Comment{
			ReportID:  &reportID,
			UserID:    user.ID,
			ParentID:  parentID,
			Content:   content,
			Upvotes:   upvotes,
			Downvotes: downvotes,
		}

		if err := commentRepo.Create(comment); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to create comment: %v\n", err)
			continue
		}

		// Track top-level comments for potential replies
		if parentID == nil {
			topLevelComments = append(topLevelComments, comment)
		}

		// Update report comment count
		db.Model(&models.Report{}).Where("id = ?", report.ID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))

		createdCount++
		if createdCount%10 == 0 {
			fmt.Printf("  ✓ Created %d comments...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d comments\n", createdCount)
	return nil
}

func seedReactions(db *gorm.DB, reactionRepo repository.ReactionRepository, reportRepo repository.ReportRepository, commentRepo repository.CommentRepository, userRepo repository.UserRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all users, reports, and comments
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	reports, _, err := reportRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch reports: %w", err)
	}

	// Get comments
	var comments []models.Comment
	if err := db.Find(&comments).Error; err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("  ⚠ No users found, skipping reaction seeding")
		return nil
	}

	reactionTypes := []models.ReactionType{
		models.ReactionUpvote,
		models.ReactionDownvote,
	}

	fmt.Println("Seeding reactions...")
	numReactions := 50
	createdCount := 0

	// Track reactions to avoid duplicates
	reactionMap := make(map[string]bool)

	for i := 0; i < numReactions; i++ {
		user := users[rng.Intn(len(users))]
		reactionType := reactionTypes[rng.Intn(len(reactionTypes))]
		// 70% upvotes, 30% downvotes
		if rng.Float32() < 0.7 {
			reactionType = models.ReactionUpvote
		} else {
			reactionType = models.ReactionDownvote
		}

		var targetType models.ReactionTargetType
		var targetID uint

		// 60% on reports, 40% on comments
		if len(comments) > 0 && rng.Float32() < 0.4 {
			targetType = models.ReactionTargetComment
			targetID = comments[rng.Intn(len(comments))].ID
		} else if len(reports) > 0 {
			targetType = models.ReactionTargetReport
			targetID = reports[rng.Intn(len(reports))].ID
		} else {
			continue
		}

		// Check for duplicate
		key := fmt.Sprintf("%d-%s-%d", user.ID, targetType, targetID)
		if reactionMap[key] {
			continue
		}
		reactionMap[key] = true

		// Check if reaction already exists
		_, err := reactionRepo.FindByUserAndTarget(user.ID, targetType, targetID)
		if err == nil {
			continue
		}

		reaction := &models.Reaction{
			UserID:       user.ID,
			TargetType:   targetType,
			TargetID:     targetID,
			ReactionType: reactionType,
		}

		if err := reactionRepo.Create(reaction); err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
				continue
			}
			fmt.Printf("  ⚠ Warning: Failed to create reaction: %v\n", err)
			continue
		}

		// Update upvotes/downvotes on target
		if targetType == models.ReactionTargetReport {
			if reactionType == models.ReactionUpvote {
				db.Model(&models.Report{}).Where("id = ?", targetID).UpdateColumn("upvotes", gorm.Expr("upvotes + 1"))
			} else {
				db.Model(&models.Report{}).Where("id = ?", targetID).UpdateColumn("downvotes", gorm.Expr("downvotes + 1"))
			}
		} else {
			if reactionType == models.ReactionUpvote {
				db.Model(&models.Comment{}).Where("id = ?", targetID).UpdateColumn("upvotes", gorm.Expr("upvotes + 1"))
			} else {
				db.Model(&models.Comment{}).Where("id = ?", targetID).UpdateColumn("downvotes", gorm.Expr("downvotes + 1"))
			}
		}

		createdCount++
		if createdCount%10 == 0 {
			fmt.Printf("  ✓ Created %d reactions...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d reactions\n", createdCount)
	return nil
}

func seedRouteChanges(db *gorm.DB, routeChangeRepo repository.RouteChangeRepository, routeRepo repository.RouteRepository, userRepo repository.UserRepository, stopRepo repository.StopRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all routes, users, and stops
	routes, _, err := routeRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch routes: %w", err)
	}
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	stops, _, err := stopRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch stops: %w", err)
	}

	if len(routes) == 0 || len(users) == 0 {
		fmt.Println("  ⚠ No routes or users found, skipping route change seeding")
		return nil
	}

	changeTypes := []models.ChangeType{
		models.ChangeTypeRouteCreated,
		models.ChangeTypeRouteUpdated,
		models.ChangeTypeStopAdded,
		models.ChangeTypeStopRemoved,
		models.ChangeTypeStopOrderChanged,
		models.ChangeTypeStopUpdated,
	}

	reasons := []string{
		"Penyesuaian rute berdasarkan kebutuhan penumpang",
		"Perbaikan infrastruktur halte",
		"Optimasi waktu tempuh",
		"Penambahan halte baru",
		"Perubahan jadwal operasional",
		"Peningkatan layanan",
		"Penyesuaian dengan perkembangan kota",
		"Perbaikan aksesibilitas",
	}

	fmt.Println("Seeding route changes...")
	numChanges := 30
	createdCount := 0

	for i := 0; i < numChanges; i++ {
		route := routes[rng.Intn(len(routes))]
		user := users[rng.Intn(len(users))]
		changeType := changeTypes[rng.Intn(len(changeTypes))]
		reason := reasons[rng.Intn(len(reasons))]

		var affectedStopID *uint
		if (changeType == models.ChangeTypeStopAdded || changeType == models.ChangeTypeStopRemoved || changeType == models.ChangeTypeStopUpdated) && len(stops) > 0 {
			stopID := stops[rng.Intn(len(stops))].ID
			affectedStopID = &stopID
		}

		routeChange := &models.RouteChange{
			RouteID:        route.ID,
			ChangedBy:      user.ID,
			ChangeType:     changeType,
			AffectedStopID: affectedStopID,
			Reason:         reason,
		}

		if err := routeChangeRepo.Create(routeChange); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to create route change: %v\n", err)
			continue
		}

		createdCount++
		if createdCount%10 == 0 {
			fmt.Printf("  ✓ Created %d route changes...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d route changes\n", createdCount)
	return nil
}

func seedUserBadges(db *gorm.DB, userBadgeRepo repository.UserBadgeRepository, badgeRepo repository.BadgeRepository, userRepo repository.UserRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all users and badges
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}
	badges, err := badgeRepo.FindAll()
	if err != nil {
		return fmt.Errorf("failed to fetch badges: %w", err)
	}

	if len(users) == 0 || len(badges) == 0 {
		fmt.Println("  ⚠ No users or badges found, skipping user badge seeding")
		return nil
	}

	fmt.Println("Seeding user badges...")
	createdCount := 0

	// Assign badges to users based on their reputation and activity
	for _, user := range users {
		// Check which badges user should have based on reputation
		for _, badge := range badges {
			shouldHave := false

			switch badge.CriteriaType {
			case models.BadgeCriteriaReputationPoints:
				if user.ReputationPoints >= badge.CriteriaValue {
					shouldHave = true
				}
			case models.BadgeCriteriaReportsAccepted:
				// Simulate: users with higher reputation likely have more accepted reports
				if user.ReputationPoints >= badge.CriteriaValue*10 {
					shouldHave = true
				}
			case models.BadgeCriteriaCommentsMade:
				// Simulate: users with higher reputation likely have more comments
				if user.ReputationPoints >= badge.CriteriaValue*2 {
					shouldHave = true
				}
			case models.BadgeCriteriaUpvotesReceived:
				// Simulate: users with higher reputation likely have more upvotes
				if user.ReputationPoints >= badge.CriteriaValue*2 {
					shouldHave = true
				}
			case models.BadgeCriteriaCheckInsCount:
				// Simulate: users with higher reputation likely have more check-ins
				if user.ReputationPoints >= badge.CriteriaValue*5 {
					shouldHave = true
				}
			case models.BadgeCriteriaUniqueRoutes:
				// Simulate: users with higher reputation likely have checked more routes
				if user.ReputationPoints >= badge.CriteriaValue*10 {
					shouldHave = true
				}
			case models.BadgeCriteriaConsecutiveDays:
				// Simulate: only very active users have consecutive days badges
				if user.ReputationPoints >= badge.CriteriaValue*15 {
					shouldHave = true
				}
			}

			if shouldHave {
				// Check if user already has this badge
				_, err := userBadgeRepo.FindByUserAndBadge(user.ID, badge.ID)
				if err == nil {
					continue
				}

				// Random earned time (within last 6 months)
				earnedAt := time.Now().Add(-time.Duration(rng.Intn(180)) * 24 * time.Hour)

				userBadge := &models.UserBadge{
					UserID:   user.ID,
					BadgeID:  badge.ID,
					EarnedAt: earnedAt,
				}

				if err := userBadgeRepo.Create(userBadge); err != nil {
					if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
						continue
					}
					fmt.Printf("  ⚠ Warning: Failed to create user badge: %v\n", err)
					continue
				}

				createdCount++
			}
		}
	}

	fmt.Printf("  ✓ Created %d user badges\n", createdCount)
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
		{
			routeNumber: "15",
			name:        "Lebak Bulus - Harmoni",
			description: "Koridor 15: Lebak Bulus ke Harmoni",
			stopNames:   []string{"Lebak Bulus", "Fatmawati", "Cipete Raya", "Haji Nawi", "Blok A", "Blok M", "ASEAN", "Senopati", "Sudirman", "Thamrin", "Harmoni"},
		},
		{
			routeNumber: "16",
			name:        "Harmoni - Ragunan",
			description: "Koridor 16: Harmoni ke Ragunan",
			stopNames:   []string{"Harmoni", "Thamrin", "Sudirman", "Senopati", "ASEAN", "Blok M", "Blok A", "Haji Nawi", "Cipete Raya", "Fatmawati", "Lebak Bulus", "Ragunan"},
		},
		{
			routeNumber: "17",
			name:        "Pulo Gadung - Cawang",
			description: "Koridor 17: Pulo Gadung ke Cawang",
			stopNames:   []string{"Pulo Gadung", "Rawamangun", "Pulomas", "Kayu Putih", "Pramuka BPKP", "Matraman", "Bidara Cina", "Stasiun Jatinegara", "Cawang UKI", "Cawang"},
		},
		{
			routeNumber: "18",
			name:        "Kampung Rambutan - Cikoko",
			description: "Koridor 18: Kampung Rambutan ke Cikoko",
			stopNames:   []string{"Kampung Rambutan", "Pasar Rebo", "Cililitan", "Bidara Cina", "Stasiun Jatinegara", "Cawang UKI", "Cikoko"},
		},
		{
			routeNumber: "19",
			name:        "Kota - Tanjung Priok",
			description: "Koridor 19: Kota ke Tanjung Priok",
			stopNames:   []string{"Kota", "Stasiun Kota", "Jembatan Merah", "Mangga Dua", "Gunung Sahari", "Kemayoran", "Pademangan", "Ancol", "Yos Sudarso", "Sunter", "Tanjung Priok"},
		},
		{
			routeNumber: "20",
			name:        "Pluit - Kalideres",
			description: "Koridor 20: Pluit ke Kalideres",
			stopNames:   []string{"Pluit", "Muara Karang", "Penjaringan", "Stasiun Kota", "Jembatan Merah", "Mangga Besar", "Olahraga", "Kota", "Kalideres"},
		},
		{
			routeNumber: "21",
			name:        "Blok M - Pasar Minggu",
			description: "Koridor 21: Blok M ke Pasar Minggu",
			stopNames:   []string{"Blok M", "Blok A", "Haji Nawi", "Cipete Raya", "Fatmawati", "Lebak Bulus", "Ragunan", "Pasar Minggu"},
		},
		{
			routeNumber: "22",
			name:        "PGC - Cakung",
			description: "Koridor 22: PGC ke Cakung",
			stopNames:   []string{"PGC", "Cakung", "Rawamangun", "Pulomas", "Kayu Putih"},
		},
		{
			routeNumber: "23",
			name:        "Tanjung Priok - Ancol",
			description: "Koridor 23: Tanjung Priok ke Ancol",
			stopNames:   []string{"Tanjung Priok", "Yos Sudarso", "Sunter", "Ancol"},
		},
		{
			routeNumber: "24",
			name:        "Kampung Melayu - Cililitan",
			description: "Koridor 24: Kampung Melayu ke Cililitan",
			stopNames:   []string{"Kampung Melayu", "Stasiun Jatinegara", "Bidara Cina", "Cililitan"},
		},
		{
			routeNumber: "25",
			name:        "Monumen Nasional - Harmoni",
			description: "Koridor 25: Monumen Nasional ke Harmoni",
			stopNames:   []string{"Monumen Nasional", "Harmoni", "Gajah Mada", "Hayam Wuruk"},
		},
		{
			routeNumber: "26",
			name:        "Pulo Gebang - Cakung",
			description: "Koridor 26: Pulo Gebang ke Cakung",
			stopNames:   []string{"Pulo Gebang", "Cakung", "Rawamangun", "Pulomas"},
		},
		{
			routeNumber: "27",
			name:        "Ciledug - Kebon Jeruk",
			description: "Koridor 27: Ciledug ke Kebon Jeruk",
			stopNames:   []string{"Ciledug", "Kebon Jeruk", "Tomang", "Grogol"},
		},
		{
			routeNumber: "28",
			name:        "Ragunan - Fatmawati",
			description: "Koridor 28: Ragunan ke Fatmawati",
			stopNames:   []string{"Ragunan", "Lebak Bulus", "Fatmawati"},
		},
		{
			routeNumber: "29",
			name:        "Pinang Ranti - Kampung Rambutan",
			description: "Koridor 29: Pinang Ranti ke Kampung Rambutan",
			stopNames:   []string{"Pinang Ranti", "Cililitan", "Pasar Rebo", "Kampung Rambutan"},
		},
		{
			routeNumber: "30",
			name:        "Harmoni - Senayan",
			description: "Koridor 30: Harmoni ke Senayan",
			stopNames:   []string{"Harmoni", "Thamrin", "Sudirman", "Senayan"},
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

func seedHashtags(hashtagRepo repository.HashtagRepository, reportRepo repository.ReportRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all reports
	reports, _, err := reportRepo.List(0, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch reports: %w", err)
	}

	if len(reports) == 0 {
		fmt.Println("  ⚠ No reports found, skipping hashtag seeding")
		return nil
	}

	// Common hashtags related to TransJakarta
	hashtagNames := []string{
		"transjakarta", "busway", "jakarta", "transportasi", "publictransport",
		"halte", "rute", "bus", "brt", "jakartautara", "jakartaselatan",
		"jakartatimur", "jakartabarat", "jakartapusat", "kota", "blokm",
		"pulogadung", "ragunan", "kampungmelayu", "ancol", "kalideres",
		"perbaikan", "laporan", "masalah", "fasilitas", "jadwal",
		"terlambat", "rusak", "perluperbaikan", "kebersihan", "keamanan",
	}

	fmt.Println("Seeding hashtags...")
	createdCount := 0

	// Create hashtags
	for _, hashtagName := range hashtagNames {
		hashtag, err := hashtagRepo.FindOrCreate(hashtagName)
		if err != nil {
			fmt.Printf("  ⚠ Warning: Failed to create/find hashtag %s: %v\n", hashtagName, err)
			continue
		}
		if hashtag.ID > 0 {
			createdCount++
		}
	}

	fmt.Printf("  ✓ Created/found %d hashtags\n", createdCount)

	// Assign hashtags to reports
	fmt.Println("Assigning hashtags to reports...")
	assignedCount := 0

	for _, report := range reports {
		// Assign 1-3 random hashtags to each report
		numHashtags := 1 + rng.Intn(3)
		usedHashtags := make(map[string]bool)

		for i := 0; i < numHashtags; i++ {
			hashtagName := hashtagNames[rng.Intn(len(hashtagNames))]
			if usedHashtags[hashtagName] {
				continue
			}
			usedHashtags[hashtagName] = true

			hashtag, err := hashtagRepo.FindOrCreate(hashtagName)
			if err != nil {
				continue
			}

			// Check if report already has this hashtag
			reportHashtags, err := hashtagRepo.GetByReport(report.ID)
			if err == nil {
				alreadyHas := false
				for _, rh := range reportHashtags {
					if rh.ID == hashtag.ID {
						alreadyHas = true
						break
					}
				}
				if alreadyHas {
					continue
				}
			}

			if err := hashtagRepo.CreateReportHashtag(report.ID, hashtag.ID); err != nil {
				if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
					continue
				}
				fmt.Printf("  ⚠ Warning: Failed to assign hashtag %s to report %d: %v\n", hashtagName, report.ID, err)
				continue
			}

			// Increment usage count
			if err := hashtagRepo.IncrementUsageCount(hashtag.ID); err != nil {
				// Non-fatal, continue
			}

			assignedCount++
		}
	}

	fmt.Printf("  ✓ Assigned %d hashtags to reports\n", assignedCount)
	return nil
}

func seedUserFollows(userFollowRepo repository.UserFollowRepository, userRepo repository.UserRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all users
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) < 2 {
		fmt.Println("  ⚠ Not enough users found, skipping user follow seeding")
		return nil
	}

	fmt.Println("Seeding user follows...")
	createdCount := 0
	followMap := make(map[string]bool)

	// Create 50-100 follow relationships
	numFollows := 50 + rng.Intn(51)

	for i := 0; i < numFollows; i++ {
		followerIdx := rng.Intn(len(users))
		followingIdx := rng.Intn(len(users))

		// Can't follow yourself
		if followerIdx == followingIdx {
			continue
		}

		follower := users[followerIdx]
		following := users[followingIdx]

		// Check for duplicate
		key := fmt.Sprintf("%d-%d", follower.ID, following.ID)
		if followMap[key] {
			continue
		}
		followMap[key] = true

		// Check if already following
		isFollowing, err := userFollowRepo.IsFollowing(follower.ID, following.ID)
		if err == nil && isFollowing {
			continue
		}

		userFollow := &models.UserFollow{
			FollowerID:  follower.ID,
			FollowingID: following.ID,
		}

		if err := userFollowRepo.Create(userFollow); err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "UNIQUE") {
				continue
			}
			fmt.Printf("  ⚠ Warning: Failed to create user follow: %v\n", err)
			continue
		}

		createdCount++
		if createdCount%10 == 0 {
			fmt.Printf("  ✓ Created %d user follows...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d user follows\n", createdCount)
	return nil
}

func seedBulkUploadLogs(bulkUploadLogRepo repository.BulkUploadLogRepository, userRepo repository.UserRepository) error {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get all users
	users, _, err := userRepo.List(0, 1000)
	if err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		fmt.Println("  ⚠ No users found, skipping bulk upload log seeding")
		return nil
	}

	entityTypes := []models.BulkUploadEntityType{
		models.BulkUploadEntityTypeRoute,
		models.BulkUploadEntityTypeStop,
		models.BulkUploadEntityTypeVehicle,
	}

	statuses := []models.BulkUploadStatus{
		models.BulkUploadStatusCompleted,
		models.BulkUploadStatusCompleted,
		models.BulkUploadStatusCompleted,
		models.BulkUploadStatusFailed,
		models.BulkUploadStatusPending,
	}

	fmt.Println("Seeding bulk upload logs...")
	numLogs := 20
	createdCount := 0

	for i := 0; i < numLogs; i++ {
		user := users[rng.Intn(len(users))]
		entityType := entityTypes[rng.Intn(len(entityTypes))]
		status := statuses[rng.Intn(len(statuses))]

		totalRows := 50 + rng.Intn(200) // 50-250 rows
		var successCount, duplicateCount, errorCount int
		var errorMessage *string

		if status == models.BulkUploadStatusCompleted {
			successCount = totalRows - rng.Intn(10)
			duplicateCount = rng.Intn(5)
			errorCount = totalRows - successCount - duplicateCount
		} else if status == models.BulkUploadStatusFailed {
			successCount = rng.Intn(10)
			errorCount = totalRows - successCount
			errMsg := "Failed to process some rows due to validation errors"
			errorMessage = &errMsg
		} else {
			successCount = 0
			duplicateCount = 0
			errorCount = 0
		}

		filePath := fmt.Sprintf("./uploads/bulk_uploads/bulk_%d/%s_%d.csv",
			time.Now().Unix(),
			entityType,
			i+1,
		)

		// Random time within last 30 days
		createdAt := time.Now().Add(-time.Duration(rng.Intn(30*24)) * time.Hour)
		lastUpdatedAt := createdAt.Add(time.Duration(rng.Intn(60)) * time.Minute)

		log := &models.BulkUploadLog{
			EntityType:       entityType,
			FilePath:         filePath,
			Status:           status,
			TotalRows:        totalRows,
			SuccessCount:     successCount,
			DuplicateCount:   duplicateCount,
			ErrorCount:       errorCount,
			ErrorMessage:     errorMessage,
			UserID:           user.ID,
			LastProcessedRow: successCount,
			LastUpdatedAt:    lastUpdatedAt,
			CreatedAt:        createdAt,
			UpdatedAt:        lastUpdatedAt,
		}

		if err := bulkUploadLogRepo.Create(log); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to create bulk upload log: %v\n", err)
			continue
		}

		createdCount++
		if createdCount%5 == 0 {
			fmt.Printf("  ✓ Created %d bulk upload logs...\n", createdCount)
		}
	}

	fmt.Printf("  ✓ Created %d bulk upload logs\n", createdCount)
	return nil
}

func seedForums(db *gorm.DB, forumRepo repository.ForumRepository, routeRepo repository.RouteRepository, forumPostRepo repository.ForumPostRepository, forumMemberRepo repository.ForumMemberRepository, userRepo repository.UserRepository) error {
	fmt.Println("\n🌐 Seeding forums...")

	// Get all routes
	routes, _, err := routeRepo.List(0, 100, map[string]interface{}{})
	if err != nil {
		return err
	}

	if len(routes) == 0 {
		fmt.Println("  ⚠ No routes found, skipping forum seeding")
		return nil
	}

	// Get some users for membership and posts
	users, _, err := userRepo.List(0, 20)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		fmt.Println("  ⚠ No users found, skipping forum seeding")
		return nil
	}

	createdForums := 0
	createdPosts := 0
	createdMembers := 0

	// Create forums for first 10 routes
	maxForums := 10
	if len(routes) < maxForums {
		maxForums = len(routes)
	}

	for i := 0; i < maxForums; i++ {
		route := routes[i]

		// Create forum
		forum := &models.Forum{
			RouteID: route.ID,
		}
		if err := forumRepo.Create(forum); err != nil {
			fmt.Printf("  ⚠ Warning: Failed to create forum for route %d: %v\n", route.ID, err)
			continue
		}

		createdForums++

		// Add some members (first 5-8 users)
		memberCount := 5 + (i % 4) // 5-8 members per forum
		if memberCount > len(users) {
			memberCount = len(users)
		}

		for j := 0; j < memberCount; j++ {
			member := &models.ForumMember{
				ForumID: forum.ID,
				UserID:  users[j].ID,
			}
			if err := forumMemberRepo.Create(member); err != nil {
				continue // Skip if already member
			}
			createdMembers++
		}

		// Create some posts (2-4 posts per forum)
		postCount := 2 + (i % 3) // 2-4 posts
		postTypes := []models.PostType{
			models.PostTypeDiscussion,
			models.PostTypeInfo,
			models.PostTypeQuestion,
			models.PostTypeAnnouncement,
		}

		for j := 0; j < postCount; j++ {
			postUser := users[j%len(users)]
			postType := postTypes[j%len(postTypes)]

			post := &models.ForumPost{
				ForumID:  forum.ID,
				UserID:   postUser.ID,
				PostType: postType,
				Title:    fmt.Sprintf("%s - %s %d", route.Name, getPostTypeTitle(postType), j+1),
				Content:  fmt.Sprintf("This is a %s post about %s. Route number: %s", postType, route.Name, route.RouteNumber),
			}

			// First post in first forum should be pinned
			if i == 0 && j == 0 {
				post.IsPinned = true
			}

			if err := forumPostRepo.Create(post); err != nil {
				fmt.Printf("  ⚠ Warning: Failed to create post: %v\n", err)
				continue
			}

			createdPosts++
		}
	}

	fmt.Printf("  ✓ Created %d forums\n", createdForums)
	fmt.Printf("  ✓ Created %d forum posts\n", createdPosts)
	fmt.Printf("  ✓ Created %d forum memberships\n", createdMembers)
	return nil
}

func getPostTypeTitle(postType models.PostType) string {
	switch postType {
	case models.PostTypeDiscussion:
		return "Discussion"
	case models.PostTypeInfo:
		return "Information"
	case models.PostTypeQuestion:
		return "Question"
	case models.PostTypeAnnouncement:
		return "Announcement"
	default:
		return "Post"
	}
}
