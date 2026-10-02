// Package backup produces an encrypted-at-rest JSON export of the data set and
// records each run in backup_exports.
package backup

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Document is the top-level export payload. Encrypted []byte fields marshal to
// base64 automatically via encoding/json.
type Document struct {
	Version     int              `json:"version"`
	GeneratedAt time.Time        `json:"generated_at"`
	Users       []userExport     `json:"users"`
	Companies   []companyExport  `json:"companies"`
	Groups      []groupExport    `json:"password_groups"`
	Passwords   []passwordExport `json:"password_entries"`
	Projects    []projectExport  `json:"app_projects"`
	Environments []envExport     `json:"app_environments"`
	Secrets     []secretExport   `json:"app_secrets"`
	KeyVersions []keyVerExport   `json:"key_versions"`
}

type userExport struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	// password_hash is intentionally omitted.
}

type companyExport struct {
	ID          string `json:"id"`
	OwnerUserID string `json:"owner_user_id"`
	Name        string `json:"name"`
}

type groupExport struct {
	ID    string `json:"id"`
	Slug  string `json:"slug"`
	Label string `json:"label"`
	Icon  string `json:"icon"`
}

type passwordExport struct {
	ID                 string     `json:"id"`
	OwnerUserID        string     `json:"owner_user_id"`
	CompanyID          *string    `json:"company_id"`
	GroupID            string     `json:"group_id"`
	Scope              string     `json:"scope"`
	Title              string     `json:"title"`
	SiteURL            *string    `json:"site_url"`
	Domain             *string    `json:"domain"`
	Login              *string    `json:"login"`
	EntryType          string     `json:"entry_type"`
	ExpiresAt          *time.Time `json:"expires_at"`
	Owner              *string    `json:"owner"`
	EncryptedPassword  []byte     `json:"encrypted_password"`
	PasswordNonce      []byte     `json:"password_nonce"`
	PasswordKeyVersion int        `json:"password_key_version"`
	EncryptedComment   []byte     `json:"encrypted_comment"`
	CommentNonce       []byte     `json:"comment_nonce"`
}

type projectExport struct {
	ID          string  `json:"id"`
	OwnerUserID string  `json:"owner_user_id"`
	CompanyID   *string `json:"company_id"`
	Name        string  `json:"name"`
}

type envExport struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
}

type secretExport struct {
	ID              string `json:"id"`
	ProjectID       string `json:"project_id"`
	EnvironmentID   string `json:"environment_id"`
	Key             string `json:"key"`
	EncryptedValue  []byte `json:"encrypted_value"`
	ValueNonce      []byte `json:"value_nonce"`
	ValueKeyVersion int    `json:"value_key_version"`
}

type keyVerExport struct {
	Version   int    `json:"version"`
	Algorithm string `json:"algorithm"`
	Status    string `json:"status"`
}

// build gathers all data into an export Document.
func build(ctx context.Context, pool *pgxpool.Pool) (*Document, error) {
	doc := &Document{Version: 1, GeneratedAt: time.Now().UTC()}

	if err := collect(ctx, pool,
		`SELECT id, email, created_at FROM users ORDER BY created_at`,
		func(rows pgx.Rows) error {
			var u userExport
			if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt); err != nil {
				return err
			}
			doc.Users = append(doc.Users, u)
			return nil
		}); err != nil {
		return nil, err
	}

	if err := collect(ctx, pool,
		`SELECT id, owner_user_id, name FROM companies ORDER BY name`,
		func(rows pgx.Rows) error {
			var c companyExport
			if err := rows.Scan(&c.ID, &c.OwnerUserID, &c.Name); err != nil {
				return err
			}
			doc.Companies = append(doc.Companies, c)
			return nil
		}); err != nil {
		return nil, err
	}

	if err := collect(ctx, pool,
		`SELECT id, slug, label, icon FROM password_groups ORDER BY sort_order`,
		func(rows pgx.Rows) error {
			var g groupExport
			if err := rows.Scan(&g.ID, &g.Slug, &g.Label, &g.Icon); err != nil {
				return err
			}
			doc.Groups = append(doc.Groups, g)
			return nil
		}); err != nil {
		return nil, err
	}

	if err := collect(ctx, pool, `
		SELECT id, owner_user_id, company_id, group_id, scope, title, site_url, domain, login,
			entry_type, expires_at, owner,
			encrypted_password, password_nonce, password_key_version, encrypted_comment, comment_nonce
		FROM password_entries WHERE deleted_at IS NULL ORDER BY created_at`,
		func(rows pgx.Rows) error {
			var p passwordExport
			if err := rows.Scan(&p.ID, &p.OwnerUserID, &p.CompanyID, &p.GroupID, &p.Scope, &p.Title, &p.SiteURL, &p.Domain, &p.Login,
				&p.EntryType, &p.ExpiresAt, &p.Owner,
				&p.EncryptedPassword, &p.PasswordNonce, &p.PasswordKeyVersion, &p.EncryptedComment, &p.CommentNonce); err != nil {
				return err
			}
			doc.Passwords = append(doc.Passwords, p)
			return nil
		}); err != nil {
		return nil, err
	}

	if err := collect(ctx, pool,
		`SELECT id, owner_user_id, company_id, name FROM app_projects ORDER BY name`,
		func(rows pgx.Rows) error {
			var p projectExport
			if err := rows.Scan(&p.ID, &p.OwnerUserID, &p.CompanyID, &p.Name); err != nil {
				return err
			}
			doc.Projects = append(doc.Projects, p)
			return nil
		}); err != nil {
		return nil, err
	}

	if err := collect(ctx, pool,
		`SELECT id, project_id, name FROM app_environments ORDER BY sort_order`,
		func(rows pgx.Rows) error {
			var e envExport
			if err := rows.Scan(&e.ID, &e.ProjectID, &e.Name); err != nil {
				return err
			}
			doc.Environments = append(doc.Environments, e)
			return nil
		}); err != nil {
		return nil, err
	}

	if err := collect(ctx, pool, `
		SELECT id, project_id, environment_id, key, encrypted_value, value_nonce, value_key_version
		FROM app_secrets WHERE deleted_at IS NULL ORDER BY key`,
		func(rows pgx.Rows) error {
			var s secretExport
			if err := rows.Scan(&s.ID, &s.ProjectID, &s.EnvironmentID, &s.Key, &s.EncryptedValue, &s.ValueNonce, &s.ValueKeyVersion); err != nil {
				return err
			}
			doc.Secrets = append(doc.Secrets, s)
			return nil
		}); err != nil {
		return nil, err
	}

	if err := collect(ctx, pool,
		`SELECT version, algorithm, status FROM app_key_versions ORDER BY version`,
		func(rows pgx.Rows) error {
			var k keyVerExport
			if err := rows.Scan(&k.Version, &k.Algorithm, &k.Status); err != nil {
				return err
			}
			doc.KeyVersions = append(doc.KeyVersions, k)
			return nil
		}); err != nil {
		return nil, err
	}

	return doc, nil
}

// collect runs a query and calls scan for each row.
func collect(ctx context.Context, pool *pgxpool.Pool, sql string, scan func(pgx.Rows) error) error {
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
