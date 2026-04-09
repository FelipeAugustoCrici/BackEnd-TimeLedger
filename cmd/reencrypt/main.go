// cmd/reencrypt/main.go
//
// Ferramenta para re-criptografar ou resetar campos financeiros.
//
// Modo re-criptografia (padrão):
//
//	OLD_KEY="chave-antiga" NEW_KEY="chave-nova" go run ./cmd/reencrypt
//
// Modo reset (quando a chave antiga é desconhecida):
//
//	RESET=true NEW_KEY="chave-nova" go run ./cmd/reencrypt
//
//	- user_settings: hourly_rate=0, daily_hours_goal=8, monthly_goal=0
//	- task_entries:  hourly_rate=0, total_amount=0
package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// ── helpers de cripto independentes da env ────────────────────────────────────

func deriveKey(raw string) []byte {
	h := sha256.Sum256([]byte(raw))
	return h[:]
}

func decryptWithKey(encoded, keyRaw string) (float64, error) {
	key := deriveKey(keyRaw)

	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		// tenta com padding
		if padded, err2 := base64.URLEncoding.DecodeString(encoded); err2 == nil {
			data = padded
		} else if v, parseErr := strconv.ParseFloat(encoded, 64); parseErr == nil {
			// valor legado em texto puro (antes da criptografia)
			return v, nil
		} else {
			return 0, fmt.Errorf("base64 decode: %w", err)
		}
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return 0, errors.New("ciphertext muito curto")
	}
	plaintext, err := gcm.Open(nil, data[:nonceSize], data[nonceSize:], nil)
	if err != nil {
		return 0, fmt.Errorf("gcm.Open: %w", err)
	}
	return strconv.ParseFloat(string(plaintext), 64)
}

func encryptWithKey(value float64, keyRaw string) (string, error) {
	key := deriveKey(keyRaw)
	plaintext := []byte(strconv.FormatFloat(value, 'f', -1, 64))

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ct), nil
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	_ = godotenv.Load()

	newKey := os.Getenv("NEW_KEY")
	resetMode := os.Getenv("RESET") == "true"

	if newKey == "" {
		log.Fatal("Defina NEW_KEY antes de executar.")
	}

	if !resetMode {
		oldKey := os.Getenv("OLD_KEY")
		if oldKey == "" {
			log.Fatal("Defina OLD_KEY (ou use RESET=true para zerar os valores).")
		}
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_SSLMODE"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("db.Ping: %v", err)
	}

	if resetMode {
		log.Println("Modo RESET: zerando valores e re-criptografando com NEW_KEY...")
		resetUserSettings(db, newKey)
		resetTaskEntries(db, newKey)
	} else {
		oldKey := os.Getenv("OLD_KEY")
		log.Println("Modo re-criptografia: descriptografando com OLD_KEY e re-criptografando com NEW_KEY...")
		reencryptUserSettings(db, oldKey, newKey)
		reencryptTaskEntries(db, oldKey, newKey)
	}

	log.Println("Concluído.")
}

// ── user_settings ─────────────────────────────────────────────────────────────

func reencryptUserSettings(db *sql.DB, oldKey, newKey string) {
	rows, err := db.Query(`SELECT id, hourly_rate, daily_hours_goal, monthly_goal FROM user_settings`)
	if err != nil {
		log.Fatalf("query user_settings: %v", err)
	}
	defer rows.Close()

	type row struct {
		id          string
		hourly      string
		dailyGoal   string
		monthlyGoal string
	}

	var records []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.hourly, &r.dailyGoal, &r.monthlyGoal); err != nil {
			log.Fatalf("scan user_settings: %v", err)
		}
		records = append(records, r)
	}

	log.Printf("user_settings: %d registro(s) encontrado(s)", len(records))

	for _, r := range records {
		hourly, err := decryptWithKey(r.hourly, oldKey)
		if err != nil {
			log.Printf("[WARN] user_settings id=%s: decrypt hourly_rate falhou (%v) – pulando", r.id, err)
			continue
		}
		daily, err := decryptWithKey(r.dailyGoal, oldKey)
		if err != nil {
			log.Printf("[WARN] user_settings id=%s: decrypt daily_hours_goal falhou (%v) – pulando", r.id, err)
			continue
		}
		monthly, err := decryptWithKey(r.monthlyGoal, oldKey)
		if err != nil {
			log.Printf("[WARN] user_settings id=%s: decrypt monthly_goal falhou (%v) – pulando", r.id, err)
			continue
		}

		newHourly, _ := encryptWithKey(hourly, newKey)
		newDaily, _ := encryptWithKey(daily, newKey)
		newMonthly, _ := encryptWithKey(monthly, newKey)

		_, err = db.Exec(`
			UPDATE user_settings
			SET hourly_rate=$1, daily_hours_goal=$2, monthly_goal=$3
			WHERE id=$4`,
			newHourly, newDaily, newMonthly, r.id,
		)
		if err != nil {
			log.Printf("[ERROR] user_settings id=%s: update falhou: %v", r.id, err)
			continue
		}
		log.Printf("user_settings id=%s: hourly_rate=%.2f re-criptografado", r.id, hourly)
	}
}

