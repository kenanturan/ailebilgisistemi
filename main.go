package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Person struct'ı aile üyesi bilgilerini temsil eder
type Person struct {
	ID          string `json:"id"`
	Ad          string `json:"ad"`
	Soyad       string `json:"soyad"`
	TC          string `json:"tc"`
	CepTelefonu string `json:"cepTelefonu"`
	AnneAdi     string `json:"anneAdi"`
	BabaAdi     string `json:"babaAdi"`
	EsID        string `json:"esId"`
	Cinsiyet    string `json:"cinsiyet"`
	Hakkinda    string `json:"hakkinda"`
	Fotograf    string `json:"fotograf"`
}

// PersonWithParents kişi ve ebeveyn bilgilerini birlikte tutar
type PersonWithParents struct {
	Person
	AnneAdSoyad string `json:"anneAdSoyad"`
	BabaAdSoyad string `json:"babaAdSoyad"`
}

// Marriage struct'ı evlilik bilgilerini temsil eder
type Marriage struct {
	ID            string    `json:"id"`
	Person1ID     string    `json:"person1_id"`
	Person2ID     string    `json:"person2_id"`
	EvlilikTarihi time.Time `json:"evlilik_tarihi"`
	Durum         string    `json:"durum"`
}

var db *sql.DB
var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"multiply": func(a, b int) int {
		return a * b
	},
	"subtract": func(a, b int) int {
		return a - b
	},
}).ParseGlob("templates/*.html"))

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./aile.db")
	if err != nil {
		log.Fatal("Veritabanı bağlantı hatası:", err)
	}

	// Bağlantıyı test et
	err = db.Ping()
	if err != nil {
		log.Fatal("Veritabanı ping hatası:", err)
	}

	// Tablo oluştur
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS people (
		id TEXT PRIMARY KEY,
		ad TEXT NOT NULL,
		soyad TEXT NOT NULL,
		tc TEXT UNIQUE NOT NULL,
		cepTelefonu TEXT,
		anneAdi TEXT,
		babaAdi TEXT,
		esId TEXT,
		cinsiyet TEXT,
		hakkinda TEXT,
		fotograf TEXT
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Tablo oluşturma hatası:", err)
	}

	// Eş kolonu ekle
	_, err = db.Exec(`ALTER TABLE people ADD COLUMN esId TEXT;`)
	if err != nil {
		// Kolon zaten varsa hata vermeyi görmezden gel
		log.Printf("Eş kolonu eklenirken hata (muhtemelen zaten var): %v", err)
	}

	// Evlilik tablosunu oluştur
	createMarriageTableSQL := `
	CREATE TABLE IF NOT EXISTS marriages (
		id TEXT PRIMARY KEY,
		person1_id TEXT NOT NULL,
		person2_id TEXT NOT NULL,
		evlilik_tarihi DATE,
		durum TEXT DEFAULT 'evli',
		FOREIGN KEY (person1_id) REFERENCES people(id),
		FOREIGN KEY (person2_id) REFERENCES people(id)
	);`

	_, err = db.Exec(createMarriageTableSQL)
	if err != nil {
		log.Fatal("Evlilik tablosu oluşturma hatası:", err)
	}

	fmt.Println("Veritabanı başarıyla oluşturuldu")

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
}

// Ana sayfa handler'ı
func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

// Kişi ekleme handler'ı
func kisiEkleHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "kisi-ekle.html", nil)
}

// Kişi listesi handler'ı
func kisiListesiHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "kisi-listesi.html", nil)
}

