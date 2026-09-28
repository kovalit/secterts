-- 010 custom icons for password entries
--
-- Adds a "custom" icon source alongside the existing "favicon"/"group" values
-- and a column to hold a user-uploaded icon as a small data URL.
ALTER TABLE password_entries
    ADD COLUMN custom_icon TEXT;

ALTER TABLE password_entries
    DROP CONSTRAINT chk_password_entries_icon_source;

ALTER TABLE password_entries
    ADD CONSTRAINT chk_password_entries_icon_source
    CHECK (icon_source IN ('favicon', 'group', 'custom'));
