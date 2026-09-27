-- 004 password groups + system defaults
CREATE TABLE password_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    label TEXT NOT NULL,
    icon TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO password_groups (slug, label, icon, sort_order) VALUES
('email', 'Почтовые ящики', 'mail', 10),
('social', 'Социальные сети', 'users', 20),
('hosting_infra', 'Хостинг и инфраструктура', 'server', 30),
('development', 'Разработка', 'code', 40)
ON CONFLICT (slug) DO NOTHING;
