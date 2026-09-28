// Command import-dashlane imports a Dashlane credentials.csv export.
// Passwords, notes, secondary usernames, and OTP URLs are encrypted before
// insertion. The command never logs credential values.
package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/kovalit/secrets-center/backend/internal/config"
	appcrypto "github.com/kovalit/secrets-center/backend/internal/crypto"
	"github.com/kovalit/secrets-center/backend/internal/db"
	"github.com/kovalit/secrets-center/backend/internal/passwords"
)

type credential struct {
	Username  string
	Username2 string
	Username3 string
	Title     string
	Password  string
	Note      string
	URL       string
	Category  string
	OTPURL    string
}

func main() {
	csvPath := flag.String("csv", "", "path to Dashlane credentials.csv")
	userEmail := flag.String("user-email", "", "owner email; optional when exactly one active user exists")
	dryRun := flag.Bool("dry-run", false, "validate and report counts without inserting")
	allowExisting := flag.Bool("allow-existing", false, "allow import when the owner already has password entries")
	flag.Parse()

	if *csvPath == "" {
		log.Fatal("-csv is required")
	}
	if err := run(*csvPath, *userEmail, *dryRun, *allowExisting); err != nil {
		log.Fatalf("import failed: %v", err)
	}
}

func run(csvPath, userEmail string, dryRun, allowExisting bool) error {
	rows, err := readCSV(csvPath)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return errors.New("CSV contains no credentials")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // no-op after commit

	ownerID, activeUsers, err := resolveOwner(ctx, tx, userEmail)
	if err != nil {
		return err
	}

	var existing int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM password_entries WHERE owner_user_id=$1 AND deleted_at IS NULL`, ownerID).Scan(&existing); err != nil {
		return err
	}

	groups, err := loadGroups(ctx, tx)
	if err != nil {
		return err
	}
	for _, slug := range []string{"email", "social", "hosting_infra", "development"} {
		if groups[slug] == "" {
			return fmt.Errorf("required password group %q is missing", slug)
		}
	}

	if dryRun {
		log.Printf("dry run: credentials=%d active_users=%d existing_entries=%d", len(rows), activeUsers, existing)
		return nil
	}
	if existing > 0 && !allowExisting {
		return fmt.Errorf("owner already has %d password entries; rerun with -allow-existing only after checking for duplicates", existing)
	}

	enc, err := appcrypto.NewEncryptor(cfg.MasterKeyBase64)
	if err != nil {
		return err
	}
	for i, row := range rows {
		if err := insertCredential(ctx, tx, enc, ownerID, groups, row); err != nil {
			return fmt.Errorf("row %d: %w", i+2, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	log.Printf("import complete: imported=%d", len(rows))
	return nil
}

func readCSV(path string) ([]credential, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	index := make(map[string]int, len(header))
	for i, name := range header {
		name = strings.TrimPrefix(strings.TrimSpace(name), "\ufeff")
		index[name] = i
	}
	for _, required := range []string{"username", "title", "password", "note", "url", "category", "otpUrl"} {
		if _, ok := index[required]; !ok {
			return nil, fmt.Errorf("missing CSV column %q", required)
		}
	}

	value := func(record []string, key string) string {
		i, ok := index[key]
		if !ok || i >= len(record) {
			return ""
		}
		return strings.TrimSpace(record[i])
	}

	var out []credential
	for {
		record, err := r.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		row := credential{
			Username: value(record, "username"), Username2: value(record, "username2"), Username3: value(record, "username3"),
			Title: value(record, "title"), Password: value(record, "password"), Note: value(record, "note"),
			URL: value(record, "url"), Category: value(record, "category"), OTPURL: value(record, "otpUrl"),
		}
		if row.Title == "" || row.Password == "" {
			return nil, fmt.Errorf("credential row %d has an empty title or password", len(out)+2)
		}
		out = append(out, row)
	}
	return out, nil
}

func resolveOwner(ctx context.Context, tx pgx.Tx, email string) (id string, count int, err error) {
	if email != "" {
		err = tx.QueryRow(ctx, `SELECT id FROM users WHERE lower(email)=lower($1) AND is_active`, strings.TrimSpace(email)).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, errors.New("active user with the requested email was not found")
		}
		return id, 1, err
	}

	rows, err := tx.Query(ctx, `SELECT id FROM users WHERE is_active ORDER BY created_at`)
	if err != nil {
		return "", 0, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return "", 0, err
		}
		ids = append(ids, userID)
	}
	if err := rows.Err(); err != nil {
		return "", 0, err
	}
	if len(ids) != 1 {
		return "", len(ids), fmt.Errorf("expected exactly one active user, found %d; specify -user-email", len(ids))
	}
	return ids[0], len(ids), nil
}

func loadGroups(ctx context.Context, tx pgx.Tx) (map[string]string, error) {
	rows, err := tx.Query(ctx, `SELECT slug, id FROM password_groups`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var slug, id string
		if err := rows.Scan(&slug, &id); err != nil {
			return nil, err
		}
		out[slug] = id
	}
	return out, rows.Err()
}

func insertCredential(ctx context.Context, tx pgx.Tx, enc appcrypto.Encryptor, ownerID string, groups map[string]string, row credential) error {
	password, passwordNonce, passwordVersion, err := enc.EncryptString(row.Password)
	if err != nil {
		return err
	}

	var encryptedComment, commentNonce []byte
	var commentVersion *int
	if comment := buildComment(row); comment != "" {
		ciphertext, nonce, version, err := enc.EncryptString(comment)
		if err != nil {
			return err
		}
		encryptedComment, commentNonce, commentVersion = ciphertext, nonce, &version
	}

	var siteURL, domain, login *string
	if row.URL != "" {
		siteURL = &row.URL
		if parsed := passwords.ExtractDomain(row.URL); parsed != "" {
			domain = &parsed
		}
	}
	if row.Username != "" {
		login = &row.Username
	} else if row.Username2 != "" {
		login = &row.Username2
	} else if row.Username3 != "" {
		login = &row.Username3
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO password_entries (
			owner_user_id, group_id, scope, title, site_url, domain, icon_source, login,
			encrypted_password, password_nonce, password_key_version,
			encrypted_comment, comment_nonce, comment_key_version)
		VALUES ($1,$2,'personal',$3,$4,$5,'group',$6,$7,$8,$9,$10,$11,$12)`,
		ownerID, groups[groupSlug(row.Category)], row.Title, siteURL, domain, login,
		password, passwordNonce, passwordVersion, encryptedComment, commentNonce, commentVersion)
	return err
}

func groupSlug(category string) string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "почтовые ящики", "email":
		return "email"
	case "социальные сети", "social":
		return "social"
	case "серверная инфраструктура", "хостинг и инфраструктура", "hosting":
		return "hosting_infra"
	default:
		return "development"
	}
}

func buildComment(row credential) string {
	parts := make([]string, 0, 4)
	if row.Note != "" {
		parts = append(parts, row.Note)
	}
	if row.Username2 != "" && row.Username2 != row.Username {
		parts = append(parts, "Дополнительный логин: "+row.Username2)
	}
	if row.Username3 != "" && row.Username3 != row.Username && row.Username3 != row.Username2 {
		parts = append(parts, "Дополнительный логин 2: "+row.Username3)
	}
	if row.OTPURL != "" {
		parts = append(parts, "OTP URL: "+row.OTPURL)
	}
	return strings.Join(parts, "\n\n")
}