// ── task_entries ──────────────────────────────────────────────────────────────

func reencryptTaskEntries(db *sql.DB, oldKey, newKey string) {
	rows, err := db.Query(`SELECT id, hourly_rate, total_amount FROM task_entries WHERE hourly_rate IS NOT NULL`)
	if err != nil {
		log.Fatalf("query task_entries: %v", err)
	}
	defer rows.Close()

	type row struct {
		id          string
		hourlyRate  string
		totalAmount string
	}

	var records []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.hourlyRate, &r.totalAmount); err != nil {
			log.Fatalf("scan task_entries: %v", err)
		}
		records = append(records, r)
	}

	log.Printf("task_entries: %d registro(s) encontrado(s)", len(records))

	for _, r := range records {
		hourly, err := decryptWithKey(r.hourlyRate, oldKey)
		if err != nil {
			log.Printf("[WARN] task_entries id=%s: decrypt hourly_rate falhou (%v) – pulando", r.id, err)
			continue
		}
		total, err := decryptWithKey(r.totalAmount, oldKey)
		if err != nil {
			log.Printf("[WARN] task_entries id=%s: decrypt total_amount falhou (%v) – pulando", r.id, err)
			continue
		}

		newHourly, _ := encryptWithKey(hourly, newKey)
		newTotal, _ := encryptWithKey(total, newKey)

		_, err = db.Exec(`
			UPDATE task_entries
			SET hourly_rate=$1, total_amount=$2
			WHERE id=$3`,
			newHourly, newTotal, r.id,
		)
		if err != nil {
			log.Printf("[ERROR] task_entries id=%s: update falhou: %v", r.id, err)
			continue
		}
		log.Printf("task_entries id=%s: hourly_rate=%.2f re-criptografado", r.id, hourly)
	}
}

// ── reset: user_settings ──────────────────────────────────────────────────────

func resetUserSettings(db *sql.DB, newKey string) {
	rows, err := db.Query(`SELECT id FROM user_settings`)
	if err != nil {
		log.Fatalf("query user_settings: %v", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("scan user_settings: %v", err)
		}
		ids = append(ids, id)
	}

	log.Printf("user_settings: %d registro(s) para resetar", len(ids))

	hourlyEnc, _ := encryptWithKey(0, newKey)
	dailyEnc, _ := encryptWithKey(8, newKey)
	monthlyEnc, _ := encryptWithKey(0, newKey)

	for _, id := range ids {
		_, err := db.Exec(`
			UPDATE user_settings
			SET hourly_rate=$1, daily_hours_goal=$2, monthly_goal=$3
			WHERE id=$4`,
			hourlyEnc, dailyEnc, monthlyEnc, id,
		)
		if err != nil {
			log.Printf("[ERROR] user_settings id=%s: %v", id, err)
			continue
		}
		log.Printf("user_settings id=%s: resetado (hourly=0, daily_goal=8, monthly=0)", id)
	}
}

// ── reset: task_entries ───────────────────────────────────────────────────────

func resetTaskEntries(db *sql.DB, newKey string) {
	rows, err := db.Query(`SELECT id, time_spent_minutes FROM task_entries WHERE hourly_rate IS NOT NULL`)
	if err != nil {
		log.Fatalf("query task_entries: %v", err)
	}
	defer rows.Close()

	type row struct {
		id      string
		minutes int
	}

	var records []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.minutes); err != nil {
			log.Fatalf("scan task_entries: %v", err)
		}
		records = append(records, r)
	}

	log.Printf("task_entries: %d registro(s) para resetar", len(records))

	for _, r := range records {
		// hourly_rate=0 → total_amount=0 independente das horas
		hourlyEnc, _ := encryptWithKey(0, newKey)
		totalEnc, _ := encryptWithKey(0, newKey)

		_, err := db.Exec(`
			UPDATE task_entries
			SET hourly_rate=$1, total_amount=$2
			WHERE id=$3`,
			hourlyEnc, totalEnc, r.id,
		)
		if err != nil {
			log.Printf("[ERROR] task_entries id=%s: %v", r.id, err)
			continue
		}
		log.Printf("task_entries id=%s: resetado (hourly=0, total=0)", r.id)
	}
}
