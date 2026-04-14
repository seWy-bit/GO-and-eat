CREATE TABLE IF NOT EXISTS restaurants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS menu_items (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price BIGINT NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (restaurant_id) REFERENCES restaurants(id) ON DELETE CASCADE
);

-- Индекс для быстрого поиска блюд по ресторану
CREATE INDEX IF NOT EXISTS idx_menu_items_restaurant_id ON menu_items(restaurant_id);

COMMENT ON TABLE restaurants IS 'Рестораны в системе';
COMMENT ON COLUMN restaurants.id IS 'Уникальный идентификатор ресторана';
COMMENT ON COLUMN restaurants.name IS 'Название ресторана';
COMMENT ON COLUMN restaurants.address IS 'Физический адрес';
COMMENT ON COLUMN restaurants.phone IS 'Контактный телефон';

COMMENT ON TABLE menu_items IS 'Блюда в меню ресторанов';
COMMENT ON COLUMN menu_items.id IS 'Уникальный ID блюда';
COMMENT ON COLUMN menu_items.restaurant_id IS 'Ссылка на ресторан';
COMMENT ON COLUMN menu_items.price IS 'Цена в копейках (например, 59900 = 599.00 руб)';
COMMENT ON COLUMN menu_items.stock IS 'Доступное количество для заказа';
COMMENT ON COLUMN menu_items.available IS 'Доступно для заказа (автоматически обновляется из stock)';