// Kişi ekleme handler'ı
func createPerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("Hatalı metod: %s", r.Method)
		http.Error(w, "Sadece POST metodu kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	// Request body'yi oku ve logla
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Request body okuma hatası: %v", err)
		http.Error(w, "Request body okuma hatası: "+err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("Gelen veri: %s", string(body))

	// Body'yi yeniden oluştur
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		log.Printf("JSON decode hatası: %v", err)
		http.Error(w, "JSON decode hatası: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Alınan kişi bilgileri: %+v", person)

	// Zorunlu alan kontrolü
	if person.Ad == "" || person.Soyad == "" || person.TC == "" || person.Cinsiyet == "" {
		http.Error(w, "Ad, Soyad, TC ve Cinsiyet alanları zorunludur", http.StatusBadRequest)
		return
	}

	// TC kontrolü
	if len(person.TC) != 11 {
		http.Error(w, "TC 11 haneli olmalıdır", http.StatusBadRequest)
		return
	}

	// Otomatik ID oluştur
	person.ID = uuid.New().String()

	// SQL sorgusunu logla
	log.Printf("SQL sorgusu çalıştırılıyor: INSERT INTO people VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
		person.ID, person.Ad, person.Soyad, person.TC, person.CepTelefonu,
		person.AnneAdi, person.BabaAdi, person.EsID, person.Cinsiyet, person.Hakkinda, "fotograf_data")

	// Veritabanına ekle
	_, err = db.Exec(`
		INSERT INTO people (id, ad, soyad, tc, cepTelefonu, anneAdi, babaAdi, esId, cinsiyet, hakkinda, fotograf) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		person.ID, person.Ad, person.Soyad, person.TC, person.CepTelefonu,
		nullToEmpty(person.AnneAdi), nullToEmpty(person.BabaAdi), nullToEmpty(person.EsID),
		person.Cinsiyet, nullToEmpty(person.Hakkinda), nullToEmpty(person.Fotograf))
	if err != nil {
		http.Error(w, "Veritabanı hatası: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Kaydedilen fotoğraf boyutu: %d byte", len(person.Fotograf))

	// Eğer eş seçildiyse, karşılıklı olarak eş bilgisini güncelle
	if person.EsID != "" {
		// Yeni eklenen kişinin eşinin bilgisini güncelle
		_, err = db.Exec("UPDATE people SET esId = ? WHERE id = ?", person.ID, person.EsID)
		if err != nil {
			http.Error(w, "Eş güncellenirken hata: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(person)
}

// Tüm kişileri listeleme handler'ı
func getPeople(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Sadece GET metodu kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	log.Println("getPeople fonksiyonu çağrıldı")

	// SQL sorgusunu logla
	sqlQuery := `
		SELECT p.id, p.ad, p.soyad, p.tc, p.cepTelefonu, 
			   COALESCE(p.anneAdi, '') as anneAdi, 
			   COALESCE(p.babaAdi, '') as babaAdi,
			   CASE WHEN p.esId IS NULL THEN '' ELSE p.esId END as esId,
			   p.cinsiyet, 
			   COALESCE(p.hakkinda, '') as hakkinda, 
			   COALESCE(p.fotograf, '') as fotograf,
			   COALESCE(anne.ad || ' ' || anne.soyad, '') as anneAdSoyad,
			   COALESCE(baba.ad || ' ' || baba.soyad, '') as babaAdSoyad
		FROM people p
		LEFT JOIN people anne ON p.anneAdi = anne.id
		LEFT JOIN people baba ON p.babaAdi = baba.id`

	log.Printf("SQL sorgusu: %s", sqlQuery)

	rows, err := db.Query(sqlQuery)
	if err != nil {
		handleError(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var people []PersonWithParents

	for rows.Next() {
		var p PersonWithParents
		err := rows.Scan(
			&p.ID, &p.Ad, &p.Soyad, &p.TC, &p.CepTelefonu,
			&p.AnneAdi, &p.BabaAdi, &p.EsID, &p.Cinsiyet, &p.Hakkinda, &p.Fotograf,
			&p.AnneAdSoyad, &p.BabaAdSoyad)
		if err != nil {
			log.Printf("Satır okuma hatası: %v", err)
			http.Error(w, "Veri okuma hatası: "+err.Error(), http.StatusInternalServerError)
			return
		}
		people = append(people, p)
	}

	log.Printf("Bulunan kişi sayısı: %d", len(people))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(people)
}

// Kişi güncelleme handler'ı
func updatePerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Sadece PUT metodu kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var person Person
	if err := json.NewDecoder(r.Body).Decode(&person); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if person.ID == "" {
		http.Error(w, "ID boş olamaz", http.StatusBadRequest)
		return
	}

	// Önce eski eş bilgisini alalım
	var oldEsID string
	var err error
	err = db.QueryRow("SELECT esId FROM people WHERE id = ?", person.ID).Scan(&oldEsID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Eğer yeni bir eş seçildiyse
	if person.EsID != "" {
		// Yeni eşin bilgisini güncelle
		_, err = db.Exec("UPDATE people SET esId = ? WHERE id = ?", person.ID, person.EsID)
		if err != nil {
			http.Error(w, "Eş güncellenirken hata: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Eski eşin bilgisini temizle (eğer varsa ve değiştiyse)
	if oldEsID != "" && oldEsID != person.EsID {
		_, err = db.Exec("UPDATE people SET esId = NULL WHERE id = ?", oldEsID)
		if err != nil {
			http.Error(w, "Eski eş güncellenirken hata: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Kişiyi güncelle
	result, err := db.Exec(`
		UPDATE people 
		SET ad=?, soyad=?, tc=?, cepTelefonu=?, 
			anneAdi=?, babaAdi=?, esId=?, 
			cinsiyet=?, hakkinda=?, fotograf=? 
		WHERE id=?`,
		person.Ad, person.Soyad, person.TC, person.CepTelefonu,
		nullToEmpty(person.AnneAdi), nullToEmpty(person.BabaAdi), nullToEmpty(person.EsID),
		person.Cinsiyet, nullToEmpty(person.Hakkinda), nullToEmpty(person.Fotograf), person.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Kişi bulunamadı", http.StatusNotFound)
		return
	}

	log.Printf("Güncellenen fotoğraf boyutu: %d byte", len(person.Fotograf))

	// Debug için fotoğraf verisini kontrol et
	log.Printf("Sunucuya gelen fotoğraf verisi uzunluğu: %d", len(person.Fotograf))
	if len(person.Fotograf) > 0 {
		log.Printf("Fotoğraf verisi başlangıcı: %s", person.Fotograf[:100])
	}

	json.NewEncoder(w).Encode(person)
}

// Kişi silme handler'ı
func deletePerson(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Sadece DELETE metodu kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parametresi gerekli", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("DELETE FROM people WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Kişi bulunamadı", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Kişi başarıyla silindi")
}

func kisiDetayHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/kisi/")

	// Kişinin kendi bilgilerini al
	var person PersonWithParents
	err := db.QueryRow(`
		SELECT p.id, p.ad, p.soyad, p.tc, p.cepTelefonu, 
			   COALESCE(p.anneAdi, '') as anneAdi, 
			   COALESCE(p.babaAdi, '') as babaAdi,
			   COALESCE(p.esId, '') as esId,
			   p.cinsiyet, 
			   COALESCE(p.hakkinda, '') as hakkinda, 
			   COALESCE(p.fotograf, '') as fotograf,
			   COALESCE(anne.ad || ' ' || anne.soyad, '') as anneAdSoyad,
			   COALESCE(baba.ad || ' ' || baba.soyad, '') as babaAdSoyad
		FROM people p
		LEFT JOIN people anne ON p.anneAdi = anne.id
		LEFT JOIN people baba ON p.babaAdi = baba.id
		WHERE p.id = ?`, id).Scan(
		&person.ID, &person.Ad, &person.Soyad, &person.TC, &person.CepTelefonu,
		&person.AnneAdi, &person.BabaAdi, &person.EsID, &person.Cinsiyet, &person.Hakkinda, &person.Fotograf,
		&person.AnneAdSoyad, &person.BabaAdSoyad)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Debug için fotoğraf verisini kontrol et
	log.Printf("Kişi detay - Fotoğraf verisi başlangıcı: %s",
		func() string {
			if len(person.Fotograf) > 100 {
				return person.Fotograf[:100]
			}
			return person.Fotograf
		}())
	log.Printf("Kişi detay - Fotoğraf verisi uzunluğu: %d", len(person.Fotograf))
	log.Printf("Kişi detay - Fotoğraf verisi data:image ile başlıyor mu?: %v",
		strings.HasPrefix(person.Fotograf, "data:image"))

	// Kişinin anne, baba ve büyükanne-büyükbabalarını bul
	rows, err := db.Query(`
		WITH RECURSIVE soy_agaci AS (
			-- Büyükanne ve büyükbabalar (anne tarafı)
			SELECT 
				p.id, p.ad, p.soyad, p.cinsiyet, -2 as nesil,
				COALESCE(anne.id, '') as anne_id, 
				COALESCE(anne.ad || ' ' || anne.soyad, '') as anne_adi,
				COALESCE(baba.id, '') as baba_id, 
				COALESCE(baba.ad || ' ' || baba.soyad, '') as baba_adi
			FROM people p
			LEFT JOIN people anne ON p.anneAdi = anne.id
			LEFT JOIN people baba ON p.babaAdi = baba.id
			WHERE p.id IN (
				SELECT anneAdi FROM people WHERE id IN (
					SELECT anneAdi FROM people WHERE id = ?
				)
				UNION
				SELECT babaAdi FROM people WHERE id IN (
					SELECT anneAdi FROM people WHERE id = ?
				)
			)

			UNION ALL

			-- Büyükanne ve büyükbabalar (baba tarafı)
			SELECT 
				p.id, p.ad, p.soyad, p.cinsiyet, -2 as nesil,
				COALESCE(anne.id, '') as anne_id, 
				COALESCE(anne.ad || ' ' || anne.soyad, '') as anne_adi,
				COALESCE(baba.id, '') as baba_id, 
				COALESCE(baba.ad || ' ' || baba.soyad, '') as baba_adi
			FROM people p
			LEFT JOIN people anne ON p.anneAdi = anne.id
			LEFT JOIN people baba ON p.babaAdi = baba.id
			WHERE p.id IN (
				SELECT anneAdi FROM people WHERE id IN (
					SELECT babaAdi FROM people WHERE id = ?
				)
				UNION
				SELECT babaAdi FROM people WHERE id IN (
					SELECT babaAdi FROM people WHERE id = ?
				)
			)

			UNION ALL

			-- Anne ve baba
			SELECT 
				p.id, p.ad, p.soyad, p.cinsiyet, -1 as nesil,
				COALESCE(anne.id, '') as anne_id, 
				COALESCE(anne.ad || ' ' || anne.soyad, '') as anne_adi,
				COALESCE(baba.id, '') as baba_id, 
				COALESCE(baba.ad || ' ' || baba.soyad, '') as baba_adi
			FROM people p
			LEFT JOIN people anne ON p.anneAdi = anne.id
			LEFT JOIN people baba ON p.babaAdi = baba.id
			WHERE p.id IN (
				SELECT anneAdi FROM people WHERE id = ?
				UNION
				SELECT babaAdi FROM people WHERE id = ?
			)

			UNION ALL

			-- Direkt çocuklar
			SELECT p.id, p.ad, p.soyad, p.cinsiyet, 1 as nesil,
				COALESCE(anne.id, '') as anne_id, 
				COALESCE(anne.ad || ' ' || anne.soyad, '') as anne_adi,
				COALESCE(baba.id, '') as baba_id, 
				COALESCE(baba.ad || ' ' || baba.soyad, '') as baba_adi
			FROM people p
			LEFT JOIN people anne ON p.anneAdi = anne.id
			LEFT JOIN people baba ON p.babaAdi = baba.id
			WHERE p.anneAdi = ? OR p.babaAdi = ?

			UNION ALL

			-- Torunlar
			SELECT p.id, p.ad, p.soyad, p.cinsiyet, sa.nesil + 1,
				COALESCE(anne.id, '') as anne_id, 
				COALESCE(anne.ad || ' ' || anne.soyad, '') as anne_adi,
				COALESCE(baba.id, '') as baba_id, 
				COALESCE(baba.ad || ' ' || baba.soyad, '') as baba_adi
			FROM people p
			LEFT JOIN people anne ON p.anneAdi = anne.id
			LEFT JOIN people baba ON p.babaAdi = baba.id
			JOIN soy_agaci sa ON p.anneAdi = sa.id OR p.babaAdi = sa.id
			WHERE sa.nesil >= 1
		)
		SELECT id, ad, soyad, cinsiyet, nesil, 
			   anne_id, anne_adi, baba_id, baba_adi
		FROM (
			SELECT * FROM soy_agaci
			UNION ALL
			-- Eş
			SELECT 
				p.id, p.ad, p.soyad, p.cinsiyet, -3 as nesil,
				COALESCE(anne.id, '') as anne_id, 
				COALESCE(anne.ad || ' ' || anne.soyad, '') as anne_adi,
				COALESCE(baba.id, '') as baba_id, 
				COALESCE(baba.ad || ' ' || baba.soyad, '') as baba_adi
			FROM people p
			LEFT JOIN people anne ON p.anneAdi = anne.id
			LEFT JOIN people baba ON p.babaAdi = baba.id
			WHERE p.id = (SELECT COALESCE(esId, '') FROM people WHERE id = ?)
			OR p.esId = ?
			UNION ALL
			-- Kardeşler
			SELECT 
				p.id, p.ad, p.soyad, p.cinsiyet, 0 as nesil,
				COALESCE(anne.id, '') as anne_id, 
				COALESCE(anne.ad || ' ' || anne.soyad, '') as anne_adi,
				COALESCE(baba.id, '') as baba_id, 
				COALESCE(baba.ad || ' ' || baba.soyad, '') as baba_adi
			FROM people p
			LEFT JOIN people anne ON p.anneAdi = anne.id
			LEFT JOIN people baba ON p.babaAdi = baba.id
			WHERE p.id != ? AND (
				p.anneAdi IN (SELECT anneAdi FROM people WHERE id = ? AND anneAdi IS NOT NULL AND anneAdi != '')
				AND p.babaAdi IN (SELECT babaAdi FROM people WHERE id = ? AND babaAdi IS NOT NULL AND babaAdi != '')
				AND p.anneAdi != ''
				AND p.babaAdi != ''
			)
		) combined
		ORDER BY 
			CASE 
				WHEN nesil = -3 THEN 1  -- Eş en üstte
				WHEN nesil = -2 THEN 2  -- Büyükanne/büyükbaba
				WHEN nesil = -1 THEN 3  -- Anne/baba
				WHEN nesil = 0 THEN 4   -- Kardeşler
				WHEN nesil = 1 THEN 5   -- Çocuklar
				WHEN nesil >= 2 THEN 6  -- Torunlar
				ELSE 7                  -- Diğer
			END,
			cinsiyet DESC, 
			ad`,
		id, id, id, id, id, id, id, id, id, id, id, id, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var children []struct {
		ID       string
		Ad       string
		Soyad    string
		Cinsiyet string
		Nesil    int
		AnneID   string
		AnneAdi  string
		BabaID   string
		BabaAdi  string
	}

	for rows.Next() {
		var child struct {
			ID       string
			Ad       string
			Soyad    string
			Cinsiyet string
			Nesil    int
			AnneID   string
			AnneAdi  string
			BabaID   string
			BabaAdi  string
		}
		if err := rows.Scan(&child.ID, &child.Ad, &child.Soyad, &child.Cinsiyet, &child.Nesil, &child.AnneID, &child.AnneAdi, &child.BabaID, &child.BabaAdi); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		children = append(children, child)
	}

	// Verileri template'e gönder
	data := struct {
		Person   PersonWithParents
		Children []struct {
			ID       string
			Ad       string
			Soyad    string
			Cinsiyet string
			Nesil    int
			AnneID   string
			AnneAdi  string
			BabaID   string
			BabaAdi  string
		}
	}{
		Person:   person,
		Children: children,
	}

	templates.ExecuteTemplate(w, "kisi-detay.html", data)
}

// Evlilik ekleme handler'ı
func createMarriage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Sadece POST metodu kabul edilir", http.StatusMethodNotAllowed)
		return
	}

	var marriage Marriage
	if err := json.NewDecoder(r.Body).Decode(&marriage); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	marriage.ID = uuid.New().String()
	marriage.Durum = "evli"

	_, err := db.Exec(`
		INSERT INTO marriages (id, person1_id, person2_id, evlilik_tarihi, durum)
		VALUES (?, ?, ?, ?, ?)`,
		marriage.ID, marriage.Person1ID, marriage.Person2ID,
		marriage.EvlilikTarihi, marriage.Durum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(marriage)
}

// nullToEmpty boş string veya nil değerleri boş string'e çevirir
func nullToEmpty(s string) string {
	if s == "" {
		return ""
	}
	return s
}

// Middleware eklenebilir
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Temel güvenlik kontrolleri
		// CSRF koruması
		// Rate limiting
		next(w, r)
	}
}

// Merkezi hata yönetimi için helper
func handleError(w http.ResponseWriter, err error, status int) {
	log.Printf("Hata: %v", err)
	http.Error(w, err.Error(), status)
}

func main() {
	initDB()
	defer db.Close()

	// Statik dosyalar için
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Sayfa yönlendirmeleri
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/kisi-listesi", http.StatusSeeOther)
	})
	http.HandleFunc("/kisi-ekle", kisiEkleHandler)
	http.HandleFunc("/kisi-listesi", kisiListesiHandler)

	// API endpointleri
	http.HandleFunc("/api/people", authMiddleware(getPeople))
	http.HandleFunc("/api/person/create", createPerson)
	http.HandleFunc("/api/person/update", updatePerson)
	http.HandleFunc("/api/person/delete", deletePerson)

	http.HandleFunc("/kisi/", kisiDetayHandler)

	fmt.Println("Sunucu 3000 portunda başlatılıyor...")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
