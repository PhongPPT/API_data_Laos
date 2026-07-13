package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NominatimResponse represents the JSON returned by OpenStreetMap API
type NominatimResponse struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

func main() {
	// 1. Connect to the local Postgres database (Running via Docker)
	// Make sure to use the exposed port from docker-compose if running from Mac host
	dsn := "postgres://user:password@localhost:5432/laosgeo?sslmode=disable"
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database. Make sure Docker is running: %v", err)
	}
	defer db.Close()

	fmt.Println("🚀 Connected to Database. Starting Geocoding process...")

	// 2. Process Districts first
	geocodeDistricts(db)

	// 3. Process Villages
	// Uncomment the line below to run villages (Note: 8000+ items will take several hours)
	geocodeVillages(db)

	fmt.Println("✅ Geocoding finished!")
}

func geocodeDistricts(db *sqlx.DB) {
	fmt.Println("--- Fetching Districts with NULL coordinates ---")

	type DistrictRecord struct {
		ID         int    `db:"dr_id"`
		Name       string `db:"dr_name"`
		NameEn     string `db:"dr_name_en"`
		Province   string `db:"pr_name"`
		ProvinceEn string `db:"pr_name_en"`
	}

	var records []DistrictRecord
	query := `
		SELECT d.dr_id, d.dr_name, d.dr_name_en, p.pr_name, p.pr_name_en 
		FROM districts d 
		JOIN provinces p ON d.pr_id = p.pr_id 
		WHERE d.lat IS NULL OR d.lng IS NULL
	`
	err := db.Select(&records, query)
	if err != nil {
		log.Fatalf("Error fetching districts: %v", err)
	}

	fmt.Printf("Found %d districts to update.\n", len(records))

	for i, rec := range records {
		// Fallback strategies for searching
		searchQueries := []string{
			fmt.Sprintf("%s, %s, Laos", rec.NameEn, rec.ProvinceEn),
			fmt.Sprintf("ເມືອງ%s, %s, Laos", rec.Name, rec.Province), // Use Lao name with 'Meuang'
			fmt.Sprintf("%s, Laos", rec.NameEn),                      // Just district and Laos
		}

		fmt.Printf("[%d/%d] Searching: %s... ", i+1, len(records), rec.NameEn)

		var finalLat, finalLng float64
		var success bool

		for _, sq := range searchQueries {
			lat, lng, err := fetchCoordinates(sq)
			if err != nil {
				fmt.Printf("❌ API Error: %v\n", err)
				break
			}
			if lat != 0 && lng != 0 {
				finalLat = lat
				finalLng = lng
				success = true
				break // Found it!
			}
			// Wait before next fallback attempt to respect rate limit
			time.Sleep(1500 * time.Millisecond)
		}

		if !success {
			fmt.Println("⚠️ Not Found on Map (Need manual update)")
		} else {
			// Update Database
			_, err = db.Exec(`UPDATE districts SET lat = $1, lng = $2 WHERE dr_id = $3`, finalLat, finalLng, rec.ID)
			if err != nil {
				fmt.Printf("❌ DB Update Failed: %v\n", err)
			} else {
				fmt.Printf("✅ Success (Lat: %f, Lng: %f)\n", finalLat, finalLng)
			}
		}

		time.Sleep(1500 * time.Millisecond)
	}
}

func geocodeVillages(db *sqlx.DB) {
	fmt.Println("--- Fetching Villages with NULL coordinates ---")

	type VillageRecord struct {
		ID         int    `db:"vill_id"`
		Village    string `db:"vill_name"`
		VillageEn  string `db:"vill_name_en"`
		DistrictID int    `db:"dr_id"`
		District   string `db:"dr_name"`
		DistrictEn string `db:"dr_name_en"`
		Province   string `db:"pr_name"`
		ProvinceEn string `db:"pr_name_en"`
	}

	var records []VillageRecord
	query := `
		SELECT v.vill_id, v.vill_name, v.vill_name_en, d.dr_id, d.dr_name, d.dr_name_en, p.pr_name, p.pr_name_en 
		FROM villages v 
		JOIN districts d ON v.dr_id = d.dr_id 
		JOIN provinces p ON d.pr_id = p.pr_id 
		WHERE v.lat IS NULL OR v.lng IS NULL
	`
	err := db.Select(&records, query)
	if err != nil {
		log.Fatalf("Error fetching villages: %v", err)
	}

	fmt.Printf("Found %d villages to update.\n", len(records))

	for i, rec := range records {
		searchQueries := []string{
			fmt.Sprintf("%s, %s, %s, Laos", rec.VillageEn, rec.DistrictEn, rec.ProvinceEn),
			fmt.Sprintf("ບ້ານ%s, ເມືອງ%s, %s, Laos", rec.Village, rec.District, rec.Province),
			fmt.Sprintf("%s, Laos", rec.VillageEn),
		}

		fmt.Printf("[%d/%d] Searching: %s... ", i+1, len(records), rec.VillageEn)

		var finalLat, finalLng float64
		var success bool

		for _, sq := range searchQueries {
			lat, lng, err := fetchCoordinates(sq)
			if err != nil {
				fmt.Printf("❌ API Error: %v\n", err)
				break
			}
			if lat != 0 && lng != 0 {
				finalLat = lat
				finalLng = lng
				success = true
				break
			}
			time.Sleep(1500 * time.Millisecond)
		}

		if !success {
			// FALLBACK: Use District's coordinates
			fmt.Print("⚠️ Not Found -> Using District coordinates fallback... ")
			var distLat, distLng float64
			err := db.QueryRow(`SELECT lat, lng FROM districts WHERE dr_id = $1`, rec.DistrictID).Scan(&distLat, &distLng)
			if err == nil && distLat != 0 && distLng != 0 {
				_, err = db.Exec(`UPDATE villages SET lat = $1, lng = $2 WHERE vill_id = $3`, distLat, distLng, rec.ID)
				if err != nil {
					fmt.Printf("❌ DB Update Failed: %v\n", err)
				} else {
					fmt.Printf("✅ Fallback Success (Lat: %f, Lng: %f)\n", distLat, distLng)
				}
			} else {
				fmt.Println("❌ District fallback failed (Need manual update)")
			}
		} else {
			_, err = db.Exec(`UPDATE villages SET lat = $1, lng = $2 WHERE vill_id = $3`, finalLat, finalLng, rec.ID)
			if err != nil {
				fmt.Printf("❌ DB Update Failed: %v\n", err)
			} else {
				fmt.Printf("✅ Success (Lat: %f, Lng: %f)\n", finalLat, finalLng)
			}
		}

		time.Sleep(1500 * time.Millisecond)
	}
}

func fetchCoordinates(query string) (float64, float64, error) {
	// URL Encode the search query
	encodedQuery := url.QueryEscape(query)
	apiUrl := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", encodedQuery)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return 0, 0, err
	}

	// ⚠️ OpenStreetMap blocks generic emails (like example.com). We use a realistic one.
	req.Header.Set("User-Agent", "LaosGeoProject/1.0 (Contact: laosgeo.developer@gmail.com)")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,lo;q=0.8")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	var results []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, err
	}

	if len(results) == 0 {
		return 0, 0, nil // Not found
	}

	lat, _ := strconv.ParseFloat(results[0].Lat, 64)
	lng, _ := strconv.ParseFloat(results[0].Lon, 64)

	return lat, lng, nil
}
