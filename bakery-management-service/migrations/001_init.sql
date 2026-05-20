CREATE TABLE IF NOT EXISTS ingredients (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    unit TEXT NOT NULL,
    cost_per_unit NUMERIC(12,2) NOT NULL CHECK (cost_per_unit >= 0),
    recipe_sections JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recipes (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id TEXT PRIMARY KEY,
    recipe_id TEXT NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id TEXT NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    amount NUMERIC(12,3) NOT NULL CHECK (amount >= 0)
);

CREATE TABLE IF NOT EXISTS recipe_calculations (
    id TEXT PRIMARY KEY,
    product_id TEXT NOT NULL,
    quantity INT NOT NULL,
    total_cost NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'TODO',
    due_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS statistics_daily (
    id TEXT PRIMARY KEY,
    stat_date DATE NOT NULL,
    product_id TEXT NOT NULL,
    product_name TEXT,
    baked INT NOT NULL DEFAULT 0,
    delivered INT NOT NULL DEFAULT 0,
    sold INT NOT NULL DEFAULT 0,
    returned INT NOT NULL DEFAULT 0,
    left_qty INT NOT NULL DEFAULT 0,
    revenue NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS statistics_weekly (
    id TEXT PRIMARY KEY,
    week_label TEXT NOT NULL,
    product_id TEXT NOT NULL,
    delivered INT NOT NULL DEFAULT 0,
    sold INT NOT NULL DEFAULT 0,
    returned INT NOT NULL DEFAULT 0,
    revenue NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS statistics_monthly (
    id TEXT PRIMARY KEY,
    month_label TEXT NOT NULL,
    product_id TEXT NOT NULL,
    baked INT NOT NULL DEFAULT 0,
    sold INT NOT NULL DEFAULT 0,
    returned INT NOT NULL DEFAULT 0,
    revenue NUMERIC(12,2) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS email_logs (
    id TEXT PRIMARY KEY,
    recipient TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
